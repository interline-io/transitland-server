package directions

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/interline-io/transitland-dbutil/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/testdata"
)

func Test_awsRouter(t *testing.T) {
	tcs := []testCase{
		{"ped", basicTests["ped"], true, 4215, 4.100, testdata.Path("directions/response/aws_ped.json")},
		{"bike", basicTests["bike"], false, 0, 0, ""}, // unsupported mode
		{"auto", basicTests["auto"], true, 671, 5.452, ""},
		{"depart_now", model.DirectionRequest{Mode: model.StepModeAuto, From: &baseFrom, To: &baseTo, DepartAt: nil}, true, 671, 4.1, ""}, // at LEAST 671s
		{"no_dest_fail", basicTests["no_dest_fail"], false, 0, 0, ""},
		{"no_routable_dest_fail", basicTests["no_routable_dest_fail"], false, 0, 0, ""},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			recorder := testutil.NewRecorder(filepath.Join(testdata.Path("directions/aws/location"), tc.name), "directions://aws")
			defer recorder.Stop()
			h, err := makeTestMockRouter(recorder)
			if err != nil {
				t.Fatal(err)
			}
			testHandler(t, h, tc)
		})
	}
}

// Mock reader
func makeTestMockRouter(tr http.RoundTripper) (*awsRouter, error) {
	// Use custom client/transport
	cn := ""
	lc := &mockLocationClient{
		Client: &http.Client{
			Transport: tr,
		},
	}
	return newAWSRouter(lc, cn), nil
}

// Regenerate results
// func makeTestAwsRouter(tr http.RoundTripper) (*awsRouter, error) {
// 	cn := os.Getenv("TL_AWS_LOCATION_CALCULATOR")
// 	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	cfg.HTTPClient = &http.Client{
// 		Transport: tr,
// 	}
// 	lc := location.NewFromConfig(cfg)
// 	return newAWSRouter(lc, cn), nil
// }

// We need to mock out the location services client
type mockLocationClient struct {
	Client *http.Client
}

func (mc *mockLocationClient) CalculateRoute(ctx context.Context, params *location.CalculateRouteInput, opts ...func(*location.Options)) (*location.CalculateRouteOutput, error) {
	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "directions://aws", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	resp, err := mc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	b, _ := ioutil.ReadAll(resp.Body)
	a := location.CalculateRouteOutput{}
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
