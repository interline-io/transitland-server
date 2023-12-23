package gql

import (
	"context"
	"testing"

	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/model"
)

func TestFvslCache(t *testing.T) {
	cfg := testconfig.Config(t, testconfig.Options{})
	c := newFvslCache()
	c.Get(model.WithConfig(context.Background(), cfg), 1)
}
