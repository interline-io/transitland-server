package auth

import (
	"context"
	"strings"
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
	roles []string
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func (user *User) WithRoles(roles ...string) *User {
	user.AddRoles(roles...)
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

func (user *User) Merge(other *User) {
	user.Name = other.Name
	user.AddRoles(other.roles...)
}

// HasRole checks if a User is allowed to use a defined role.
func (user *User) HasRole(role string) bool {
	checkRole := strings.ToLower(role)
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
	for _, r := range user.roles {
		if r == checkRole {
			return true
		}
	}
	return false
}

func (user *User) AddRoles(roles ...string) {
	merged := map[string]bool{}
	for _, r := range user.roles {
		merged[r] = true
	}
	for _, r := range roles {
		merged[r] = true
	}
	var rr []string
	for k := range merged {
		rr = append(rr, k)
	}
	user.roles = rr
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *User {
	raw, _ := ctx.Value(userCtxKey).(*User)
	return raw
}
