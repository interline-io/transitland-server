package authz

import (
	"context"
	"errors"
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
	Check(context.Context, TupleKey) (bool, error)
	ListObjects(context.Context, TupleKey) ([]TupleKey, error)
	GetObjectTuples(context.Context, TupleKey) ([]TupleKey, error)
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
	var authn AuthnProvider
	var authz AuthzProvider
	if cfg.Auth0Domain != "" {
		var err error
		authn, err = NewAuth0Client(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)
		if err != nil {
			return nil, err
		}
	} else {
		authn = NewMockAuthnClient()
	}
	if cfg.FGAEndpoint != "" {
		var err error
		authz, err = NewFGAClient(cfg.FGAStoreID, cfg.FGAModelID, cfg.FGAEndpoint)
		if err != nil {
			return nil, err
		}
	}
	checker := NewChecker(authn, authz, nil, nil)
	return checker, nil
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

// USERS

func (c *Checker) ListUsers(ctx context.Context, user auth.User, query string) ([]User, error) {
	// TODO: filter users
	users, err := c.authn.Users(ctx, query)
	if err != nil {
		return nil, err
	}
	var ret []User
	for _, user := range users {
		if user != nil {
			ret = append(ret, *user)
		}
	}
	return ret, nil
}

func (c *Checker) User(ctx context.Context, user auth.User, userId string) (*User, error) {
	// TODO: filter users
	ret, err := c.authn.UserByID(ctx, userId)
	return ret, err
}

func (c *Checker) hydrateUsers(ctx context.Context, user auth.User, users []User) ([]User, error) {
	// TODO: filter users
	ret := []User{}
	for _, u := range users {
		uu, err := c.authn.UserByID(ctx, u.ID)
		if err == nil && uu != nil {
			ret = append(ret, *uu)
		}
	}
	return ret, nil
}

// TENANTS

type TenantPermissionsResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Users struct {
		Admins  []User `json:"admins"`
		Members []User `json:"members"`
	} `json:"users"`
	Actions struct {
		CanEditMembers bool `json:"can_edit_members"`
		CanView        bool `json:"can_view"`
		CanEdit        bool `json:"can_edit"`
		CanCreateOrg   bool `json:"can_create_org"`
		CanDeleteOrg   bool `json:"can_delete_org"`
	} `json:"actions"`
}

func (c *Checker) ListTenants(ctx context.Context, user auth.User) ([]string, error) {
	return c.listObjectNames(ctx, user, TenantType, CanView)
}

func (c *Checker) TenantPermissions(ctx context.Context, user auth.User, tenantId int) (*TenantPermissionsResponse, error) {
	entTk := TupleKey{}.WithObject(TenantType, itoa(tenantId))
	ret := &TenantPermissionsResponse{ID: tenantId}
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == AdminRelation {
			ret.Users.Admins = append(ret.Users.Admins, User{ID: tk.UserName})
		}
		if tk.Relation == MemberRelation {
			ret.Users.Members = append(ret.Users.Members, User{ID: tk.UserName})
		}
	}
	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanEditMembers))
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanEdit))
	ret.Actions.CanCreateOrg, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanCreateOrg))
	ret.Actions.CanDeleteOrg, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanDeleteOrg))
	ret.Users.Admins, _ = c.hydrateUsers(ctx, user, ret.Users.Admins)
	ret.Users.Members, _ = c.hydrateUsers(ctx, user, ret.Users.Members)
	return ret, nil
}

func (c *Checker) AddTenantPermission(ctx context.Context, user auth.User, addUser string, groupId int, relation Relation) error {
	return c.addObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(addUser).WithObject(TenantType, itoa(groupId)))
}

func (c *Checker) RemoveTenantPermission(ctx context.Context, user auth.User, removeUser string, groupId int, relation Relation) error {
	return c.removeObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(removeUser).WithObject(TenantType, itoa(groupId)).WithRelation(relation))
}

// GROUPS

func (c *Checker) ListGroups(ctx context.Context, user auth.User) ([]int, error) {
	return c.listObjectIds(ctx, user, GroupType, CanView)
}

