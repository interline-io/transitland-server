package gbfs

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGbfsFetch(t *testing.T) {
	ts := httptest.NewServer(&TestGbfsServer{Language: "en", Path: testutil.RelPath("test/data/gbfs")})
	defer ts.Close()
	opts := Options{}
	opts.FeedURL = fmt.Sprintf("%s/%s", ts.URL, "gbfs.json")
	feeds, _, err := Fetch(nil, opts)
	if err != nil {
		t.Fatal(err)
	}
	fids := []string{}
	for _, ent := range feeds {
		fids = append(fids, ent.SystemInformation.Name.Val)
	}
	assert.ElementsMatch(t, []string{"Bay Wheels"}, fids)
}
