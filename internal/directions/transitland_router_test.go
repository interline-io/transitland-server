package directions

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/interline-io/transitland-dbutil/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/testdata"
)

func Test_transitlandRouter(t *testing.T) {
	// Ensure the directory for test fixtures exists
	fdir := testdata.Path("directions/transitland")
	tcs := []testCase{
		{
			"transit",
			model.DirectionRequest{
				Mode:     model.StepModeTransit,
				From:     &model.WaypointInput{Lat: 37.7757, Lon: -122.47996},
				To:       &model.WaypointInput{Lat: 37.76845, Lon: -122.23508},
				DepartAt: &baseDepartAt,
			},
			true,
			5389, // Duration from request.yaml
			26412.81, // Distance from request.yaml
			testdata.Path("directions/response/tl_transit.json"),
		},
		{
			"transit_with_walk",
			model.DirectionRequest{
				Mode:     model.StepModeTransit,
				From:     &baseFrom,
				To:       &baseTo,
				DepartAt: &baseDepartAt,
			},
			true,
			3000, // Example duration, adjust based on actual data
			6000, // Example distance, adjust based on actual data
			"",
		},
		{
			"walk_only",
			basicTests["walk"],
			true,
			3600, // Example duration, adjust based on actual data
			4500, // Example distance, adjust based on actual data
			"",
		},
		{
			"auto",
			basicTests["auto"],
			false, // Should fail as unsupported mode
			0,
			0,
			"",
		},
		{
			"no_dest_fail",
			basicTests["no_dest_fail"],
			false,
			0,
			0,
			"",
		},
		{
			"no_service",
			model.DirectionRequest{
				Mode: model.StepModeTransit,
				From: &model.WaypointInput{Lat: 0, Lon: 0},
				To:   &model.WaypointInput{Lat: 1, Lon: 1},
			},
			false,
			0,
			0,
			"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			recorder := testutil.NewRecorder(filepath.Join(fdir, tc.name), "directions://transitland_router")
			defer recorder.Stop()
			h, err := makeTestTransitlandRouter(recorder)
			if err != nil {
				t.Fatal(err)
			}
			testHandler(t, h, tc)
		})
	}
}

func makeTestTransitlandRouter(tr http.RoundTripper) (*transitlandRouter, error) {
	endpoint := os.Getenv("TL_TRANSITLAND_ENDPOINT")
	apikey := os.Getenv("TL_TRANSITLAND_API_KEY")
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}
	return newTransitlandRouter(client, endpoint, apikey), nil
}

// Test specific response parsing
func Test_transitlandResponseParsing(t *testing.T) {
	resp := &transitlandResponse{
		Plan: struct {
			Itineraries []struct {
				Duration     float64 `json:"duration"`
				StartTime    int64   `json:"startTime"`
				EndTime      int64   `json:"endTime"`
				WalkTime     float64 `json:"walkTime"`
				TransitTime  float64 `json:"transitTime"`
				WaitingTime  float64 `json:"waitingTime"`
				WalkDistance float64 `json:"walkDistance"`
				Transfers    int     `json:"transfers"`
				Legs        []struct {
					StartTime   int64           `json:"startTime"`
					EndTime     int64           `json:"endTime"`
					Mode        string          `json:"mode"`
					Duration    float64         `json:"duration"`
					Distance    float64         `json:"distance"`
					From        transitlandStop `json:"from"`
					To          transitlandStop `json:"to"`
					LegGeometry struct {
						Points string `json:"points"`
					} `json:"legGeometry"`
				} `json:"legs"`
			}{
				{
					Duration:     5389, // Duration from request.yaml
					StartTime:    1732759804000,
					EndTime:      1732765193000,
					WalkTime:     927,
					TransitTime:  3367,
					WaitingTime:  1095,
					WalkDistance: 1111.97,
					Transfers:    2,
					Legs: []struct {
						StartTime   int64           `json:"startTime"`
						EndTime     int64           `json:"endTime"`
						 Mode        string          `json:"mode"`
						 Duration    float64         `json:"duration"`
						 Distance    float64         `json:"distance"`
						 From        transitlandStop `json:"from"`
						 To          transitlandStop `json:"to"`
						 LegGeometry struct {
							 Points string `json:"points"`
						 } `json:"legGeometry"`
					}{
						{
							StartTime: 1732759804000,
							EndTime:   1732760402000,
							Mode:      "WALK",
							Duration:  598,
							Distance:  598.44,
							From: transitlandStop{
								Name: "Origin",
								Lat:  37.7757,
								Lon:  -122.47996,
							},
							To: transitlandStop{
								Name: "Fulton St & 22nd Ave",
								Lat:  37.77266,
								Lon:  -122.48061,
							},
						},
						{
							StartTime: 1732760402000,
							EndTime:   1732762269000,
							Mode:      "BUS",
							Duration:  1867,
							Distance:  6738.26,
							From: transitlandStop{
								Name: "Fulton St & 22nd Ave",
								Lat:  37.77266,
								Lon:  -122.48061,
							},
							To: transitlandStop{
								Name: "Bus Stop B",
								Lat:  37.7751,
								Lon:  -122.4196,
							},
						},
					},
				},
			},
		},
	}

	directions := convertTransitlandResponse(resp)

	if !directions.Success {
		t.Error("Expected successful conversion")
	}

	if len(directions.Legs) != 2 {
		t.Errorf("Expected 2 legs, got %d", len(directions.Legs))
	}

	if directions.Duration.Duration != 5389 {
		t.Errorf("Expected duration 5389, got %f", directions.Duration.Duration)
	}

	if directions.Legs[0].Mode != model.StepModeWalk {
		t.Errorf("Expected first leg mode WALK, got %s", directions.Legs[0].Mode)
	}

	if directions.Legs[1].Mode != model.StepModeTransit {
		t.Errorf("Expected second leg mode TRANSIT, got %s", directions.Legs[1].Mode)
	}
} 