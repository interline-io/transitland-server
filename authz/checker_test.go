package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.TenantList(context.Background(), newTestUser(ltk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.Tenants {
					gotNames = append(gotNames, v.Name.Val)
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
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
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
		checker := newTestChecker(t, checkerTestData)
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.TenantAddPermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Object.ID(),
					ltk.Subject.Name,
					ltk.Relation,
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("TenantRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				CheckAsUser: "tl-tenant-admin",
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.TenantRemovePermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Object.ID(),
					ltk.Subject.Name,
					ltk.Relation,
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("TenantSave", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.TenantSave(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Object.ID(),
					tc.Object.Name,
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
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:            NewEntityKey(TenantType, "tl-tenant"),
				Object:             NewEntityKey(GroupType, "new-group"),
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
			{
				Subject:     NewEntityKey(TenantType, "tl-tenant"),
				Object:      NewEntityKey(GroupType, "new-group2"),
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(TenantType, "tl-tenant"),
				Object:      NewEntityKey(GroupType, "new-group3"),
				CheckAsUser: "global_admin",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				_, err := checker.TenantCreateGroup(
					context.Background(),
					newTestUser(tc.CheckAsUser),
					ltk.Subject.ID(),
					tc.Object.Name,
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.GroupList(context.Background(), newTestUser(ltk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.Groups {
					gotNames = append(gotNames, v.Name.Val)
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
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
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
		checker := newTestChecker(t, checkerTestData)
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.GroupAddPermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, tc.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("GroupRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				Notes:       "already deleted",
				ExpectError: true,
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.GroupRemovePermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedList(context.Background(), newTestUser(ltk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.Feeds {
					gotNames = append(gotNames, v.OnestopID.Val)
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
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
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
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedVersionList(context.Background(), newTestUser(ltk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret.FeedVersions {
					gotNames = append(gotNames, v.SHA1.Val)
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
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedVersionPermissions(
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
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
		checker := newTestChecker(t, checkerTestData)
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
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.FeedVersionAddPermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
				)
				checkErrUnauthorized(t, err, tc.ExpectError, tc.ExpectUnauthorized)
			})
		}
	})

	t.Run("FeedVersionRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t, checkerTestData)
		checks := []testTuple{
			{
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:     NewEntityKey(UserType, "tl-tenant-member"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				Notes:       "already deleted",
				ExpectError: true,
				CheckAsUser: "tl-tenant-admin",
			},
			{
				Subject:            NewEntityKey(UserType, "ian"),
				Object:             NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:           ViewerRelation,
				ExpectUnauthorized: true,
				CheckAsUser:        "ian",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.FeedVersionRemovePermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
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

	auth0c := NewMockAuthnClient()
	auth0c.AddUser("ian", User{Name: "Ian", ID: "ian", Email: "ian@example.com"})
	auth0c.AddUser("drew", User{Name: "Drew", ID: "drew", Email: "drew@example.com"})
	auth0c.AddUser("tl-tenant-member", User{Name: "Nisar", ID: "tl-tenant-member", Email: "tl-tenant-member@example.com"})

	fgac, err := newTestCheckerFGA(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range testData {
		if err := fgac.WriteTuple(context.Background(), dbTupleLookup(t, dbx, tc.TupleKey())); err != nil {
			t.Fatal(err)
		}
	}
	checker := NewChecker(auth0c, fgac, dbx, nil)
	checker.globalAdmins = append(checker.globalAdmins, "global_admin")
	return checker
}

func newTestCheckerFGA(t testing.TB) (*FGAClient, error) {
	cfg := AuthzConfig{
		FGAEndpoint:      os.Getenv("TL_TEST_FGA_ENDPOINT"),
		FGALoadModelFile: "../test/authz/tls.json",
	}
	fgac, err := NewFGAClient(cfg.FGAEndpoint, "", "")
	if err != nil {
		return nil, err
	}
	if _, err := fgac.CreateStore(context.Background(), "test"); err != nil {
		return nil, err
	}
	if _, err := fgac.CreateModel(context.Background(), cfg.FGALoadModelFile); err != nil {
		return nil, err
	}
	return fgac, nil
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
	if expectUnauthorized {
		if err == ErrUnauthorized {
			return true
		}
		t.Errorf("got error '%s', expected unauthorized", err.Error())
		return false
	}
	if expectError {
		if err != ErrUnauthorized {
			return true
		}
		if err == ErrUnauthorized {
			t.Errorf("got unauthorized error, expected other error type")
		} else {
			t.Errorf("got no error, expected error")
		}
		return false
	}
	if err != nil {
		t.Errorf("got error '%s', expected no error", err.Error())
		return false
	}
	return true
}
