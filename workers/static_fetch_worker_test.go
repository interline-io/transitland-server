package workers

import (
	"context"
	"testing"
	"time"

	"github.com/interline-io/transitland-dbutil/testutil"
	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/testdata"
	"github.com/stretchr/testify/assert"
)

func TestStaticFetchWorker(t *testing.T) {
	ctx := context.Background()
	ts := testutil.NewTestServer(testdata.Path("external"))
	defer ts.Close()

	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
		a := "BA"
		aurl := ts.URL + "/bart.zip"
		jobQueue := cfg.JobQueue
		jobQueue.Use(newCfgMiddleware(cfg))
		jobQueue.AddQueue("default", 1)
		jobQueue.AddJobType(func() jobs.JobWorker { return &StaticFetchWorker{} })
		go func() {
			jobQueue.Run(ctx)
		}()
		jobQueue.AddJob(ctx, jobs.Job{
			JobType: "static-fetch",
			JobArgs: map[string]any{
				"feed_id":  a,
				"feed_url": aurl,
			},
		})
		time.Sleep(2 * time.Second)

		// Check that we fetched from this url
		ctx := model.WithConfig(context.Background(), cfg)
		feeds, err := cfg.Finder.FindFeeds(ctx, nil, nil, nil, &model.FeedFilter{OnestopID: &a})
		if err != nil {
			t.Fatal(err)
		}
		if len(feeds) != 1 {
			t.Fatal("expected one feed")
		}
		fetches, _ := cfg.Finder.FeedFetchesByFeedID(ctx, []model.FeedFetchParam{{FeedID: feeds[0].ID}})
		if len(fetches) == 0 {
			t.Error("expected at least one fetch")
		} else if len(fetches[0]) == 0 {
			t.Error("expected at least one fetch")
		} else {
			fetch := fetches[0][0]
			assert.Equal(t, aurl, fetch.URL, "expected url")
		}
	})
}
