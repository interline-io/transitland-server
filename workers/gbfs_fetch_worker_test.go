package workers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-mw/jobs"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/interline-io/transitland-server/model"
)

func TestGbfsFetchWorker(t *testing.T) {
	ts := httptest.NewServer(&gbfs.TestGbfsServer{Language: "en", Path: testutil.RelPath("test/data/gbfs")})
	defer ts.Close()

	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
		job := jobs.Job{}
		w := GbfsFetchWorker{
			Url:    ts.URL + "/gbfs.json",
			FeedID: "test-gbfs",
		}
		ctx := model.WithConfig(context.Background(), cfg)
		err := w.Run(ctx, job)
		if err != nil {
			t.Fatal(err)
		}
		// Test
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
