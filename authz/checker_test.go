package authz

import (
	"context"
	"encoding/json"
	"fmt"
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
			if tk.Test != "list" {
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
		// tcs := []listTest{
		// 	{
		// 		user:      "admin",
		// 		expectIds: []int{1, 2, 3, 4},
		// 	},
		// 	{
		// 		user:      "ian",
		// 		expectIds: []int{1, 2, 3},
		// 	},
		// 	{
		// 		user:      "nisar",
		// 		expectIds: []int{3},
		// 	},
		// 	{
		// 		user:      "drew",
		// 		expectIds: []int{1, 3},
		// 	},
		// }
		// for _, tc := range tcs {
		// 	t.Run(tc.user, func(t *testing.T) {
		// 		ret, err := checker.ListFeeds(context.Background(), newTestUser(tc.user))
		// 		if err != nil {
		// 			t.Fatal(err)
		// 			return
		// 		}
		// 		assert.ElementsMatch(t, tc.expectIds, ret, "feed ids")
		// 	})
		// }
	})

	t.Run("ListFeedVersions", func(t *testing.T) {
		tcs := []listTest{
			{
				user:      "ian",
				expectIds: []int{1},
			},
			{
				user:      "nisar",
				expectIds: []int{1},
			},
			{
				user:      "drew",
				expectIds: []int{},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.user, func(t *testing.T) {
				ret, err := checker.ListFeedVersions(context.Background(), newTestUser(tc.user))
				if err != nil {
					t.Fatal(err)
					return
				}
				assert.ElementsMatch(t, tc.expectIds, ret, "feed version ids")
			})
		}
	})

	t.Run("ListGroups", func(t *testing.T) {
		tcs := []listTest{
			{
				user:      "ian",
				expectIds: []int{1, 2, 3},
			},
			{
				user:      "nisar",
				expectIds: []int{3},
			},
			{
				user:      "drew",
				expectIds: []int{1, 3},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.user, func(t *testing.T) {
				ret, err := checker.ListGroups(context.Background(), newTestUser(tc.user))
				if err != nil {
					t.Fatal(err)
					return
				}
				assert.ElementsMatch(t, tc.expectIds, ret, "group ids")
			})
		}
	})

	t.Run("ListTenants", func(t *testing.T) {
		tcs := []listTest{
			{
				user:      "ian",
				expectIds: []int{1},
			},
			{
				user:      "nisar",
				expectIds: []int{1},
			},
			{
				user:      "drew",
				expectIds: []int{1},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.user, func(t *testing.T) {
				ret, err := checker.ListTenants(context.Background(), newTestUser(tc.user))
				if err != nil {
					t.Fatal(err)
					return
				}
				assert.ElementsMatch(t, tc.expectIds, ret, "tenant ids")
			})
		}
	})

	type permTest struct {
		user          string
		id            int
		expectActions map[string]bool
		expectError   bool
		rawJson       string
	}

	t.Run("TenantPermissions", func(t *testing.T) {
		tcs := []permTest{
			{
				user: "drew",
				id:   1,
				expectActions: map[string]bool{
					"can_create_org":   false,
					"can_delete_org":   false,
					"can_edit":         false,
					"can_edit_members": false,
					"can_view":         true,
				},
			},
			{
				user:    "ian",
				id:      1,
				rawJson: `{"id":1,"name":"","users":{"admins":[],"members":[{"id":"ian","name":"Ian","email":"ian@example.com"},{"id":"drew","name":"Drew","email":"drew@example.com"},{"id":"nisar","name":"Nisar","email":"nisar@example.com"}]},"actions":{"can_edit_members":false,"can_view":true,"can_edit":false,"can_create_org":false,"can_delete_org":false}}`,
			},
		}
		for _, tc := range tcs {
			t.Run(fmt.Sprintf("%s:%d", tc.user, tc.id), func(t *testing.T) {
				ret, err := checker.TenantPermissions(context.Background(), newTestUser(tc.user), tc.id)
				if err != nil && !tc.expectError {
					t.Errorf("got error '%s', did not expect error", err.Error())
				}
				if err == nil && tc.expectError {
					t.Errorf("got no error, expected error")
				}
				if err != nil {
					return
				}
				if len(tc.expectActions) > 0 {
					compareAsJson(t, tc.expectActions, ret.Actions)
				}
				if tc.rawJson != "" {
					jj2, _ := json.Marshal(ret)
					if !assert.JSONEq(t, tc.rawJson, string(jj2)) {
						t.Logf("actual raw json: %s", string(jj2))
					}
				}
			})
		}
	})

	t.Run("FeedPermissions", func(t *testing.T) {
		tcs := []permTest{
			{
				user:        "nisar",
				id:          1,
				expectError: true,
			},
			{
				user: "drew",
				id:   1,
				expectActions: map[string]bool{
					"can_create_feed_version": true,
					"can_delete_feed_version": true,
					"can_edit":                true,
					"can_view":                true,
				},
			},
			{
				user:    "ian",
				id:      1,
				rawJson: `{"id":1,"group":{"id":1,"name":"","tenant":{"id":1,"name":"","users":{"admins":[],"members":[{"id":"ian","name":"Ian","email":"ian@example.com"},{"id":"drew","name":"Drew","email":"drew@example.com"},{"id":"nisar","name":"Nisar","email":"nisar@example.com"}]},"actions":{"can_edit_members":false,"can_view":true,"can_edit":false,"can_create_org":false,"can_delete_org":false}},"users":{"viewers":[{"id":"ian","name":"Ian","email":"ian@example.com"}],"editors":[{"id":"drew","name":"Drew","email":"drew@example.com"}],"managers":[]},"actions":{"can_view":true,"can_edit_members":false,"can_create_feed":false,"can_delete_feed":false,"can_edit":false}},"users":{"viewers":[]},"actions":{"can_view":true,"can_edit":false,"can_create_feed_version":false,"can_delete_feed_version":false}}`,
			},
			{
				user: "ian",
				id:   1,
				expectActions: map[string]bool{
					"can_view":                true,
					"can_edit":                false,
					"can_create_feed_version": false,
					"can_delete_feed_version": false,
				},
			},
			{
				user: "ian",
				id:   2,
				expectActions: map[string]bool{
					"can_view":                true,
					"can_edit":                true,
					"can_create_feed_version": true,
					"can_delete_feed_version": true,
				},
			},
			{
				user: "ian",
				id:   3,
				expectActions: map[string]bool{
					"can_create_feed_version": false,
					"can_delete_feed_version": false,
					"can_edit":                false,
					"can_view":                true,
				},
			},
			{
				user:        "ian",
				id:          4,
				expectError: true,
			},
			{
				user:        "ian",
				id:          5,
				expectError: true,
			},
		}
		for _, tc := range tcs {
			t.Run(fmt.Sprintf("%s:%d", tc.user, tc.id), func(t *testing.T) {
				ret, err := checker.FeedPermissions(context.Background(), newTestUser(tc.user), tc.id)
				if err != nil && !tc.expectError {
					t.Errorf("got error '%s', did not expect error", err.Error())
				}
				if err == nil && tc.expectError {
					t.Errorf("got no error, expected error")
				}
				if err != nil {
					return
				}
				if len(tc.expectActions) > 0 {
					compareAsJson(t, tc.expectActions, ret.Actions)
				}
				if tc.rawJson != "" {
					jj2, _ := json.Marshal(ret)
					if !assert.JSONEq(t, tc.rawJson, string(jj2)) {
						t.Logf("actual raw json: %s", string(jj2))
					}
				}
			})
		}
	})
}

func compareAsJson(t testing.TB, j1 any, j2 any) {
	// Compare as json to make test simpler
	jj1, _ := json.Marshal(j1)
	jj2, _ := json.Marshal(j2)
	assert.JSONEq(t, string(jj1), string(jj2))
}

func mapStrInt(v string) []int {
	var ret []int
	for _, a := range strings.Split(v, " ") {
		ret = append(ret, atoi(a))
	}
	return ret
}
