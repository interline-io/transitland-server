package actions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	sq "github.com/Masterminds/squirrel"

	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/testdata"
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
			serveFile:          "external/bart.zip",
			expectResponseCode: 200,
			expectResponseSize: 456139,
			expectResponseSHA1: "e535eb2b3b9ac3ef15d82c56575e914575e732e0",
			expectSuccess:      true,
		},
		{
			name:               "bart existing old",
			feedId:             "BA",
			serveFile:          "external/bart-old.zip",
			expectResponseCode: 200,
			expectResponseSize: 429721,
			expectResponseSHA1: "dd7aca4a8e4c90908fd3603c097fabee75fea907",
			expectSuccess:      true,
		},
		{
			name:               "bart invalid",
			feedId:             "BA",
			serveFile:          "external/invalid.zip",
			expectResponseCode: 200,
			expectResponseSize: 12,
			expectResponseSHA1: "88af471a23dfdc103e67752dd56128ae77b8debe",
			expectError:        false,
			expectSuccess:      false,
		},
		{
			name:               "bart new",
			feedId:             "BA",
			serveFile:          "external/bart-new.zip",
			expectResponseCode: 200,
			expectResponseSize: 1151609,
			expectResponseSHA1: "b40aa01814bf92dba06dbccdebcc3aefa6208248",
			expectError:        false,
			expectSuccess:      true,
		},
		{
			name:               "hart existing",
			feedId:             "HA",
			serveFile:          "external/hart.zip",
			expectResponseCode: 200,
			expectResponseSize: 3543136,
			expectResponseSHA1: "c969427f56d3a645195dd8365cde6d7feae7e99b",
			expectSuccess:      true,
		},
		{
			name:               "404",
			feedId:             "BA",
			serveFile:          "example.zip",
			expectError:        false,
			expectResponseCode: 404,
			expectSuccess:      false,
		},
		{
			name:        "invalid feed",
			feedId:      "unknown",
			serveFile:   "example.zip",
			expectError: true,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Setup http
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/"+tc.serveFile {
					http.Error(w, "404", 404)
					return

				}
				buf, err := os.ReadFile(testdata.Path(tc.serveFile))
				if err != nil {
					http.Error(w, "404", 404)
					return
				}
				w.Write(buf)
			}))
			defer ts.Close()

			// Setup job
			feedUrl := ts.URL + "/" + tc.serveFile
			testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
				cfg.Checker = nil // disable checker for this test
				ctx := model.WithConfig(context.Background(), cfg)
				// Run job
				if result, err := StaticFetch(ctx, tc.feedId, nil, feedUrl); err != nil && !tc.expectError {
					_ = result
					t.Fatal("unexpected error", err)
				} else if err == nil && tc.expectError {
					t.Fatal("expected responseError")
				} else if err != nil && tc.expectError {
					return
				}
				// Check output
				ff := dmfr.FeedFetch{}
				if err := dbutil.Get(
					ctx,
					cfg.Finder.DBX(),
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

		})
	}
}
