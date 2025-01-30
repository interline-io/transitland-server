package linerouter

import (
	"testing"

	dt "github.com/interline-io/transitland-server/internal/directions/directionstest"
	"github.com/interline-io/transitland-server/testdata"
)

func Test_lineRouter(t *testing.T) {
	tcs := []dt.TestCase{
		{
			Name:     "ped",
			Req:      dt.BasicTests["ped"],
			Success:  true,
			Duration: 4116,
			Distance: 4.116,
			ResJson:  testdata.Path("directions/response/line_ped.json"),
		},
		{
			Name:     "bike",
			Req:      dt.BasicTests["bike"],
			Success:  true,
			Duration: 1029,
			Distance: 4.116,
			ResJson:  "",
		},
		{
			Name:     "auto",
			Req:      dt.BasicTests["auto"],
			Success:  true,
			Duration: 411,
			Distance: 4.116,
			ResJson:  "",
		},
		{
			Name:     "transit",
			Req:      dt.BasicTests["transit"],
			Success:  true,
			Duration: 823,
			Distance: 4.116,
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
	}
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			h := &lineRouter{}
			dt.HandlerTest(t, h, tc)
		})
	}
}
