package authz

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/find"
)

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

func (c *Checker) UserList(ctx context.Context, req *UserListRequest) (*UserListResponse, error) {
	query := ""
	if req != nil {
		query = req.Q
	}
	// TODO: filter users
	users, err := c.authn.Users(ctx, query)
	if err != nil {
		return nil, err
	}
	return &UserListResponse{Users: users}, nil
}

func (c *Checker) User(ctx context.Context, req *UserRequest) (*UserResponse, error) {
	// TODO: filter users
	ret, err := c.authn.UserByID(ctx, req.Id)
	return &UserResponse{User: ret}, err
}

func (c *Checker) getUsers(ctx context.Context, users []*User) ([]*User, error) {
	// Must already be filtered for permissions
	var ret []*User
	for _, u := range users {
		uu, err := c.authn.UserByID(ctx, u.Id)
		if err == nil && uu != nil {
			ret = append(ret, uu)
		}
	}
	return ret, nil
}

// ///////////////////
// TENANTS
// ///////////////////

func (c *Checker) getTenants(ctx context.Context, ids []int64) ([]*Tenant, error) {
	return gets[*Tenant](ctx, c.db, ids, "tl_tenants", "id", "coalesce(tenant_name,'') as name")
}

func (c *Checker) TenantList(ctx context.Context, req *TenantListRequest) (*TenantListResponse, error) {
	ids, err := c.listCtxObjects(ctx, TenantType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := c.getTenants(ctx, ids)
	return &TenantListResponse{Tenants: t}, err
}

func (c *Checker) Tenant(ctx context.Context, req *TenantRequest) (*TenantResponse, error) {
	tenantId := req.Id
	if err := c.checkObjectOrError(ctx, CanView, NewEntityID(TenantType, tenantId)); err != nil {
		return nil, err
	}
	t, err := c.getTenants(ctx, []int64{tenantId})
	return &TenantResponse{Tenant: first(t)}, err
}

func (c *Checker) TenantPermissions(ctx context.Context, req *TenantRequest) (*TenantPermissionsResponse, error) {
	ent, err := c.Tenant(ctx, req)
	if err != nil {
		return nil, err
	}
	ret := &TenantPermissionsResponse{
		Tenant:  ent.Tenant,
		Actions: &TenantPermissionsResponse_Actions{},
		Users:   &TenantPermissionsResponse_Users{},
	}

	// Actions
	entKey := NewEntityID(TenantType, req.Id)
	groupIds, _ := c.listSubjectRelations(ctx, entKey, GroupType, ParentRelation)
	ret.Groups, _ = c.getGroups(ctx, groupIds)
	ret.Actions.CanView, _ = c.checkObject(ctx, CanView, entKey)
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, CanEditMembers, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, CanEdit, entKey)
	ret.Actions.CanCreateOrg, _ = c.checkObject(ctx, CanCreateOrg, entKey)
	ret.Actions.CanDeleteOrg, _ = c.checkObject(ctx, CanDeleteOrg, entKey)

	// Get tenant metadata
	tps, err := c.getObjectTuples(ctx, CanView, entKey)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == AdminRelation {
			ret.Users.Admins = append(ret.Users.Admins, newUser(tk.Subject.Name))
		}
		if tk.Relation == MemberRelation {
			ret.Users.Members = append(ret.Users.Members, newUser(tk.Subject.Name))
		}
	}
	ret.Users.Admins, _ = c.getUsers(ctx, ret.Users.Admins)
	ret.Users.Members, _ = c.getUsers(ctx, ret.Users.Members)
	return ret, nil
}