type GroupPermissionsResponse struct {
	ID     int                        `json:"id"`
	Name   string                     `json:"name"`
	Tenant *TenantPermissionsResponse `json:"tenant,omitempty"`
	Users  struct {
		Viewers  []User `json:"viewers"`
		Editors  []User `json:"editors"`
		Managers []User `json:"managers"`
	} `json:"users"`
	Actions struct {
		CanView        bool `json:"can_view"`
		CanEditMembers bool `json:"can_edit_members"`
		CanCreateFeed  bool `json:"can_create_feed"`
		CanDeleteFeed  bool `json:"can_delete_feed"`
		CanEdit        bool `json:"can_edit"`
	} `json:"actions"`
}

func (c *Checker) GroupPermissions(ctx context.Context, user auth.User, groupId int) (*GroupPermissionsResponse, error) {
	entTk := TupleKey{}.WithObject(GroupType, itoa(groupId))
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}
	ret := &GroupPermissionsResponse{
		ID: groupId,
	}
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ret.Tenant, _ = c.TenantPermissions(ctx, user, atoi(tk.UserName))
		}
		if tk.Relation == ManagerRelation {
			ret.Users.Managers = append(ret.Users.Managers, User{ID: tk.UserName})
		}
		if tk.Relation == EditorRelation {
			ret.Users.Editors = append(ret.Users.Editors, User{ID: tk.UserName})
		}
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, User{ID: tk.UserName})
		}
	}
	ret.Users.Managers, _ = c.hydrateUsers(ctx, user, ret.Users.Managers)
	ret.Users.Editors, _ = c.hydrateUsers(ctx, user, ret.Users.Editors)
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, user, ret.Users.Viewers)
	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanEditMembers))
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanEdit))
	ret.Actions.CanCreateFeed, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanCreateFeed))
	ret.Actions.CanDeleteFeed, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanDeleteFeed))
	return ret, nil
}

func (c *Checker) AddGroupPermission(ctx context.Context, user auth.User, addUser string, groupId int, relation Relation) error {
	return c.addObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(addUser).WithObject(GroupType, itoa(groupId)))
}

func (c *Checker) RemoveGroupPermission(ctx context.Context, user auth.User, removeUser string, groupId int, relation Relation) error {
	return c.removeObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(removeUser).WithObject(GroupType, itoa(groupId)).WithRelation(relation))
}

// FEEDS

func (c *Checker) ListFeeds(ctx context.Context, user auth.User) ([]int, error) {
	return c.listObjectIds(ctx, user, FeedType, CanView)
}

type FeedPermissionsResponse struct {
	ID    int                       `json:"id"`
	Group *GroupPermissionsResponse `json:"group,omitempty"`
	Users struct {
		Viewers []User `json:"viewers"`
	} `json:"users"`
	Actions struct {
		CanView              bool `json:"can_view"`
		CanEdit              bool `json:"can_edit"`
		CanCreateFeedVersion bool `json:"can_create_feed_version"`
		CanDeleteFeedVersion bool `json:"can_delete_feed_version"`
	} `json:"actions"`
}

func (c *Checker) FeedPermissions(ctx context.Context, user auth.User, feedId int) (*FeedPermissionsResponse, error) {
	entTk := TupleKey{}.WithObject(FeedType, itoa(feedId))
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}
	ret := &FeedPermissionsResponse{
		ID: feedId,
	}
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ret.Group, _ = c.GroupPermissions(ctx, user, atoi(tk.UserName))
		}
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, User{ID: tk.UserName})
		}
	}
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, user, ret.Users.Viewers)
	ret.Actions.CanView = true
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanEdit))
	ret.Actions.CanCreateFeedVersion, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanCreateFeedVersion))
	ret.Actions.CanDeleteFeedVersion, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanDeleteFeedVersion))
	return ret, nil
}

func (c *Checker) AddFeedPermission(ctx context.Context, user auth.User, addUser string, feedId int, relation Relation) error {
	return c.addObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(addUser).WithObject(FeedType, itoa(feedId)))
}

