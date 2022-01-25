package directions

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/interline-io/transitland-server/internal/testutil"
)

func Test_valhallaHandler(t *testing.T) {
	fdir := "../test/fixtures/valhalla"
	tcs := []testCase{
		{"ped", basicTests["ped"], 3130, 4.387, "../test/fixtures/response/val_ped.json"},
		{"bike", basicTests["bike"], 1132, 4.912, ""},
		{"auto", basicTests["auto"], 1037, 5.133, ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			recorder := testutil.NewRecorder(filepath.Join(fdir, tc.name), "directions://valhalla")
			defer recorder.Stop()
			hcl := &http.Client{
				Transport: recorder,
			}
			h := newValhallaHandler(hcl)
			testHandler(t, h, tc)
		})
	}
}
