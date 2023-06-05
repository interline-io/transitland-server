package authz

import (
	"context"
	"errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/find"
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
	authn        AuthnProvider
	authz        AuthzProvider
	db           sqlx.Ext
	globalAdmins []string
}

func NewCheckerFromConfig(cfg AuthzConfig, db sqlx.Ext, redisClient *redis.Client) (*Checker, error) {
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
	checker := NewChecker(authn, authz, db, redisClient)
	if cfg.GlobalAdmin != "" {
		checker.globalAdmins = append(checker.globalAdmins, cfg.GlobalAdmin)
	}
	return checker, nil
}

func NewChecker(n AuthnProvider, p AuthzProvider, db sqlx.Ext, redisClient *redis.Client) *Checker {
	return &Checker{
		authn: n,
		authz: p,
		db:    db,
	}
}

// ///////////////////
// USERS
// ///////////////////

func (c *Checker) UserList(ctx context.Context, checkUser auth.User, query string) ([]User, error) {
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

func (c *Checker) User(ctx context.Context, checkUser auth.User, userId string) (*User, error) {
	// TODO: filter users
	ret, err := c.authn.UserByID(ctx, userId)
	return ret, err
}

func (c *Checker) hydrateUsers(ctx context.Context, checkUser auth.User, users []User) ([]User, error) {
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
	Name tt.String `json:"name" db:"tenant_name"`
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

func (c *Checker) TenantList(ctx context.Context, checkUser auth.User) ([]TenantResponse, error) {
	ids, err := c.listUserObjects(ctx, checkUser, TenantType, CanView)
	if err != nil {
		return nil, err
	}
	return hydrates[TenantResponse](ctx, c.db, ids)
}

func (c *Checker) Tenant(ctx context.Context, checkUser auth.User, tenantId int) (*TenantResponse, error) {
	if err := c.checkObjectOrError(ctx, checkUser, CanView, NewEntityID(TenantType, tenantId)); err != nil {
		return nil, err
	}
	r, err := hydrate[TenantResponse](ctx, c.db, tenantId)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Checker) TenantPermissions(ctx context.Context, checkUser auth.User, tenantId int) (*TenantPermissionsResponse, error) {
	// Check tenant access
	entKey := NewEntityID(TenantType, tenantId)
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey)
	if err != nil {
		return nil, err
	}

	// Get tenant metadata
	ret := &TenantPermissionsResponse{}
	ret.TenantResponse, _ = hydrate[TenantResponse](ctx, c.db, tenantId)
	for _, tk := range tps {
		if tk.Relation == AdminRelation {
			ret.Users.Admins = append(ret.Users.Admins, User{ID: tk.Subject.Name})
		}
		if tk.Relation == MemberRelation {
			ret.Users.Members = append(ret.Users.Members, User{ID: tk.Subject.Name})
		}
	}

	groupIds, _ := c.listSubjectRelations(ctx, entKey, GroupType, ParentRelation)
	ret.Groups, _ = hydrates[GroupResponse](ctx, c.db, groupIds)

	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, checkUser, CanEditMembers, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey)
	ret.Actions.CanCreateOrg, _ = c.checkObject(ctx, checkUser, CanCreateOrg, entKey)
	ret.Actions.CanDeleteOrg, _ = c.checkObject(ctx, checkUser, CanDeleteOrg, entKey)
	ret.Users.Admins, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Admins)
	ret.Users.Members, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Members)
	return ret, nil
}

