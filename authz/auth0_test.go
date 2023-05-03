package authz

import "testing"

func newTestAuth0Client(t testing.TB, cfg AuthzConfig) (*Auth0Client, error) {
	auth0c, err := NewAuth0Client(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)
	if err != nil {
		return nil, err
	}
	return auth0c, err
}

func TestAuth0Client(t *testing.T) {

}
