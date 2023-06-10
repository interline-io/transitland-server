package authz

import (
	"os"

	"github.com/interline-io/transitland-server/auth"
)

func newTestConfig() AuthzConfig {
	cfg := AuthzConfig{
		Auth0Domain:       os.Getenv("TL_AUTH0_DOMAIN"),
		Auth0ClientID:     os.Getenv("TL_AUTH0_CLIENT_ID"),
		Auth0ClientSecret: os.Getenv("TL_AUTH0_CLIENT_SECRET"),
		FGAEndpoint:       "http://localhost:8090", // os.Getenv("TL_FGA_ENDPOINT"),
		FGALoadModelFile:  "../test/authz/tls.json",
		FGALoadTupleFile:  "../test/authz/tls.csv",
	}
	return cfg
}

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
