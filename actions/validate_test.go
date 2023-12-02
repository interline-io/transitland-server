package actions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
)

type hw map[string]any

func TestValidateUpload(t *testing.T) {
	baseDir := testutil.RelPath("test/data")
	ts200 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := os.ReadFile(filepath.Join(baseDir, r.URL.Path))
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		w.Write(buf)
	}))
	testfinder.FindersTxRollback(t, nil, nil, func(te model.Finders) {
		url := ts200.URL + "/external/caltrain.zip"
		rtUrls := []string{ts200.URL + "/rt/CT.json"}
		vr, err := ValidateUpload(context.Background(), te.Config, nil, &url, rtUrls, nil)
		if err != nil {
			t.Fatal(err)
		}
		_ = vr
	})

}