func (c *Checker) TenantSave(ctx context.Context, req *TenantSaveRequest) (*TenantSaveResponse, error) {
	tenantId := req.Tenant.Id
	if check, err := c.TenantPermissions(ctx, &TenantRequest{Id: tenantId}); err != nil {
		return nil, err
	} else if !check.Actions.CanEdit {
		return nil, ErrUnauthorized
	}
	newName := req.Tenant.Name
	log.Trace().Str("tenantName", newName).Int64("id", tenantId).Msg("TenantSave")
	_, err := sq.StatementBuilder.
		RunWith(c.db).
		PlaceholderFormat(sq.Dollar).
		Update("tl_tenants").
		SetMap(map[string]any{
			"tenant_name": newName,
		}).
		Where("id = ?", tenantId).Exec()
	return &TenantSaveResponse{}, err
}

func (c *Checker) TenantAddPermission(ctx context.Context, req *TenantModifyPermissionRequest) (*TenantSaveResponse, error) {
	tenantId := req.Id
	if check, err := c.TenantPermissions(ctx, &TenantRequest{Id: tenantId}); err != nil {
		return nil, err
	} else if !check.Actions.CanEditMembers {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(req.UserRelation.UserId).WithObjectID(TenantType, tenantId).WithRelation(req.UserRelation.Relation)
	log.Trace().Str("tk", tk.String()).Int64("id", tenantId).Msg("TenantAddPermission")
	return &TenantSaveResponse{}, c.authz.ReplaceTuple(ctx, tk)
}

func (c *Checker) TenantRemovePermission(ctx context.Context, req *TenantModifyPermissionRequest) (*TenantSaveResponse, error) {
	tenantId := req.Id
	if check, err := c.TenantPermissions(ctx, &TenantRequest{Id: tenantId}); err != nil {
		return nil, err
	} else if !check.Actions.CanEditMembers {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(req.UserRelation.UserId).WithObjectID(TenantType, tenantId).WithRelation(req.UserRelation.Relation)
	log.Trace().Str("tk", tk.String()).Int64("id", tenantId).Msg("TenantRemovePermission")
	return &TenantSaveResponse{}, c.authz.DeleteTuple(ctx, tk)
}

func (c *Checker) TenantCreate(ctx context.Context, req *TenantCreateRequest) (*TenantSaveResponse, error) {
	return &TenantSaveResponse{}, nil
}

func (c *Checker) TenantCreateGroup(ctx context.Context, req *TenantCreateGroupRequest) (*GroupSaveResponse, error) {
	tenantId := req.Id
	groupName := req.Group.Name
	if check, err := c.TenantPermissions(ctx, &TenantRequest{Id: tenantId}); err != nil {
		return nil, err
	} else if !check.Actions.CanCreateOrg {
		return nil, ErrUnauthorized
	}
	log.Trace().Str("groupName", groupName).Int64("id", tenantId).Msg("TenantCreateGroup")
	groupId := int64(0)
	err := sq.StatementBuilder.
		RunWith(c.db).
		PlaceholderFormat(sq.Dollar).
		Insert("tl_groups").
		Columns("group_name").
		Values(groupName).
		Suffix(`RETURNING "id"`).
		QueryRow().
		Scan(&groupId)
	if err != nil {
		return nil, err
	}
	addTk := NewTupleKey().WithSubjectID(TenantType, tenantId).WithObjectID(GroupType, groupId).WithRelation(ParentRelation)
	if err := c.authz.WriteTuple(ctx, addTk); err != nil {
		return nil, err
	}
	return &GroupSaveResponse{Group: &Group{Id: groupId}}, err
}

// ///////////////////
// GROUPS
// ///////////////////

func (c *Checker) getGroups(ctx context.Context, ids []int64) ([]*Group, error) {
	return gets[*Group](ctx, c.db, ids, "tl_groups", "id", "coalesce(group_name,'') as name")
}

func (c *Checker) GroupList(ctx context.Context, req *GroupListRequest) (*GroupListResponse, error) {
	ids, err := c.listCtxObjects(ctx, GroupType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := c.getGroups(ctx, ids)
	return &GroupListResponse{Groups: t}, err
}

func (c *Checker) Group(ctx context.Context, req *GroupRequest) (*GroupResponse, error) {
	groupId := req.Id
	if err := c.checkObjectOrError(ctx, CanView, NewEntityID(GroupType, groupId)); err != nil {
		return nil, err
	}
	t, err := c.getGroups(ctx, []int64{groupId})
	return &GroupResponse{Group: first(t)}, err
}

func (c *Checker) GroupPermissions(ctx context.Context, req *GroupRequest) (*GroupPermissionsResponse, error) {
	groupId := req.Id
	ent, err := c.Group(ctx, req)
	if err != nil {
		return nil, err
	}
	ret := &GroupPermissionsResponse{
		Group:   ent.Group,
		Users:   &GroupPermissionsResponse_Users{},
		Actions: &GroupPermissionsResponse_Actions{},
	}

	// Actions
	entKey := NewEntityID(GroupType, groupId)
	ret.Actions.CanView, _ = c.checkObject(ctx, CanView, entKey)
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, CanEditMembers, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, CanEdit, entKey)
	ret.Actions.CanCreateFeed, _ = c.checkObject(ctx, CanCreateFeed, entKey)
	ret.Actions.CanDeleteFeed, _ = c.checkObject(ctx, CanDeleteFeed, entKey)

	// Get feeds
	feedIds, _ := c.listSubjectRelations(ctx, entKey, FeedType, ParentRelation)
	ret.Feeds, _ = c.getFeeds(ctx, feedIds)

	// Get group metadata
	tps, err := c.getObjectTuples(ctx, CanView, entKey)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ct, _ := c.Tenant(ctx, &TenantRequest{Id: tk.Subject.ID()})
			ret.Tenant = ct.Tenant
		}
		if tk.Relation == ManagerRelation {
			ret.Users.Managers = append(ret.Users.Managers, newUser(tk.Subject.Name))
		}
		if tk.Relation == EditorRelation {
			ret.Users.Editors = append(ret.Users.Editors, newUser(tk.Subject.Name))
		}
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, newUser(tk.Subject.Name))
		}
	}
	ret.Users.Managers, _ = c.getUsers(ctx, ret.Users.Managers)
	ret.Users.Editors, _ = c.getUsers(ctx, ret.Users.Editors)
	ret.Users.Viewers, _ = c.getUsers(ctx, ret.Users.Viewers)
	return ret, nil
}

