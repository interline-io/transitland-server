package fga

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/interline-io/transitland-server/internal/generated/azpb"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// For less typing

type Action = azpb.Action
type ObjectType = azpb.ObjectType
type Relation = azpb.Relation

var FeedType = azpb.FeedType
var UserType = azpb.UserType
var TenantType = azpb.TenantType
var GroupType = azpb.GroupType
var FeedVersionType = azpb.FeedVersionType

var ViewerRelation = azpb.ViewerRelation
var MemberRelation = azpb.MemberRelation
var AdminRelation = azpb.AdminRelation
var ManagerRelation = azpb.ManagerRelation
var ParentRelation = azpb.ParentRelation
var EditorRelation = azpb.EditorRelation

var CanEdit = azpb.CanEdit
var CanView = azpb.CanView
var CanCreateFeedVersion = azpb.CanCreateFeedVersion
var CanDeleteFeedVersion = azpb.CanDeleteFeedVersion
var CanCreateFeed = azpb.CanCreateFeed
var CanDeleteFeed = azpb.CanDeleteFeed
var CanSetGroup = azpb.CanSetGroup
var CanCreateOrg = azpb.CanCreateOrg
var CanEditMembers = azpb.CanEditMembers
var CanDeleteOrg = azpb.CanDeleteOrg
var CanSetTenant = azpb.CanSetTenant

// Tests

type testCase struct {
	Subject            azpb.EntityKey
	Object             azpb.EntityKey
	Action             azpb.Action
	Relation           azpb.Relation
	Expect             string
	Notes              string
	ExpectError        bool
	ExpectUnauthorized bool
	CheckAsUser        string
	ExpectActions      []azpb.Action
	ExpectKeys         []azpb.EntityKey
}

func (tk *testCase) TupleKey() azpb.TupleKey {
	return azpb.TupleKey{Subject: tk.Subject, Object: tk.Object, Relation: tk.Relation, Action: tk.Action}
}

func (tk *testCase) String() string {
	if tk.Notes != "" {
		return tk.Notes
	}
	a := tk.TupleKey().String()
	if tk.CheckAsUser != "" {
		a = a + "|checkuser:" + tk.CheckAsUser
	}
	return a
}

