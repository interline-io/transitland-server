package directions

import (
	"testing"

	"github.com/interline-io/transitland-server/testdata"
)

func Test_lineRouter(t *testing.T) {
	tcs := []testCase{
		{"ped", basicTests["ped"], true, 4116, 4.116, testdata.DataPath("fixtures/response/line_ped.json")},
		{"bike", basicTests["bike"], true, 1029, 4.116, ""},
		{"auto", basicTests["auto"], true, 411, 4.116, ""},
		{"transit", basicTests["transit"], true, 823, 4.116, ""},
		{"no_dest_fail", basicTests["no_dest_fail"], false, 0, 0, ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			h := &lineRouter{}
			testHandler(t, h, tc)
		})
	}
}
