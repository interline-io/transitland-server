package auth

import (
	"context"
	"strings"
)

// ctxUser is the base user implementation.
type ctxUser struct {
	name        string
	valid       bool
	roles       map[string]bool
	externalIds map[string]string
}

func newCtxUser(name string) *ctxUser {
	return newCtxUserWith(name, nil, nil)
}

func newCtxUserWith(name string, roles map[string]bool, externalIds map[string]string) *ctxUser {
	u := ctxUser{
		name:        name,
		valid:       true,
		roles:       map[string]bool{},
		externalIds: map[string]string{},
	}
	for k, v := range roles {
		u.roles[k] = v
	}
	for k, v := range externalIds {
		u.externalIds[k] = v
	}
	return &u
}

func (user ctxUser) clone() *ctxUser {
	return newCtxUserWith(user.name, user.roles, user.externalIds)
}

func (user ctxUser) Name() string {
	return user.name
}

func (user ctxUser) IsValid() bool {
	return user.valid
}

func (user ctxUser) GetExternalID(eid string) (string, bool) {
	a, ok := user.externalIds[eid]
	return a, ok
}

func (user ctxUser) WithExternalIDs(m map[string]string) User {
	newUser := user.clone()
	for k, v := range m {
		newUser.externalIds[k] = v
	}
	return newUser
}

func (user ctxUser) WithRoles(roles ...string) User {
	newUser := user.clone()
	for _, v := range roles {
		newUser.roles[v] = true
	}
	return newUser
}

// HasRole checks if a User is allowed to use a defined role.
func (user ctxUser) HasRole(role string) bool {
	checkRole := strings.ToLower(role)
	// Check for original roles
	switch checkRole {
	case "anon":
		return true
	case "user":
		return user.name != ""
	}
	// Check all other roles
	return user.hasRole(checkRole)
}

func (user ctxUser) Roles() []string {
	var keys []string
	for k := range user.roles {
		keys = append(keys, k)
	}
	return keys
}

func (user ctxUser) hasRole(checkRole string) bool {
	return user.roles[checkRole]
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) User {
	raw, _ := ctx.Value(userCtxKey).(User)
	return raw
}
