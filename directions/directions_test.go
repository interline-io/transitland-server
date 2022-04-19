package directions

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
)

var baseFrom = model.WaypointInput{Lon: -122.401001, Lat: 37.789001}
var baseTo = model.WaypointInput{Lon: -122.446999, Lat: 37.782001}
var baseTime = time.Unix(1234567890, 0)

var basicTests = map[string]model.DirectionRequest{
	"ped":          {Mode: model.StepModeWalk, From: &baseFrom, To: &baseTo, DepartAt: &baseTime},
	"bike":         {Mode: model.StepModeBicycle, From: &baseFrom, To: &baseTo, DepartAt: &baseTime},
	"auto":         {Mode: model.StepModeAuto, From: &baseFrom, To: &baseTo, DepartAt: &baseTime},
	"transit":      {Mode: model.StepModeTransit, From: &baseFrom, To: &baseTo, DepartAt: &baseTime},
	"no_dest_fail": {Mode: model.StepModeWalk, From: &baseFrom, DepartAt: &baseTime},
	"no_routable_dest_fail": {
		Mode:     model.StepModeWalk,
		From:     &baseFrom,
		To:       &model.WaypointInput{Lon: -123.54949951171876, Lat: 37.703380457832374},
		DepartAt: &baseTime,
	},
}

type testCase struct {
	name     string
	req      model.DirectionRequest
	success  bool
	duration float64
	distance float64
	resJson  string
}

func testHandler(t *testing.T, h Handler, tc testCase) *model.Directions {
	ret, err := h.Request(tc.req)
	if err != nil {
		t.Fatal(err)
	}
	if ret.Success != tc.success {
		t.Errorf("got success '%t', expected '%t'", ret.Success, tc.success)
	} else if ret.Success {
		assert.InEpsilon(t, ret.Duration.Duration, tc.duration, 1.0, "duration")
		assert.InEpsilon(t, ret.Distance.Distance, tc.distance, 1.0, "distance")
	}
	_ = time.Now()
	if tc.resJson != "" {
		resJson, err := json.Marshal(ret)
		if err != nil {
			t.Fatal(err)
		}
		a, err := ioutil.ReadFile(tc.resJson)
		if err != nil {
			t.Fatal(err)
		}
		if !assert.JSONEq(t, string(a), string(resJson)) {
			t.Log("json response was:")
			t.Log(string(resJson))
		}
	}
	return ret
}
