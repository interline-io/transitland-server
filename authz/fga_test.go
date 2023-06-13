package authz

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestFGAClient(t testing.TB, testTuples []fgaTestTuple) *FGAClient {
	cfg := AuthzConfig{
		FGAEndpoint:      os.Getenv("TL_TEST_FGA_ENDPOINT"),
		FGALoadModelFile: "../test/authz/tls.json",
	}
	fgac, err := NewFGAClient(cfg.FGAEndpoint, "", "")
	if err != nil {
		t.Fatal(err)
		return nil
	}
	if _, err := fgac.CreateStore(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := fgac.CreateModel(context.Background(), cfg.FGALoadModelFile); err != nil {
		t.Fatal(err)
	}
	for _, tk := range testTuples {
		if err := fgac.WriteTuple(context.Background(), tk.TupleKey()); err != nil {
			t.Fatal(err)
		}
	}
	return fgac
}

type fgaTestTuple struct {
	Subject           EntityKey
	Object            EntityKey
	Action            Action
	Relation          Relation
	Checks            []string
	Test              string
	Expect            string
	Notes             string
	ExpectError       bool
	CheckAsUser       string
	ExpectErrorAsUser bool
	ExpectActions     []Action
	ExpectIds         []int
	ExpectNames       []EntityKey
	ListAction        Action
}

func (tk *fgaTestTuple) TupleKey() TupleKey {
	return TupleKey{Subject: tk.Subject, Object: tk.Object, Relation: tk.Relation, Action: tk.Action}
}

func (tk *fgaTestTuple) String() string {
	return tk.TupleKey().String() + "|checkuser:" + tk.CheckAsUser
}

func TestFGAClient(t *testing.T) {
	testData := []fgaTestTuple{
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
		{
			Subject:  NewEntityKey(FeedType, "BA"),
			Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ParentRelation,
		},
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

	if os.Getenv("TL_TEST_FGA_ENDPOINT") == "" {
		t.Skip("no TL_TEST_FGA_ENDPOINT set, skipping")
		return
	}

	t.Run("GetObjectTuples", func(t *testing.T) {
		fgac := newTestFGAClient(t, testData)
		checks := []fgaTestTuple{
			{
				Object: NewEntityKey(TenantType, "tl-tenant"),
				Expect: "user:admin:admin user:ian:member user:drew:member user:nisar:member",
			},
			{
				Object: NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Expect: "feed:BA:parent user:nisar:viewer",
			},
			{
				Object: NewEntityKey(FeedType, "CT"),
				Expect: "org:CT-group:parent",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				tks, err := fgac.GetObjectTuples(context.Background(), tc.TupleKey())
				if err != nil {
					t.Error(err)
				}
				expect := strings.Split(tc.Expect, " ")
				var got []string
				for _, vtk := range tks {
					got = append(got, fmt.Sprintf("%s:%s:%s", vtk.Subject.Type, vtk.Subject.Name, vtk.Relation))
				}
				assert.ElementsMatch(t, expect, got, "usertype:username:relation does not match")

			})
		}
	})

	t.Run("Check", func(t *testing.T) {
		fgac := newTestFGAClient(t, testData)
		checks := []fgaTestTuple{
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
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(GroupType, "test-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
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
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "EX"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "test"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
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
				Test:          "check",
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "test-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
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
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
				Test:          "check",
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "EX"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "test"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "test-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "EX"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "test"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				for _, checkAction := range tc.ExpectActions {
					expect := true
					if checkAction < 0 {
						expect = false
						checkAction = checkAction * -1
					}
					var err error
					ltk := tc.TupleKey()
					ltk.Action = checkAction
					ok, err := fgac.Check(context.Background(), ltk)
					if err != nil {
						t.Fatal(err)
					}
					if ok && !expect {
						t.Errorf("for %s got %t, expected %t", checkAction.String(), ok, expect)
					}
					if !ok && expect {
						t.Errorf("got %s %t, expected %t", checkAction.String(), ok, expect)
					}
				}
			})
		}
	})

	t.Run("ListObjects", func(t *testing.T) {
		fgac := newTestFGAClient(t, testData)
		checks := []fgaTestTuple{
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1, 2, 3, 4},
				ExpectNames: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
			},
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanEdit,
				ExpectIds:   []int{1, 2, 3, 4},
				ExpectNames: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
			},

			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1, 2, 3, 4},
				ExpectNames: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group", "EX-group"),
			},
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanEdit,
				ExpectIds:   []int{1, 2, 3, 4},
				ExpectNames: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group", "EX-group"),
			},

			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1},
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:     NewEntityKey(UserType, "admin"),
				Object:      NewEntityKey(FeedVersionType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1},
				ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanEdit,
				ExpectIds:   []int{2},
				ExpectNames: newEntityKeys(FeedType, "BA"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1, 2, 3},
				ExpectNames: newEntityKeys(FeedType, "CT", "BA", "HA"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1, 2, 3},
				ExpectNames: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1},
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1},
				ExpectNames: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanEdit,
				ExpectIds:   []int{1},
				ExpectNames: newEntityKeys(FeedType, "CT"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1, 3},
				ExpectNames: newEntityKeys(FeedType, "CT", "HA"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1},
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1, 3},
				ExpectNames: newEntityKeys(GroupType, "CT-group", "HA-group"),
			},
			{
				Subject:     NewEntityKey(UserType, "drew"),
				Object:      NewEntityKey(FeedVersionType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{},
				ExpectNames: newEntityKeys(FeedVersionType),
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(GroupType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{3},
				ExpectNames: newEntityKeys(GroupType, "HA-group"),
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(FeedType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{3},
				ExpectNames: newEntityKeys(FeedType, "HA"),
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(TenantType, ""),
				ListAction:  CanView,
				ExpectIds:   []int{1},
				ExpectNames: newEntityKeys(TenantType, "tl-tenant"),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := tc.TupleKey()
				ltk.Action = tc.ListAction
				objs, err := fgac.ListObjects(context.Background(), ltk)
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range objs {
					gotNames = append(gotNames, v.Object.Name)
				}
				var expectNames []string
				for _, ek := range tc.ExpectNames {
					expectNames = append(expectNames, ek.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "object ids")
			})
		}
	})

	t.Run("WriteTuple", func(t *testing.T) {
		fgac := newTestFGAClient(t, testData)
		checks := []fgaTestTuple{
			{
				Subject:  NewEntityKey(UserType, "test100"),
				Object:   NewEntityKey(TenantType, "tl-tenant"),
				Relation: MemberRelation,
			},
			{
				Subject:  NewEntityKey(UserType, "test100"),
				Object:   NewEntityKey(GroupType, "CT-group"),
				Relation: ViewerRelation,
			},
			{
				Subject:     NewEntityKey(UserType, "test100"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    ViewerRelation,
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "test100"),
				Object:   NewEntityKey(GroupType, "HA-group"),
				Relation: ViewerRelation,
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "HA-group"),
				Notes:       "invalid relation",
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "test102"),
				Object:   NewEntityKey(GroupType, "100"),
				Relation: ViewerRelation,
			},
			{
				Subject:  NewEntityKey(UserType, "ian"),
				Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: ViewerRelation,
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				Notes:       "already exists",
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "test1"),
				Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: ViewerRelation,
			},
			{
				Subject:  NewEntityKey(UserType, "test2"),
				Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: ViewerRelation,
			},
			{
				Subject:  NewEntityKey(UserType, "test3"),
				Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: ViewerRelation,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Write tuple and check if error was expected
				ltk := tc.TupleKey()
				err := fgac.WriteTuple(context.Background(), ltk)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
				// Check was written
				tks, err := fgac.GetObjectTuples(context.Background(), ltk)
				if err != nil {
					t.Error(err)
				}
				var gotTks []string
				for _, v := range tks {
					gotTks = append(gotTks, fmt.Sprintf("%s:%s:%s", v.Subject.Type, v.Subject.Name, v.Relation))
				}
				checkTk := fmt.Sprintf("%s:%s:%s", ltk.Subject.Type, ltk.Subject.Name, ltk.Relation)
				assert.Contains(t, gotTks, checkTk, "written tuple not found in updated object tuples")
			})
		}
	})

	t.Run("DeleteTuple", func(t *testing.T) {
		fgac := newTestFGAClient(t, testData)
		checks := []fgaTestTuple{
			{
				Subject:  NewEntityKey(UserType, "ian"),
				Object:   NewEntityKey(GroupType, "CT-group"),
				Relation: 4,
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "CT-group"),
				Relation:    4,
				Notes:       "already deleted",
				ExpectError: true,
			},
			{
				Subject:     NewEntityKey(UserType, "test102"),
				Object:      NewEntityKey(GroupType, "100"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "nisar"),
				Object:   NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: 4,
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    4,
				Notes:       "already deleted",
				ExpectError: true,
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "test2"),
				Object:   NewEntityKey(TenantType, "restricted-tenant"),
				Relation: 2,
			},
			{
				Subject:     NewEntityKey(UserType, "test101"),
				Object:      NewEntityKey(GroupType, "BA-group"),
				Relation:    4,
				Notes:       "does not exist",
				ExpectError: true,
			},
			{
				Subject:     NewEntityKey(UserType, "test101"),
				Object:      NewEntityKey(GroupType, "BA-group"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := tc.TupleKey()
				err := fgac.DeleteTuple(context.Background(), ltk)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
			})
		}
	})
}

func checkExpectError(t testing.TB, err error, expect bool) bool {
	if err != nil && !expect {
		t.Errorf("got error '%s', did not expect error", err.Error())
		return false
	}
	if err == nil && expect {
		t.Errorf("got no error, expected error")
		return false
	}
	if err != nil {
		return false
	}
	return true
}
