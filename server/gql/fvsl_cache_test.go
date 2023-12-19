package gql

import (
	"testing"

	"github.com/interline-io/transitland-server/internal/testconfig"
)

func TestFvslCache(t *testing.T) {
	te := testconfig.Config(t, testconfig.Options{})
	c := newFvslCache(te.Finder)
	c.Get(1)
}
