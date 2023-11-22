package gql

import (
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
)

func TestFvslCache(t *testing.T) {
	te := testfinder.Finders(t, nil, nil)
	c := newFvslCache(te.Finder)
	c.Get(1)
}
