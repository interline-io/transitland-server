package fvsl

import (
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
)

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	os.Exit(m.Run())
}

func TestFVSLCache(t *testing.T) {
	te := testfinder.Finders(t, nil, nil)
	c := NewFVSLCache(te.Finder)
	c.Get(1)
}
