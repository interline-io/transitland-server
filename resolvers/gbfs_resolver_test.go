package resolvers

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/testutil"
)

func setupGbfs() error {
	// Setup
	sourceFeedId := "gbfs-test"
	ts := httptest.NewServer(&gbfs.TestGbfsServer{Language: "en", Path: testutil.RelPath("test/data/gbfs")})
	defer ts.Close()
	opts := gbfs.Options{}
	opts.FeedURL = fmt.Sprintf("%s/%s", ts.URL, "gbfs.json")
	feeds, _, err := gbfs.Fetch(opts)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		key := fmt.Sprintf("%s:%s", sourceFeedId, feed.SystemInformation.Language.Val)
		TestGbfsFinder.AddData(context.Background(), key, feed)
	}
	return nil
}

func TestGbfsBikeResolver(t *testing.T) {
	setupGbfs()
	testcases := []testcase{
		{
			"basic",
			`{
				bikes(where: {near:{lon: -122.396445, lat:37.793250, radius:100}}) {
				  bike_id
				}
			}`,
			hw{},
			``,
			"bikes.#.bike_id",
			[]string{"2e09a0ed99c8ad32cca516661618645e"},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
