package directions

import (
	"errors"
	"os"

	"github.com/interline-io/transitland-server/model"
)

// PROOF OF CONCEPT

type Handler interface {
	Request(model.DirectionRequest) (*model.Directions, error)
}

func HandleRequest(preferredHandler string, req model.DirectionRequest) (*model.Directions, error) {
	if req.From == nil || req.To == nil {
		return nil, errors.New("from and to waypoints required")
	}
	var handler Handler
	switch preferredHandler {
	case "valhalla":
		handler = &valhallaHandler{}
	case "aws":
		handler = newAWSHandler(nil, os.Getenv("TL_AWS_ROUTE_CALCULATOR_NAME"))
	default:
		return nil, errors.New("unknown handler")
	}
	return handler.Request(req)
}

func validateDirectionRequest(req model.DirectionRequest) error {
	if req.From == nil || req.To == nil {
		return errors.New("from and to waypoints required")
	}
	return nil
}
