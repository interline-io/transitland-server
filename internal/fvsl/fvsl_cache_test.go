package fvsl

import (
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/internal/testutil"
)

func TestMain(m *testing.M) {
	if a, ok := testutil.CheckTestDB(); !ok {
		log.Print(a)
		return
	}
	os.Exit(m.Run())
}

func TestFVSLCache(t *testing.T) {
	te := testfinder.Finders(t, nil, nil)
	c := NewFVSLCache(te.Finder)
	c.Get(1)
}