func (c *Checker) TenantSave(ctx context.Context, checkUser auth.User, tenantId int, newName string) (int, error) {
	log.Trace().Str("tenantName", newName).Int("id", tenantId).Msg("TenantSave")
	id := 0
	err := sq.StatementBuilder.
		RunWith(c.db).
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
	tk := NewTupleKey().WithUser(addUser).WithObjectID(TenantType, tenantId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", tenantId).Msg("TenantAddPermission")
	return c.replaceObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

func (c *Checker) TenantRemovePermission(ctx context.Context, checkUser auth.User, tenantId int, removeUser string, relation Relation) error {
	tk := NewTupleKey().WithUser(removeUser).WithObjectID(TenantType, tenantId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", tenantId).Msg("TenantRemovePermission")
	return c.removeObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

func (c *Checker) TenantCreate(ctx context.Context, checkUser auth.User, tenantName string) (int, error) {
	return 0, nil
}

func (c *Checker) TenantCreateGroup(ctx context.Context, checkUser auth.User, tenantId int, groupName string) (int, error) {
	entKey := NewEntityID(TenantType, tenantId)
	if err := c.checkObjectOrError(ctx, checkUser, CanCreateOrg, entKey); err != nil {
		return 0, err
	}
	log.Trace().Str("groupName", groupName).Int("id", tenantId).Msg("TenantCreateGroup")
	groupId := 0
	err := sq.StatementBuilder.
		RunWith(c.db).
		PlaceholderFormat(sq.Dollar).
		Insert("tl_groups").
		Columns("id", "group_name").
		Values(sq.Expr("(select max(id)+1 from tl_groups)"), groupName).
		Suffix(`RETURNING "id"`).
		QueryRow().Scan(&groupId)
	if err != nil {
		return 0, err
	}
	addTk := NewTupleKey().WithSubject(entKey.Type, entKey.Name).WithObjectID(GroupType, groupId).WithRelation(ParentRelation)
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
	Name tt.String `json:"name" db:"group_name"`
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

func (c *Checker) Group(ctx context.Context, checkUser auth.User, groupId int) (*GroupResponse, error) {
	entKey := NewEntityID(GroupType, groupId)
	if err := c.checkObjectOrError(ctx, checkUser, CanView, entKey); err != nil {
		return nil, err
	}
	r, err := hydrate[GroupResponse](ctx, c.db, groupId)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Checker) GroupList(ctx context.Context, checkUser auth.User) ([]GroupResponse, error) {
	ids, err := c.listUserObjects(ctx, checkUser, GroupType, CanView)
	if err != nil {
		return nil, err
	}
	return hydrates[GroupResponse](ctx, c.db, ids)
}

func (c *Checker) GroupPermissions(ctx context.Context, checkUser auth.User, groupId int) (*GroupPermissionsResponse, error) {
	// Check group access
	entKey := NewEntityID(GroupType, groupId)
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey)
	if err != nil {
		return nil, err
	}

	// Get group metadata
	ret := &GroupPermissionsResponse{}
	ret.GroupResponse, _ = hydrate[GroupResponse](ctx, c.db, groupId)

	// Get group relations
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ret.Tenant, _ = c.Tenant(ctx, checkUser, tk.Subject.ID())
		}
		if tk.Relation == ManagerRelation {
			ret.Users.Managers = append(ret.Users.Managers, User{ID: tk.Subject.Name})
		}
		if tk.Relation == EditorRelation {
			ret.Users.Editors = append(ret.Users.Editors, User{ID: tk.Subject.Name})
		}
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, User{ID: tk.Subject.Name})
		}
	}

	// Get feeds
	feedIds, _ := c.listSubjectRelations(ctx, entKey, FeedType, ParentRelation)
	ret.Feeds, _ = hydrates[FeedResponse](ctx, c.db, feedIds)

	// Prepare response
	ret.Users.Managers, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Managers)
	ret.Users.Editors, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Editors)
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Viewers)
	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, checkUser, CanEditMembers, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey)
	ret.Actions.CanCreateFeed, _ = c.checkObject(ctx, checkUser, CanCreateFeed, entKey)
	ret.Actions.CanDeleteFeed, _ = c.checkObject(ctx, checkUser, CanDeleteFeed, entKey)
	return ret, nil
}

