package authz

import (
	"github.com/interline-io/transitland-server/auth"
)

type testUser struct {
	name string
}

func newTestUser(name string) *testUser {
	return &testUser{name: name}
}

func (u testUser) Name() string {
	return u.name
}

func (u testUser) GetExternalID(string) (string, bool) {
	return "test", true
}

func (u testUser) HasRole(string) bool { return true }

func (u testUser) IsValid() bool { return true }

func (u testUser) Roles() []string { return nil }

func (u testUser) WithExternalIDs(map[string]string) auth.User {
	return u
}

func (u testUser) WithRoles(...string) auth.User {
	return u
}
