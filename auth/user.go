package auth

import (
	"context"
	"strings"

	"github.com/interline-io/transitland-server/model"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// User defines role access methods.
type User struct {
	Name  string
	Roles []string
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func (user *User) WithRoles(roles ...string) *User {
	user.Roles = roles
	return user
}

func (user *User) IsAnon() bool {
	return user.HasRole("anon")
}

func (user *User) IsUser() bool {
	return user.HasRole("user")
}
func (user *User) IsAdmin() bool {
	return user.HasRole("admin")
}

// HasRole checks if a User is allowed to use a defined role.
func (user *User) HasRole(role model.Role) bool {
	checkRole := strings.ToLower(string(role))
	// Check for original roles
	switch checkRole {
	case "anon":
		return user.hasRole("anon") || user.hasRole("user") || user.hasRole("admin")
	case "user":
		return user.hasRole("user") || user.hasRole("admin")
	case "admin":
		return user.hasRole("admin")
	}
	// Check all other roles
	return user.hasRole(checkRole)
}

func (user *User) hasRole(checkRole string) bool {
	for _, r := range user.Roles {
		if r == checkRole {
			return true
		}
	}
	return false
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *User {
	raw, _ := ctx.Value(userCtxKey).(*User)
	return raw
}
