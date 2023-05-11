package authz

import (
	"context"
	"errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/find"
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
	ReplaceTuple(context.Context, TupleKey) error
	DeleteTuple(context.Context, TupleKey) error
}

type Checker struct {
	authn     AuthnProvider
	authz     AuthzProvider
	feedCache *ecache.Cache[int]
	fvidCache *ecache.Cache[int]
	finder    model.Finder
}

func NewCheckerFromConfig(cfg AuthzConfig, finder model.Finder, redisClient *redis.Client) (*Checker, error) {
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
	checker := NewChecker(authn, authz, finder, redisClient)
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

// ///////////////////
// USERS
// ///////////////////

func (c *Checker) UserList(ctx context.Context, user auth.User, query string) ([]User, error) {
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
	// Must already be filtered for permissions
	ret := []User{}
	for _, u := range users {
		uu, err := c.authn.UserByID(ctx, u.ID)
		if err == nil && uu != nil {
			ret = append(ret, *uu)
		}
	}
	return ret, nil
}

// ///////////////////
// TENANTS
// ///////////////////

type TenantResponse struct {
	responseId
	Name string `json:"name" db:"tenant_name"`
}

func (t TenantResponse) TableName() string {
	return "tl_tenants"
}

type TenantPermissionsResponse struct {
	TenantResponse
	Groups []GroupResponse `json:"groups"`
	Users  struct {
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

func (c *Checker) TenantList(ctx context.Context, user auth.User) ([]TenantResponse, error) {
	ids, err := c.listObjectIds(ctx, user, TenantType, CanView)
	if err != nil {
		return nil, err
	}
	return hydrates[TenantResponse](ctx, c.finder.DBX(), "tl_tenants", ids)
}

func (c *Checker) Tenant(ctx context.Context, user auth.User, tenantId int) (*TenantResponse, error) {
	entTk := TupleKey{}.WithObjectID(TenantType, tenantId)
	if ok, err := c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanView)); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("unauthorized")
	}
	r, err := hydrate[TenantResponse](ctx, c.finder.DBX(), "tl_tenants", tenantId)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Checker) TenantPermissions(ctx context.Context, user auth.User, tenantId int) (*TenantPermissionsResponse, error) {
	// Check tenant access
	entTk := TupleKey{}.WithObjectID(TenantType, tenantId)
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}

	// Get tenant metadata
	ret := &TenantPermissionsResponse{}
	ret.TenantResponse, _ = hydrate[TenantResponse](ctx, c.finder.DBX(), "tl_tenants", tenantId)
	for _, tk := range tps {
		if tk.Relation == AdminRelation {
			ret.Users.Admins = append(ret.Users.Admins, User{ID: tk.UserName})
		}
		if tk.Relation == MemberRelation {
			ret.Users.Members = append(ret.Users.Members, User{ID: tk.UserName})
		}
	}

	groupTks, _ := c.authz.ListObjects(ctx, TupleKey{}.WithObject(GroupType, "").WithUserID(TenantType, tenantId).WithRelation(ParentRelation))
	var groupIds []int
	for _, ftk := range groupTks {
		groupIds = append(groupIds, ftk.ObjectID())
	}
	ret.Groups, _ = hydrates[GroupResponse](ctx, c.finder.DBX(), "tl_groups", groupIds)

	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanEditMembers))
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanEdit))
	ret.Actions.CanCreateOrg, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanCreateOrg))
	ret.Actions.CanDeleteOrg, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanDeleteOrg))
	ret.Users.Admins, _ = c.hydrateUsers(ctx, user, ret.Users.Admins)
	ret.Users.Members, _ = c.hydrateUsers(ctx, user, ret.Users.Members)
	return ret, nil
}

func (c *Checker) TenantSave(ctx context.Context, checkUser auth.User, tenantId int, newName string) (int, error) {
	log.Trace().Str("tenantName", newName).Int("id", tenantId).Msg("TenantSave")
	id := 0
	err := sq.StatementBuilder.
		RunWith(c.finder.DBX()).
		PlaceholderFormat(sq.Dollar).
		Insert("tl_tenants").
		Columns("id", "tenant_name").
		Values(tenantId, newName).
		Suffix("on conflict (id) do update set tenant_name = ?", newName).
		Suffix(`RETURNING "id"`).
		QueryRow().Scan(&id)
	return id, err
}

