package actions

import (
	"context"
	"testing"

	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-geom"
)

func TestValidateUpload(t *testing.T) {
	tcs := []struct {
		name        string
		serveFile   string
		rtUrls      []string
		expectError bool
		f           func(*testing.T, *model.ValidationResult)
	}{
		{
			name:      "ct",
			serveFile: "test/data/external/caltrain.zip",
			rtUrls:    []string{"test/data/rt/CT-vp-error.json"},
			f: func(t *testing.T, result *model.ValidationResult) {
				if len(result.Errors) != 1 {
					t.Fatal("expected errors")
					return
				}
				if len(result.Errors[0].Errors) != 1 {
					t.Fatal("expected errors")
					return
				}
				g := result.Errors[0].Errors[0]
				if v, ok := g.Geometry.Geometry.(*geom.GeometryCollection); ok {
					ggs := v.Geoms()
					assert.Equal(t, len(ggs), 2)
					assert.Equal(t, len(ggs[0].FlatCoords()), 1112)
					assert.Equal(t, len(ggs[1].FlatCoords()), 2)
				}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Setup http
			ts := testutil.NewTestFileserver()
			defer ts.Close()

			// Setup job
			testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
				cfg.Checker = nil // disable checker for this test
				ctx := model.WithConfig(context.Background(), cfg)
				// Run job
				feedUrl := ts.URL + "/" + tc.serveFile
				var rturls []string
				for _, v := range tc.rtUrls {
					rturls = append(rturls, ts.URL+"/"+v)
				}
				result, err := ValidateUpload(ctx, nil, &feedUrl, rturls)
				if err != nil && !tc.expectError {
					_ = result
					t.Fatal("unexpected error", err)
				} else if err == nil && tc.expectError {
					t.Fatal("expected responseError")
				} else if err != nil && tc.expectError {
					return
				}
				if tc.f != nil {
					tc.f(t, result)
				}
				// jj, _ := json.MarshalIndent(result, "", "  ")
				// fmt.Println(string(jj))
			})
		})
	}
}
