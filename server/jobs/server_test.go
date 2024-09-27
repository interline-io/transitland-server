package jobs

import (
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-dbutil/testutil"
)

func TestMain(m *testing.M) {
	if a, ok := testutil.CheckTestDB(); !ok {
		log.Print(a)
		return
	}
	os.Exit(m.Run())
}
