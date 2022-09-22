package workers

import (
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/gbfsfinder"
	"github.com/interline-io/transitland-server/internal/rtfinder"
	"github.com/interline-io/transitland-server/model"
)

var TestDBFinder model.Finder
var TestRTFinder model.RTFinder
var TestGbfsFinder model.GbfsFinder

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	db := find.MustOpenDB(g)
	dbf := find.NewDBFinder(db)
	TestDBFinder = dbf
	TestRTFinder = rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	TestGbfsFinder = gbfsfinder.NewFinder(nil)
	os.Exit(m.Run())
}