func (c *Checker) TenantAddPermission(ctx context.Context, checkUser auth.User, tenantId int, addUser string, relation Relation) error {
	tk := TupleKey{}.WithUserName(addUser).WithObjectID(TenantType, tenantId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", tenantId).Msg("TenantAddPermission")
	return c.replaceObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

func (c *Checker) TenantRemovePermission(ctx context.Context, checkUser auth.User, tenantId int, removeUser string, relation Relation) error {
	tk := TupleKey{}.WithUserName(removeUser).WithObjectID(TenantType, tenantId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", tenantId).Msg("TenantRemovePermission")
	return c.removeObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

func (c *Checker) TenantCreateGroup(ctx context.Context, checkUser auth.User, tenantId int, groupName string) (int, error) {
	entTk := TupleKey{}.WithObjectID(TenantType, tenantId)
	if check, err := c.checkObject(ctx, entTk.WithUserName(checkUser.Name()).WithAction(CanCreateOrg)); err != nil {
		return 0, err
	} else if !check {
		return 0, errors.New("unauthorized")
	}
	log.Trace().Str("groupName", groupName).Int("id", tenantId).Msg("TenantCreateGroup")
	groupId := 0
	err := sq.StatementBuilder.
		RunWith(c.finder.DBX()).
		PlaceholderFormat(sq.Dollar).
		Insert("tl_groups").
		Columns("id", "group_name").
		Values(sq.Expr("(select max(id)+1 from tl_groups)"), groupName).
		Suffix(`RETURNING "id"`).
		QueryRow().Scan(&groupId)
	if err != nil {
		return 0, err
	}
	addTk := TupleKey{}.WithUserID(TenantType, tenantId).WithObjectID(GroupType, groupId).WithRelation(ParentRelation)
	if err := c.authz.WriteTuple(ctx, addTk); err != nil {
		return 0, err
	}
	return groupId, err
}

// ///////////////////
// GROUPS
// ///////////////////

type GroupResponse struct {
	responseId
	Name string `json:"name" db:"group_name"`
}

func (t GroupResponse) TableName() string {
	return "tl_groups"
}

type GroupPermissionsResponse struct {
	GroupResponse
	Tenant *TenantResponse `json:"tenant,omitempty"`
	Feeds  []FeedResponse  `json:"feeds"`
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

func (c *Checker) Group(ctx context.Context, user auth.User, groupId int) (*GroupResponse, error) {
	entTk := TupleKey{}.WithObjectID(TenantType, groupId)
	if ok, err := c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanView)); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("unauthorized")
	}
	r, err := hydrate[GroupResponse](ctx, c.finder.DBX(), "tl_groups", groupId)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Checker) GroupList(ctx context.Context, user auth.User) ([]GroupResponse, error) {
	ids, err := c.listObjectIds(ctx, user, GroupType, CanView)
	if err != nil {
		return nil, err
	}
	return hydrates[GroupResponse](ctx, c.finder.DBX(), "tl_groups", ids)
}

func (c *Checker) GroupPermissions(ctx context.Context, user auth.User, groupId int) (*GroupPermissionsResponse, error) {
	// Check group access
	entTk := TupleKey{}.WithObjectID(GroupType, groupId)
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}

	// Get group metadata
	ret := &GroupPermissionsResponse{}
	ret.GroupResponse, _ = hydrate[GroupResponse](ctx, c.finder.DBX(), "tl_groups", groupId)

	// Get group relations
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ret.Tenant, _ = c.Tenant(ctx, user, tk.UserID())
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
	feedTks, _ := c.authz.ListObjects(ctx, TupleKey{ObjectType: FeedType}.WithUserID(GroupType, groupId).WithRelation(ParentRelation))
	ret.Feeds, _ = hydrates[FeedResponse](ctx, c.finder.DBX(), "current_feeds", tkObjectIds(feedTks))

	// Prepare response
	ret.Users.Managers, _ = c.hydrateUsers(ctx, user, ret.Users.Managers)
	ret.Users.Editors, _ = c.hydrateUsers(ctx, user, ret.Users.Editors)
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, user, ret.Users.Viewers)
	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanEditMembers))
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanEdit))
	ret.Actions.CanCreateFeed, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanCreateFeed))
	ret.Actions.CanDeleteFeed, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanDeleteFeed))
	return ret, nil
}

func (c *Checker) GroupSave(ctx context.Context, checkUser auth.User, groupId int, newName string) (int, error) {
	log.Trace().Str("groupName", newName).Int("id", groupId).Msg("GroupSave")
	id := 0
	err := sq.StatementBuilder.
		RunWith(c.finder.DBX()).
		PlaceholderFormat(sq.Dollar).
		Insert("tl_groups").
		Columns("id", "group_name").
		Values(groupId, newName).
		Suffix("on conflict (id) do update set group_name = ?", newName).
		Suffix(`RETURNING "id"`).
		QueryRow().Scan(&id)
	return id, err
}

