package workers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/interline-io/transitland-server/model"
)

func TestGbfsFetchWorker(t *testing.T) {
	ts := httptest.NewServer(&gbfs.TestGbfsServer{Language: "en", Path: testutil.RelPath("test/data/gbfs")})
	defer ts.Close()

	testfinder.FindersTxRollback(t, nil, nil, func(te model.Finders) {
		job := jobs.Job{}
		job.Opts.Finder = te.Finder
		job.Opts.RTFinder = te.RTFinder
		job.Opts.GbfsFinder = te.GbfsFinder
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
		bikes, err := te.GbfsFinder.FindBikes(
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
