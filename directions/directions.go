package directions

import (
	"errors"
	"os"

	"github.com/interline-io/transitland-server/model"
)

type Handler interface {
	Request(model.DirectionRequest) (*model.Directions, error)
}

func HandleRequest(preferredHandler string, req model.DirectionRequest) (*model.Directions, error) {
	var handler Handler
	switch preferredHandler {
	case "valhalla":
		handler = &valhallaHandler{}
	case "aws":
		handler = newAWSHandler(nil, os.Getenv("TL_AWS_LOCATION_CALCULATOR"))
	default:
		handler = &lineHandler{}
	}
	return handler.Request(req)
}

func validateDirectionRequest(req model.DirectionRequest) error {
	if req.From == nil || req.To == nil {
		return errors.New("from and to waypoints required")
	}
	return nil
}

func wpiWaypoint(w *model.WaypointInput) *model.Waypoint {
	if w == nil {
		return nil
	}
	return &model.Waypoint{
		Lon:  w.Lon,
		Lat:  w.Lat,
		Name: w.Name,
	}
}
