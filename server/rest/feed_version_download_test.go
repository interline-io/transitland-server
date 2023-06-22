package rest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/authn"
)

func TestFeedVersionDownloadRequest(t *testing.T) {
	g := os.Getenv("TL_TEST_STORAGE")
	if g == "" {
		t.Skip("TL_TEST_STORAGE not set - skipping")
	}
	srv, te := testRestConfig(t)
	te.Config.Storage = g
	restSrv, err := testRestServer(t, te.Config, srv)
	if err != nil {
		t.Fatal(err)
	}
	restSrv = authn.AdminDefaultMiddleware("test")(restSrv)

	t.Run("ok", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/feed_versions/d2813c293bcfd7a97dde599527ae6c62c98e66c6/download", nil)
		rr := httptest.NewRecorder()
		restSrv.ServeHTTP(rr, req)
		if sc := rr.Result().StatusCode; sc != 200 {
			t.Errorf("got status code %d, expected 200", sc)
		}
		if sc := len(rr.Body.Bytes()); sc != 59324 {
			t.Errorf("got %d bytes, expected 59324", sc)
		}
	})
	t.Run("not authorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/feed_versions/dd7aca4a8e4c90908fd3603c097fabee75fea907/download", nil)
		rr := httptest.NewRecorder()
		restSrv.ServeHTTP(rr, req)
		if sc := rr.Result().StatusCode; sc != 401 {
			t.Errorf("got status code %d, expected 401", sc)
		}
	})
	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/feed_versions/asdxyz/download", nil)
		rr := httptest.NewRecorder()
		restSrv.ServeHTTP(rr, req)
		if sc := rr.Result().StatusCode; sc != 404 {
			t.Errorf("got status code %d, expected 404", sc)
		}
	})
}

func TestFeedDownloadLatestRequest(t *testing.T) {
	g := os.Getenv("TL_TEST_STORAGE")
	if g == "" {
		t.Skip("TL_TEST_STORAGE not set - skipping")
	}
	srv, te := testRestConfig(t)
	te.Config.Storage = g
	restSrv, err := testRestServer(t, te.Config, srv)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ok", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/feeds/CT/download_latest_feed_version", nil)
		rr := httptest.NewRecorder()
		restSrv.ServeHTTP(rr, req)
		if sc := rr.Result().StatusCode; sc != 200 {
			t.Errorf("got status code %d, expected 200", sc)
		}
		if sc := len(rr.Body.Bytes()); sc != 59324 {
			t.Errorf("got %d bytes, expected 59324", sc)
		}
	})
	t.Run("not authorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/feeds/BA/download_latest_feed_version", nil)
		rr := httptest.NewRecorder()
		restSrv.ServeHTTP(rr, req)
		if sc := rr.Result().StatusCode; sc != 401 {
			t.Errorf("got status code %d, expected 401", sc)
		}
	})
	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/feeds/asdxyz/download_latest_feed_version", nil)
		rr := httptest.NewRecorder()
		restSrv.ServeHTTP(rr, req)
		if sc := rr.Result().StatusCode; sc != 404 {
			t.Errorf("got status code %d, expected 404", sc)
		}
	})
}