func (c *Checker) GroupSave(ctx context.Context, req *GroupSaveRequest) (*GroupSaveResponse, error) {
	groupId := req.Group.Id
	newName := req.Group.Name
	if check, err := c.GroupPermissions(ctx, &GroupRequest{Id: req.Group.Id}); err != nil {
		return nil, err
	} else if !check.Actions.CanEdit {
		return nil, ErrUnauthorized
	}
	log.Trace().Str("groupName", newName).Int64("id", groupId).Msg("GroupSave")
	_, err := sq.StatementBuilder.
		RunWith(c.db).
		PlaceholderFormat(sq.Dollar).
		Update("tl_groups").
		SetMap(map[string]any{
			"group_name": newName,
		}).
		Where("id = ?", groupId).Exec()
	return &GroupSaveResponse{}, err
}

func (c *Checker) GroupAddPermission(ctx context.Context, req *GroupModifyPermissionRequest) (*GroupSaveResponse, error) {
	groupId := req.Id
	if check, err := c.GroupPermissions(ctx, &GroupRequest{Id: groupId}); err != nil {
		return nil, err
	} else if !check.Actions.CanEditMembers {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(req.UserRelation.UserId).WithObjectID(GroupType, groupId).WithRelation(req.UserRelation.Relation)
	log.Trace().Str("tk", tk.String()).Int64("id", groupId).Msg("GroupAddPermission")
	return &GroupSaveResponse{}, c.authz.ReplaceTuple(ctx, tk)
}

func (c *Checker) GroupRemovePermission(ctx context.Context, req *GroupModifyPermissionRequest) (*GroupSaveResponse, error) {
	groupId := req.Id
	if check, err := c.GroupPermissions(ctx, &GroupRequest{Id: groupId}); err != nil {
		return nil, err
	} else if !check.Actions.CanEditMembers {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(req.UserRelation.UserId).WithObjectID(GroupType, groupId).WithRelation(req.UserRelation.Relation)
	log.Trace().Str("tk", tk.String()).Int64("id", groupId).Msg("GroupRemovePermission")
	return &GroupSaveResponse{}, c.authz.DeleteTuple(ctx, tk)
}

// ///////////////////
// FEEDS
// ///////////////////

func (c *Checker) getFeeds(ctx context.Context, ids []int64) ([]*Feed, error) {
	return gets[*Feed](ctx, c.db, ids, "current_feeds", "id", "onestop_id", "coalesce(name,'') as name")
}

func (c *Checker) FeedList(ctx context.Context, req *FeedListRequest) (*FeedListResponse, error) {
	feedIds, err := c.listCtxObjects(ctx, FeedType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := c.getFeeds(ctx, feedIds)
	return &FeedListResponse{Feeds: t}, err
}

func (c *Checker) Feed(ctx context.Context, req *FeedRequest) (*FeedResponse, error) {
	feedId := req.Id
	if err := c.checkObjectOrError(ctx, CanView, NewEntityID(FeedType, feedId)); err != nil {
		return nil, err
	}
	t, err := c.getFeeds(ctx, []int64{feedId})
	return &FeedResponse{Feed: first(t)}, err
}

func (c *Checker) FeedPermissions(ctx context.Context, req *FeedRequest) (*FeedPermissionsResponse, error) {
	ent, err := c.Feed(ctx, req)
	if err != nil {
		return nil, err
	}
	ret := &FeedPermissionsResponse{
		Feed:    ent.Feed,
		Actions: &FeedPermissionsResponse_Actions{},
	}

	// Actions
	feedId := req.Id
	entKey := NewEntityID(FeedType, feedId)
	ret.Actions.CanView, _ = c.checkObject(ctx, CanView, entKey)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, CanEdit, entKey)
	ret.Actions.CanSetGroup, _ = c.checkObject(ctx, CanSetGroup, entKey)
	ret.Actions.CanCreateFeedVersion, _ = c.checkObject(ctx, CanCreateFeedVersion, entKey)
	ret.Actions.CanDeleteFeedVersion, _ = c.checkObject(ctx, CanDeleteFeedVersion, entKey)

	// Get feed metadata
	tps, err := c.getObjectTuples(ctx, CanView, entKey)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == ParentRelation {
			ct, _ := c.Group(ctx, &GroupRequest{Id: tk.Subject.ID()})
			ret.Group = ct.Group
		}
	}
	return ret, nil
}

func (c *Checker) FeedSetGroup(ctx context.Context, req *FeedSetGroupRequest) (*FeedSaveResponse, error) {
	feedId := req.Id
	newGroup := req.GroupId
	if check, err := c.FeedPermissions(ctx, &FeedRequest{Id: feedId}); err != nil {
		return nil, err
	} else if !check.Actions.CanSetGroup {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithSubjectID(GroupType, newGroup).WithObjectID(FeedType, feedId).WithRelation(ParentRelation)
	log.Trace().Str("tk", tk.String()).Int64("id", feedId).Msg("FeedSetGroup")
	return &FeedSaveResponse{}, c.authz.ReplaceTuple(ctx, tk)
}

/////////////////////
// FEED VERSIONS
/////////////////////

func (c *Checker) getFeedVersions(ctx context.Context, ids []int64) ([]*FeedVersion, error) {
	return gets[*FeedVersion](ctx, c.db, ids, "feed_versions", "id", "feed_id", "sha1", "coalesce(name,'') as name")
}

func (c *Checker) FeedVersionList(ctx context.Context, req *FeedVersionListRequest) (*FeedVersionListResponse, error) {
	fvids, err := c.listCtxObjects(ctx, FeedVersionType, CanView)
	if err != nil {
		return nil, err
	}
	t, err := c.getFeedVersions(ctx, fvids)
	return &FeedVersionListResponse{FeedVersions: t}, err
}

func (c *Checker) FeedVersion(ctx context.Context, req *FeedVersionRequest) (*FeedVersionResponse, error) {
	fvid := req.Id
	feedId := int64(0)
	// We need to get feed id before any other checks
	// If there is a "not found" error here, save it for after the global admin check
	// This is for consistency with other permission checks
	t, fvErr := c.getFeedVersions(ctx, []int64{fvid})
	fv := first(t)
	if fv != nil {
		feedId = fv.FeedId
	}
	ctxTk := NewTupleKey().WithObjectID(FeedVersionType, fvid).WithSubjectID(FeedType, feedId).WithRelation(ParentRelation)
	if err := c.checkObjectOrError(ctx, CanView, NewEntityID(FeedVersionType, fvid), ctxTk); err != nil {
		return nil, err
	}
	// Now return deferred fvErr
	if fvErr != nil {
		return nil, fvErr
	}
	return &FeedVersionResponse{FeedVersion: fv}, nil
}

func (c *Checker) FeedVersionPermissions(ctx context.Context, req *FeedVersionRequest) (*FeedVersionPermissionsResponse, error) {
	ent, err := c.FeedVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	ret := &FeedVersionPermissionsResponse{
		FeedVersion: ent.FeedVersion,
		Users:       &FeedVersionPermissionsResponse_Users{},
		Actions:     &FeedVersionPermissionsResponse_Actions{},
	}

	// Actions
	fvid := req.Id
	ctxTk := NewTupleKey().WithObjectID(FeedVersionType, ent.FeedVersion.Id).WithSubjectID(FeedType, ent.FeedVersion.Id).WithRelation(ParentRelation)
	entKey := NewEntityID(FeedVersionType, fvid)
	ret.Actions.CanView, _ = c.checkObject(ctx, CanView, entKey, ctxTk)
	ret.Actions.CanEditMembers, _ = c.checkObject(ctx, CanEditMembers, entKey, ctxTk)
	ret.Actions.CanEdit, _ = c.checkObject(ctx, CanEdit, entKey, ctxTk)

	// Get fv metadata
	tps, err := c.getObjectTuples(ctx, CanView, entKey, ctxTk)
	if err != nil {
		return nil, err
	}
	for _, tk := range tps {
		if tk.Relation == EditorRelation {
			ret.Users.Editors = append(ret.Users.Editors, newUser(tk.Subject.Name))
		}
		if tk.Relation == ViewerRelation {
			ret.Users.Viewers = append(ret.Users.Viewers, newUser(tk.Subject.Name))
		}
	}
	ret.Users.Editors, _ = c.getUsers(ctx, ret.Users.Editors)
	ret.Users.Viewers, _ = c.getUsers(ctx, ret.Users.Viewers)
	return ret, nil
}

func (c *Checker) FeedVersionAddPermission(ctx context.Context, req *FeedVersionModifyPermissionRequest) (*FeedVersionSaveResponse, error) {
	fvid := req.Id
	if check, err := c.FeedVersionPermissions(ctx, &FeedVersionRequest{Id: fvid}); err != nil {
		return nil, err
	} else if !check.Actions.CanEditMembers {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(req.UserRelation.UserId).WithObjectID(FeedVersionType, fvid).WithRelation(req.UserRelation.Relation)
	log.Trace().Str("tk", tk.String()).Int64("id", fvid).Msg("FeedVersionAddPermission")
	return &FeedVersionSaveResponse{}, c.authz.ReplaceTuple(ctx, tk)
}

func (c *Checker) FeedVersionRemovePermission(ctx context.Context, req *FeedVersionModifyPermissionRequest) (*FeedVersionSaveResponse, error) {
	fvid := req.Id
	if check, err := c.FeedVersionPermissions(ctx, &FeedVersionRequest{Id: fvid}); err != nil {
		return nil, err
	} else if !check.Actions.CanEditMembers {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(req.UserRelation.UserId).WithObjectID(FeedVersionType, fvid).WithRelation(req.UserRelation.Relation)
	log.Trace().Str("tk", tk.String()).Int64("id", fvid).Msg("FeedVersionRemovePermission")
	return &FeedVersionSaveResponse{}, c.authz.DeleteTuple(ctx, tk)
}

// ///////////////////
// internal
// ///////////////////

func (c *Checker) listCtxObjects(ctx context.Context, objectType ObjectType, action Action) ([]int64, error) {
	checkUser := auth.ForContext(ctx)
	if checkUser == nil {
		return nil, ErrUnauthorized
	}
	tk := NewTupleKey().WithUser(checkUser.Name()).WithObject(objectType, "").WithAction(action)
	objTks, err := c.authz.ListObjects(ctx, tk)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	var ret []int64
	for _, tk := range objTks {
		ret = append(ret, tk.Object.ID())
	}
	return ret, nil
}

func (c *Checker) listSubjectRelations(ctx context.Context, sub EntityKey, objectType ObjectType, relation Relation) ([]int64, error) {
	tk := NewTupleKey().WithSubject(sub.Type, sub.Name).WithObject(objectType, "").WithRelation(relation)
	rels, err := c.authz.ListObjects(ctx, tk)
	if err != nil {
		return nil, err
	}
	var ret []int64
	for _, v := range rels {
		ret = append(ret, v.Object.ID())
	}
	return ret, nil
}

func (c *Checker) getObjectTuples(ctx context.Context, checkAction Action, obj EntityKey, ctxtk ...TupleKey) ([]TupleKey, error) {
	if err := c.checkObjectOrError(ctx, checkAction, obj, ctxtk...); err != nil {
		return nil, err
	}
	return c.authz.GetObjectTuples(ctx, NewTupleKey().WithObject(obj.Type, obj.Name))
}

func (c *Checker) checkObjectOrError(ctx context.Context, checkAction Action, obj EntityKey, ctxtk ...TupleKey) error {
	ok, err := c.checkObject(ctx, checkAction, obj, ctxtk...)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnauthorized
	}
	return nil
}

func (c *Checker) checkObject(ctx context.Context, checkAction Action, obj EntityKey, ctxtk ...TupleKey) (bool, error) {
	checkUser := auth.ForContext(ctx)
	if checkUser == nil {
		return false, ErrUnauthorized
	}
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

type hasId interface {
	GetId() int64
}

func checkIds[T hasId](ents []T, ids []int64) error {
	if len(ents) != len(ids) {
		return errors.New("not found")
	}
	check := map[int64]bool{}
	for _, ent := range ents {
		check[ent.GetId()] = true
	}
	for _, id := range ids {
		if _, ok := check[id]; !ok {
			return errors.New("not found")
		}
	}
	return nil
}

func gets[T hasId](ctx context.Context, db sqlx.Ext, ids []int64, table string, cols ...string) ([]T, error) {
	var t []T
	q := sq.StatementBuilder.Select(cols...).From(table).Where(sq.Eq{"id": ids})
	if err := find.Select(ctx, db, q, &t); err != nil {
		log.Trace().Err(err)
		return nil, err
	}
	if err := checkIds(t, ids); err != nil {
		return nil, err
	}
	return t, nil
}

func first[T any](v []T) T {
	var xt T
	if len(v) > 0 {
		return v[0]
	}
	return xt
}

func newUser(id string) *User {
	return &User{Id: id}
}

func newUserRel(userId string, rel Relation) *UserRelation {
	return &UserRelation{UserId: userId, Relation: rel}
}
