package actions

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	sq "github.com/Masterminds/squirrel"

	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStaticFetchWorker(t *testing.T) {
	tcs := []struct {
		name               string
		feedId             string
		serveFile          string
		expectError        bool
		expectResponseCode int64
		expectResponseSize int64
		expectResponseSHA1 string
		expectSuccess      bool
	}{
		{
			name:               "bart existing",
			feedId:             "BA",
			serveFile:          "test/data/external/bart.zip",
			expectResponseCode: 200,
			expectResponseSize: 456139,
			expectResponseSHA1: "e535eb2b3b9ac3ef15d82c56575e914575e732e0",
			expectSuccess:      true,
		},
		{
			name:               "bart existing old",
			feedId:             "BA",
			serveFile:          "test/data/external/bart-old.zip",
			expectResponseCode: 200,
			expectResponseSize: 429721,
			expectResponseSHA1: "dd7aca4a8e4c90908fd3603c097fabee75fea907",
			expectSuccess:      true,
		},
		{
			name:               "bart invalid",
			feedId:             "BA",
			serveFile:          "test/data/external/invalid.zip",
			expectResponseCode: 200,
			expectResponseSize: 12,
			expectResponseSHA1: "88af471a23dfdc103e67752dd56128ae77b8debe",
			expectError:        true,
			expectSuccess:      false,
		},
		// Needs FeedVersionDelete to be re-implemented
		// {
		// 	name:               "bart new",
		// 	feedId:             "BA",
		// 	serveFile:          "test/data/external/bart-new.zip",
		// 	expectResponseCode: 200,
		// 	expectResponseSize: 12,
		// 	expectResponseSHA1: "88af471a23dfdc103e67752dd56128ae77b8debe",
		// 	expectError:        true,
		// 	expectSuccess:      false,
		// },
		{
			name:               "hart existing",
			feedId:             "HA",
			serveFile:          "test/data/external/hart.zip",
			expectResponseCode: 200,
			expectResponseSize: 3543136,
			expectResponseSHA1: "c969427f56d3a645195dd8365cde6d7feae7e99b",
			expectSuccess:      true,
		},
		{
			name:               "404",
			feedId:             "BA",
			serveFile:          "test/data/example.zip",
			expectError:        true,
			expectResponseCode: 404,
			expectSuccess:      false,
		},
	}
	cfg, dbf, _, _ := testfinder.Finders(t, nil, nil)
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Setup http
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/"+tc.serveFile {
					http.Error(w, "404", 404)
					return
				}
				buf, err := ioutil.ReadFile(testutil.RelPath(tc.serveFile))
				if err != nil {
					http.Error(w, "404", 404)
					return
				}
				w.Write(buf)
			}))
			defer ts.Close()

			// Setup job
			feedUrl := ts.URL + "/" + tc.serveFile

			// Run job
			if err := StaticFetch(context.Background(), cfg, dbf, tc.feedId, feedUrl); err != nil && !tc.expectError {
				t.Fatal(err)
			} else if err == nil && tc.expectError {
				t.Fatal(errors.New("expected responseError"))
			}

			// Check output
			ff := dmfr.FeedFetch{}
			if err := find.Get(
				context.Background(),
				dbf.DBX(),
				sq.StatementBuilder.
					Select("ff.*").
					From("feed_fetches ff").
					Join("current_feeds cf on cf.id = ff.feed_id").
					Where(sq.Eq{"cf.onestop_id": tc.feedId}).
					Where(sq.Eq{"ff.url": feedUrl}).
					OrderBy("ff.id desc").
					Limit(1),
				&ff,
			); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectResponseCode, ff.ResponseCode.Val, "expect response_code")
			assert.Equal(t, tc.expectSuccess, ff.Success, "expect success")
			assert.Equal(t, tc.expectResponseSize, ff.ResponseSize.Val, "expect response_size")
			if tc.expectResponseSHA1 != "" {
				assert.Equal(t, tc.expectResponseSHA1, ff.ResponseSHA1.Val, "expect response_sha1")
			}
		})
	}
}