package authz

import (
	"context"
	"encoding/json"
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

	t.Run("ListFeeds", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "list" || tk.ObjectType != FeedType {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk.Action, _ = ActionString(checkAction)
				if tk.Action.String() != "can_view" {
					continue
				}
				t.Run(tk.String(), func(t *testing.T) {
					ret, err := checker.ListFeeds(context.Background(), newTestUser(tk.UserName))
					if err != nil {
						t.Fatal(err)
					}
					assert.ElementsMatch(t, mapStrInt(tk.Expect), ret, "feed ids")
				})
			}
		}
	})

	t.Run("ListFeedVersions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "list" || tk.ObjectType != FeedVersionType {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk.Action, _ = ActionString(checkAction)
				if tk.Action.String() != "can_view" {
					continue
				}
				t.Run(tk.String(), func(t *testing.T) {
					ret, err := checker.ListFeedVersions(context.Background(), newTestUser(tk.UserName))
					if err != nil {
						t.Fatal(err)
					}
					assert.ElementsMatch(t, mapStrInt(tk.Expect), ret, "feed version ids")
				})
			}
		}
	})

	t.Run("GroupList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "list" || tk.ObjectType != GroupType {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk.Action, _ = ActionString(checkAction)
				if tk.Action.String() != "can_view" {
					continue
				}
				t.Run(tk.String(), func(t *testing.T) {
					ret, err := checker.GroupList(context.Background(), newTestUser(tk.UserName))
					if err != nil {
						t.Fatal(err)
					}
					assert.ElementsMatch(t, mapStrInt(tk.Expect), ret, "group ids")
				})
			}
		}
	})

	t.Run("TenantList", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "list" || tk.ObjectType != TenantType {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk.Action, _ = ActionString(checkAction)
				if tk.Action.String() != "can_view" {
					continue
				}
				t.Run(tk.String(), func(t *testing.T) {
					ret, err := checker.TenantList(context.Background(), newTestUser(tk.UserName))
					if err != nil {
						t.Fatal(err)
					}
					assert.ElementsMatch(t, mapStrInt(tk.Expect), ret, "tenant ids")
				})
			}
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "check" || tk.ObjectType != FeedType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedPermissions(
					context.Background(),
					newTestUser(tk.UserName),
					atoi(tk.ObjectName),
				)
				checkExpectError(t, err, tk.ExpectErrorAsUser)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "check" || tk.ObjectType != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.FeedVersionPermissions(
					context.Background(),
					newTestUser(tk.UserName),
					atoi(tk.ObjectName),
				)
				checkExpectError(t, err, tk.ExpectErrorAsUser)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "check" || tk.ObjectType != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.GroupPermissions(
					context.Background(),
					newTestUser(tk.UserName),
					atoi(tk.ObjectName),
				)
				checkExpectError(t, err, tk.ExpectErrorAsUser)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "check" || tk.ObjectType != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				ret, err := checker.TenantPermissions(
					context.Background(),
					newTestUser(tk.UserName),
					atoi(tk.ObjectName),
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
			if tk.Test != "write" || tk.ObjectType != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.FeedVersionAddPermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
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
			if tk.Test != "delete" || tk.ObjectType != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.FeedVersionRemovePermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
					return
				}
			})
		}
	})

	t.Run("TenantAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "write" || tk.ObjectType != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.TenantAddPermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.UserName)),
					atoi(tk.ObjectName),
					tk.UserName,
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
		for _, tk := range checks {
			if tk.Test != "delete" || tk.ObjectType != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.TenantRemovePermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.UserName)),
					atoi(tk.ObjectName),
					tk.UserName,
					tk.Relation,
				)
				if !checkExpectError(t, err, tk.ExpectErrorAsUser) {
					return
				}
			})
		}
	})

	t.Run("GroupAddPermission", func(t *testing.T) {
		checker := newTestChecker(t)
		for _, tk := range checks {
			if tk.Test != "write" || tk.ObjectType != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.GroupAddPermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
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
			if tk.Test != "delete" || tk.ObjectType != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.GroupRemovePermission(
					context.Background(),
					newTestUser(stringOr(tk.CheckAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
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
		ret = append(ret, atoi(a))
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
