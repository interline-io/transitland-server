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
		FGAModelID:        os.Getenv("TL_FGA_MODEL_ID"),
		FGAEndpoint:       os.Getenv("TL_FGA_ENDPOINT"),
		FGATestModelPath:  os.Getenv("TL_FGA_TEST_MODEL_PATH"),
		FGATestTuplesPath: os.Getenv("TL_FGA_TEST_TUPLES_PATH"),
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
