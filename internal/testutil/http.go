package testutil

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
)

func NewTestServer(baseDir string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		buf, err := os.ReadFile(filepath.Join(baseDir, p))
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		w.Write(buf)
	}))
	return ts
}