func (c *Checker) GroupSave(ctx context.Context, checkUser auth.User, groupId int, newName string) (int, error) {
	log.Trace().Str("groupName", newName).Int("id", groupId).Msg("GroupSave")
	id := 0
	err := sq.StatementBuilder.
		RunWith(c.db).
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
	tk := NewTupleKey().WithUser(addUser).WithObjectID(GroupType, groupId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", groupId).Msg("GroupAddPermission")
	return c.replaceObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

func (c *Checker) GroupRemovePermission(ctx context.Context, checkUser auth.User, removeUser string, groupId int, relation Relation) error {
	tk := NewTupleKey().WithUser(removeUser).WithObjectID(GroupType, groupId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", groupId).Msg("GroupRemovePermission")
	return c.removeObjectTuple(ctx, checkUser, CanEditMembers, tk)
}

/////////////////////
// FEEDS
/////////////////////

type FeedResponse struct {
	responseId
	OnestopID tt.String `json:"onestop_id" db:"onestop_id"`
	Name      tt.String `json:"name" db:"name"`
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

func (c *Checker) FeedList(ctx context.Context, checkUser auth.User) ([]FeedResponse, error) {
	feedIds, err := c.listUserObjects(ctx, checkUser, FeedType, CanView)
	if err != nil {
		return nil, err
	}
	return hydrates[FeedResponse](ctx, c.db, feedIds)
}

func (c *Checker) FeedPermissions(ctx context.Context, checkUser auth.User, feedId int) (*FeedPermissionsResponse, error) {
	entKey := NewEntityID(FeedType, feedId)
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey)
	if err != nil {
		return nil, err
	}
	ret := &FeedPermissionsResponse{}
	ret.FeedResponse, _ = hydrate[FeedResponse](ctx, c.db, feedId)
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ret.Group, _ = c.Group(ctx, checkUser, tk.Subject.ID())
		}
	}
	ret.Actions.CanView = true
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey)
	ret.Actions.CanCreateFeedVersion, _ = c.checkObject(ctx, checkUser, CanCreateFeedVersion, entKey)
	ret.Actions.CanDeleteFeedVersion, _ = c.checkObject(ctx, checkUser, CanDeleteFeedVersion, entKey)
	return ret, nil
}

func (c *Checker) FeedSetGroup(ctx context.Context, checkUser auth.User, feedId int, newGroup int) error {
	tk := NewTupleKey().WithSubjectID(FeedType, feedId).WithObjectID(GroupType, newGroup).WithRelation(ParentRelation)
	log.Trace().Str("tk", tk.String()).Int("id", feedId).Msg("FeedSetGroup")
	return c.replaceObjectTuple(ctx, checkUser, CanEdit, tk)
}

/////////////////////
// FEED VERSIONS
/////////////////////

type FeedVersionResponse struct {
	responseId
	Name string `json:"name" db:"-"`
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

func (c *Checker) FeedVersionList(ctx context.Context, user auth.User) ([]FeedVersionResponse, error) {
	var ret []FeedVersionResponse
	feedIds, err := c.listUserObjects(ctx, user, FeedVersionType, CanView)
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

func (c *Checker) FeedVersionPermissions(ctx context.Context, checkUser auth.User, fvid int) (*FeedVersionPermissionsResponse, error) {
	entKey := NewEntityID(FeedVersionType, fvid)
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey)
	if err != nil {
		return nil, err
	}
	ret := &FeedVersionPermissionsResponse{}
	ret.ID = fvid
	for _, tk := range tps {
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, User{ID: tk.Subject.Name})
		}
	}
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Viewers)
	ret.Actions.CanView = true
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, checkUser, CanEditMembers, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey)
	return ret, nil
}

func (c *Checker) FeedVersionAddPermission(ctx context.Context, user auth.User, addUser string, fvid int, relation Relation) error {
	tk := NewTupleKey().WithUser(addUser).WithObjectID(FeedVersionType, fvid).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", fvid).Msg("FeedVersionAddPermission")
	return c.replaceObjectTuple(ctx, user, CanEditMembers, tk)
}

func (c *Checker) FeedVersionRemovePermission(ctx context.Context, user auth.User, removeUser string, fvid int, relation Relation) error {
	tk := NewTupleKey().WithUser(removeUser).WithObjectID(FeedVersionType, fvid).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", fvid).Msg("FeedVersionRemovePermission")
	return c.removeObjectTuple(ctx, user, CanEditMembers, tk)
}

func (c *Checker) FeedVersionSetFeed(ctx context.Context, user auth.User, fvid int, feedId int) error {
	tk := NewTupleKey().WithSubjectID(FeedType, feedId).WithObjectID(FeedVersionType, fvid).WithRelation(ParentRelation)
	log.Trace().Str("tk", tk.String()).Int("id", fvid).Msg("FeedVersionSetFeed")
	return c.replaceObjectTuple(ctx, user, CanCreateFeedVersion, tk)
}

