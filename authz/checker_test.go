package authz

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/stretchr/testify/assert"
)

func TestChecker(t *testing.T) {
	// Test assertions
	checks, err := LoadTuples("../test/authz/tls.csv")
	if err != nil {
		t.Fatal(err)
	}

	// TENANTS
	t.Run("TenantList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "list", TenantType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.TenantList(context.Background(), newTestUser(tk.Subject.Name))
				if err != nil {
					t.Fatal(err)
				}
				var retIds []int
				for _, v := range ret {
					retIds = append(retIds, v.ID)
				}
				assert.ElementsMatch(t, mapStrInt(tk.Expect), retIds, "tenant ids")
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "check", TenantType, CanView) {
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.TenantPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectErrorAsUser)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("TenantAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "write", TenantType, 0) {
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.TenantAddPermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Object.ID(),
					tk.Subject.Name,
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
					return
				}
			})
		}
	})

	t.Run("TenantRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range filterTestTuple(checks, "delete", TenantType, 0) {
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.TenantRemovePermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.Subject.Name)),
					tk.Object.ID(),
					tk.Subject.Name,
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
					return
				}
			})
		}
	})

	// GROUPS
	t.Run("GroupList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "list" || tk.Object.Type != GroupType {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk.Action, _ = ActionString(checkAction)
				if tk.Action.String() != "can_view" {
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
					assert.ElementsMatch(t, mapStrInt(tk.Expect), retIds, "group ids")
				})
			}
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "check" || tk.Object.Type != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.GroupPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectErrorAsUser)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("GroupAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "write" || tk.Object.Type != GroupType {
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
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
					return
				}
			})
		}
	})

	t.Run("GroupRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "delete" || tk.Object.Type != GroupType {
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
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
					return
				}
			})
		}
	})

	// FEEDS
	t.Run("FeedList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "list" || tk.Object.Type != FeedType {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk.Action, _ = ActionString(checkAction)
				if tk.Action.String() != "can_view" {
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
					assert.ElementsMatch(t, mapStrInt(tk.Expect), retIds, "feed ids")
				})
			}
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "check" || tk.Object.Type != FeedType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectErrorAsUser)
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
		for _, tk := range checks {
			if tk.Test != "list" || tk.Object.Type != FeedVersionType {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk.Action, _ = ActionString(checkAction)
				if tk.Action.String() != "can_view" {
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
					assert.ElementsMatch(t, mapStrInt(tk.Expect), retIds, "feed version ids")
				})
			}
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "check" || tk.Object.Type != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedVersionPermissions(
					context.Background(),
					newTestUser(tk.Subject.Name),
					tk.Object.ID(),
				)
				checkExpectError(t, err, tk.ExpectErrorAsUser)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("FeedVersionAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "write" || tk.Object.Type != FeedVersionType {
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
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
					return
				}
			})
		}
	})

	t.Run("FeedVersionRemovePermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "delete" || tk.Object.Type != FeedVersionType {
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
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
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

func mapStrInt(v string) []int {
	var ret []int
	for _, a := range strings.Split(v, " ") {
		if a == "" {
			continue
		}
		ai, _ := strconv.Atoi(a)
		ret = append(ret, ai)
	}
	return ret
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
	fgac := newTestFGAClient(t)
	checker := NewChecker(auth0c, fgac, te.Finder, nil)
	return checker
}
