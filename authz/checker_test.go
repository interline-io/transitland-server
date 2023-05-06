package authz

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
)

func newTestChecker(t testing.TB, cfg AuthzConfig, finder model.Finder) (*Checker, error) {
	auth0c := NewMockAuthnClient()
	fgac, err := newTestFGAClient(t, cfg)
	if err != nil {
		return nil, err
	}
	checker := NewChecker(auth0c, fgac, finder, nil)
	return checker, err
}

func TestChecker(t *testing.T) {
	te := testfinder.Finders(t, nil, nil)
	cfg := newTestConfig()
	cfg.FGAEndpoint = "http://localhost:8090"
	cfg.FGATestModelPath = "../test/authz/tls.model"
	cfg.FGATestTuplesPath = "../test/authz/tls.csv"
	checker, err := newTestChecker(t, cfg, te.Finder)
	if err != nil {
		t.Fatal(err)
	}

	// Test assertions
	checks, err := LoadTuples("../test/authz/tls.csv")
	if err != nil {
		t.Fatal(err)
	}

	type listTest struct {
		user      string
		expectIds []int
	}

	t.Run("ListFeeds", func(t *testing.T) {
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

	t.Run("ListGroups", func(t *testing.T) {
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
					ret, err := checker.ListGroups(context.Background(), newTestUser(tk.UserName))
					if err != nil {
						t.Fatal(err)
					}
					assert.ElementsMatch(t, mapStrInt(tk.Expect), ret, "feed ids")
				})
			}
		}
	})

	t.Run("ListTenants", func(t *testing.T) {
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
					ret, err := checker.ListTenants(context.Background(), newTestUser(tk.UserName))
					if err != nil {
						t.Fatal(err)
					}
					assert.ElementsMatch(t, mapStrInt(tk.Expect), ret, "feed ids")
				})
			}
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
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
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("FeedVersionPermissions", func(t *testing.T) {
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
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("GroupPermissions", func(t *testing.T) {
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
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("TenantPermissions", func(t *testing.T) {
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
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
				checkActionSubset(t, ret.Actions, tk.Checks)
			})
		}
	})

	t.Run("AddFeedPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "write" || tk.ObjectType != FeedType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.AddFeedVersionPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
			})
		}
	})

	t.Run("RemoveFeedPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "delete" || tk.ObjectType != FeedType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.RemoveFeedVersionPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
			})
		}
	})

	t.Run("AddFeedVersionPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "write" || tk.ObjectType != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.AddFeedVersionPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
			})
		}
	})

	t.Run("RemoveFeedVersionPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "delete" || tk.ObjectType != FeedVersionType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.RemoveFeedVersionPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
			})
		}
	})

	t.Run("AddTenantPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "write" || tk.ObjectType != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.AddTenantPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
			})
		}
	})

	t.Run("RemoveTenantPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "delete" || tk.ObjectType != TenantType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.RemoveTenantPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
			})
		}
	})

	t.Run("AddGroupPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "write" || tk.ObjectType != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.AddGroupPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
					return
				}
			})
		}
	})

	t.Run("RemoveGroupPermission", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "delete" || tk.ObjectType != GroupType {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := checker.RemoveGroupPermission(
					context.Background(),
					newTestUser(stringOr(tk.TestAsUser, tk.UserName)),
					tk.UserName,
					atoi(tk.ObjectName),
					tk.Relation,
				)
				checkExpectError(t, err, tk.ExpectError)
				if err != nil {
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

func compareAsJson(t testing.TB, j1 any, j2 any) {
	// Compare as json to make test simpler
	jj1, _ := json.Marshal(j1)
	jj2, _ := json.Marshal(j2)
	assert.JSONEq(t, string(jj1), string(jj2))
}

func checkExpectError(t testing.TB, err error, expect bool) {
	if err != nil && !expect {
		t.Errorf("got error '%s', did not expect error", err.Error())
	}
	if err == nil && expect {
		t.Errorf("got no error, expected error")
	}
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
