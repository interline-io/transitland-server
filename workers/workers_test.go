package workers

import (
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/internal/dbutil"
)

func TestMain(m *testing.M) {
	if a, ok := dbutil.CheckTestDB(); !ok {
		log.Print(a)
		return
	}
	os.Exit(m.Run())
}
