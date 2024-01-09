package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

func NewTestFileserver() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		buf, err := os.ReadFile(RelPath(r.URL.Path))
		if err != nil {
			http.Error(w, "404", 404)
			return
		}
		w.Write(buf)
	}))
	return ts
}
