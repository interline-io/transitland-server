package authz

import (
	"context"
	"fmt"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
)

func newTestChecker(t testing.TB, cfg AuthzConfig, finder model.Finder) (*Checker, error) {
	auth0c, err := newTestAuth0Client(t, cfg)
	if err != nil {
		return nil, err
	}
	fgac, err := newTestFGAClient(t, cfg)
	if err != nil {
		return nil, err
	}
	checker := NewChecker(auth0c, fgac, finder, nil)
	return checker, err
}

func TestChecker(t *testing.T) {
	te := testfinder.Finders(t, nil, nil)
	cfg := newTestConfig()
	cfg.FGAEndpoint = "http://localhost:8090"
	cfg.FGATestModelPath = "../test/authz/tls.model"
	cfg.FGATestTuplesPath = "../test/authz/tls.csv"
	checker, err := newTestChecker(t, cfg, te.Finder)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("ListFeeds", func(t *testing.T) {
		ret, err := checker.ListFeeds(context.Background(), newTestUser("ian"))
		if err != nil {
			t.Fatal(err)
		}
		assert.ElementsMatch(t, []int{1, 2, 3}, ret, "feed ids")
	})
	t.Run("ListFeedVersions", func(t *testing.T) {
		ret, err := checker.ListFeedVersions(context.Background(), newTestUser("ian"))
		if err != nil {
			t.Fatal(err)
		}
		assert.ElementsMatch(t, []int{1}, ret, "feed version ids")
	})
	t.Run("FeedPermissions", func(t *testing.T) {
		ret, err := checker.FeedPermissions(context.Background(), newTestUser("ian"), 1)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("ret:", ret)
	})
}
