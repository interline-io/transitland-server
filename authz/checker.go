package authz

import (
	"context"
	"errors"

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
	Check(context.Context, TupleKey, ...TupleKey) (bool, error)
	ListObjects(context.Context, TupleKey) ([]TupleKey, error)
	GetObjectTuples(context.Context, TupleKey) ([]TupleKey, error)
	WriteTuple(context.Context, TupleKey) error
	ReplaceTuple(context.Context, TupleKey) error
	DeleteTuple(context.Context, TupleKey) error
}

var ErrUnauthorized = errors.New("unauthorized")

type Checker struct {
	authn        AuthnProvider
	authz        AuthzProvider
	db           sqlx.Ext
	globalAdmins []string
}

func NewCheckerFromConfig(cfg AuthzConfig, db sqlx.Ext, redisClient *redis.Client) (*Checker, error) {
	var authn AuthnProvider
	authn = NewMockAuthnClient()
	var authz AuthzProvider
	authz = NewMockAuthzClient()

	// Load Auth0 if configured
	if cfg.Auth0Domain != "" {
		var err error
		authn, err = NewAuth0Client(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)
		if err != nil {
			return nil, err
		}
	}

	// Load FGA if configured
	if cfg.FGAEndpoint != "" {
		fgac, err := NewFGAClient(cfg.FGAEndpoint, cfg.FGAStoreID, cfg.FGAModelID)
		if err != nil {
			return nil, err
		}
		authz = fgac
		// Create test FGA environment
		if cfg.FGALoadModelFile != "" {
			if cfg.FGAStoreID == "" {
				if _, err := fgac.CreateStore(context.Background(), "test"); err != nil {
					return nil, err
				}
			}
			if _, err := fgac.CreateModel(context.Background(), cfg.FGALoadModelFile); err != nil {
				return nil, err
			}
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

type TenantResponseList struct {
	Tenants []TenantResponse `json:"tenants"`
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

func (c *Checker) TenantList(ctx context.Context, checkUser auth.User) (*TenantResponseList, error) {
	ids, err := c.listUserObjects(ctx, checkUser, TenantType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := hydrates[TenantResponse](ctx, c.db, ids)
	return &TenantResponseList{Tenants: t}, err
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
	entKey := NewEntityID(TenantType, tenantId)
	ent, err := c.Tenant(ctx, checkUser, tenantId)
	if err != nil {
		return nil, err
	}
	ret := &TenantPermissionsResponse{TenantResponse: *ent}

	// Actions
	groupIds, _ := c.listSubjectRelations(ctx, entKey, GroupType, ParentRelation)
	ret.Groups, _ = hydrates[GroupResponse](ctx, c.db, groupIds)
	ret.Actions.CanView, _ = c.checkObject(ctx, checkUser, CanView, entKey)
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, checkUser, CanEditMembers, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey)
	ret.Actions.CanCreateOrg, _ = c.checkObject(ctx, checkUser, CanCreateOrg, entKey)
	ret.Actions.CanDeleteOrg, _ = c.checkObject(ctx, checkUser, CanDeleteOrg, entKey)

	// Get tenant metadata
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == AdminRelation {
			ret.Users.Admins = append(ret.Users.Admins, User{ID: tk.Subject.Name})
		}
		if tk.Relation == MemberRelation {
			ret.Users.Members = append(ret.Users.Members, User{ID: tk.Subject.Name})
		}
	}
	ret.Users.Admins, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Admins)
	ret.Users.Members, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Members)
	return ret, nil
}

func (c *Checker) TenantSave(ctx context.Context, checkUser auth.User, tenantId int, newName string) error {
	if tenantCheck, err := c.TenantPermissions(ctx, checkUser, tenantId); err != nil {
		return err
	} else if !tenantCheck.Actions.CanEdit {
		return ErrUnauthorized
	}
	log.Trace().Str("tenantName", newName).Int("id", tenantId).Msg("TenantSave")
	_, err := sq.StatementBuilder.
		RunWith(c.db).
		PlaceholderFormat(sq.Dollar).
		Update("tl_tenants").
		SetMap(map[string]any{
			"tenant_name": newName,
		}).
		Where("id = ?", tenantId).Exec()
	return err
}

func (c *Checker) TenantAddPermission(ctx context.Context, checkUser auth.User, tenantId int, addUser string, relation Relation) error {
	if check, err := c.TenantPermissions(ctx, checkUser, tenantId); err != nil {
		return err
	} else if !check.Actions.CanEditMembers {
		return ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(addUser).WithObjectID(TenantType, tenantId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", tenantId).Msg("TenantAddPermission")
	return c.authz.ReplaceTuple(ctx, tk)
}

func (c *Checker) TenantRemovePermission(ctx context.Context, checkUser auth.User, tenantId int, removeUser string, relation Relation) error {
	if check, err := c.TenantPermissions(ctx, checkUser, tenantId); err != nil {
		return err
	} else if !check.Actions.CanEditMembers {
		return ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(removeUser).WithObjectID(TenantType, tenantId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", tenantId).Msg("TenantRemovePermission")
	return c.authz.DeleteTuple(ctx, tk)
}

func (c *Checker) TenantCreate(ctx context.Context, checkUser auth.User, tenantName string) (int, error) {
	return 0, nil
}

func (c *Checker) TenantCreateGroup(ctx context.Context, checkUser auth.User, tenantId int, groupName string) (int, error) {
	if check, err := c.TenantPermissions(ctx, checkUser, tenantId); err != nil {
		return 0, err
	} else if !check.Actions.CanCreateOrg {
		return 0, ErrUnauthorized
	}
	entKey := NewEntityID(TenantType, tenantId)
	log.Trace().Str("groupName", groupName).Int("id", tenantId).Msg("TenantCreateGroup")
	groupId := 0
	err := sq.StatementBuilder.
		RunWith(c.db).
		PlaceholderFormat(sq.Dollar).
		Insert("tl_groups").
		Columns("id", "group_name").
		Values(sq.Expr("(select max(id)+1 from tl_groups)"), groupName).
		Suffix(`RETURNING "id"`).
		QueryRow().
		Scan(&groupId)
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

type GroupResponseList struct {
	Groups []GroupResponse `json:"groups"`
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

func (c *Checker) GroupList(ctx context.Context, checkUser auth.User) (*GroupResponseList, error) {
	ids, err := c.listUserObjects(ctx, checkUser, GroupType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := hydrates[GroupResponse](ctx, c.db, ids)
	return &GroupResponseList{Groups: t}, err
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

func (c *Checker) GroupPermissions(ctx context.Context, checkUser auth.User, groupId int) (*GroupPermissionsResponse, error) {
	entKey := NewEntityID(GroupType, groupId)
	ent, err := c.Group(ctx, checkUser, groupId)
	if err != nil {
		return nil, err
	}
	ret := &GroupPermissionsResponse{GroupResponse: *ent}

	// Actions
	ret.Actions.CanView, _ = c.checkObject(ctx, checkUser, CanView, entKey)
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, checkUser, CanEditMembers, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey)
	ret.Actions.CanCreateFeed, _ = c.checkObject(ctx, checkUser, CanCreateFeed, entKey)
	ret.Actions.CanDeleteFeed, _ = c.checkObject(ctx, checkUser, CanDeleteFeed, entKey)

	// Get feeds
	feedIds, _ := c.listSubjectRelations(ctx, entKey, FeedType, ParentRelation)
	ret.Feeds, _ = hydrates[FeedResponse](ctx, c.db, feedIds)

	// Get group metadata
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey)
	if err != nil {
		return nil, err
	}
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
	ret.Users.Managers, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Managers)
	ret.Users.Editors, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Editors)
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Viewers)
	return ret, nil
}

func (c *Checker) GroupSave(ctx context.Context, checkUser auth.User, groupId int, newName string) error {
	if check, err := c.GroupPermissions(ctx, checkUser, groupId); err != nil {
		return err
	} else if !check.Actions.CanEdit {
		return ErrUnauthorized
	}
	log.Trace().Str("groupName", newName).Int("id", groupId).Msg("GroupSave")
	_, err := sq.StatementBuilder.
		RunWith(c.db).
		PlaceholderFormat(sq.Dollar).
		Update("tl_groups").
		SetMap(map[string]any{
			"group_name": newName,
		}).
		Where("id = ?", groupId).Exec()
	return err
}

func (c *Checker) GroupAddPermission(ctx context.Context, checkUser auth.User, addUser string, groupId int, relation Relation) error {
	if check, err := c.GroupPermissions(ctx, checkUser, groupId); err != nil {
		return err
	} else if !check.Actions.CanEditMembers {
		return ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(addUser).WithObjectID(GroupType, groupId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", groupId).Msg("GroupAddPermission")
	return c.authz.ReplaceTuple(ctx, tk)
}

func (c *Checker) GroupRemovePermission(ctx context.Context, checkUser auth.User, removeUser string, groupId int, relation Relation) error {
	if check, err := c.GroupPermissions(ctx, checkUser, groupId); err != nil {
		return err
	} else if !check.Actions.CanEditMembers {
		return ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(removeUser).WithObjectID(GroupType, groupId).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", groupId).Msg("GroupRemovePermission")
	return c.authz.DeleteTuple(ctx, tk)
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

type FeedResponseList struct {
	Feeds []FeedResponse `json:"feeds"`
}

type FeedPermissionsResponse struct {
	FeedResponse
	Group *GroupResponse `json:"group"`
	Users struct {
	} `json:"users"`
	Actions struct {
		CanView              bool `json:"can_view"`
		CanEdit              bool `json:"can_edit"`
		CanSetGroup          bool `json:"can_set_group"`
		CanCreateFeedVersion bool `json:"can_create_feed_version"`
		CanDeleteFeedVersion bool `json:"can_delete_feed_version"`
	} `json:"actions"`
}

func (c *Checker) FeedList(ctx context.Context, checkUser auth.User) (*FeedResponseList, error) {
	feedIds, err := c.listUserObjects(ctx, checkUser, FeedType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := hydrates[FeedResponse](ctx, c.db, feedIds)
	return &FeedResponseList{Feeds: t}, err
}

func (c *Checker) Feed(ctx context.Context, checkUser auth.User, feedId int) (*FeedResponse, error) {
	if err := c.checkObjectOrError(ctx, checkUser, CanView, NewEntityID(FeedType, feedId)); err != nil {
		return nil, err
	}
	r, err := hydrate[FeedResponse](ctx, c.db, feedId)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Checker) FeedPermissions(ctx context.Context, checkUser auth.User, feedId int) (*FeedPermissionsResponse, error) {
	entKey := NewEntityID(FeedType, feedId)
	ent, err := c.Feed(ctx, checkUser, feedId)
	if err != nil {
		return nil, err
	}
	ret := &FeedPermissionsResponse{FeedResponse: *ent}

	// Actions
	ret.Actions.CanView, _ = c.checkObject(ctx, checkUser, CanView, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey)
	ret.Actions.CanSetGroup, _ = c.checkObject(ctx, checkUser, CanSetGroup, entKey)
	ret.Actions.CanCreateFeedVersion, _ = c.checkObject(ctx, checkUser, CanCreateFeedVersion, entKey)
	ret.Actions.CanDeleteFeedVersion, _ = c.checkObject(ctx, checkUser, CanDeleteFeedVersion, entKey)

	// Get feed metadata
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ret.Group, _ = c.Group(ctx, checkUser, tk.Subject.ID())
		}
	}
	return ret, nil
}

func (c *Checker) FeedSetGroup(ctx context.Context, checkUser auth.User, feedId int, newGroup int) error {
	if check, err := c.FeedPermissions(ctx, checkUser, feedId); err != nil {
		return err
	} else if !check.Actions.CanSetGroup {
		return ErrUnauthorized
	}
	tk := NewTupleKey().WithSubjectID(GroupType, newGroup).WithObjectID(FeedType, feedId).WithRelation(ParentRelation)
	log.Trace().Str("tk", tk.String()).Int("id", feedId).Msg("FeedSetGroup")
	return c.authz.ReplaceTuple(ctx, tk)
}

/////////////////////
// FEED VERSIONS
/////////////////////

type FeedVersionResponse struct {
	responseId
	Name   tt.String `json:"name" db:"name"`
	SHA1   tt.String `json:"sha1" db:"sha1"`
	FeedID int       `json:"feed_id" db:"feed_id"`
}

func (t FeedVersionResponse) TableName() string {
	return "feed_versions"
}

type FeedVersionResponseList struct {
	FeedVersions []FeedVersionResponse `json:"feed_versions"`
}

type FeedVersionPermissionsResponse struct {
	FeedVersionResponse
	Users struct {
		Viewers []User `json:"viewers"`
		Editors []User `json:"editors"`
	} `json:"users"`
	Actions struct {
		CanView        bool `json:"can_view"`
		CanEditMembers bool `json:"can_edit_members"`
		CanEdit        bool `json:"can_edit"`
	} `json:"actions"`
}

func (c *Checker) FeedVersionList(ctx context.Context, user auth.User) (*FeedVersionResponseList, error) {
	fvids, err := c.listUserObjects(ctx, user, FeedVersionType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := hydrates[FeedVersionResponse](ctx, c.db, fvids)
	return &FeedVersionResponseList{FeedVersions: t}, err
}

func (c *Checker) FeedVersion(ctx context.Context, checkUser auth.User, fvid int) (*FeedVersionResponse, error) {
	// We need to get feed id before any other checks
	// If there is a "not found" error here, save it for after the global admin check
	// This is for consistency with other permission checks
	r, fvErr := hydrate[FeedVersionResponse](ctx, c.db, fvid)
	ctxTk := NewTupleKey().WithObjectID(FeedVersionType, r.ID).WithSubjectID(FeedType, r.FeedID).WithRelation(ParentRelation)
	if err := c.checkObjectOrError(ctx, checkUser, CanView, NewEntityID(FeedVersionType, fvid), ctxTk); err != nil {
		return nil, err
	}
	// Now return deferred fvErr
	if fvErr != nil {
		return nil, fvErr
	}
	return &r, nil
}

func (c *Checker) FeedVersionPermissions(ctx context.Context, checkUser auth.User, fvid int) (*FeedVersionPermissionsResponse, error) {
	entKey := NewEntityID(FeedVersionType, fvid)
	ent, err := c.FeedVersion(ctx, checkUser, fvid)
	if err != nil {
		return nil, err
	}
	ret := &FeedVersionPermissionsResponse{FeedVersionResponse: *ent}
	ctxTk := NewTupleKey().WithObjectID(FeedVersionType, ent.ID).WithSubjectID(FeedType, ent.FeedID).WithRelation(ParentRelation)

	// Actions
	ret.Actions.CanView, _ = c.checkObject(ctx, checkUser, CanView, entKey, ctxTk)
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, checkUser, CanEditMembers, entKey, ctxTk)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, checkUser, CanEdit, entKey, ctxTk)

	// Get fv metadata
	tps, err := c.getObjectTuples(ctx, checkUser, CanView, entKey, ctxTk)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == EditorRelation {
			ret.Users.Editors = append(ret.Users.Editors, User{ID: tk.Subject.Name})
		}
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, User{ID: tk.Subject.Name})
		}
	}
	ret.Users.Editors, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Editors)
	ret.Users.Viewers, _ = c.hydrateUsers(ctx, checkUser, ret.Users.Viewers)
	return ret, nil
}

func (c *Checker) FeedVersionAddPermission(ctx context.Context, checkUser auth.User, addUser string, fvid int, relation Relation) error {
	if check, err := c.FeedVersionPermissions(ctx, checkUser, fvid); err != nil {
		return err
	} else if !check.Actions.CanEditMembers {
		return ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(addUser).WithObjectID(FeedVersionType, fvid).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", fvid).Msg("FeedVersionAddPermission")
	return c.authz.ReplaceTuple(ctx, tk)
}

func (c *Checker) FeedVersionRemovePermission(ctx context.Context, checkUser auth.User, removeUser string, fvid int, relation Relation) error {
	if check, err := c.FeedVersionPermissions(ctx, checkUser, fvid); err != nil {
		return err
	} else if !check.Actions.CanEditMembers {
		return ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(removeUser).WithObjectID(FeedVersionType, fvid).WithRelation(relation)
	log.Trace().Str("tk", tk.String()).Int("id", fvid).Msg("FeedVersionRemovePermission")
	return c.authz.DeleteTuple(ctx, tk)
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

func (c *Checker) getObjectTuples(ctx context.Context, checkUser auth.User, checkAction Action, obj EntityKey, ctxtk ...TupleKey) ([]TupleKey, error) {
	if err := c.checkObjectOrError(ctx, checkUser, checkAction, obj, ctxtk...); err != nil {
		return nil, err
	}
	return c.authz.GetObjectTuples(ctx, NewTupleKey().WithObject(obj.Type, obj.Name))
}

func (c *Checker) checkObjectOrError(ctx context.Context, checkUser auth.User, checkAction Action, obj EntityKey, ctxtk ...TupleKey) error {
	ok, err := c.checkObject(ctx, checkUser, checkAction, obj, ctxtk...)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnauthorized
	}
	return nil
}

func (c *Checker) checkObject(ctx context.Context, checkUser auth.User, checkAction Action, obj EntityKey, ctxtk ...TupleKey) (bool, error) {
	userName := checkUser.Name()
	for _, v := range c.globalAdmins {
		if v == userName {
			log.Debug().Str("check_user", userName).Str("obj", obj.String()).Str("check_action", checkAction.String()).Msg("global admin action")
			return true, nil
		}
	}
	checkTk := NewTupleKey().WithUser(userName).WithObject(obj.Type, obj.Name).WithAction(checkAction)
	ret, err := c.authz.Check(ctx, checkTk, ctxtk...)
	log.Trace().Str("tk", checkTk.String()).Bool("result", ret).Err(err).Msg("checkObject")
	return ret, err
}

type responseId struct {
	ID int `json:"id" db:"id"`
}

func (e *responseId) GetID() int {
	return e.ID
}

type hydratable interface {
	GetID() int
	TableName() string
}

func hydrate[T any, PT interface {
	*T
	hydratable
}](ctx context.Context, db sqlx.Ext, id int) (T, error) {
	var ret T
	r, err := hydrates[T, PT](ctx, db, []int{id})
	if err != nil {
		return ret, err
	}
	if len(r) == 0 {
		return ret, errors.New("not found")
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
		log.Trace().Err(err).Msg("hydrates")
	}
	byId := map[int]PT{}
	for _, f := range dbr {
		byId[f.GetID()] = f
	}
	ret := make([]PT, len(ids))
	for i, id := range ids {
		if b, ok := byId[id]; !ok {
			return nil, errors.New("not found")
		} else if b == nil {
			return nil, errors.New("not found (got nil)")
		} else {
			ret[i] = b
		}
	}
	ret2 := make([]T, 0, len(ids))
	for _, r := range ret {
		ret2 = append(ret2, *r)
	}
	return ret2, nil
}
