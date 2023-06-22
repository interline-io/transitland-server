package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

var checkerGetTests = []testTuple{
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(TenantType, "tl-tenant"),

		Expect: "user:tl-tenant-admin:admin user:ian:member user:drew:member user:tl-tenant-member:member",
	},
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Expect: "feed:BA:parent user:tl-tenant-member:viewer",
	},
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(FeedType, "CT"),
		Expect: "org:tl-tenant:parent",
	},
}

func TestChecker(t *testing.T) {
	if os.Getenv("TL_TEST_FGA_ENDPOINT") == "" {
		t.Skip("no TL_TEST_FGA_ENDPOINT set, skipping")
		return
	}

	te := testfinder.Finders(t, nil, nil)
	dbx := te.Finder.DBX()

	checkerTestData := []testTuple{
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "CT-group"),
			Relation: ParentRelation,
			Notes:    "org:CT-group is belongs to tenant:tl-tenant",
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "BA-group"),
			Relation: ParentRelation,
			Notes:    "org:BA-group belongs to tenant:tl-tenant",
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "HA-group"),
			Relation: ParentRelation,
			Notes:    "org:HA-group belongs to tenant:tl-tenant",
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "EX-group"),
			Relation: ParentRelation,
			Notes:    "org:EX-group will be for admins only",
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant#member"),
			Object:   NewEntityKey(GroupType, "HA-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(TenantType, "restricted-tenant"),
			Object:   NewEntityKey(GroupType, "test-group"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(GroupType, "CT-group"),
			Object:   NewEntityKey(FeedType, "CT"),
			Relation: ParentRelation,
			Notes:    "feed:CT should be viewable to members of org:CT-group (ian drew) and editable by org:CT-group editors (drew)",
		},
		{
			Subject:  NewEntityKey(GroupType, "BA-group"),
			Object:   NewEntityKey(FeedType, "BA"),
			Relation: ParentRelation,
			Notes:    "feed:BA should be viewable to members of org:BA-group () and editable by org:BA-group editors (ian)",
		},
		{
			Subject:  NewEntityKey(GroupType, "HA-group"),
			Object:   NewEntityKey(FeedType, "HA"),
			Relation: ParentRelation,
			Notes:    "feed:HA-group should be viewable to all members of tenant:tl-tenant (tl-tenant-admin tl-tenant-member ian drew) and editable by org:HA-group editors ()",
		},
		{
			Subject:  NewEntityKey(GroupType, "EX-group"),
			Object:   NewEntityKey(FeedType, "EX"),
			Relation: ParentRelation,
			Notes:    "feed:EX should only be viewable to admins of tenant:tl-tenant (tl-tenant-admin)",
		},
		{
			Subject:  NewEntityKey(UserType, "tl-tenant-admin"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: AdminRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "ian"),
			Object:   NewEntityKey(GroupType, "CT-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "ian"),
			Object:   NewEntityKey(GroupType, "BA-group"),
			Relation: EditorRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "ian"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "drew"),
			Object:   NewEntityKey(GroupType, "CT-group"),
			Relation: EditorRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "drew"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "tl-tenant-member"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "tl-tenant-member"),
			Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "test2"),
			Object:   NewEntityKey(TenantType, "restricted-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "test-group-viewer"),
			Object:   NewEntityKey(GroupType, "test-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "test-group-editor"),
			Object:   NewEntityKey(GroupType, "test-group"),
			Relation: EditorRelation,
		},
		{
			Subject:  NewEntityKey(GroupType, "test-group#viewer"),
			Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: EditorRelation,
		},
	}

	// Users

	t.Run("UserList", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		tcs := []struct {
			CheckAsUser string
			ExpectUsers []string
			ExpectError bool
			Query       string
		}{
			{
				CheckAsUser: "ian",
				ExpectUsers: []string{"ian", "drew", "tl-tenant-member"},
			},
			{
				CheckAsUser: "ian",
				Query:       "drew",
				ExpectUsers: []string{"drew"},
			},
			// TODO: user filtering
			// {
			// 	CheckAsUser: "no-one",
			// 	ExpectUsers: []string{},
			// 	ExpectError: true,
			// },
		}
		for _, tc := range tcs {
			t.Run("", func(t *testing.T) {
				ents, err := checker.UserList(testUserCtx(tc.CheckAsUser), &UserListRequest{Q: tc.Query})
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
				var entNames []string
				for _, ent := range ents.Users {
					entNames = append(entNames, ent.Id)
				}
				assert.ElementsMatch(t, tc.ExpectUsers, entNames)
			})
		}
	})

	t.Run("User", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		tcs := []struct {
			CheckAsUser  string
			ExpectUserId string
			ExpectError  bool
		}{
			{
				CheckAsUser:  "ian",
				ExpectUserId: "drew",
			},
			{
				CheckAsUser:  "ian",
				ExpectUserId: "not-found",
				ExpectError:  true,
			},
		}
		for _, tc := range tcs {
			t.Run("", func(t *testing.T) {
				ent, err := checker.User(
					testUserCtx(tc.CheckAsUser),
					&UserRequest{Id: tc.ExpectUserId},
				)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
				if ent == nil {
					t.Fatal("got no result")
				}
				assert.Equal(t, tc.ExpectUserId, ent.User.Id)
			})
		}

	})

	// TENANTS
	t.Run("TenantList", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-member"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:    NewEntityKey(UserType, "unknown"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType),
			},
		}

		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ret, err := checker.TenantList(
					testUserCtx(tc.CheckAsUser, tc.Subject.Name),
					&TenantListRequest{},
				)
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.Tenants {
					gotNames = append(gotNames, v.Name)
				}
				var expectNames []string
				for _, v := range tc.ExpectKeys {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "tenant names")
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(TenantType, "not-found"),
				ExpectUnauthorized: true,
				Notes:              "not found",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.TenantPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&TenantRequest{Id: ltk.Object.ID()},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.ExpectActions)
			})
		}
	})

	t.Run("TenantAddPermission", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "test100"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
				ExpectError: true,
				Notes:       "already exists",
			},
			{
				Subject:            NewEntityKey(UserType, "test100"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				Relation:           MemberRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "test100"),
				Object:             NewEntityKey(TenantType, "not-found"),
				Relation:           MemberRelation,
				CheckAsUser:        "tl-tenant-admin",
				ExpectUnauthorized: true,
				Notes:              "not found",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantAddPermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&TenantModifyPermissionRequest{
						Id:           ltk.Object.ID(),
						UserRelation: newUserRel(ltk.Subject.Name, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("TenantRemovePermission", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				Relation:           MemberRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},

			{
				Subject:            NewEntityKey(UserType, "test2"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				Relation:           MemberRelation,
				CheckAsUser:        "tl-tenant-admin",
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantRemovePermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&TenantModifyPermissionRequest{
						Id:           ltk.Object.ID(),
						UserRelation: newUserRel(ltk.Subject.Name, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("TenantSave", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Subject: NewEntityKey(UserType, "tl-tenant-admin"),
				Object:  NewEntityKey(TenantType, "tl-tenant"),
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(TenantType, "not found"),
				ExpectUnauthorized: true,
			},
			{
				Subject:     NewEntityKey(UserType, "tl-tenant-admin"),
				Object:      NewEntityKey(TenantType, "new tenant"),
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantSave(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&TenantSaveRequest{
						Tenant: &Tenant{
							Id:   ltk.Object.ID(),
							Name: tc.Object.Name,
						},
					},
				)
				if checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized) {
					return
				}
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	})

	t.Run("TenantCreateGroup", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:            NewEntityKey(TenantType, "tl-tenant"),
				Object:             NewEntityKey(GroupType, "new-group"),
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			{
				Subject:     NewEntityKey(TenantType, "tl-tenant"),
				Object:      NewEntityKey(GroupType, fmt.Sprintf("new-group2-%d", time.Now().UnixNano())),
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(TenantType, "tl-tenant"),
				Object:      NewEntityKey(GroupType, fmt.Sprintf("new-group3-%d", time.Now().UnixNano())),
				CheckAsUser: "global_admin",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantCreateGroup(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&TenantCreateGroupRequest{
						Id:    ltk.Subject.ID(),
						Group: &Group{Name: tc.Object.Name},
					},
				)
				if checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized) {
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				// TODO: DELETE GROUP
			})
		}
	})

	// GROUPS
	t.Run("GroupList", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group", "EX-group"),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group"),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "CT-group", "HA-group"),
			},
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-member"),
				Object:     NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "HA-group"),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ret, err := checker.GroupList(
					testUserCtx(tc.CheckAsUser, tc.Subject.Name),
					&GroupListRequest{},
				)
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.Groups {
					gotNames = append(gotNames, v.Name)
				}
				var expectNames []string
				for _, v := range tc.ExpectKeys {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "group names")
			})
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(GroupType, "test-group"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(GroupType, "test-group"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(GroupType, "BA-group"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(GroupType, "test-group"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.GroupPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&GroupRequest{Id: ltk.Object.ID()},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.ExpectActions)
			})
		}
	})

	t.Run("GroupAddPermission", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "test100"),
				Object:      NewEntityKey(GroupType, "HA-group"),
				Relation:    ViewerRelation,
				Notes:       "does not exist",
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "HA-group"),
				Notes:       "invalid relation",
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "test102"),
				Object:             NewEntityKey(GroupType, "100"),
				Relation:           ViewerRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			{
				Subject:     NewEntityKey(UserType, "test100"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "test100"),
				Object:             NewEntityKey(GroupType, "CT-group"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.GroupAddPermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&GroupModifyPermissionRequest{
						Id:           ltk.Object.ID(),
						UserRelation: newUserRel(ltk.Subject.Name, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("GroupRemovePermission", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "test102"),
				Object:             NewEntityKey(GroupType, "100"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			{
				Subject:     NewEntityKey(UserType, "test101"),
				Object:      NewEntityKey(GroupType, "BA-group"),
				Relation:    ViewerRelation,
				Notes:       "does not exist",
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "test101"),
				Object:             NewEntityKey(GroupType, "BA-group"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.GroupRemovePermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&GroupModifyPermissionRequest{
						Id:           ltk.Object.ID(),
						UserRelation: newUserRel(ltk.Subject.Name, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("GroupSave", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(GroupType, "CT-group"),
				ExpectUnauthorized: true,
			},
			{
				Subject: NewEntityKey(UserType, "tl-tenant-admin"),
				Object:  NewEntityKey(GroupType, "BA-group"),
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(GroupType, "not found"),
				ExpectUnauthorized: true,
			},
			{
				Subject:     NewEntityKey(UserType, "tl-tenant-admin"),
				Object:      NewEntityKey(GroupType, "not found"),
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.GroupSave(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&GroupSaveRequest{
						Group: &Group{
							Id:   ltk.Object.ID(),
							Name: tc.Object.Name,
						},
					},
				)
				if checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized) {
					return
				}
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	})

	// FEEDS
	t.Run("FeedList", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "BA", "HA"),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "HA"),
			},

			{
				Subject:    NewEntityKey(UserType, "tl-tenant-member"),
				Object:     NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "HA"),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ret, err := checker.FeedList(
					testUserCtx(tc.CheckAsUser, tc.Subject.Name),
					&FeedListRequest{},
				)
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.Feeds {
					gotNames = append(gotNames, v.OnestopId)
				}
				var expectNames []string
				for _, v := range tc.ExpectKeys {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "feed names")
			})
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedType, "EX"),
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedType, "test"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion},
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(FeedType, "BA"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(FeedType, "EX"),
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(FeedType, "test"),
				ExpectActions:      []Action{-CanView, -CanEdit},
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(FeedType, "CT"),
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(FeedType, "BA"),
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(FeedType, "EX"),
				ExpectUnauthorized: true,
			},
			{
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(FeedType, "test"),
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&FeedRequest{Id: ltk.Object.ID()},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.ExpectActions)
			})
		}
	})

	t.Run("FeedSetGroup", func(t *testing.T) {
		tcs := []testTuple{
			{
				Subject:     NewEntityKey(FeedType, "BA"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				CheckAsUser: "global_admin",
			},
			{
				Subject:            NewEntityKey(FeedType, "EX"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				CheckAsUser:        "drew",
				ExpectUnauthorized: true,
			},
		}
		for _, tc := range tcs {
			t.Run("", func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.FeedSetGroup(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&FeedSetGroupRequest{Id: ltk.Subject.ID(), GroupId: ltk.Object.ID()},
				)
				if checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized) {
					return
				}
				// Verify write
				fr, err := checker.FeedPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&FeedRequest{Id: ltk.Subject.ID()},
				)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tc.Object.Name, fr.Group.Name)
			})
		}
	})

	// FEED VERSIONS
	t.Run("FeedVersionList", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		// Only user:tl-tenant-member has permissions explicitly defined
		checks := []testTuple{
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType),
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType),
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType),
			},
			{
				Subject:    NewEntityKey(UserType, "tl-tenant-member"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ret, err := checker.FeedVersionList(
					testUserCtx(tc.CheckAsUser, tc.Subject.Name),
					&FeedVersionListRequest{},
				)
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.FeedVersions {
					gotNames = append(gotNames, v.Sha1)
				}
				var expectNames []string
				for _, v := range tc.ExpectKeys {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "feed version names")
			})
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, CanEdit},
			},
			{
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions:      []Action{-CanView},
				Notes:              "only feed:BA readers and tl-tenant-member",
				ExpectUnauthorized: true,
			},
			{
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "test-group-viewer"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers},
			},
			{
				Subject:            NewEntityKey(UserType, "test-group-viewer"),
				Object:             NewEntityKey(FeedVersionType, "not-found"),
				ExpectUnauthorized: true,
			},
			{
				Subject:     NewEntityKey(UserType, "global_admin"),
				Object:      NewEntityKey(FeedVersionType, "not-found"),
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedVersionPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&FeedVersionRequest{Id: ltk.Object.ID()},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.ExpectActions)
			})
		}
	})

	t.Run("FeedVersionAddPermission", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				Notes:       "already exists",
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(UserType, "test1"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(UserType, "test2"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "test3"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:           ViewerRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			{
				Subject:     NewEntityKey(UserType, "test3"),
				Object:      NewEntityKey(FeedVersionType, "not found"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.FeedVersionAddPermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&FeedVersionModifyPermissionRequest{
						Id:           ltk.Object.ID(),
						UserRelation: newUserRel(ltk.Subject.Name, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("FeedVersionRemovePermission", func(t *testing.T) {
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, "not found"),
				Relation:    ViewerRelation,
				ExpectError: true,
				CheckAsUser: "global_admin",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.FeedVersionRemovePermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&FeedVersionModifyPermissionRequest{
						Id:           ltk.Object.ID(),
						UserRelation: newUserRel(ltk.Subject.Name, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

}

func stringOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func checkActionSubset(t testing.TB, actions any, checks []Action) {
	checkA, err := actionsToMap(actions)
	if err != nil {
		t.Error(err)
	}
	checkActions := checkActionsToMap(checks)
	checkMapSubset(t, checkA, checkActions)
}

func checkMapSubset(t testing.TB, got map[string]bool, expect map[string]bool) {
	var keys = map[string]bool{}
	for k := range got {
		keys[k] = true
	}
	for k := range expect {
		keys[k] = true
	}
	for k := range keys {
		if got[k] != expect[k] {
			t.Errorf("key %s mismatch, got %t expect %t", k, got[k], expect[k])
		}
	}
}

func actionsToMap(actions any) (map[string]bool, error) {
	jj, err := json.Marshal(actions)
	if err != nil {
		return nil, err
	}
	ret := map[string]bool{}
	if err := json.Unmarshal(jj, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func checkActionsToMap(v []Action) map[string]bool {
	ret := map[string]bool{}
	for _, checkAction := range v {
		expect := true
		if checkAction < 0 {
			expect = false
			checkAction *= -1
		}
		ret[checkAction.String()] = expect
	}
	return ret
}

func newTestChecker(t testing.TB, testData []testTuple) *Checker {
	te := testfinder.Finders(t, nil, nil)
	dbx := te.Finder.DBX()
	cfg := AuthzConfig{
		FGAEndpoint:      os.Getenv("TL_TEST_FGA_ENDPOINT"),
		FGALoadModelFile: testutil.RelPath("test/authz/tls.json"),
		GlobalAdmin:      "global_admin",
	}

	checker, err := NewCheckerFromConfig(cfg, dbx, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add test data
	for _, tc := range testData {
		if err := checker.authz.WriteTuple(context.Background(), dbTupleLookup(t, dbx, tc.TupleKey())); err != nil {
			t.Fatal(err)
		}
	}

	// Override AuthnProvider
	auth0c := NewMockAuthnClient()
	auth0c.AddUser("ian", &User{Name: "Ian", Id: "ian", Email: "ian@example.com"})
	auth0c.AddUser("drew", &User{Name: "Drew", Id: "drew", Email: "drew@example.com"})
	auth0c.AddUser("tl-tenant-member", &User{Name: "Tenant Member", Id: "tl-tenant-member", Email: "tl-tenant-member@example.com"})
	checker.authn = auth0c
	return checker
}

func dbTupleLookup(t testing.TB, dbx sqlx.Ext, tk TupleKey) TupleKey {
	tk.Subject = dbNameToEntityKey(t, dbx, tk.Subject)
	tk.Object = dbNameToEntityKey(t, dbx, tk.Object)
	return tk
}

func dbNameToEntityKey(t testing.TB, dbx sqlx.Ext, ek EntityKey) EntityKey {
	if ek.Name == "" {
		return ek
	}
	nsplit := strings.Split(ek.Name, "#")
	oname := nsplit[0]
	nname := ek.Name
	var err error
	switch ek.Type {
	case TenantType:
		err = sqlx.Get(dbx, &nname, "select id from tl_tenants where tenant_name = $1", oname)
	case GroupType:
		err = sqlx.Get(dbx, &nname, "select id from tl_groups where group_name = $1", oname)
	case FeedType:
		err = sqlx.Get(dbx, &nname, "select id from current_feeds where onestop_id = $1", oname)
	case FeedVersionType:
		err = sqlx.Get(dbx, &nname, "select id from feed_versions where sha1 = $1", oname)
	case UserType:
	}
	if err == sql.ErrNoRows {
		t.Log("lookup warning:", ek.Type, "name:", ek.Name, "not found")
		err = nil
	}
	if err != nil {
		t.Fatal(err)
	}
	nsplit[0] = nname
	ek.Name = strings.Join(nsplit, "#")
	return ek
}

func newEntityKeys(t ObjectType, keys ...string) []EntityKey {
	var ret []EntityKey
	for _, k := range keys {
		ret = append(ret, NewEntityKey(t, k))
	}
	return ret
}

func checkErrUnauthorized(t testing.TB, err error, expectError bool, expectUnauthorized bool) bool {
	// return true if there was an error
	// log unexpected errors
	if err == nil {
		if expectUnauthorized {
			t.Errorf("expected unauthorized, got no error")
		} else if expectError {
			t.Errorf("expected error, got no error")
		}
	} else {
		if expectUnauthorized && err != ErrUnauthorized {
			t.Errorf("expected unauthorized, got error '%s'", err.Error())
		}
	}
	return err != nil
}

// test user

type testUser struct {
	name string
}

func newTestUser(name string) *testUser {
	return &testUser{name: name}
}

func (u testUser) Name() string {
	return u.name
}

func (u testUser) GetExternalID(string) (string, bool) {
	return "test", true
}

func (u testUser) HasRole(string) bool { return true }

func (u testUser) IsValid() bool { return true }

func (u testUser) Roles() []string { return nil }

func (u testUser) WithExternalIDs(map[string]string) authn.User {
	return u
}

func (u testUser) WithRoles(...string) authn.User {
	return u
}

func testUserCtx(first ...string) context.Context {
	for _, u := range first {
		if u != "" {
			return authn.WithUser(context.Background(), newTestUser(u))
		}
	}
	return context.Background()
}
