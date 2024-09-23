package workers

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/testdata"
	"github.com/stretchr/testify/assert"
)

func TestGbfsFetchWorker(t *testing.T) {
	ts := httptest.NewServer(&gbfs.TestGbfsServer{Language: "en", Path: testdata.Path("gbfs")})
	defer ts.Close()
	ctx := context.Background()
	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
		jobQueue := cfg.JobQueue
		jobQueue.Use(newCfgMiddleware(cfg))
		jobQueue.AddQueue("default", 1)
		jobQueue.AddJobType(func() jobs.JobWorker { return &GbfsFetchWorker{} })
		go func() {
			jobQueue.Run(ctx)
			time.Sleep(1 * time.Second)
		}()
		jobQueue.AddJob(ctx, jobs.Job{
			JobType: "gbfs-fetch",
			JobArgs: map[string]any{
				"url":     ts.URL + "/gbfs.json",
				"feed_id": "test-gbfs",
			},
		})
		time.Sleep(1 * time.Second)

		// Test
		ctx := model.WithConfig(context.Background(), cfg)
		bikes, err := cfg.GbfsFinder.FindBikes(
			ctx,
			nil,
			&model.GbfsBikeRequest{
				Near: &model.PointRadius{
					Lon:    -122.396445,
					Lat:    37.793250,
					Radius: 100,
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		bikeids := []string{}
		for _, ent := range bikes {
			bikeids = append(bikeids, ent.BikeID.Val)
		}
		assert.ElementsMatch(t, []string{"2e09a0ed99c8ad32cca516661618645e"}, bikeids)
	})
}
