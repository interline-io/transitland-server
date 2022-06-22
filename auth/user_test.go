package auth

import (
	"testing"

	"github.com/interline-io/transitland-server/model"
)

func TestUser_HasRole(t *testing.T) {
	testcases := []struct {
		name    string
		user    *User
		role    model.Role
		hasRole bool
	}{
		{"anon", NewUser("").WithRoles("anon"), model.RoleAnon, true},
		{"anon", NewUser("").WithRoles("anon", "user"), model.RoleAnon, true},
		{"anon", NewUser("").WithRoles("anon", "user", "admin"), model.RoleAnon, true},
		{"anon", NewUser("").WithRoles("user"), model.RoleAnon, true},
		{"anon", NewUser("").WithRoles("admin"), model.RoleAnon, true},

		{"user", NewUser("").WithRoles("anon"), model.RoleUser, false},
		{"user", NewUser("").WithRoles("user"), model.RoleUser, true},
		{"user", NewUser("").WithRoles("admin"), model.RoleUser, true},

		{"admin", NewUser("").WithRoles("anon"), model.RoleAdmin, false},
		{"admin", NewUser("").WithRoles("anon", "user"), model.RoleAdmin, false},
		{"admin", NewUser("").WithRoles("anon"), model.RoleAdmin, false},
		{"admin", NewUser("").WithRoles("user"), model.RoleAdmin, false},
		{"admin", NewUser("").WithRoles("admin"), model.RoleAdmin, true},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.user.HasRole(string(tc.role)) != tc.hasRole {
				t.Errorf("expected role %s to be %t", tc.role, tc.hasRole)
			}
		})
	}
}
