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

var checkerTestData = []fgaTestTuple{
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
		Notes:    "org:no-one",
	},
	{
		Subject:  NewEntityKey(GroupType, "CT-group"),
		Object:   NewEntityKey(FeedType, "CT"),
		Relation: ParentRelation,
		Notes:    "feed:1 should be viewable to members of org:1 (ian drew) and editable by org:1 editors (drew)",
	},
	{
		Subject:  NewEntityKey(GroupType, "BA-group"),
		Object:   NewEntityKey(FeedType, "BA"),
		Relation: ParentRelation,
		Notes:    "feed:2 should be viewable to members of org:2 () and editable by org:2 editors (ian)",
	},
	{
		Subject:  NewEntityKey(GroupType, "HA-group"),
		Object:   NewEntityKey(FeedType, "HA"),
		Relation: ParentRelation,
		Notes:    "feed:3 should be viewable to all members of tenant:1 (admin nisar ian drew) and editable by org:3 editors ()",
	},
	{
		Subject:  NewEntityKey(GroupType, "EX-group"),
		Object:   NewEntityKey(FeedType, "EX"),
		Relation: ParentRelation,
		Notes:    "feed:4 should only be viewable to admins of tenant:1 (admin)",
	},
	// loaded from postgres
	// {
	// 	Subject:  NewEntityKey(FeedType, "BA"),
	// 	Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
	// 	Relation: ParentRelation,
	// },
	{
		Subject:  NewEntityKey(UserType, "admin"),
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
		Subject:  NewEntityKey(UserType, "nisar"),
		Object:   NewEntityKey(TenantType, "tl-tenant"),
		Relation: MemberRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "nisar"),
		Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation: ViewerRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "test2"),
		Object:   NewEntityKey(TenantType, "restricted-tenant"),
		Relation: MemberRelation,
	},
}

