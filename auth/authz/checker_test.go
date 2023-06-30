package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/generated/azpb"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if a, ok := dbutil.CheckTestDB(); !ok {
		log.Print(a)
		return
	}
	os.Exit(m.Run())
}

var checkerGetTests = []TestTuple{
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
	fgaUrl, a, ok := dbutil.CheckEnv("TL_TEST_FGA_ENDPOINT")
	if !ok {
		t.Skip(a)
		return
	}
	if a, ok := dbutil.CheckTestDB(); !ok {
		t.Skip(a)
		return
	}
	dbx := dbutil.MustOpenTestDB()
	checkerTestData := []TestTuple{
		// Assign users to tenants
		{
			Notes:    "all users can access all-users-tenant",
			Subject:  NewEntityKey(UserType, "*"),
			Object:   NewEntityKey(TenantType, "all-users-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "tl-tenant-admin"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: AdminRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "ian"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "drew"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "test2"),
			Object:   NewEntityKey(TenantType, "restricted-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  NewEntityKey(UserType, "tl-tenant-member"),
			Object:   NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		// Assign groups to tenants
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "CT-group"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "BA-group"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "HA-group"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant"),
			Object:   NewEntityKey(GroupType, "EX-group"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant").WithRefRel(MemberRelation),
			Object:   NewEntityKey(GroupType, "HA-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(TenantType, "restricted-tenant"),
			Object:   NewEntityKey(GroupType, "test-group"),
			Relation: ParentRelation,
		},
		// Assign users to groups
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
			Subject:  NewEntityKey(UserType, "drew"),
			Object:   NewEntityKey(GroupType, "CT-group"),
			Relation: ManagerRelation,
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
		// Assign feeds to groups
		{
			Subject:  NewEntityKey(GroupType, "CT-group"),
			Object:   NewEntityKey(FeedType, "CT"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(GroupType, "BA-group"),
			Object:   NewEntityKey(FeedType, "BA"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(GroupType, "HA-group"),
			Object:   NewEntityKey(FeedType, "HA"),
			Relation: ParentRelation,
		},
		{
			Subject:  NewEntityKey(GroupType, "EX-group"),
			Object:   NewEntityKey(FeedType, "EX"),
			Relation: ParentRelation,
		},
		// Assign feed versions
		{
			Subject:  NewEntityKey(UserType, "tl-tenant-member"),
			Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(GroupType, "test-group").WithRefRel(ViewerRelation),
			Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ViewerRelation,
		},
		{
			Subject:  NewEntityKey(TenantType, "tl-tenant").WithRefRel(MemberRelation),
			Object:   NewEntityKey(FeedVersionType, "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			Relation: ViewerRelation,
		},
	}

	// Users

	t.Run("UserList", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		tcs := []struct {
			Notes       string
			CheckAsUser string
			ExpectUsers []string
			ExpectError bool
			Query       string
		}{
			{
				Notes:       "user ian can see all users",
				CheckAsUser: "ian",
				ExpectUsers: []string{"ian", "drew", "tl-tenant-member", "new-user"},
			},
			{
				Notes:       "user ian can filter with query=drew",
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
			t.Run(tc.Notes, func(t *testing.T) {
				ents, err := checker.UserList(testUserCtx(tc.CheckAsUser), &azpb.UserListRequest{Q: tc.Query})
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		tcs := []struct {
			Notes        string
			CheckAsUser  string
			ExpectUserId string
			ExpectError  bool
		}{
			{
				Notes:        "ok",
				CheckAsUser:  "ian",
				ExpectUserId: "drew",
			},
			{
				Notes:        "not found",
				CheckAsUser:  "ian",
				ExpectUserId: "not found",
				ExpectError:  true,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.Notes, func(t *testing.T) {
				ent, err := checker.User(
					testUserCtx(tc.CheckAsUser),
					&azpb.UserRequest{Id: tc.ExpectUserId},
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		checks := []TestTuple{
			{
				Notes:      "user tl-tenant-admin is admin of tl-tenant and user:* on all-users-tenant",
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant", "all-users-tenant"),
			},
			{
				Notes:      "user ian is member of tl-tenant and user:* on all-users-tenant",
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant", "all-users-tenant"),
			},
			{
				Notes:      "user tl-tenant-member is member of tl-tenant and user:* on all-users-tenant",
				Subject:    NewEntityKey(UserType, "tl-tenant-member"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant", "all-users-tenant"),
			},
			{
				Notes:      "user new-user is user:* on all-users-tenant",
				Subject:    NewEntityKey(UserType, "new-user"),
				Object:     NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "all-users-tenant"),
			},
		}

		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ret, err := checker.TenantList(
					testUserCtx(tc.CheckAsUser, tc.Subject.Name),
					&azpb.TenantListRequest{},
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		checks := []TestTuple{
			// User checks
			{
				Notes:         "user tl-tenant-admin is an admin of tl-tenant",
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
			},
			{
				Notes:              "user tl-tenant-admin is unauthorized for restricted-tenant",
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user ian is viewer of tl-tenant",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Notes:              "user ian is unauthorized for restricted-tenant",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user drew is a viewer of tl-tenant",
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Notes:              "user drew is unauthorized for restricted-tenant",
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user tl-tenant-member is a viewer of tl-tenant",
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Notes:              "user tl-tenant-member is unauthorized for restricted-tenant",
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user tl-tenant-member is a viewer of tl-tenant",
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Notes:              "user tl-tenant-member is unauthorized for restricted-tenant",
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user tl-tenant-member expects unauthorized error for non-existing tenant",
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(TenantType, "not found"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user ian is viewer of all-users-tenant through user:*",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "all-users-tenant"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Notes:         "user new-user is viewer of all-users-tenant through user:*",
				Subject:       NewEntityKey(UserType, "new-user"),
				Object:        NewEntityKey(TenantType, "all-users-tenant"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			// General checks
			{
				Notes:         "global admins are admins of all tenants",
				Subject:       NewEntityKey(UserType, "global_admin"),
				Object:        NewEntityKey(TenantType, "all-users-tenant"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
			},
			{
				Notes:       "global admins get not found on not found tenant",
				Subject:     NewEntityKey(UserType, "global_admin"),
				Object:      NewEntityKey(TenantType, "not found"),
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.TenantPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.TenantRequest{Id: ltk.Object.ID()},
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
		checks := []TestTuple{
			// User checks
			{
				Notes:       "user tl-tenant-admin is an admin of tl-tenant and can add a user",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "user tl-tenant-admin is an admin of tl-tenant and can add user:*",
				Subject:     NewEntityKey(UserType, "*"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:              "user ian is a vewier of tl-tenant is not authorized to add a user",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				Relation:           MemberRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			// General checks
			{
				Notes:       "error for invalid relation",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    ParentRelation,
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "error for disallowed relation",
				Subject:     NewEntityKey(UserType, "*"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    AdminRelation,
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "replaces relation if it already exists",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:              "users get unauthorized when attempting to add user to not found tenant",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(TenantType, "not found"),
				Relation:           MemberRelation,
				CheckAsUser:        "tl-tenant-admin",
				ExpectUnauthorized: true,
			},
			{
				Notes:       "global admins can add users to all tenants",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(TenantType, "restricted-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admins get not found when adding user to a not found tenant",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(TenantType, "not found"),
				Relation:    MemberRelation,
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
			// TODO
			// {
			// 	Notes:       "user tl-tenant-admin gets an error when attempting to add a user that does not exist",
			// 	Subject:     NewEntityKey(UserType, "not found"),
			// 	Object:      NewEntityKey(TenantType, "tl-tenant"),
			// 	Relation:    MemberRelation,
			// 	CheckAsUser: "tl-tenant-admin",
			// 	ExpectError: true,
			// },
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantAddPermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.TenantModifyPermissionRequest{
						Id:             ltk.Object.ID(),
						EntityRelation: azpb.NewEntityRelation(ltk.Subject, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("TenantRemovePermission", func(t *testing.T) {
		checks := []TestTuple{
			// User checks
			{
				Notes:       "tl-tenant-admin is a admin of tl-tenant and can remove a user",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:              "user ian is a viewer of tl-tenant and is not authorized to remove a user",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				Relation:           MemberRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user tl-tenant-admin is not a member of restricted-tenant and is not authorized to remove a user",
				Subject:            NewEntityKey(UserType, "test2"),
				Object:             NewEntityKey(TenantType, "restricted-tenant"),
				Relation:           MemberRelation,
				CheckAsUser:        "tl-tenant-admin",
				ExpectUnauthorized: true,
			},
			// General checks
			{
				Notes:       "error if relation does not exist",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
				ExpectError: true,
			}, {
				Notes:              "users get unauthorized when attemping to add user to not found tenant",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(TenantType, "not found"),
				Relation:           MemberRelation,
				CheckAsUser:        "tl-tenant-admin",
				ExpectUnauthorized: true,
			},
			{
				Notes:       "global admins can remove users from all tenants",
				Subject:     NewEntityKey(UserType, "test2"),
				Object:      NewEntityKey(TenantType, "restricted-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admins get error when removing user from not found tenant",
				Subject:     NewEntityKey(UserType, "test2"),
				Object:      NewEntityKey(TenantType, "not found"),
				Relation:    MemberRelation,
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
			// TODO
			// {
			// 	Notes:       "removing a non-existing user causes an error",
			// 	Subject:     NewEntityKey(UserType, "asd123"),
			// 	Object:      NewEntityKey(TenantType, "tl-tenant"),
			// 	Relation:    MemberRelation,
			// 	CheckAsUser: "tl-tenant-admin",
			// 	ExpectError: true,
			// },
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantRemovePermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.TenantModifyPermissionRequest{
						Id:             ltk.Object.ID(),
						EntityRelation: azpb.NewEntityRelation(ltk.Subject, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("TenantSave", func(t *testing.T) {
		checks := []TestTuple{
			// User checks
			{
				Notes:              "user ian is a viewer of tl-tenant and is not authorized to edit",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user new-user is not a viewer of tl-tenant",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(TenantType, "tl-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user new-user is a viewer of all-users-tenant through user:* but not admin",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(TenantType, "all-users-tenant"),
				ExpectUnauthorized: true,
			},
			{
				Notes:   "user tl-tenant-admin is admin of tl-tenant and can edit",
				Subject: NewEntityKey(UserType, "tl-tenant-admin"),
				Object:  NewEntityKey(TenantType, "tl-tenant"),
			},
			// General checks
			{
				Notes:              "users get unauthorized for tenant that does not exist",
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(TenantType, "not found"),
				ExpectUnauthorized: true,
			},
			{
				Notes:       "global admins get error for not found tenant",
				Subject:     NewEntityKey(UserType, "tl-tenant-admin"),
				Object:      NewEntityKey(TenantType, "new tenant"),
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantSave(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.TenantSaveRequest{
						Tenant: &azpb.Tenant{
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
		checks := []TestTuple{
			// User checks
			{
				Notes:              "user ian is viewer of tl-tenant and not authorized to create groups",
				Subject:            NewEntityKey(TenantType, "tl-tenant"),
				Object:             NewEntityKey(GroupType, "new-group"),
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			{
				Notes:              "user new-user is viewer of all-users-tenant and not authorized to create groups",
				Subject:            NewEntityKey(TenantType, "all-users-tenant"),
				Object:             NewEntityKey(GroupType, "new-group"),
				ExpectUnauthorized: true,
				CheckAsUser:        "new-user",
			},
			{
				Notes:       "user tl-tenant-admin is admin of tl-tenant and can create groups",
				Subject:     NewEntityKey(TenantType, "tl-tenant"),
				Object:      NewEntityKey(GroupType, fmt.Sprintf("new-group2-%d", time.Now().UnixNano())),
				CheckAsUser: "tl-tenant-admin",
			},
			// General checks
			{
				Notes:       "global admins can create groups in all tenants",
				Subject:     NewEntityKey(TenantType, "tl-tenant"),
				Object:      NewEntityKey(GroupType, fmt.Sprintf("new-group3-%d", time.Now().UnixNano())),
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admins can create groups in all tenants",
				Subject:     NewEntityKey(TenantType, "restricted-tenant"),
				Object:      NewEntityKey(GroupType, fmt.Sprintf("new-group4-%d", time.Now().UnixNano())),
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admins get not found for tenant that does not exist",
				Subject:     NewEntityKey(TenantType, "not found"),
				Object:      NewEntityKey(GroupType, fmt.Sprintf("new-group5-%d", time.Now().UnixNano())),
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantCreateGroup(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.TenantCreateGroupRequest{
						Id:    ltk.Subject.ID(),
						Group: &azpb.Group{Name: tc.Object.Name},
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		checks := []TestTuple{
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
					&azpb.GroupListRequest{},
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		checks := []TestTuple{
			// User checks
			{
				Notes:         "user tl-tenant-admin is admin of parent tenant to CT-group",
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Notes:         "user tl-tenant-admin is admin of parent tenant to BA-group",
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Notes:         "user tl-tenant-admin is admin of parent tenant to BA-group",
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Notes:         "user tl-tenant-admin is admin of parent tenant to BA-group",
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Notes:         "user ian is a viewer of CT-group",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Notes:         "user ian is a editor of CT-group",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Notes:         "user ian is a viewer of HA-group through tl-tenant#member",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Notes:              "user ian is not authorized for EX-group",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user ian is not authorized for test-group",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(GroupType, "test-group"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user drew is a manager of CT-group",
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Notes:              "user drew is not authrozied for BA-group",
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(GroupType, "BA-group"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user drew is a viewer of HA-group through tl-tenant#member",
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Notes:              "user drew is not authorized for EX-group",
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user drew is not authorized for group test-group",
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(GroupType, "test-group"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user tl-tenant-member is a viewer of HA-group through tl-tenant#member",
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Notes:              "tl-tenant-member is not authorized to access EX-group",
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				ExpectUnauthorized: true,
			},
			// General checks
			{
				Notes:              "users get unauthorized for groups that are not found",
				Subject:            NewEntityKey(UserType, "tl-tenant-admin"),
				Object:             NewEntityKey(GroupType, "test-group"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "global admins are managers of all groups",
				Subject:       NewEntityKey(UserType, "global_admin"),
				Object:        NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed, CanSetTenant},
			},
			{
				Notes:       "global admins get not found for not found groups",
				Subject:     NewEntityKey(UserType, "global_admin"),
				Object:      NewEntityKey(GroupType, "not found"),
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.GroupPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.GroupRequest{Id: ltk.Object.ID()},
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
		checks := []TestTuple{
			// User checks
			// TODO
			// {
			// 	Notes:       "user tl-tenant-admin gets error when adding user that does not exist",
			// 	Subject:     NewEntityKey(UserType, "test100"),
			// 	Object:      NewEntityKey(GroupType, "HA-group"),
			// 	Relation:    ViewerRelation,
			// 	ExpectError: true,
			// 	CheckAsUser: "tl-tenant-admin",
			// },
			{
				Notes:       "tl-tenant-admin is manager of CT-group through tl-tenant and can add user",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "tl-tenant-admin is manager of CT-group through tl-tenant and can add tenant#member as viewer",
				Subject:     NewEntityKey(TenantType, "tl-tenant").WithRefRel(MemberRelation),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "tl-tenant-admin is manager of CT-group through tl-tenant and can add tenant#member as editor",
				Subject:     NewEntityKey(TenantType, "tl-tenant").WithRefRel(MemberRelation),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    EditorRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:              "user ian is viewer of CT-group and not authorized to add users",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(GroupType, "CT-group"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			{
				Notes:       "user drew is a manager of CT-group and can add users",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "drew",
			},
			// General checks
			{
				Notes:              "users get unauthorized for groups that do not exist",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(GroupType, "not found"),
				Relation:           ViewerRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			{
				Notes:       "error for invalid relation",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ParentRelation,
				CheckAsUser: "tl-tenant-admin",
				ExpectError: true,
			},
			{
				Notes:       "error for invalid relation",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "HA-group"),
				Relation:    ParentRelation,
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "error for disallowed relation",
				Subject:     NewEntityKey(GroupType, "BA-group").WithRefRel(MemberRelation),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
				ExpectError: true,
			},
			{
				Notes:       "error for disallowed relation",
				Subject:     NewEntityKey(TenantType, "tl-tenant#admin"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    EditorRelation,
				CheckAsUser: "tl-tenant-admin",
				ExpectError: true,
			},
			{
				Notes:       "global admin can add users to any group",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admin gets not found for groups that do not exist",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(GroupType, "not found"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.GroupAddPermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.GroupModifyPermissionRequest{
						Id:             ltk.Object.ID(),
						EntityRelation: azpb.NewEntityRelation(ltk.Subject, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("GroupRemovePermission", func(t *testing.T) {
		checks := []TestTuple{
			// User checks
			{
				Notes:       "user tl-tenant-admin is manager of CT-group through tl-tenant",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:              "user ian is viewer of BA-group and is not authorized to add users",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(GroupType, "BA-group"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			// General checks
			{
				Notes:              "users get authorized for groups that do not exist",
				Subject:            NewEntityKey(UserType, "new-user"),
				Object:             NewEntityKey(GroupType, "not found"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			{
				Notes:       "users get error for removing tuple that does not exist",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(GroupType, "BA-group"),
				Relation:    ViewerRelation,
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "global admins can remove users from any group",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admins get error for removing tuples that do not exist",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    EditorRelation,
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
			{
				Notes:       "global admins get not found for groups that do not exist",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(GroupType, "not found"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.GroupRemovePermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.GroupModifyPermissionRequest{
						Id:             ltk.Object.ID(),
						EntityRelation: azpb.NewEntityRelation(ltk.Subject, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("GroupSave", func(t *testing.T) {
		checks := []TestTuple{
			// User checks
			{
				Notes:              "user ian is a viewer of CT-group and cannot edit",
				CheckAsUser:        "ian",
				Object:             NewEntityKey(GroupType, "CT-group"),
				ExpectUnauthorized: true,
			},
			{
				Notes:       "user tl-tenant-admin is a manager of BA-group through tl-tenant and can edit",
				Object:      NewEntityKey(GroupType, "BA-group"),
				CheckAsUser: "tl-tenant-admin",
			},
			// General checks
			{
				Notes:              "users get unauthorized for groups that are not found",
				Object:             NewEntityKey(GroupType, "not found"),
				CheckAsUser:        "tl-tenant-admin",
				ExpectUnauthorized: true,
			},
			{
				Notes:       "global admins can edit any group",
				Object:      NewEntityKey(GroupType, "BA-group"),
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admins get not found for groups that are not found",
				Object:      NewEntityKey(GroupType, "not found"),
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.GroupSave(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.GroupSaveRequest{
						Group: &azpb.Group{
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		checks := []TestTuple{
			{
				Notes:      "user tl-tenant-admin can see all feeds with groups that are in tl-tenant",
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
			},
			{
				Notes:      "user ian is viewer in CT-group, BA-group, and also HA-group through tl-tenant#member",
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "BA", "HA"),
			},
			{
				Notes:      "user drew is editor of CT-group and can see feed CT and also feed HA through tl-tenant#member on HA-group",
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "HA"),
			},
			{
				Notes:      "user tl-tenant-member can see feed HA through tl-tenant#member on HA-group",
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
					&azpb.FeedListRequest{},
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		checks := []TestTuple{
			// User checks
			{
				Notes:         "user ian is a viewer of CT through CT-group",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Notes:         "user ian is a editor of BA through BA-group",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion},
			},
			{
				Notes:         "user ian is a viewer of HA through HA-group through tl-tenant#member",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Notes:              "user ian is unauthorized for feed EX",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedType, "EX"),
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user ian is unauthorized for feed test",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedType, "test"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user drew is manager of feed CT through CT-group",
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion, CanSetGroup},
			},
			{
				Notes:              "user drew is unauthorized for feed BA",
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(FeedType, "BA"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user drew is viewer of feed HA through HA-group through tl-tenant#member",
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Notes:              "user tl-tenant-member is unauthorized for feed BA",
				Subject:            NewEntityKey(UserType, "tl-tenant-member"),
				Object:             NewEntityKey(FeedType, "BA"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "user tl-tenant-member is viewer for feed HA through HA-group through tl-tenant#member",
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			// General checks
			{
				Notes:              "users get unauthorized for a feed that is not found",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedType, "not found"),
				ExpectUnauthorized: true,
			},
			{
				Notes:         "global admins are manager for all feeds",
				Subject:       NewEntityKey(UserType, "global_admin"),
				Object:        NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion, CanSetGroup},
			},
			{
				Notes:       "global admins get not found for feed that does not exist",
				Subject:     NewEntityKey(UserType, "global_admin"),
				Object:      NewEntityKey(FeedType, "not found"),
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.FeedRequest{Id: ltk.Object.ID()},
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
		tcs := []TestTuple{
			// User checks
			// TODO!!
			// {
			// 	Notes:       "user drew is a manager of feed CT and can assign it to a different group",
			// 	Subject:     NewEntityKey(FeedType, "CT"),
			// 	Object:      NewEntityKey(GroupType, "test-group"),
			// 	CheckAsUser: "drew",
			// },
			{
				Notes:              "user ian is an editor of group BA and is not authorized to assign to a different group",
				Subject:            NewEntityKey(FeedType, "BA"),
				Object:             NewEntityKey(GroupType, "test-group"),
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			{
				Notes:              "user drew is not authorized to assign feed EX to a group",
				Subject:            NewEntityKey(FeedType, "EX"),
				Object:             NewEntityKey(GroupType, "EX-group"),
				CheckAsUser:        "drew",
				ExpectUnauthorized: true,
			},
			// General checks
			{
				Notes:       "user global_admin is a global admin and can assign a feed to a group",
				Subject:     NewEntityKey(FeedType, "BA"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				CheckAsUser: "global_admin",
			},
		}
		for _, tc := range tcs {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.FeedSetGroup(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.FeedSetGroupRequest{Id: ltk.Subject.ID(), GroupId: ltk.Object.ID()},
				)
				if checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized) {
					return
				}
				// Verify write
				fr, err := checker.FeedPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.FeedRequest{Id: ltk.Subject.ID()},
				)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("tc.Object:", tc.Object, "fr.Group:", fr.Group)
				assert.Equal(t, tc.Object.Name, fr.Group.Name)
			})
		}
	})

	// FEED VERSIONS
	t.Run("FeedVersionList", func(t *testing.T) {
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		// Only user:tl-tenant-member has permissions explicitly defined
		checks := []TestTuple{
			// User checks
			{
				Notes:      "tl-tenant-admin has no explicit feed versions but can access d281 through tenant#member",
				Subject:    NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType, "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			},
			{
				Notes:      "ian has no explicit feed versions but can access d281 through tenant#member",
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType, "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			},
			{
				Notes:      "drew has no explicit feed versions but can access d281 through tenant#member",
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType, "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			},
			{
				Notes:      "tl-tenant-admin has explicit access to e535 and can access d281 through tenant#member",
				Subject:    NewEntityKey(UserType, "tl-tenant-member"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ExpectKeys: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0", "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			},
			// General checks
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ret, err := checker.FeedVersionList(
					testUserCtx(tc.CheckAsUser, tc.Subject.Name),
					&azpb.FeedVersionListRequest{},
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
		checker := newTestChecker(t, fgaUrl, checkerTestData)
		checks := []TestTuple{
			// User checks
			{
				Notes:         "tl-tenant-admin is a editor of e535 through tenant",
				Subject:       NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers},
			},
			{
				Notes:         "user ian is an editor of e535 through feed BA through group BA-group",
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, CanEdit},
			},
			{
				Notes:              "drew is not authorized to read e535",
				Subject:            NewEntityKey(UserType, "drew"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions:      []Action{-CanView},
				ExpectUnauthorized: true,
			},
			{
				Notes:         "tl-tenant-member is directly granted viewer on e535",
				Subject:       NewEntityKey(UserType, "tl-tenant-member"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Notes:         "user test-group-viewer is viewer on e535 through grant to test-group#viewer",
				Subject:       NewEntityKey(UserType, "test-group-viewer"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers},
			},
			// General checks
			{
				Notes:              "users get unauthorized for feed versions that do not exist",
				Subject:            NewEntityKey(UserType, "test-group-viewer"),
				Object:             NewEntityKey(FeedVersionType, "not found"),
				ExpectUnauthorized: true,
			},
			{
				Notes:       "global admins get error for feed versions that do not exist",
				Subject:     NewEntityKey(UserType, "global_admin"),
				Object:      NewEntityKey(FeedVersionType, "not found"),
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedVersionPermissions(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.FeedVersionRequest{Id: ltk.Object.ID()},
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
		checks := []TestTuple{
			// User checks
			{
				Notes:       "user tl-tenant-admin is a manager of e535 through tenant#admin and can edit users",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:              "user ian is editor of e535 through BA and BA-group and can not edit users",
				Subject:            NewEntityKey(UserType, "test3"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:           ViewerRelation,
				CheckAsUser:        "ian",
				ExpectUnauthorized: true,
			},
			// General checks
			{
				Notes:       "existing tuple will still remove other subject matched tuples",
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "invalid relation returns error",
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ParentRelation,
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "disallowed relation returns error",
				Subject:     NewEntityKey(GroupType, "BA-group#editor"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:       "global admins get error when editing feed version that does not exist",
				Subject:     NewEntityKey(UserType, "test3"),
				Object:      NewEntityKey(FeedVersionType, "not found"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
				ExpectError: true,
			},
			{
				Notes:       "global admins can edit users of any feed version",
				Subject:     NewEntityKey(UserType, "new-user"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.FeedVersionAddPermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.FeedVersionModifyPermissionRequest{
						Id:             ltk.Object.ID(),
						EntityRelation: azpb.NewEntityRelation(ltk.Subject, ltk.Relation),
					},
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("FeedVersionRemovePermission", func(t *testing.T) {
		checks := []TestTuple{
			// User checks
			{
				Notes:       "user tl-tenant-admin is a manager of feed version e535 through tenant#admin and can edit permissions",
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Notes:              "user ian is not a manager of feed version e535",
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			// General checks
			{
				Notes:       "global admins get not found when editing permissions of a feed version that does not exist",
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, "not found"),
				Relation:    ViewerRelation,
				ExpectError: true,
				CheckAsUser: "global_admin",
			},
			{
				Notes:       "global admins can edit permissions of any feed version",
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "global_admin",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test - initialize for each test
				checker := newTestChecker(t, fgaUrl, checkerTestData)
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.FeedVersionRemovePermission(
					testUserCtx(tc.CheckAsUser, ltk.Subject.Name),
					&azpb.FeedVersionModifyPermissionRequest{
						Id:             ltk.Object.ID(),
						EntityRelation: azpb.NewEntityRelation(ltk.Subject, ltk.Relation),
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

func newTestChecker(t testing.TB, url string, testData []TestTuple) *Checker {
	dbx := dbutil.MustOpenTestDB()
	cfg := AuthzConfig{
		FGAEndpoint:      url,
		FGALoadModelFile: testutil.RelPath("test/authz/tls.json"),
		GlobalAdmin:      "global_admin",
	}

	checker, err := NewCheckerFromConfig(cfg, dbx, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add test data
	for _, tc := range testData {
		if err := checker.fgaClient.WriteTuple(context.Background(), dbTupleLookup(t, dbx, tc.TupleKey())); err != nil {
			t.Fatal(err)
		}
	}

	// Override UserProvider
	userClient := NewMockUserProvider()
	userClient.AddUser("ian", &azpb.User{Name: "Ian", Id: "ian", Email: "ian@example.com"})
	userClient.AddUser("drew", &azpb.User{Name: "Drew", Id: "drew", Email: "drew@example.com"})
	userClient.AddUser("tl-tenant-member", &azpb.User{Name: "Tenant Member", Id: "tl-tenant-member", Email: "tl-tenant-member@example.com"})
	userClient.AddUser("new-user", &azpb.User{Name: "Unassigned Member", Id: "new-user", Email: "new-user@example.com"})
	checker.userClient = userClient
	return checker
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
		if !(expectUnauthorized || expectError) {
			t.Errorf("got error '%s', expected no error", err.Error())
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
