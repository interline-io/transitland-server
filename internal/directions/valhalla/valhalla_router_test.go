package valhalla

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/interline-io/transitland-dbutil/testutil"
	dt "github.com/interline-io/transitland-server/internal/directions/directionstest"
	"github.com/interline-io/transitland-server/testdata"
)

func Test_valhallaRouter(t *testing.T) {
	fdir := testdata.Path("directions/valhalla")
	tcs := []dt.TestCase{
		{
			Name:     "ped",
			Req:      dt.BasicTests["ped"],
			Success:  true,
			Duration: 3130,
			Distance: 4.387,
			ResJson:  testdata.Path("directions/response/val_ped.json"),
		},
		{
			Name:     "bike",
			Req:      dt.BasicTests["bike"],
			Success:  true,
			Duration: 1132,
			Distance: 4.912,
			ResJson:  "",
		},
		{
			Name:     "auto",
			Req:      dt.BasicTests["auto"],
			Success:  true,
			Duration: 1037,
			Distance: 5.133,
			ResJson:  "",
		},
		{
			Name:     "transit",
			Req:      dt.BasicTests["transit"],
			Success:  false,
			Duration: 0,
			Distance: 0,
			ResJson:  "",
		},
		{
			Name:     "no_dest_fail",
			Req:      dt.BasicTests["no_dest_fail"],
			Success:  false,
			Duration: 0,
			Distance: 0,
			ResJson:  "",
		},
		{
			Name:     "no_routable_dest_fail",
			Req:      dt.BasicTests["no_routable_dest_fail"],
			Success:  false,
			Duration: 0,
			Distance: 0,
			ResJson:  "",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			recorder := testutil.NewRecorder(filepath.Join(fdir, tc.Name), "directions://valhalla")
			defer recorder.Stop()
			h, err := makeTestvalhallaRouter(recorder)
			if err != nil {
				t.Fatal(err)
			}
			dt.HandlerTest(t, h, tc)
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
