package auth

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
		{"anon", newCtxUser("").WithRoles("anon"), model.RoleAnon, true},
		{"anon", newCtxUser("").WithRoles("anon", "user"), model.RoleAnon, true},
		{"anon", newCtxUser("").WithRoles("anon", "user", "admin"), model.RoleAnon, true},
		{"anon", newCtxUser("").WithRoles("user"), model.RoleAnon, true},
		{"anon", newCtxUser("").WithRoles("admin"), model.RoleAnon, true},

		{"user", newCtxUser("").WithRoles("anon"), model.RoleUser, false},
		{"user", newCtxUser("").WithRoles("user"), model.RoleUser, true},
		{"user", newCtxUser("").WithRoles("admin"), model.RoleUser, true},

		{"admin", newCtxUser("").WithRoles("anon"), model.RoleAdmin, false},
		{"admin", newCtxUser("").WithRoles("anon", "user"), model.RoleAdmin, false},
		{"admin", newCtxUser("").WithRoles("anon"), model.RoleAdmin, false},
		{"admin", newCtxUser("").WithRoles("user"), model.RoleAdmin, false},
		{"admin", newCtxUser("").WithRoles("admin"), model.RoleAdmin, true},

		{"other roles", newCtxUser("").WithRoles("tlv2-admin"), model.Role("tlv2-admin"), true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.user.HasRole(string(tc.role)) != tc.hasRole {
				t.Errorf("expected role %s to be %t", tc.role, tc.hasRole)
			}
		})
	}
}
