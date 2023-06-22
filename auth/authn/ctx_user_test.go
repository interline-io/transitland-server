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
		{"anon", newCtxUser(""), model.RoleAnon, true},
		{"anon", newCtxUser("test"), model.RoleAnon, true},
		{"anon", newCtxUser("test").WithRoles("admin"), model.RoleAnon, true},

		{"user", newCtxUser(""), model.RoleUser, false},
		{"user", newCtxUser("test"), model.RoleUser, true},
		{"user", newCtxUser("test").WithRoles("admin"), model.RoleUser, true},

		{"admin", newCtxUser(""), model.RoleAdmin, false},
		{"admin", newCtxUser(""), model.RoleAdmin, false},
		{"admin", newCtxUser("test"), model.RoleAdmin, false},
		{"admin", newCtxUser("test"), model.RoleAdmin, false},
		{"admin", newCtxUser("test").WithRoles("admin"), model.RoleAdmin, true},

		{"other roles", newCtxUser("test").WithRoles("tlv2-admin"), model.Role("tlv2-admin"), true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.user.HasRole(string(tc.role)) != tc.hasRole {
				t.Errorf("expected role %s to be %t", tc.role, tc.hasRole)
			}
		})
	}
}
