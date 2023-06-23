package authz

import (
	"testing"

	"github.com/interline-io/transitland-server/internal/dbutil"
)

func TestAuth0Client(t *testing.T) {
	_, a, ok := dbutil.CheckEnv("TL_TEST_AUTH0_DOMAIN")
	if !ok {
		t.Skip(a)
		return
	}
}
