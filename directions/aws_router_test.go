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

func Test_awsRouter(t *testing.T) {
	tcs := []testCase{
		{"ped", basicTests["ped"], true, 4215, 4.100, "../test/fixtures/response/aws_ped.json"},
		{"bike", basicTests["bike"], false, 0, 0, ""}, // unsupported mode
		{"auto", basicTests["auto"], true, 671, 5.452, ""},
		{"depart_now", model.DirectionRequest{Mode: model.StepModeAuto, From: &baseFrom, To: &baseTo, DepartAt: nil}, true, 671, 4.1, ""}, // at LEAST 671s
		{"no_dest_fail", basicTests["no_dest_fail"], false, 0, 0, ""},
		{"no_routable_dest_fail", basicTests["no_routable_dest_fail"], false, 0, 0, ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			recorder := testutil.NewRecorder(filepath.Join("../test/fixtures/aws/location", tc.name), "directions://aws")
			defer recorder.Stop()
			h, err := makeTestawsRouter(recorder)
			if err != nil {
				t.Fatal(err)
			}
			testHandler(t, h, tc)
		})
	}
}

func makeTestawsRouter(tr http.RoundTripper) (*awsRouter, error) {
	// Use custom client/transport
	cn := os.Getenv("TL_AWS_LOCATION_CALCULATOR")
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	hcl := &http.Client{
		Transport: tr,
	}
	cfg.HTTPClient = hcl
	lc := location.NewFromConfig(cfg)
	return newAWSRouter(lc, cn), nil
}
