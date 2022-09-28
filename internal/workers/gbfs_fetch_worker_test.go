package workers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/gbfsfinder"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/internal/testutil"

	"github.com/interline-io/transitland-server/model"
)

func TestGbfsFetchWorker(t *testing.T) {
	ts := httptest.NewServer(&gbfs.TestGbfsServer{Language: "en", Path: testutil.RelPath("test/data/gbfs")})
	defer ts.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	gbfsFinder := gbfsfinder.NewFinder(redisClient)
	job := jobs.Job{}
	job.Opts.Finder = TestDBFinder
	job.Opts.RTFinder = TestRTFinder
	job.Opts.GbfsFinder = gbfsFinder
	w := GbfsFetchWorker{
		Url:          ts.URL + "/gbfs.json",
		SourceType:   "gbfs",
		SourceFeedID: "test-gbfs",
	}
	err := w.Run(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	// Test
	gbfsFinder.FindBikes(
		context.Background(),
		nil,
		&model.GbfsBikeRequest{
			Near: &model.PointRadius{
				Lon:    -122.396445,
				Lat:    37.793250,
				Radius: 100,
			},
		},
	)
}
