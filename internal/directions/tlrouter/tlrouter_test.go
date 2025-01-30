package tlrouter

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

func Test_tlrouterRouter(t *testing.T) {
	fdir := testdata.Path("directions/tlrouter")
	tcs := []dt.TestCase{
		{
			Name:     "ped",
			Req:      dt.BasicTests["ped"],
			Success:  true,
			Duration: 3130,
			Distance: 4.387,
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
			recorder := testutil.NewRecorder(filepath.Join(fdir, tc.Name), "directions://tlrouter")
			defer recorder.Stop()
			h, err := makeTestRouter(recorder)
			if err != nil {
				t.Fatal(err)
			}
			dt.HandlerTest(t, h, tc)
		})
	}
}

func makeTestRouter(tr http.RoundTripper) (*Router, error) {
	endpoint := os.Getenv("TL_TLROUTER_ENDPOINT")
	apikey := os.Getenv("TL_TLROUTER_APIKEY")
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
	return NewRouter(client, endpoint, apikey), nil
}
