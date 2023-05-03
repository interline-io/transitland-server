package authz

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/interline-io/transitland-server/model"
)

type AuthnProvider interface {
	Users(context.Context, string) ([]*User, error)
	UserByID(context.Context, string) (*User, error)
}

type AuthzProvider interface {
	ListObjects(context.Context, TupleKey) ([]TupleKey, error)
	GetObjectTuples(context.Context, TupleKey) ([]TupleKey, error)
	Check(context.Context, TupleKey) (bool, error)
	WriteTuple(context.Context, TupleKey) error
	DeleteTuple(context.Context, TupleKey) error
}

type Checker struct {
	authn     AuthnProvider
	authz     AuthzProvider
	feedCache *ecache.Cache[int]
	fvidCache *ecache.Cache[int]
	finder    model.Finder
}

func NewCheckerFromConfig(cfg AuthzConfig) (*Checker, error) {
	auth0c, err := NewAuth0Client(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)
	if err != nil {
		return nil, err
	}
	fgac, err := NewFGAClient(cfg.FGAStoreID, cfg.FGAModelID, cfg.FGAEndpoint)
	if err != nil {
		return nil, err
	}
	checker := NewChecker(auth0c, fgac, nil, nil)
	return checker, err
}

func NewChecker(n AuthnProvider, p AuthzProvider, finder model.Finder, redisClient *redis.Client) *Checker {
	return &Checker{
		authn:     n,
		authz:     p,
		finder:    finder,
		feedCache: ecache.NewCache[int](redisClient, "checker:feeds"),
		fvidCache: ecache.NewCache[int](redisClient, "checker:fvids"),
	}
}

func (c *Checker) Check(ctx context.Context, tk TupleKey) (bool, error) {
	return c.authz.Check(ctx, tk)
}

func (c *Checker) ListFeeds(ctx context.Context, user auth.User) ([]int, error) {
	return c.listObjectIds(ctx, user, "feed", "can_view")
}

func (c *Checker) ListFeedVersions(ctx context.Context, user auth.User) ([]int, error) {
	return c.listObjectIds(ctx, user, "feed_version", "can_view")
}

type FeedPermissionsResponse struct {
	Parent  string
	Viewers []string
}

func (c *Checker) FeedPermissions(ctx context.Context, user auth.User, feedId int) (FeedPermissionsResponse, error) {
	ret := FeedPermissionsResponse{
		Viewers: []string{},
	}
	tps, err := c.getObjectTuples(ctx, user, "can_view", TupleKey{}.WithObject("feed", itoa(feedId)))
	if err != nil {
		return ret, err
	}
	for _, tk := range tps {
		fmt.Println("tk:", tk)
		if tk.Relation == "parent" {
			ret.Parent = tk.ObjectName
		}
		if tk.Relation == "viewer" {
			ret.Viewers = append(ret.Viewers, tk.UserName)
		}
	}
	return ret, nil
}

func (c *Checker) FeedVersionPermissions(ctx context.Context, user auth.User, fvid int) ([]TupleKey, error) {
	return c.getObjectTuples(ctx, user, "can_view", TupleKey{}.WithObject("feed_version", itoa(fvid)))
}

func (c *Checker) AddFeedPermission(ctx context.Context, user auth.User, addUser string, feedId int, relation string) error {
	return c.addObjectTuple(ctx, user, "can_edit_members", TupleKey{}.WithUser(addUser).WithObject("feed", itoa(feedId)))
}

func (c *Checker) AddFeedVersionPermission(ctx context.Context, user auth.User, addUser string, fvid int, relation string) error {
	return c.addObjectTuple(ctx, user, "can_edit_members", TupleKey{}.WithUser(addUser).WithObject("feed_version", itoa(fvid)))
}

func (c *Checker) RemoveFeedPermission(ctx context.Context, user auth.User, removeUser string, feedId int, relation string) error {
	return c.removeObjectTuple(ctx, user, "can_edit_members", TupleKey{}.WithUser(removeUser).WithObject("feed_version", itoa(feedId)).WithRelation(relation))
}

func (c *Checker) RemoveFeedVersionPermission(ctx context.Context, user auth.User, removeUser string, fvid int, relation string) error {
	return c.removeObjectTuple(ctx, user, "can_edit_members", TupleKey{}.WithUser(removeUser).WithObject("feed_version", itoa(fvid)).WithRelation(relation))
}

func (c *Checker) listObjects(ctx context.Context, user auth.User, objectType string, relation string) ([]TupleKey, error) {
	tk := TupleKey{ObjectType: objectType, Relation: relation}.WithUser(user.Name())
	objTks, err := c.authz.ListObjects(ctx, tk)
	if err != nil {
		return nil, err
	}
	return objTks, nil
}

func (c *Checker) listObjectIds(ctx context.Context, user auth.User, objectType string, relation string) ([]int, error) {
	objTks, err := c.listObjects(ctx, user, objectType, relation)
	if err != nil {
		return nil, err
	}
	var ret []int
	for _, tk := range objTks {
		kid, err := strconv.Atoi(tk.ObjectName)
		if err != nil {
			return nil, err
		}
		ret = append(ret, kid)
	}
	return ret, nil
}

func (c *Checker) checkObject(ctx context.Context, tk TupleKey) (bool, error) {
	return c.authz.Check(ctx, tk)
}

func (c *Checker) getObjectTuples(ctx context.Context, user auth.User, checkRelation string, tk TupleKey) ([]TupleKey, error) {
	if ok, err := c.checkObject(ctx, tk.WithUser(user.Name()).WithRelation(checkRelation)); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("unauthorized")
	}
	return c.authz.GetObjectTuples(ctx, tk)
}

func (c *Checker) addObjectTuple(ctx context.Context, user auth.User, checkRelation string, tk TupleKey) error {
	if ok, err := c.checkObject(ctx, tk.WithUser(user.Name()).WithRelation(checkRelation)); err != nil {
		return err
	} else if !ok {
		return errors.New("unauthorized")
	}
	return c.authz.WriteTuple(ctx, tk)
}

func (c *Checker) removeObjectTuple(ctx context.Context, user auth.User, checkRelation string, tk TupleKey) error {
	if ok, err := c.checkObject(ctx, tk.WithUser(user.Name()).WithRelation(checkRelation)); err != nil {
		return err
	} else if !ok {
		return errors.New("unauthorized")
	}
	return c.authz.DeleteTuple(ctx, tk)
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func atoi(v string) int {
	a, _ := strconv.Atoi(v)
	return a
}