func (c *Checker) RemoveFeedPermission(ctx context.Context, user auth.User, removeUser string, feedId int, relation Relation) error {
	return c.removeObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(removeUser).WithObject(FeedType, itoa(feedId)).WithRelation(relation))
}

// FEED VERSIONS

func (c *Checker) ListFeedVersions(ctx context.Context, user auth.User) ([]int, error) {
	return c.listObjectIds(ctx, user, FeedVersionType, CanView)
}

type FeedVersionPermissionsResponse struct {
	ID    int `json:"id"`
	Users struct {
		Viewers []User `json:"viewers"`
	} `json:"users"`
	Actions struct {
		CanEditMembers bool `json:"can_edit_members"`
		CanEdit        bool `json:"can_edit"`
	} `json:"actions"`
}

func (c *Checker) FeedVersionPermissions(ctx context.Context, user auth.User, fvid int) (*FeedVersionPermissionsResponse, error) {
	entTk := TupleKey{}.WithObject(FeedVersionType, itoa(fvid))
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}
	ret := &FeedVersionPermissionsResponse{
		ID: fvid,
	}
	for _, tk := range tps {
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, User{ID: tk.UserName})
		}
	}
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, user, ret.Users.Viewers)
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanEditMembers))
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUser(user.Name()).WithAction(CanEdit))
	return ret, nil
}

func (c *Checker) AddFeedVersionPermission(ctx context.Context, user auth.User, addUser string, fvid int, relation string) error {
	return c.addObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(addUser).WithObject(FeedVersionType, itoa(fvid)))
}

func (c *Checker) RemoveFeedVersionPermission(ctx context.Context, user auth.User, removeUser string, fvid int, relation Relation) error {
	return c.removeObjectTuple(ctx, user, CanEditMembers, TupleKey{}.WithUser(removeUser).WithObject(FeedVersionType, itoa(fvid)).WithRelation(relation))
}

// internal

func (c *Checker) listObjects(ctx context.Context, user auth.User, objectType ObjectType, action Action) ([]TupleKey, error) {
	tk := TupleKey{ObjectType: objectType}.WithAction(action).WithUser(user.Name())
	objTks, err := c.authz.ListObjects(ctx, tk)
	if err != nil {
		return nil, err
	}
	return objTks, nil
}

func (c *Checker) listObjectIds(ctx context.Context, user auth.User, objectType ObjectType, action Action) ([]int, error) {
	objTks, err := c.listObjects(ctx, user, objectType, action)
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

func (c *Checker) listObjectNames(ctx context.Context, user auth.User, objectType ObjectType, action Action) ([]string, error) {
	objTks, err := c.listObjects(ctx, user, objectType, action)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, tk := range objTks {
		if err != nil {
			return nil, err
		}
		ret = append(ret, tk.ObjectName)
	}
	return ret, nil
}

func (c *Checker) checkObject(ctx context.Context, tk TupleKey) (bool, error) {
	return c.authz.Check(ctx, tk)
}

func (c *Checker) getObjectTuples(ctx context.Context, user auth.User, checkAction Action, tk TupleKey) ([]TupleKey, error) {
	if ok, err := c.checkObject(ctx, tk.WithUser(user.Name()).WithAction(checkAction)); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("unauthorized")
	}
	return c.authz.GetObjectTuples(ctx, tk)
}

func (c *Checker) addObjectTuple(ctx context.Context, user auth.User, checkAction Action, tk TupleKey) error {
	if ok, err := c.checkObject(ctx, tk.WithUser(user.Name()).WithAction(checkAction)); err != nil {
		return err
	} else if !ok {
		return errors.New("unauthorized")
	}
	return c.authz.WriteTuple(ctx, tk)
}

func (c *Checker) removeObjectTuple(ctx context.Context, user auth.User, checkAction Action, tk TupleKey) error {
	if ok, err := c.checkObject(ctx, tk.WithUser(user.Name()).WithAction(checkAction)); err != nil {
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