func (c *Checker) GroupAddPermission(ctx context.Context, checkUser auth.User, addUser string, groupId int, relation Relation) error {
	tk := TupleKey{}.WithUserName(addUser).WithObjectID(GroupType, groupId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", groupId).Msg("GroupAddPermission")
	return c.replaceObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

func (c *Checker) GroupRemovePermission(ctx context.Context, checkUser auth.User, removeUser string, groupId int, relation Relation) error {
	tk := TupleKey{}.WithUserName(removeUser).WithObjectID(GroupType, groupId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", groupId).Msg("GroupRemovePermission")
	return c.removeObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

/////////////////////
// FEEDS
/////////////////////

type FeedResponse struct {
	responseId
	OnestopID string `json:"onestop_id" db:"onestop_id"`
	Name      string `json:"name" db:"name"`
}

func (t FeedResponse) TableName() string {
	return "current_feeds"
}

type FeedPermissionsResponse struct {
	FeedResponse
	Group *GroupResponse `json:"group"`
	Users struct {
	} `json:"users"`
	Actions struct {
		CanView              bool `json:"can_view"`
		CanEdit              bool `json:"can_edit"`
		CanCreateFeedVersion bool `json:"can_create_feed_version"`
		CanDeleteFeedVersion bool `json:"can_delete_feed_version"`
	} `json:"actions"`
}

func (c *Checker) ListFeeds(ctx context.Context, user auth.User) ([]FeedResponse, error) {
	feedIds, err := c.listObjectIds(ctx, user, FeedType, CanView)
	if err != nil {
		return nil, err
	}
	return hydrates[FeedResponse](ctx, c.finder.DBX(), "current_feeds", feedIds)
}

func (c *Checker) FeedPermissions(ctx context.Context, user auth.User, feedId int) (*FeedPermissionsResponse, error) {
	entTk := TupleKey{}.WithObjectID(FeedType, feedId)
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}
	ret := &FeedPermissionsResponse{}
	ret.FeedResponse, _ = hydrate[FeedResponse](ctx, c.finder.DBX(), "current_feeds", feedId)
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ret.Group, _ = c.Group(ctx, user, tk.UserID())
		}
	}
	ret.Actions.CanView = true
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanEdit))
	ret.Actions.CanCreateFeedVersion, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanCreateFeedVersion))
	ret.Actions.CanDeleteFeedVersion, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanDeleteFeedVersion))
	return ret, nil
}

func (c *Checker) FeedSetGroup(ctx context.Context, checkUser auth.User, feedId int, newGroup int) error {
	tk := TupleKey{}.WithUserID(FeedType, feedId).WithObjectID(GroupType, newGroup).WithRelation(ParentRelation)
	log.Trace().Str("tk", tk.String()).Int("id", feedId).Msg("FeedSetGroup")
	return c.replaceObjectTuple(ctx, checkUser, CanEdit, tk)
}

/////////////////////
// FEED VERSIONS
/////////////////////

type FeedVersionResponse struct {
	responseId
	Name string `json:"name"`
}

type FeedVersionPermissionsResponse struct {
	FeedVersionResponse
	Users struct {
		Viewers []User `json:"viewers"`
	} `json:"users"`
	Actions struct {
		CanView        bool `json:"can_view"`
		CanEditMembers bool `json:"can_edit_members"`
		CanEdit        bool `json:"can_edit"`
	} `json:"actions"`
}

func (c *Checker) ListFeedVersions(ctx context.Context, user auth.User) ([]FeedVersionResponse, error) {
	var ret []FeedVersionResponse
	feedIds, err := c.listObjectIds(ctx, user, FeedVersionType, CanView)
	if err != nil {
		return nil, err
	}
	for _, feedId := range feedIds {
		r := FeedVersionResponse{}
		r.ID = feedId
		ret = append(ret, r)
	}
	return ret, nil
}

func (c *Checker) FeedVersionPermissions(ctx context.Context, user auth.User, fvid int) (*FeedVersionPermissionsResponse, error) {
	entTk := TupleKey{}.WithObjectID(FeedVersionType, fvid)
	tps, err := c.getObjectTuples(ctx, user, CanView, entTk)
	if err != nil {
		return nil, err
	}
	ret := &FeedVersionPermissionsResponse{}
	ret.ID = fvid
	for _, tk := range tps {
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, User{ID: tk.UserName})
		}
	}
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, user, ret.Users.Viewers)
	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanEditMembers))
	ret.Actions.CanEdit, _ = c.checkObject(ctx, entTk.WithUserName(user.Name()).WithAction(CanEdit))
	return ret, nil
}

