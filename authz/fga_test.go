package authz

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	ListAction        Action
}

func (tk *fgaTestTuple) TupleKey() TupleKey {
	return TupleKey{Subject: tk.Subject, Object: tk.Object, Relation: tk.Relation, Action: tk.Action}
}

func (tk *fgaTestTuple) String() string {
	return tk.TupleKey().String() + "|checkuser:" + tk.CheckAsUser
}

func TestFGAClient(t *testing.T) {
	if os.Getenv("TL_TEST_FGA_ENDPOINT") == "" {
		t.Skip("no TL_TEST_FGA_ENDPOINT set, skipping")
		return
	}

	t.Run("GetObjectTuples", func(t *testing.T) {
		fgac := newTestFGAClient(t, fgaTestData)
		checks := []fgaTestTuple{
			{
				Object: NewEntityKey(TenantType, "1"),
				Expect: "user:admin:admin user:ian:member user:drew:member user:nisar:member",
			},
			{
				Object: NewEntityKey(FeedVersionType, "1"),
				Expect: "feed:2:parent user:nisar:viewer",
			},
			{
				Object: NewEntityKey(FeedType, "1"),
				Expect: "org:1:parent",
			},
		}
		for _, tk := range checks {
			t.Run(tk.String(), func(t *testing.T) {
				tks, err := fgac.GetObjectTuples(context.Background(), tk.TupleKey())
				if err != nil {
					t.Error(err)
				}
				expect := strings.Split(tk.Expect, " ")
				var got []string
				for _, v := range tks {
					got = append(got, fmt.Sprintf("%s:%s:%s", v.Subject.Type, v.Subject.Name, v.Relation))
				}
				assert.ElementsMatch(t, expect, got, "usertype:username:relation does not match")

			})
		}
	})

	t.Run("Check", func(t *testing.T) {
		fgac := newTestFGAClient(t, fgaTestData)
		checks := []fgaTestTuple{
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(FeedVersionType, "1"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(GroupType, "1"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(GroupType, "2"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(GroupType, "3"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(GroupType, "4"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateFeed, CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(GroupType, "5"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(TenantType, "1"),
				ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "admin"),
				Object:        NewEntityKey(TenantType, "2"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedVersionType, "1"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "1"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "2"),
				ExpectActions: []Action{CanView, CanEdit, CanCreateFeedVersion, CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "3"),
				ExpectActions: []Action{CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "4"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(FeedType, "5"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanCreateFeedVersion, -CanDeleteFeedVersion},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "1"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "2"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "3"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
				Test:          "check",
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "4"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(GroupType, "5"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "1"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "ian"),
				Object:        NewEntityKey(TenantType, "2"),
				ExpectActions: []Action{-CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedVersionType, "1"),
				ExpectActions: []Action{-CanView},
				Notes:         "only feed:2 readers and nisar",
				ExpectError:   true,
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "1"),
				ExpectActions: []Action{CanView, CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "2"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "3"),
				ExpectActions: []Action{CanView, -CanEdit},
				Test:          "check",
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "4"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(FeedType, "5"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "1"),
				ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "2"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "3"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "4"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(GroupType, "5"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(TenantType, "1"),
				ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "drew"),
				Object:        NewEntityKey(TenantType, "2"),
				ExpectActions: []Action{-CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "1"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "2"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedVersionType, "1"),
				ExpectActions: []Action{CanView},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "1"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "2"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "3"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "4"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(FeedType, "5"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(GroupType, "3"),
				ExpectActions: []Action{CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(GroupType, "4"),
				ExpectActions: []Action{-CanView, -CanEdit},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "1"),
				ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
			{
				Subject:       NewEntityKey(UserType, "nisar"),
				Object:        NewEntityKey(TenantType, "2"),
				ExpectActions: []Action{-CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
			},
		}
		for _, tk := range checks {
			t.Run(tk.String(), func(t *testing.T) {
				for _, checkAction := range tk.ExpectActions {
					expect := true
					if checkAction < 0 {
						expect = false
						checkAction = checkAction * -1
					}
					var err error
					tk := tk
					tk.Action = checkAction
					ok, err := fgac.Check(context.Background(), tk.TupleKey())
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
		fgac := newTestFGAClient(t, fgaTestData)
		checks := []fgaTestTuple{
			{
				Subject:    NewEntityKey(UserType, "admin"),
				Object:     NewEntityKey(FeedType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1, 2, 3, 4},
			},
			{
				Subject:    NewEntityKey(UserType, "admin"),
				Object:     NewEntityKey(FeedType, ""),
				ListAction: CanEdit,
				ExpectIds:  []int{1, 2, 3, 4},
			},

			{
				Subject:    NewEntityKey(UserType, "admin"),
				Object:     NewEntityKey(GroupType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1, 2, 3, 4},
			},
			{
				Subject:    NewEntityKey(UserType, "admin"),
				Object:     NewEntityKey(GroupType, ""),
				ListAction: CanEdit,
				ExpectIds:  []int{1, 2, 3, 4},
			},

			{
				Subject:    NewEntityKey(UserType, "admin"),
				Object:     NewEntityKey(TenantType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1},
			},
			{
				Subject:    NewEntityKey(UserType, "admin"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1},
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(FeedType, ""),
				ListAction: CanEdit,
				ExpectIds:  []int{2},
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(FeedType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1, 2, 3},
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(GroupType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1, 2, 3},
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(TenantType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1},
			},
			{
				Subject:    NewEntityKey(UserType, "ian"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1},
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(FeedType, ""),
				ListAction: CanEdit,
				ExpectIds:  []int{1},
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(FeedType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1, 3},
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(TenantType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1},
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(GroupType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1, 3},
			},
			{
				Subject:    NewEntityKey(UserType, "drew"),
				Object:     NewEntityKey(FeedVersionType, ""),
				ListAction: CanView,
				ExpectIds:  []int{},
			},
			{
				Subject:    NewEntityKey(UserType, "nisar"),
				Object:     NewEntityKey(GroupType, ""),
				ListAction: CanView,
				ExpectIds:  []int{3},
			},
			{
				Subject:    NewEntityKey(UserType, "nisar"),
				Object:     NewEntityKey(FeedType, ""),
				ListAction: CanView,
				ExpectIds:  []int{3},
			},
			{
				Subject:    NewEntityKey(UserType, "nisar"),
				Object:     NewEntityKey(TenantType, ""),
				ListAction: CanView,
				ExpectIds:  []int{1},
			},
		}
		for _, tk := range checks {
			t.Run(tk.String(), func(t *testing.T) {
				tk := tk
				tk.Action = tk.ListAction
				objs, err := fgac.ListObjects(context.Background(), tk.TupleKey())
				if err != nil {
					t.Fatal(err)
				}
				var gotIds []int
				for _, v := range objs {
					gotIds = append(gotIds, v.Object.ID())
				}
				assert.ElementsMatch(t, tk.ExpectIds, gotIds, "object ids")
			})
		}
	})

	t.Run("WriteTuple", func(t *testing.T) {
		fgac := newTestFGAClient(t, fgaTestData)
		checks := []fgaTestTuple{
			{
				Subject:  NewEntityKey(UserType, "test100"),
				Object:   NewEntityKey(TenantType, "1"),
				Relation: MemberRelation,
			},
			{
				Subject:  NewEntityKey(UserType, "test100"),
				Object:   NewEntityKey(GroupType, "1"),
				Relation: ViewerRelation,
			},
			{
				Subject:     NewEntityKey(UserType, "test100"),
				Object:      NewEntityKey(GroupType, "1"),
				Relation:    ViewerRelation,
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "test100"),
				Object:   NewEntityKey(GroupType, "3"),
				Relation: ViewerRelation,
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "3"),
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
				Object:   NewEntityKey(FeedVersionType, "1"),
				Relation: ViewerRelation,
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(FeedVersionType, "1"),
				Relation:    ViewerRelation,
				Notes:       "already exists",
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "test1"),
				Object:   NewEntityKey(FeedVersionType, "1"),
				Relation: ViewerRelation,
			},
			{
				Subject:  NewEntityKey(UserType, "test2"),
				Object:   NewEntityKey(FeedVersionType, "1"),
				Relation: ViewerRelation,
			},
			{
				Subject:  NewEntityKey(UserType, "test3"),
				Object:   NewEntityKey(FeedVersionType, "1"),
				Relation: ViewerRelation,
			},
		}
		for _, tk := range checks {
			t.Run(tk.String(), func(t *testing.T) {
				// Write tuple and check if error was expected
				err := fgac.WriteTuple(context.Background(), tk.TupleKey())
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
				// Check was written
				tks, err := fgac.GetObjectTuples(context.Background(), tk.TupleKey())
				if err != nil {
					t.Error(err)
				}
				var gotTks []string
				for _, v := range tks {
					gotTks = append(gotTks, fmt.Sprintf("%s:%s:%s", v.Subject.Type, v.Subject.Name, v.Relation))
				}
				checkTk := fmt.Sprintf("%s:%s:%s", tk.Subject.Type, tk.Subject.Name, tk.Relation)
				assert.Contains(t, gotTks, checkTk, "written tuple not found in updated object tuples")
			})
		}
	})

	t.Run("DeleteTuple", func(t *testing.T) {
		fgac := newTestFGAClient(t, fgaTestData)
		checks := []fgaTestTuple{
			{
				Subject:  NewEntityKey(UserType, "ian"),
				Object:   NewEntityKey(GroupType, "1"),
				Relation: 4,
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(GroupType, "1"),
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
				Object:   NewEntityKey(FeedVersionType, "1"),
				Relation: 4,
			},
			{
				Subject:     NewEntityKey(UserType, "nisar"),
				Object:      NewEntityKey(FeedVersionType, "1"),
				Relation:    4,
				Notes:       "already deleted",
				ExpectError: true,
			},
			{
				Subject:     NewEntityKey(UserType, "ian"),
				Object:      NewEntityKey(FeedVersionType, "1"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
			{
				Subject:  NewEntityKey(UserType, "test2"),
				Object:   NewEntityKey(TenantType, "2"),
				Relation: 2,
			},
			{
				Subject:     NewEntityKey(UserType, "test101"),
				Object:      NewEntityKey(GroupType, "2"),
				Relation:    4,
				Notes:       "does not exist",
				ExpectError: true,
			},
			{
				Subject:     NewEntityKey(UserType, "test101"),
				Object:      NewEntityKey(GroupType, "2"),
				Relation:    4,
				Notes:       "unauthorized",
				ExpectError: true,
			},
		}
		for _, tk := range checks {
			t.Run(tk.String(), func(t *testing.T) {
				err := fgac.DeleteTuple(context.Background(), tk.TupleKey())
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
			})
		}
	})
}

func newTestFGAClient(t testing.TB, testTuples []fgaTestTuple) *FGAClient {
	cfg := newTestConfig()
	fgac, err := NewFGAClient(cfg.FGAEndpoint, cfg.FGAStoreID, cfg.FGAModelID)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	if cfg.FGALoadModelFile != "" {
		if _, err := fgac.CreateStore(context.Background(), "test"); err != nil {
			t.Fatal(err)
		}
		if _, err := fgac.CreateModel(context.Background(), cfg.FGALoadModelFile); err != nil {
			t.Fatal(err)
		}
	}
	for _, tk := range testTuples {
		if err := fgac.WriteTuple(context.Background(), tk.TupleKey()); err != nil {
			t.Fatal(err)
		}
	}
	return fgac
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
