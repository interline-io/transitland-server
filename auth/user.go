package auth

import (
	"strings"
)

// User is the base user implementation.
type CtxUser struct {
	id          string
	name        string
	email       string
	valid       bool
	roles       map[string]bool
	externalIds map[string]string
}

func NewCtxUser(id string, name string, email string) CtxUser {
	a := newCtxUserWith(id, name, email, nil, nil)
	return a
}

func newCtxUserWith(id string, name string, email string, roles map[string]bool, externalIds map[string]string) CtxUser {
	u := CtxUser{
		id:          id,
		name:        name,
		email:       email,
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
	return u
}

func (user CtxUser) clone() CtxUser {
	return newCtxUserWith(user.id, user.name, user.email, user.roles, user.externalIds)
}

func (user CtxUser) Name() string {
	return user.name
}

func (user CtxUser) ID() string {
	return user.id
}

func (user CtxUser) Email() string {
	return user.email
}

func (user CtxUser) IsValid() bool {
	return user.valid
}

func (user CtxUser) GetExternalID(eid string) (string, bool) {
	a, ok := user.externalIds[eid]
	return a, ok
}

func (user CtxUser) WithExternalIDs(m map[string]string) CtxUser {
	newUser := user.clone()
	for k, v := range m {
		newUser.externalIds[k] = v
	}
	return newUser
}

func (user CtxUser) WithRoles(roles ...string) CtxUser {
	newUser := user.clone()
	for _, v := range roles {
		newUser.roles[v] = true
	}
	return newUser
}

// HasRole checks if a User is allowed to use a defined role.
func (user CtxUser) HasRole(role string) bool {
	if user.hasRole("admin") {
		return true
	}
	checkRole := strings.ToLower(role)
	// Check for original roles
	switch checkRole {
	case "anon":
		return true
	case "user":
		return user.id != ""
	}
	// Check all other roles
	return user.hasRole(checkRole)
}

func (user CtxUser) hasRole(checkRole string) bool {
	return user.roles[checkRole]
}

func (user CtxUser) Roles() []string {
	var keys []string
	for k := range user.roles {
		keys = append(keys, k)
	}
	return keys
}