// ///////////////////
// internal
// ///////////////////

func (c *Checker) listUserObjects(ctx context.Context, user auth.User, objectType ObjectType, action Action) ([]int, error) {
	tk := NewTupleKey().WithUser(user.Name()).WithObject(objectType, "").WithAction(action)
	objTks, err := c.authz.ListObjects(ctx, tk)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var ret []int
	for _, tk := range objTks {
		ret = append(ret, tk.Object.ID())
	}
	return ret, nil
}

func (c *Checker) listSubjectRelations(ctx context.Context, sub EntityKey, objectType ObjectType, relation Relation) ([]int, error) {
	tk := NewTupleKey().WithSubject(sub.Type, sub.Name).WithObject(objectType, "").WithRelation(relation)
	rels, err := c.authz.ListObjects(ctx, tk)
	if err != nil {
		return nil, err
	}
	var ret []int
	for _, v := range rels {
		ret = append(ret, v.Object.ID())
	}
	return ret, nil
}

func (c *Checker) getObjectTuples(ctx context.Context, checkUser auth.User, checkAction Action, obj EntityKey) ([]TupleKey, error) {
	if err := c.checkObjectOrError(ctx, checkUser, checkAction, obj); err != nil {
		return nil, err
	}
	return c.authz.GetObjectTuples(ctx, NewTupleKey().WithObject(obj.Type, obj.Name))
}

func (c *Checker) addObjectTuple(ctx context.Context, checkUser auth.User, checkAction Action, tk TupleKey) error {
	if err := c.checkObjectOrError(ctx, checkUser, checkAction, tk.Object); err != nil {
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("addObjectTuple")
	return c.authz.WriteTuple(ctx, tk)
}

func (c *Checker) removeObjectTuple(ctx context.Context, checkUser auth.User, checkAction Action, tk TupleKey) error {
	if err := c.checkObjectOrError(ctx, checkUser, checkAction, tk.Object); err != nil {
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("removeObjectTuple")
	return c.authz.DeleteTuple(ctx, tk)
}

func (c *Checker) replaceObjectTuple(ctx context.Context, checkUser auth.User, checkAction Action, tk TupleKey) error {
	if err := c.checkObjectOrError(ctx, checkUser, checkAction, tk.Object); err != nil {
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("replaceObjectTuple")
	return c.authz.ReplaceTuple(ctx, tk)
}

func (c *Checker) checkObjectOrError(ctx context.Context, checkUser auth.User, checkAction Action, obj EntityKey) error {
	ok, err := c.checkObject(ctx, checkUser, checkAction, obj)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("unauthorized")
	}
	return nil
}

func (c *Checker) checkObject(ctx context.Context, checkUser auth.User, checkAction Action, obj EntityKey) (bool, error) {
	userName := checkUser.Name()
	for _, v := range c.globalAdmins {
		if v == userName {
			return true, nil
		}
	}
	checkTk := NewTupleKey().WithUser(userName).WithObject(obj.Type, obj.Name).WithAction(checkAction)
	return c.authz.Check(ctx, checkTk)
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
}](ctx context.Context, db sqlx.Ext, id int) (T, error) {
	var ret T
	r, err := hydrates[T, PT](ctx, db, []int{id})
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
}](ctx context.Context, db sqlx.Ext, ids []int) ([]T, error) {
	var dbr []PT
	// TODO: not *
	var xt PT = new(T)
	q := sq.StatementBuilder.Select("*").From(xt.TableName()).Where(sq.Eq{"id": ids})
	if err := find.Select(ctx, db, q, &dbr); err != nil {
		log.Trace().Err(err).Msg("hydrateFeeds")
	}
	byId := map[int]PT{}
	for _, f := range dbr {
		byId[f.GetID()] = f
	}
	ret := make([]PT, len(ids))
	for i, id := range ids {
		if b := byId[id]; b != nil {
			ret[i] = b
		} else {
			ret[i] = new(T)
			ret[i].SetID(id)
		}
	}
	ret2 := make([]T, 0, len(ids))
	for _, r := range ret {
		ret2 = append(ret2, *r)
	}
	return ret2, nil
}
