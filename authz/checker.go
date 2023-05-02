package authz

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	Check(context.Context, TupleKey) (bool, error)
	ListObjects(context.Context, TupleKey) ([]string, error)
}

type Checker struct {
	authn     AuthnProvider
	authz     AuthzProvider
	feedCache *ecache.Cache[int]
	fvidCache *ecache.Cache[int]
	finder    model.Finder
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

func (c *Checker) Feeds(ctx context.Context, user auth.User) ([]int, error) {
	// Check cache
	// Use ListObjects, map back to feed IDs using current_feeds.authn_id
	userKey := "user:" + user.Name()
	fmt.Println("userKey:", userKey)
	feedKeys, err := c.authz.ListObjects(ctx, TupleKey{User: userKey, Object: "feed", Relation: "can_view"})
	if err != nil {
		return nil, err
	}
	var feedIds []int
	for _, k := range feedKeys {
		kk := strings.Split(k, ":")
		if len(kk) > 1 {
			kid, err := strconv.Atoi(kk[1])
			if err != nil {
				return nil, err
			}
			feedIds = append(feedIds, kid)
		}
	}
	return feedIds, nil
}