func (c *Checker) FeedVersionAddPermission(ctx context.Context, user auth.User, addUser string, fvid int, relation Relation) error {
	tk := TupleKey{}.WithUserName(addUser).WithObjectID(FeedVersionType, fvid).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", fvid).Msg("FeedVersionAddPermission")
	return c.addObjectTuple(ctx, user, CanEditMembers, tk)
}

func (c *Checker) FeedVersionRemovePermission(ctx context.Context, user auth.User, removeUser string, fvid int, relation Relation) error {
	tk := TupleKey{}.WithUserName(removeUser).WithObjectID(FeedVersionType, fvid).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", fvid).Msg("FeedVersionRemovePermission")
	return c.removeObjectTuple(ctx, user, CanEditMembers, tk)
}

// ///////////////////
// internal
// ///////////////////

func (c *Checker) listObjects(ctx context.Context, user auth.User, objectType ObjectType, action Action) ([]TupleKey, error) {
	tk := TupleKey{ObjectType: objectType}.WithAction(action).WithUserName(user.Name())
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

func (c *Checker) checkObject(ctx context.Context, tk TupleKey) (bool, error) {
	return c.authz.Check(ctx, tk)
}

func (c *Checker) getObjectTuples(ctx context.Context, user auth.User, checkAction Action, tk TupleKey) ([]TupleKey, error) {
	if ok, err := c.checkObject(ctx, tk.WithUserName(user.Name()).WithAction(checkAction)); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("unauthorized")
	}
	return c.authz.GetObjectTuples(ctx, tk)
}

func (c *Checker) addObjectTuple(ctx context.Context, checkUser auth.User, checkAction Action, tk TupleKey) error {
	log.Trace().Str("tk", tk.String()).Msg("addObjectTuple")
	checkTk := TupleKey{}.WithUserName(checkUser.Name()).WithObject(tk.ObjectType, tk.ObjectName).WithAction(checkAction)
	if ok, err := c.checkObject(ctx, checkTk); err != nil {
		return err
	} else if !ok {
		return errors.New("unauthorized")
	}
	return c.authz.WriteTuple(ctx, tk)
}

func (c *Checker) removeObjectTuple(ctx context.Context, checkUser auth.User, checkAction Action, tk TupleKey) error {
	log.Trace().Str("tk", tk.String()).Msg("removeObjectTuple")
	checkTk := TupleKey{}.WithUserName(checkUser.Name()).WithObject(tk.ObjectType, tk.ObjectName).WithAction(checkAction)
	if ok, err := c.checkObject(ctx, checkTk); err != nil {
		return err
	} else if !ok {
		return errors.New("unauthorized")
	}
	return c.authz.DeleteTuple(ctx, tk)
}

func (c *Checker) replaceObjectTuple(ctx context.Context, checkUser auth.User, checkAction Action, tk TupleKey) error {
	log.Trace().Str("tk", tk.String()).Msg("replaceObjectTuple")
	checkTk := TupleKey{}.WithUserName(checkUser.Name()).WithObject(tk.ObjectType, tk.ObjectName).WithAction(checkAction)
	if ok, err := c.checkObject(ctx, checkTk); err != nil {
		return err
	} else if !ok {
		return errors.New("unauthorized")
	}
	return c.authz.ReplaceTuple(ctx, tk)
}

type responseId struct {
	ID int `json:"id" db:"id"`
}

func (e *responseId) GetID() int {
	return e.ID
}

func (e *responseId) SetID(v int) {
	e.ID = v
}

func (e *responseId) SetIDString(v string) {
	e.ID, _ = strconv.Atoi(v)
}

type hydratable interface {
	GetID() int
	SetID(int)
	TableName() string
}

func hydrate[T any, PT interface {
	*T
	hydratable
}](ctx context.Context, db sqlx.Ext, table string, id int) (T, error) {
	var ret T
	r, err := hydrates[T, PT](ctx, db, table, []int{id})
	if err != nil {
		return ret, nil
	}
	if len(r) > 0 {
		ret = r[0]
	}
	return ret, nil
}

func hydrates[T any, PT interface {
	*T
	hydratable
}](ctx context.Context, db sqlx.Ext, table string, ids []int) ([]T, error) {
	var dbr []PT
	q := sq.StatementBuilder.Select("id", "onestop_id", "name").From(table).Where(sq.Eq{"id": ids})
	if err := find.Select(ctx, db, q, &dbr); err != nil {
		log.Trace().Err(err).Msg("hydrateFeeds")
	}
	byId := map[int]PT{}
	for _, f := range dbr {
		byId[f.GetID()] = f
	}
	var ret []T
	for _, id := range ids {
		a := byId[id]
		a.SetID(id)
		ret = append(ret, *a)
	}
	return ret, nil
}
