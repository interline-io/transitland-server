package fvsl

import (
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
)

var TestDBFinder model.Finder

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	db := find.MustOpenDB(g)
	dbf := find.NewDBFinder(db)
	TestDBFinder = dbf
	os.Exit(m.Run())
}

func TestFVSLCache(t *testing.T) {
	c := FVSLCache{Finder: TestDBFinder}
	c.Get(1)
}
