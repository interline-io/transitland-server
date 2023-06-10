package authz

import (
	"os"
	"testing"
)

func TestAuth0Client(t *testing.T) {
	if os.Getenv("TL_TEST_AUTH0_DOMAIN") == "" {
		t.Skip("no TL_TEST_AUTH0_DOMAIN set, skipping")
		return
	}
}
