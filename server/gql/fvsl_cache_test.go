package gql

import (
	"testing"

	"github.com/interline-io/transitland-server/internal/testconfig"
)

func TestFvslCache(t *testing.T) {
	cfg := testconfig.Config(t, testconfig.Options{})
	c := newFvslCache(cfg.Finder)
	c.Get(1)
}
