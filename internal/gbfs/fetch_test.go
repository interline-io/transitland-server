package gbfs

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-server/internal/testutil"
)

func TestGbfsFetchWorker(t *testing.T) {
	ts := httptest.NewServer(&TestGbfsServer{Language: "en", Path: testutil.RelPath("test/data/gbfs")})
	defer ts.Close()
	opts := Options{}
	opts.FeedURL = fmt.Sprintf("%s/%s", ts.URL, "gbfs.json")
	feed, _, err := Fetch(opts)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("feed:", feed)
}
