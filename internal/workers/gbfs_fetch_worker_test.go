package workers

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/gbfscache"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/internal/rtcache"
	"github.com/interline-io/transitland-server/internal/testutil"

	"github.com/interline-io/transitland-server/model"
)

var TestDBFinder model.Finder
var TestRTFinder model.RTFinder
var TestGbfsFinder model.GbfsFinder

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		log.Print("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	db := find.MustOpenDB(g)
	dbf := find.NewDBFinder(db)
	TestDBFinder = dbf
	TestRTFinder = rtcache.NewRTFinder(rtcache.NewLocalCache(), db)
	TestGbfsFinder = gbfscache.NewGbfsFinder()
	os.Exit(m.Run())
}

func TestGbfsFetchWorker(t *testing.T) {
	ts := httptest.NewServer(&gbfs.TestGbfsServer{Language: "en", Path: testutil.RelPath("test/data/gbfs")})
	defer ts.Close()
	job := jobs.Job{}
	job.Opts.Finder = TestDBFinder
	job.Opts.RTFinder = TestRTFinder
	job.Opts.GbfsFinder = TestGbfsFinder
	w := GbfsFetchWorker{
		Url:          ts.URL + "/gbfs.json",
		SourceType:   "gbfs",
		SourceFeedID: "test-gbfs",
	}
	err := w.Run(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
}