func TestFGAClient(t *testing.T) {
	fgaUrl, a, ok := testutil.CheckEnv("TL_TEST_FGA_ENDPOINT")
	if !ok {
		t.Skip(a)
		return
	}

	testData := []testCase{
		// Assign users to tenants
		{
			Notes:    "All users can access all-users-tenant",
			Subject:  azpb.NewEntityKey(UserType, "*"),
			Object:   azpb.NewEntityKey(TenantType, "all-users-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "tl-tenant-admin"),
			Object:   azpb.NewEntityKey(TenantType, "tl-tenant"),
			Relation: AdminRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "ian"),
			Object:   azpb.NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "drew"),
			Object:   azpb.NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "tl-tenant-member"),
			Object:   azpb.NewEntityKey(TenantType, "tl-tenant"),
			Relation: MemberRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "test2"),
			Object:   azpb.NewEntityKey(TenantType, "restricted-tenant"),
			Relation: MemberRelation,
		},
		// Assign groups to tenants
		{
			Notes:    "org:CT-group belongs to tenant:tl-tenant",
			Subject:  azpb.NewEntityKey(TenantType, "tl-tenant"),
			Object:   azpb.NewEntityKey(GroupType, "CT-group"),
			Relation: ParentRelation,
		},
		{
			Notes:    "org:BA-group belongs to tenant:tl-tenant",
			Subject:  azpb.NewEntityKey(TenantType, "tl-tenant"),
			Object:   azpb.NewEntityKey(GroupType, "BA-group"),
			Relation: ParentRelation,
		},
		{
			Notes:    "org:HA-group belongs to tenant:tl-tenant",
			Subject:  azpb.NewEntityKey(TenantType, "tl-tenant"),
			Object:   azpb.NewEntityKey(GroupType, "HA-group"),
			Relation: ParentRelation,
		},
		{
			Notes:    "org:EX-group will be for admins only",
			Subject:  azpb.NewEntityKey(TenantType, "tl-tenant"),
			Object:   azpb.NewEntityKey(GroupType, "EX-group"),
			Relation: ParentRelation,
		},
		{
			Notes:    "all tl-tenant members can view HA-group",
			Subject:  azpb.NewEntityKey(TenantType, "tl-tenant#member"),
			Object:   azpb.NewEntityKey(GroupType, "HA-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  azpb.NewEntityKey(TenantType, "restricted-tenant"),
			Object:   azpb.NewEntityKey(GroupType, "test-group"),
			Relation: ParentRelation,
		},
		// Assign users to groups
		{
			Subject:  azpb.NewEntityKey(UserType, "ian"),
			Object:   azpb.NewEntityKey(GroupType, "CT-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "ian"),
			Object:   azpb.NewEntityKey(GroupType, "BA-group"),
			Relation: EditorRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "drew"),
			Object:   azpb.NewEntityKey(GroupType, "CT-group"),
			Relation: EditorRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "test-group-viewer"),
			Object:   azpb.NewEntityKey(GroupType, "test-group"),
			Relation: ViewerRelation,
		},
		{
			Subject:  azpb.NewEntityKey(UserType, "test-group-editor"),
			Object:   azpb.NewEntityKey(GroupType, "test-group"),
			Relation: EditorRelation,
		},
		// Assign feeds to groups
		{
			Subject:  azpb.NewEntityKey(GroupType, "CT-group"),
			Object:   azpb.NewEntityKey(FeedType, "CT"),
			Relation: ParentRelation,
			Notes:    "feed:CT should be viewable to members of org:CT-group (ian drew) and editable by org:CT-group editors (drew)",
		},
		{
			Subject:  azpb.NewEntityKey(GroupType, "BA-group"),
			Object:   azpb.NewEntityKey(FeedType, "BA"),
			Relation: ParentRelation,
			Notes:    "feed:BA should be viewable to members of org:BA-group () and editable by org:BA-group editors (ian)",
		},
		{
			Subject:  azpb.NewEntityKey(GroupType, "HA-group"),
			Object:   azpb.NewEntityKey(FeedType, "HA"),
			Relation: ParentRelation,
			Notes:    "feed:HA should be viewable to all members of tenant:tl-tenant",
		},
		{
			Subject:  azpb.NewEntityKey(GroupType, "EX-group"),
			Object:   azpb.NewEntityKey(FeedType, "EX"),
			Relation: ParentRelation,
			Notes:    "feed:EX should only be viewable to admins of tenant:tl-tenant",
		},
		// Assign feed version specific permissions
		// NOTE: This assignment is necessary for FGA tests
		// This relation is implicit in full Checker tests
		{
			Subject:  azpb.NewEntityKey(FeedType, "BA"),
			Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ParentRelation,
		},
		// Assign users to feed versions
		{
			Subject:  azpb.NewEntityKey(UserType, "tl-tenant-member"),
			Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ViewerRelation,
		},
		{
			Subject:  azpb.NewEntityKey(GroupType, "test-group").WithRefRel(ViewerRelation),
			Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
			Relation: ViewerRelation,
		},
		{
			Subject:  azpb.NewEntityKey(TenantType, "tl-tenant").WithRefRel(MemberRelation),
			Object:   azpb.NewEntityKey(FeedVersionType, "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			Relation: ViewerRelation,
		},
	}

	t.Run("GetObjectTuples", func(t *testing.T) {
		fgac := newTestFGAClient(t, fgaUrl, testData)
		checks := []testCase{
			{
				Object: azpb.NewEntityKey(TenantType, "tl-tenant"),
				Expect: "user:tl-tenant-admin:admin user:ian:member user:drew:member user:tl-tenant-member:member",
			},
			{
				Object: azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Expect: "feed:BA:parent user:tl-tenant-member:viewer org:test-group#viewer:viewer",
			},
			{
				Object: azpb.NewEntityKey(FeedType, "CT"),
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
					got = append(got, fmt.Sprintf("%s:%s", vtk.Subject.String(), vtk.Relation))
				}
				assert.ElementsMatch(t, expect, got, "usertype:username:relation does not match")
			})
		}
	})

	t.Run("Check", func(t *testing.T) {
		fgac := newTestFGAClient(t, fgaUrl, testData)
		checks := []testCase{
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(GroupType, "test-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:        azpb.NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(FeedType, "EX"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(FeedType, "test"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(GroupType, "test-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "ian"),
				Object:        azpb.NewEntityKey(TenantType, "all-users-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(TenantType, "all-users-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{-CanView},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{CanView, CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(FeedType, "EX"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(FeedType, "test"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(GroupType, "CT-group"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(GroupType, "BA-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(GroupType, "test-group"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "drew"),
				Object:        azpb.NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(FeedType, "CT"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(FeedType, "BA"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(FeedType, "HA"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(FeedType, "EX"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(FeedType, "test"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(GroupType, "HA-group"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(GroupType, "EX-group"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(TenantType, "tl-tenant"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:        azpb.NewEntityKey(TenantType, "restricted-tenant"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "test-group-viewer"),
				Object:        azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers},
			},
			{
				Subject:       azpb.NewEntityKey(UserType, "test-group-editor"),
				Object:        azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers},
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
		fgac := newTestFGAClient(t, fgaUrl, testData)
		checks := []testCase{
			{
				Notes:      "tl-tenant-admin can access all feeds in tl-tenant",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     azpb.NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
			},
			{
				Notes:      "tl-tenant-admin can edit all feeds in tl-tenant",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     azpb.NewEntityKey(FeedType, ""),
				Action:     CanEdit,
				ExpectKeys: newEntityKeys(FeedType, "CT", "BA", "HA", "EX"),
			},
			{
				Notes:      "tl-tenant-admin can view all groups in tl-tenant",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     azpb.NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group", "EX-group"),
			},
			{
				Notes:      "tl-tenant-admin can edit all groups in tl-tenant",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     azpb.NewEntityKey(GroupType, ""),
				Action:     CanEdit,
				ExpectKeys: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group", "EX-group"),
			},
			{
				Notes:      "tl-tenant-admin can view tenants tl-tenant and all-users-tenant",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     azpb.NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant", "all-users-tenant"),
			},
			{
				Notes:      "tl-tenant-admin can view a feed version that belongs to a feed or group in tl-tenant or d281 which viewable to all tenant members",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-admin"),
				Object:     azpb.NewEntityKey(FeedVersionType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0", "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			},
			{
				Notes:      "ian can edit feed BA in tl-tenant",
				Subject:    azpb.NewEntityKey(UserType, "ian"),
				Object:     azpb.NewEntityKey(FeedType, ""),
				Action:     CanEdit,
				ExpectKeys: newEntityKeys(FeedType, "BA"),
			},
			{
				Notes:      "ian can view feeds CT, BA, HA",
				Subject:    azpb.NewEntityKey(UserType, "ian"),
				Object:     azpb.NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "BA", "HA"),
			},
			{
				Notes:      "ian can view groups CT-group BA-group HA-group",
				Subject:    azpb.NewEntityKey(UserType, "ian"),
				Object:     azpb.NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "CT-group", "BA-group", "HA-group"),
			},
			{
				Notes:      "ian can view tenants tl-tenant (member explicitly) and all-users-tenant (user:*)",
				Subject:    azpb.NewEntityKey(UserType, "ian"),
				Object:     azpb.NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant", "all-users-tenant"),
			},
			{
				Notes:      "ian can view feed version e535 because of access to feed BA, group BA-group or d281 which is viewable to all tenant members",
				Subject:    azpb.NewEntityKey(UserType, "ian"),
				Object:     azpb.NewEntityKey(FeedVersionType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0", "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			},
			{
				Notes:      "drew can edit feed CT because editor of CT-group",
				Subject:    azpb.NewEntityKey(UserType, "drew"),
				Object:     azpb.NewEntityKey(FeedType, ""),
				Action:     CanEdit,
				ExpectKeys: newEntityKeys(FeedType, "CT"),
			},
			{
				Notes:      "drew can view feed CT because editor of CT-group and HA because HA has all tenant members",
				Subject:    azpb.NewEntityKey(UserType, "drew"),
				Object:     azpb.NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "CT", "HA"),
			},
			{
				Notes:      "drew can access tl-tenant because member and all-users-tenant because user:*",
				Subject:    azpb.NewEntityKey(UserType, "drew"),
				Object:     azpb.NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant", "all-users-tenant"),
			},
			{
				Notes:      "drew can access group CT-group because member and HA-group through tenant#member",
				Subject:    azpb.NewEntityKey(UserType, "drew"),
				Object:     azpb.NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "CT-group", "HA-group"),
			},
			{
				Notes:      "drew is not explicitly assigned any feed versions but can access d281 because it is viewable to all tenant members",
				Subject:    azpb.NewEntityKey(UserType, "drew"),
				Object:     azpb.NewEntityKey(FeedVersionType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedVersionType, "d2813c293bcfd7a97dde599527ae6c62c98e66c6"),
			},
			{
				Notes:      "tl-tenant-member can access HA-group through HA-group#viewer:tl-tenant#member",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:     azpb.NewEntityKey(GroupType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(GroupType, "HA-group"),
			},
			{
				Notes:      "tl-tenant-member can access feed HA through group:HA-group",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:     azpb.NewEntityKey(FeedType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(FeedType, "HA"),
			},
			{
				Notes:      "tl-tenant-member can view tl-tenant through member and all-users-tenant through user:*",
				Subject:    azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:     azpb.NewEntityKey(TenantType, ""),
				Action:     CanView,
				ExpectKeys: newEntityKeys(TenantType, "tl-tenant", "all-users-tenant"),
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				ltk := tc.TupleKey()
				ltk.Action = tc.Action
				objs, err := fgac.ListObjects(context.Background(), ltk)
				if err != nil {
					t.Fatal(err)
				}
				var gotNames []string
				for _, v := range objs {
					gotNames = append(gotNames, v.Object.Name)
				}
				var expectNames []string
				for _, ek := range tc.ExpectKeys {
					expectNames = append(expectNames, ek.Name)
				}
				assert.ElementsMatch(t, expectNames, gotNames, "object ids")
			})
		}
	})

	t.Run("WriteTuple", func(t *testing.T) {
		checks := []testCase{
			{
				Notes:    "user:* can be a member of a tenant",
				Subject:  azpb.NewEntityKey(UserType, "*"),
				Object:   azpb.NewEntityKey(TenantType, "tl-tenants"),
				Relation: MemberRelation,
			},
			{
				Notes:       "user:* cannot be an admin of a tenant",
				Subject:     azpb.NewEntityKey(UserType, "*"),
				Object:      azpb.NewEntityKey(TenantType, "tl-tenants"),
				Relation:    AdminRelation,
				ExpectError: true,
			},
			{
				Notes:    "a tenant#member can be a viewer of a group",
				Subject:  azpb.NewEntityKey(TenantType, "tl-tenant#member"),
				Object:   azpb.NewEntityKey(GroupType, "BA-group"),
				Relation: ViewerRelation,
			},
			{
				Notes:       "a tenant#admin cannot be a viewer of a group",
				Subject:     azpb.NewEntityKey(TenantType, "tl-tenant#admin"),
				Object:      azpb.NewEntityKey(GroupType, "BA-group"),
				Relation:    ViewerRelation,
				ExpectError: true,
			},
			{
				Notes:    "a tenant#member can be an editor of a group",
				Subject:  azpb.NewEntityKey(TenantType, "tl-tenant#member"),
				Object:   azpb.NewEntityKey(GroupType, "BA-group"),
				Relation: EditorRelation,
				// Formerly disallowed, now OK
				// ExpectError: true,
			},
			{
				Notes:    "user can be a member of a tenant",
				Subject:  azpb.NewEntityKey(UserType, "test100"),
				Object:   azpb.NewEntityKey(TenantType, "tl-tenant"),
				Relation: MemberRelation,
			},
			{
				Notes:    "user can be an admin of a tenant",
				Subject:  azpb.NewEntityKey(UserType, "test100"),
				Object:   azpb.NewEntityKey(TenantType, "tl-tenant"),
				Relation: AdminRelation,
			},
			{
				Notes:       "already exists",
				Subject:     azpb.NewEntityKey(UserType, "ian"),
				Object:      azpb.NewEntityKey(TenantType, "tl-tenant"),
				Relation:    MemberRelation,
				ExpectError: true,
			},
			{
				Notes:    "a user can be a viewer of a group",
				Subject:  azpb.NewEntityKey(UserType, "test100"),
				Object:   azpb.NewEntityKey(GroupType, "HA-group"),
				Relation: ViewerRelation,
			},
			{
				Notes:    "a user can be an editor of a group",
				Subject:  azpb.NewEntityKey(UserType, "test100"),
				Object:   azpb.NewEntityKey(GroupType, "HA-group"),
				Relation: EditorRelation,
			},
			{
				Notes:    "a user can be a manager of a group",
				Subject:  azpb.NewEntityKey(UserType, "test100"),
				Object:   azpb.NewEntityKey(GroupType, "HA-group"),
				Relation: ManagerRelation,
			},
			{
				Notes:       "invalid relation",
				Subject:     azpb.NewEntityKey(UserType, "ian"),
				Object:      azpb.NewEntityKey(GroupType, "HA-group"),
				Relation:    ParentRelation,
				ExpectError: true,
			},
			{
				Notes:    "a user can be a viewer of a group",
				Subject:  azpb.NewEntityKey(UserType, "test102"),
				Object:   azpb.NewEntityKey(GroupType, "100"),
				Relation: ViewerRelation,
			},
			{
				Notes:    "a user can be a viewer of a feed version",
				Subject:  azpb.NewEntityKey(UserType, "ian"),
				Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: ViewerRelation,
			},
			{
				Notes:    "a user can be a editor of a feed version",
				Subject:  azpb.NewEntityKey(UserType, "ian"),
				Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: EditorRelation,
			},
			{
				Notes:       "already exists",
				Subject:     azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:      azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				ExpectError: true,
			},
			{
				Notes:    "a tenant#member can be a viewer of a feed version",
				Subject:  azpb.NewEntityKey(TenantType, "tl-tenant#member"),
				Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: ViewerRelation,
			},
			{
				Notes:    "a tenant#member can be an editor of a feed version",
				Subject:  azpb.NewEntityKey(TenantType, "tl-tenant#member"),
				Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: EditorRelation,
				// Formerly disallowed, now OK
				// ExpectError: true,
			},
			{
				Notes:       "a tenant#admin can be a viewer of a feed version",
				Subject:     azpb.NewEntityKey(TenantType, "tl-tenant#admin"),
				Object:      azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				ExpectError: true,
			},
			{
				Notes:    "a group#member can be a viewer of a feed version",
				Subject:  azpb.NewEntityKey(TenantType, "HA-group#member"),
				Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: ViewerRelation,
			},
			{
				Notes:       "a group#editor cannot be a viewer of a feed version",
				Subject:     azpb.NewEntityKey(GroupType, "HA-group#editor"),
				Object:      azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    ViewerRelation,
				ExpectError: true,
			},
			{
				Notes:    "a group#viewer can be an editor of a feed version",
				Subject:  azpb.NewEntityKey(GroupType, "HA-group").WithRefRel(ViewerRelation),
				Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: EditorRelation,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating, so create fresh each test
				fgac := newTestFGAClient(t, fgaUrl, testData)
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
					gotTks = append(gotTks, fmt.Sprintf("%s:%s", v.Subject.String(), v.Relation))
				}
				checkTk := fmt.Sprintf("%s:%s", ltk.Subject.String(), ltk.Relation)
				assert.Contains(t, gotTks, checkTk, "written tuple not found in updated object tuples")
			})
		}
	})

	t.Run("DeleteTuple", func(t *testing.T) {
		checks := []testCase{
			{
				Subject:  azpb.NewEntityKey(UserType, "ian"),
				Object:   azpb.NewEntityKey(GroupType, "CT-group"),
				Relation: 4,
			},
			{
				Subject:     azpb.NewEntityKey(UserType, "test102"),
				Object:      azpb.NewEntityKey(GroupType, "100"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
			{
				Subject:  azpb.NewEntityKey(UserType, "tl-tenant-member"),
				Object:   azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation: 4,
			},
			{
				Subject:     azpb.NewEntityKey(UserType, "ian"),
				Object:      azpb.NewEntityKey(FeedVersionType, "e535eb2b3b9ac3ef15d82c56575e914575e732e0"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
			{
				Subject:  azpb.NewEntityKey(UserType, "test2"),
				Object:   azpb.NewEntityKey(TenantType, "restricted-tenant"),
				Relation: 2,
			},
			{
				Subject:     azpb.NewEntityKey(UserType, "test101"),
				Object:      azpb.NewEntityKey(GroupType, "BA-group"),
				Relation:    4,
				Notes:       "does not exist",
				ExpectError: true,
			},
			{
				Subject:     azpb.NewEntityKey(UserType, "test101"),
				Object:      azpb.NewEntityKey(GroupType, "BA-group"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test
				fgac := newTestFGAClient(t, fgaUrl, testData)
				ltk := tc.TupleKey()
				err := fgac.DeleteTuple(context.Background(), ltk)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
			})
		}
	})

	t.Run("SetExclusiveSubjectRelation", func(t *testing.T) {
		checks := []testCase{
			{
				Notes:    "changes ian permissions from Viewer to Manager",
				Subject:  azpb.NewEntityKey(UserType, "ian"),
				Object:   azpb.NewEntityKey(GroupType, "CT-group"),
				Relation: ManagerRelation,
				Expect:   "user:ian:manager user:drew:editor",
			},
			{
				Notes:    "changes drew permissions from Editor to Viewer",
				Subject:  azpb.NewEntityKey(UserType, "drew"),
				Object:   azpb.NewEntityKey(GroupType, "CT-group"),
				Relation: ViewerRelation,
				Expect:   "user:drew:viewer user:ian:viewer",
			},
			{
				Notes:    "assigns ian permissions as Manager, nothing to delete",
				Subject:  azpb.NewEntityKey(UserType, "ian"),
				Object:   azpb.NewEntityKey(GroupType, "HA-group"),
				Relation: ManagerRelation,
				Expect:   "user:ian:manager tenant:tl-tenant#member:viewer",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test
				fgac := newTestFGAClient(t, fgaUrl, testData)
				ltk := tc.TupleKey()
				checkRelTypes := []Relation{ViewerRelation, EditorRelation, ManagerRelation}
				err := fgac.SetExclusiveSubjectRelation(context.Background(), ltk, checkRelTypes...)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
				newTks, err := fgac.GetObjectTuples(context.Background(), azpb.NewTupleKey().WithObject(ltk.Object.Type, ltk.Object.Name))
				if err != nil {
					t.Error(err)
				}
				expect := strings.Split(tc.Expect, " ")
				var got []string
				for _, vtk := range newTks {
					ok := false
					for _, checkRel := range checkRelTypes {
						if vtk.Relation == checkRel {
							ok = true
						}
					}
					if !ok {
						continue
					}
					got = append(got, fmt.Sprintf("%s:%s", vtk.Subject.String(), vtk.Relation))
				}
				assert.ElementsMatch(t, expect, got, "usertype:username:relation does not match")
			})
		}
	})

	t.Run("SetExclusiveRelation", func(t *testing.T) {
		checks := []testCase{
			{
				Notes:    "changes feed parent",
				Object:   azpb.NewEntityKey(FeedType, "CT"),
				Subject:  azpb.NewEntityKey(GroupType, "BA-group"),
				Relation: ParentRelation,
				Expect:   "org:BA-group:parent",
			},
			{
				Notes:    "changes group tenant",
				Object:   azpb.NewEntityKey(GroupType, "CT-group"),
				Subject:  azpb.NewEntityKey(TenantType, "all-users-tenant"),
				Relation: ParentRelation,
				Expect:   "tenant:all-users-tenant:parent",
			},
			{
				Notes:    "assigns group to tenant",
				Object:   azpb.NewEntityKey(GroupType, "new-group"),
				Subject:  azpb.NewEntityKey(TenantType, "all-users-tenant"),
				Relation: ParentRelation,
				Expect:   "tenant:all-users-tenant:parent",
			},
		}
		for _, tc := range checks {
			t.Run(tc.String(), func(t *testing.T) {
				// Mutating test
				fgac := newTestFGAClient(t, fgaUrl, testData)
				ltk := tc.TupleKey()
				checkRelTypes := []Relation{ParentRelation}
				err := fgac.SetExclusiveRelation(context.Background(), ltk)
				if !checkExpectError(t, err, tc.ExpectError) {
					return
				}
				newTks, err := fgac.GetObjectTuples(context.Background(), azpb.NewTupleKey().WithObject(ltk.Object.Type, ltk.Object.Name))
				if err != nil {
					t.Error(err)
				}
				expect := strings.Split(tc.Expect, " ")
				var got []string
				for _, vtk := range newTks {
					ok := false
					for _, checkRel := range checkRelTypes {
						if vtk.Relation == checkRel {
							ok = true
						}
					}
					if !ok {
						continue
					}
					got = append(got, fmt.Sprintf("%s:%s", vtk.Subject.String(), vtk.Relation))
				}
				assert.ElementsMatch(t, expect, got, "usertype:username:relation does not match")
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

func newEntityKeys(t azpb.ObjectType, keys ...string) []azpb.EntityKey {
	var ret []azpb.EntityKey
	for _, k := range keys {
		ret = append(ret, azpb.NewEntityKey(t, k))
	}
	return ret
}

func newTestFGAClient(t testing.TB, url string, testTuples []testCase) *FGAClient {
	fgac, err := NewFGAClient(url, "", "")
	if err != nil {
		t.Fatal(err)
		return nil
	}
	if _, err := fgac.CreateStore(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := fgac.CreateModel(context.Background(), testutil.RelPath("test/authz/tls.json")); err != nil {
		t.Fatal(err)
	}
	for _, tk := range testTuples {
		if err := fgac.WriteTuple(context.Background(), tk.TupleKey()); err != nil {
			t.Fatal(err)
		}
	}
	return fgac
}
