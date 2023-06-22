package directions

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/testutil"
)

func Test_valhallaRouter(t *testing.T) {
	fdir := testutil.RelPath("test/fixtures/valhalla")
	tcs := []testCase{
		{"ped", basicTests["ped"], true, 3130, 4.387, testutil.RelPath("test/fixtures/response/val_ped.json")},
		{"bike", basicTests["bike"], true, 1132, 4.912, ""},
		{"auto", basicTests["auto"], true, 1037, 5.133, ""},
		{"transit", basicTests["transit"], false, 0, 0, ""}, // unsupported mode
		{"no_dest_fail", basicTests["no_dest_fail"], false, 0, 0, ""},
		{"no_routable_dest_fail", basicTests["no_routable_dest_fail"], false, 0, 0, ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			recorder := testutil.NewRecorder(filepath.Join(fdir, tc.name), "directions://valhalla")
			defer recorder.Stop()
			h, err := makeTestvalhallaRouter(recorder)
			if err != nil {
				t.Fatal(err)
			}
			testHandler(t, h, tc)
		})
	}
}

func makeTestvalhallaRouter(tr http.RoundTripper) (*valhallaRouter, error) {
	endpoint := os.Getenv("TL_VALHALLA_ENDPOINT")
	apikey := os.Getenv("TL_VALHALLA_API_KEY")
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
	return newValhallaRouter(client, endpoint, apikey), nil
}
