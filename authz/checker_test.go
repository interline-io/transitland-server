package authz

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/stretchr/testify/assert"
)

var checkerCheck = []fgaTestTuple{
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
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(GroupType, "5"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "admin"),
		Object:        NewEntityKey(TenantType, "1"),
		ExpectActions: []Action{CanView, CanEdit, CanEditMembers, CanCreateOrg, CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "admin"),
		Object:      NewEntityKey(TenantType, "2"),
		ExpectError: true,
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
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedType, "4"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedType, "5"),
		ExpectError: true,
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
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "4"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "5"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "ian"),
		Object:        NewEntityKey(TenantType, "1"),
		ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(TenantType, "2"),
		ExpectError: true,
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
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(FeedType, "2"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(FeedType, "3"),
		ExpectActions: []Action{CanView, -CanEdit},
		Test:          "check",
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(FeedType, "4"),
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
		Object:        NewEntityKey(GroupType, "1"),
		ExpectActions: []Action{CanView, CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(GroupType, "2"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(GroupType, "3"),
		ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateFeed, -CanDeleteFeed},
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(GroupType, "4"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(GroupType, "5"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "drew"),
		Object:        NewEntityKey(TenantType, "1"),
		ExpectActions: []Action{CanView, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "drew"),
		Object:      NewEntityKey(TenantType, "2"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(TenantType, "1"),
		ExpectActions: []Action{CanView, -CanEdit},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(TenantType, "2"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(FeedVersionType, "1"),
		ExpectActions: []Action{CanView},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "1"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "2"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(FeedType, "3"),
		ExpectActions: []Action{CanView, -CanEdit},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "4"),
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedType, "5"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(GroupType, "3"),
		ExpectActions: []Action{CanView, -CanEdit},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(GroupType, "4"),
		ExpectError: true,
	},
	{
		Subject:       NewEntityKey(UserType, "nisar"),
		Object:        NewEntityKey(TenantType, "1"),
		ExpectActions: []Action{CanView, -CanEdit, -CanEditMembers, -CanCreateOrg, -CanDeleteOrg},
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(TenantType, "2"),
		ExpectError: true,
	},
}

var checkerListTests = []fgaTestTuple{
	{
		Subject:    NewEntityKey(UserType, "admin"),
		Object:     NewEntityKey(FeedType, ""),
		ListAction: CanView,
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
		ListAction: CanView,
		ExpectIds:  []int{1, 3},
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
		Subject:   NewEntityKey(UserType, "drew"),
		Object:    NewEntityKey(FeedVersionType, ""),
		ExpectIds: []int{4},
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

var checkerWriteTests = []fgaTestTuple{
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(GroupType, "3"),
		Relation:    ViewerRelation,
		Notes:       "does not exist",
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "3"),
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
		Object:      NewEntityKey(FeedVersionType, "1"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:  NewEntityKey(UserType, "nisar"),
		Object:   NewEntityKey(FeedVersionType, "1"),
		Relation: ViewerRelation,

		Expect:      "fail",
		Notes:       "already exists",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test1"),
		Object:      NewEntityKey(FeedVersionType, "1"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test2"),
		Object:      NewEntityKey(FeedVersionType, "1"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test3"),
		Object:      NewEntityKey(FeedVersionType, "1"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		CheckAsUser: "ian",
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(TenantType, "1"),
		Relation:    MemberRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(GroupType, "1"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test100"),
		Object:      NewEntityKey(GroupType, "1"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		ExpectError: true,
		CheckAsUser: "ian",
	},
}

var checkerDeleteTests = []fgaTestTuple{
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "1"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(GroupType, "1"),
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
		Object:      NewEntityKey(FeedVersionType, "1"),
		Relation:    ViewerRelation,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "nisar"),
		Object:      NewEntityKey(FeedVersionType, "1"),
		Relation:    ViewerRelation,
		Notes:       "already deleted",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "ian"),
		Object:      NewEntityKey(FeedVersionType, "1"),
		Relation:    ViewerRelation,
		Notes:       "unauthorized",
		ExpectError: true,
		CheckAsUser: "ian",
	},
	{
		Subject:     NewEntityKey(UserType, "test2"),
		Object:      NewEntityKey(TenantType, "2"),
		Relation:    MemberRelation,
		Notes:       "unauthorized",
		CheckAsUser: "admin",
		ExpectError: true,
	},
	{
		Subject:     NewEntityKey(UserType, "test101"),
		Object:      NewEntityKey(GroupType, "2"),
		Relation:    ViewerRelation,
		Notes:       "does not exist",
		ExpectError: true,
		CheckAsUser: "admin",
	},
	{
		Subject:     NewEntityKey(UserType, "test101"),
		Object:      NewEntityKey(GroupType, "2"),
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
		Object: NewEntityKey(TenantType, "1"),

		Expect: "user:admin:admin user:ian:member user:drew:member user:nisar:member",
	},
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(FeedVersionType, "1"),
		Expect: "feed:2:parent user:nisar:viewer",
	},
	{
		Subject: EntityKey{
			Type: 0,
			Name: "",
		},
		Object: NewEntityKey(FeedType, "1"),
		Expect: "org:1:parent",
	},
}

func TestChecker(t *testing.T) {
	if os.Getenv("TL_TEST_FGA_ENDPOINT") == "" {
		t.Skip("no TL_TEST_FGA_ENDPOINT set, skipping")
		return
	}

	// TENANTS
	t.Run("TenantList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerListTests {
			if tk.Object.Type != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.TenantList(context.Background(), newTestUser(tk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var retIds []int
				for _, v := range ret {
					retIds = append(retIds, v.ID)
				}
				assert.ElementsMatch(t, tk.ExpectIds, retIds, "tenant ids")
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerCheck {
			if tk.Object.Type != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.TenantPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("TenantAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerWriteTests {
			if tk.Object.Type != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.TenantAddPermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Object.ID(),
					tk.Subject.Name,
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
			})
		}
	})

	t.Run("TenantRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerDeleteTests {
			if tk.Object.Type != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.TenantRemovePermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Object.ID(),
					tk.Subject.Name,
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
			})
		}
	})

	// GROUPS
	t.Run("GroupList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerListTests {
			if tk.Object.Type != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.GroupList(context.Background(), newTestUser(tk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var retIds []int
				for _, v := range ret {
					retIds = append(retIds, v.ID)
				}
				assert.ElementsMatch(t, tk.ExpectIds, retIds, "tenant ids")
			})
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerCheck {
			if tk.Object.Type != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.GroupPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("GroupAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerWriteTests {
			if tk.Object.Type != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.GroupAddPermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Subject.Name,
					tk.Object.ID(),
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
			})
		}
	})

	t.Run("GroupRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerDeleteTests {
			if tk.Object.Type != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.GroupRemovePermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Subject.Name,
					tk.Object.ID(),
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
			})
		}
	})

	// FEEDS
	t.Run("FeedList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerListTests {
			if tk.Object.Type != FeedType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedList(context.Background(), newTestUser(tk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var retIds []int
				for _, v := range ret {
					retIds = append(retIds, v.ID)
				}
				assert.ElementsMatch(t, tk.ExpectIds, retIds, "tenant ids")
			})
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerCheck {
			if tk.Object.Type != FeedType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	// FEED VERSIONS
	t.Run("FeedVersionList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerListTests {
			if tk.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedVersionList(context.Background(), newTestUser(tk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var retIds []int
				for _, v := range ret {
					retIds = append(retIds, v.ID)
				}
				assert.ElementsMatch(t, tk.ExpectIds, retIds, "tenant ids")
			})
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerCheck {
			if tk.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedVersionPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("FeedVersionAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerWriteTests {
			if tk.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.FeedVersionAddPermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Subject.Name,
					tk.Object.ID(),
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
			})
		}
	})

	t.Run("FeedVersionRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checkerDeleteTests {
			if tk.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.FeedVersionRemovePermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Subject.Name,
					tk.Object.ID(),
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectError) {
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
	auth0c := NewMockAuthnClient()
	auth0c.AddUser("ian", User{Name: "Ian", ID: "ian", Email: "ian@example.com"})
	auth0c.AddUser("drew", User{Name: "Drew", ID: "drew", Email: "drew@example.com"})
	auth0c.AddUser("nisar", User{Name: "Nisar", ID: "nisar", Email: "nisar@example.com"})
	fgac := newTestFGAClient(t, te.Finder.DBX(), fgaTestData)
	checker := NewChecker(auth0c, fgac, te.Finder.DBX(), nil)
	return checker
}