var checkerCheck = []fgaTestTuple{
	{
		Subject:       NewEntityKey(UserType, "admin"),
		Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		ExpectActions: []Action{CanView},
	},
	{
		Subject:       NewEntityKey(UserType, "admin"),
		Object:        NewEntityKey(GroupType, "CT-group"),
		ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
	},
	{
		Subject:       NewEntityKey(UserType, "admin"),
		Object:        NewEntityKey(GroupType, "BA-group"),
		ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
	},
	{
		Subject:       NewEntityKey(UserType, "admin"),
		Object:        NewEntityKey(GroupType, "HA-group"),
		ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
	},
	{
		Subject:       NewEntityKey(UserType, "admin"),
		Object:        NewEntityKey(GroupType, "EX-group"),
		ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
	},
	{
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(GroupType, "test-group"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "admin"),
		Object:        NewEntityKey(TenantType, "tl-tenant"),
		ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(TenantType, "restricted-tenant"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "ian"),
		Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		ExpectActions: []Action{CanView},
	},
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
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedType, "EX"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedType, "5"),
		ExpectError: true,
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
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "EX-group"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "test-group"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "ian"),
		Object:        NewEntityKey(TenantType, "tl-tenant"),
		ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(TenantType, "restricted-tenant"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		ExpectActions: []Action{-CanView},
		Notes:         "only feed:2 readers and nisar",
		ExpectError:   true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(FeedType, "CT"),
		ExpectActions: []Action{CanView, CanEdit},
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(FeedType, "BA"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(FeedType, "HA"),
		ExpectActions: []Action{CanView, -CanEdit},
		Test:          "check",
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(FeedType, "EX"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(FeedType, "5"),
		ExpectActions: []Action{-CanView, -CanEdit},
		ExpectError:   true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(GroupType, "CT-group"),
		ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(GroupType, "BA-group"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(GroupType, "HA-group"),
		ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(GroupType, "EX-group"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(GroupType, "test-group"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(TenantType, "tl-tenant"),
		ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(TenantType, "restricted-tenant"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(TenantType, "tl-tenant"),
		ExpectActions: []Action{CanView, -CanEdit},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(TenantType, "restricted-tenant"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		ExpectActions: []Action{CanView},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "CT"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "BA"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(FeedType, "HA"),
		ExpectActions: []Action{CanView, -CanEdit},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "EX"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "5"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(GroupType, "HA-group"),
		ExpectActions: []Action{CanView, -CanEdit},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(GroupType, "EX-group"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(TenantType, "tl-tenant"),
		ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(TenantType, "restricted-tenant"),
		ExpectError: true,
	},
}

var checkerListTests = []fgaTestTuple{
	{
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(FeedType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
	},
	{
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(GroupType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group", "EX-group"),
	},
	{
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(TenantType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
	},
	{
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(FeedVersionType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(FeedType, "CT", "BA", "HA"),
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group"),
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(TenantType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedVersionType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(FeedType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(FeedType, "CT", "HA"),
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(TenantType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(GroupType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(GroupType, "CT-group", "HA-group"),
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(FeedVersionType, ""),
		ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(GroupType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(GroupType, "HA-group"),
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(FeedType, "HA"),
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(TenantType, ""),
		ListAction:  CanView,
		ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
	},
}

var checkerWriteTests = []fgaTestTuple{
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(GroupType, "HA-group"),
		Relation:    ViewerRelation,
		Notes:       "does not exist",
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "HA-group"),
		Notes:       "invalid relation",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test102"),
		Object:      NewEntityKey(GroupType, "100"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		CheckAsUser: "ian",
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:  NewEntityKey(UserType, "nisar"),
		Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation: ViewerRelation,

		Expect:      "fail",
		Notes:       "already exists",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test1"),
		Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test2"),
		Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test3"),
		Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		CheckAsUser: "ian",
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(TenantType, "tl-tenant"),
		Relation:    MemberRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(GroupType, "CT-group"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(GroupType, "CT-group"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		ExpectError: true,
		CheckAsUser: "ian",
	},
}

var checkerDeleteTests = []fgaTestTuple{
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "CT-group"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "CT-group"),
		Relation:    ViewerRelation,
		Notes:       "already deleted",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test102"),
		Object:      NewEntityKey(GroupType, "100"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		ExpectError: true,
		CheckAsUser: "ian",
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation:    ViewerRelation,
		Notes:       "already deleted",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		ExpectError: true,
		CheckAsUser: "ian",
	},
	{
		Subject:     NewEntityKey(UserType, "test2"),
		Object:      NewEntityKey(TenantType, "restricted-tenant"),
		Relation:    MemberRelation,
		Notes:       "unauthorized",
		CheckAsUser: "admin",
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "test101"),
		Object:      NewEntityKey(GroupType, "BA-group"),
		Relation:    ViewerRelation,
		Notes:       "does not exist",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test101"),
		Object:      NewEntityKey(GroupType, "BA-group"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		ExpectError: true,
		CheckAsUser: "ian",
	},
}

var checkerGetTests = []fgaTestTuple{
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(TenantType, "tl-tenant"),

		Expect: "user:admin:admin user:ian:member user:drew:member user:nisar:member",
	},
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
		Expect: "feed:2:parent user:nisar:viewer",
	},
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(FeedType, "CT"),
		Expect: "org:1:parent",
	},
}

func TestChecker(t *testing.T) {
	te := testfinder.Finders(t, nil, nil)
	dbx := te.Finder.DBX()

	if os.Getenv("TL_TEST_FGA_ENDPOINT") == "" {
		t.Skip("no TL_TEST_FGA_ENDPOINT set, skipping")
		return
	}

	// TENANTS
	t.Run("TenantList", func(t *testing.T) {
		checker := newTestChecker(t)
		checks := []fgaTestTuple{
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
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
				for _, v := range ret {
					gotNames = append(gotNames, v.Name.Val)
				}
				var expectNames []string
				for _, v := range tc.ExpectNames {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "tenant names")
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerCheck {
			if tc.Object.Type != TenantType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.TenantPermissions(
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
				)
				checkExpectError(t, err, tc.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.Checks)
			})
		}
	})

	t.Run("TenantAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerWriteTests {
			if tc.Object.Type != TenantType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.TenantAddPermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Object.ID(),
					ltk.Subject.Name,
					ltk.Relation,
				)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
			})
		}
	})

	t.Run("TenantRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerDeleteTests {
			if tc.Object.Type != TenantType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.TenantRemovePermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Object.ID(),
					ltk.Subject.Name,
					ltk.Relation,
				)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
			})
		}
	})

	// GROUPS
	t.Run("GroupList", func(t *testing.T) {
		checker := newTestChecker(t)
		checks := []fgaTestTuple{
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group", "EX-group"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(GroupType, "CT-group", "HA-group"),
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(GroupType, "HA-group"),
			},
		}
		for _, tc := range checks {
			if tc.Object.Type != GroupType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.GroupList(context.Background(), newTestUser(ltk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret {
					gotNames = append(gotNames, v.Name.Val)
				}
				var expectNames []string
				for _, v := range tc.ExpectNames {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "group names")
			})
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerCheck {
			if tc.Object.Type != GroupType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.GroupPermissions(
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
				)
				checkExpectError(t, err, tc.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.Checks)
			})
		}
	})

	t.Run("GroupAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerWriteTests {
			if tc.Object.Type != GroupType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.GroupAddPermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, tc.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
				)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
			})
		}
	})

	t.Run("GroupRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerDeleteTests {
			if tc.Object.Type != GroupType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.GroupRemovePermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
				)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
			})
		}
	})

	// FEEDS
	t.Run("FeedList", func(t *testing.T) {
		checker := newTestChecker(t)
		checks := []fgaTestTuple{
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(FeedType, "CT", "BA", "HA"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(FeedType, "CT", "HA"),
			},

			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(FeedType, "HA"),
			},
		}
		for _, tc := range checks {
			if tc.Object.Type != FeedType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedList(context.Background(), newTestUser(ltk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret {
					gotNames = append(gotNames, v.OnestopID.Val)
				}
				var expectNames []string
				for _, v := range tc.ExpectNames {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "feed names")

			})
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerCheck {
			if tc.Object.Type != FeedType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedPermissions(
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
				)
				checkExpectError(t, err, tc.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.Checks)
			})
		}
	})

	// FEED VERSIONS
	t.Run("FeedVersionList", func(t *testing.T) {
		checker := newTestChecker(t)
		checks := []fgaTestTuple{
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(FeedVersionType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, ""),
				ListAction:  CanView,
				ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(FeedVersionType, ""),
				ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			},
		}
		for _, tc := range checks {
			if tc.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedVersionList(context.Background(), newTestUser(ltk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range ret {
					gotNames = append(gotNames, v.SHA1.Val)
				}
				var expectNames []string
				for _, v := range tc.ExpectNames {
					expectNames = append(expectNames, v.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "feed version names")
			})
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerCheck {
			if tc.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				ret, err := checker.FeedVersionPermissions(
					context.Background(),
					newTestUser(ltk.Subject.Name),
					ltk.Object.ID(),
				)
				checkExpectError(t, err, tc.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tc.Checks)
			})
		}
	})

	t.Run("FeedVersionAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerWriteTests {
			if tc.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.FeedVersionAddPermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
				)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
			})
		}
	})

	t.Run("FeedVersionRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tc := range checkerDeleteTests {
			if tc.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tc.String(), func(t *testing.T) {
				ltk := dbTupleLookup(t, dbx, tc.TupleKey())
				err := checker.FeedVersionRemovePermission(
					context.Background(),
					newTestUser(stringOr(tc.CheckAsUser, ltk.Subject.Name)),
					ltk.Subject.Name,
					ltk.Object.ID(),
					ltk.Relation,
				)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
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

func checkActionSubset(t testing.TB, actions any, checks []string) {
	checkA, err := actionsToMap(actions)
	if err != nil {
		t.Error(err)
	}
	checkActions := checkActionsToMap(checks)
	checkMapSubset(t, checkA, checkActions)
}

func checkMapSubset(t testing.TB, value map[string]bool, contains map[string]bool) {
	for k, v := range contains {
		a, ok := value[k]
		if !ok {
			t.Errorf("expected %#v to contain key %s", value, k)
		} else if a != v {
			t.Errorf("expected %#v to contain %s=%t, did not", value, k, v)
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

func checkActionsToMap(v []string) map[string]bool {
	ret := map[string]bool{}
	for _, checkAction := range v {
		expect := true
		if strings.HasPrefix(checkAction, "+") {
			checkAction = strings.TrimPrefix(checkAction, "+")
		} else if strings.HasPrefix(checkAction, "-") {
			expect = false
			checkAction = strings.TrimPrefix(checkAction, "-")
		}
		ret[checkAction] = expect
	}
	return ret
}

func newTestChecker(t testing.TB) *Checker {
	te := testfinder.Finders(t, nil, nil)
	dbx := te.Finder.DBX()

	auth0c := NewMockAuthnClient()
	auth0c.AddUser("ian", User{Name: "Ian", ID: "ian", Email: "ian@example.com"})
	auth0c.AddUser("drew", User{Name: "Drew", ID: "drew", Email: "drew@example.com"})
	auth0c.AddUser("nisar", User{Name: "Nisar", ID: "nisar", Email: "nisar@example.com"})

	fgac, err := newTestCheckerFGA(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range checkerTestData {
		if err := fgac.WriteTuple(context.Background(), dbTupleLookup(t, dbx, tc.TupleKey())); err != nil {
			t.Fatal(err)
		}
	}
	checker := NewChecker(auth0c, fgac, dbx, nil)
	return checker
}

func newTestCheckerFGA(t testing.TB) (*FGAClient, error) {
	cfg := AuthzConfig{
		FGAEndpoint:      os.Getenv("TL_TEST_FGA_ENDPOINT"),
		FGALoadModelFile: "../test/authz/tls.json",
		GlobalAdmin:      "global_admin",
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
		t.Log("lookup:", ek.Type, "name:", ek.Name, "not found")
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
