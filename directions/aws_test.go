package directions

import (
	"fmt"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/model"
)

func TestAws(t *testing.T) {
	req := model.DirectionRequest{
		From: &model.WaypointInput{Lon: -122.7565, Lat: 49.0021},
		To:   &model.WaypointInput{Lon: -122.3394, Lat: 47.6159},
		Mode: model.StepModeAuto,
	}
	h := newAWSHandler("us-east-1", os.Getenv("TL_AWS_ROUTE_CALCULATOR_NAME"))

	res, err := h.Request(req)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("got:", res)
}
