package directions

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
)

func Test_awsHandler(t *testing.T) {
	tcs := []testCase{
		{"ped", basicTests["ped"], 4215, 4.100, "../test/fixtures/response/aws_ped.json"},
		{"auto", basicTests["auto"], 671, 5.452, ""},
		{"depart_now", model.DirectionRequest{Mode: model.StepModeAuto, From: &baseFrom, To: &baseTo, DepartAt: nil}, 1, 1, ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			recorder := testutil.NewRecorder(filepath.Join("../test/fixtures/aws/location", tc.name), "directions://aws")
			defer recorder.Stop()
			h, err := makeTestAwsHandler(recorder)
			if err != nil {
				t.Fatal(err)
			}
			testHandler(t, h, tc)
		})
	}
}

func makeTestAwsHandler(tr http.RoundTripper) (*awsHandler, error) {
	// Use custom client/transport
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	hcl := &http.Client{
		Transport: tr,
	}
	cfg.HTTPClient = hcl
	lc := location.NewFromConfig(cfg)
	return newAWSHandler(lc, os.Getenv("TL_AWS_LOCATION_CALCULATOR")), nil
}
