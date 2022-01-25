package directions

import (
	"encoding/json"
	"fmt"
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
	"ped":  {Mode: model.StepModeWalk, From: &baseFrom, To: &baseTo, DepartAt: &baseTime},
	"bike": {Mode: model.StepModeBicycle, From: &baseFrom, To: &baseTo, DepartAt: &baseTime},
	"auto": {Mode: model.StepModeAuto, From: &baseFrom, To: &baseTo, DepartAt: &baseTime},
}

type testCase struct {
	name     string
	req      model.DirectionRequest
	duration float64
	distance float64
	resJson  string
}

func testHandler(t *testing.T, h Handler, tc testCase) *model.Directions {
	// Use custom client/transport
	ret, err := h.Request(tc.req)
	if err != nil {
		t.Fatal(err)
	}
	resJson, err := json.Marshal(ret)
	if err != nil {
		t.Fatal(err)
	}
	assert.GreaterOrEqual(t, ret.Duration.Duration, tc.duration)
	assert.GreaterOrEqual(t, ret.Distance.Distance, tc.distance)
	if tc.resJson != "" {
		a, err := ioutil.ReadFile(tc.resJson)
		if err != nil {
			t.Fatal(err)
		}
		if !assert.JSONEq(t, string(resJson), string(a)) {
			fmt.Println("json response was:")
			fmt.Println(string(resJson))
		}
	}
	return ret
}
