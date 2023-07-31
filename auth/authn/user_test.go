package authn

import (
	"testing"

	"github.com/interline-io/transitland-server/model"
)

func TestUser_HasRole(t *testing.T) {
	testcases := []struct {
		name    string
		user    User
		role    model.Role
		hasRole bool
	}{
		{"anon", NewCtxUser("", "", ""), model.RoleAnon, true},
		{"anon", NewCtxUser("test", "", ""), model.RoleAnon, true},
		{"anon", NewCtxUser("test", "", "").WithRoles("admin"), model.RoleAnon, true},

		{"user", NewCtxUser("", "", ""), model.RoleUser, false},
		{"user", NewCtxUser("test", "", ""), model.RoleUser, true},
		{"user", NewCtxUser("test", "", "").WithRoles("admin"), model.RoleUser, true},

		{"admin", NewCtxUser("", "", ""), model.RoleAdmin, false},
		{"admin", NewCtxUser("", "", ""), model.RoleAdmin, false},
		{"admin", NewCtxUser("test", "", ""), model.RoleAdmin, false},
		{"admin", NewCtxUser("test", "", ""), model.RoleAdmin, false},
		{"admin", NewCtxUser("test", "", "").WithRoles("admin"), model.RoleAdmin, true},

		{"other roles", NewCtxUser("test", "", "").WithRoles("tlv2-admin"), model.Role("tlv2-admin"), true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.user.HasRole(string(tc.role)) != tc.hasRole {
				t.Errorf("expected role %s to be %t", tc.role, tc.hasRole)
			}
		})
	}
}
