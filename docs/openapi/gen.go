package openapi

import (
	oa "github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/server/rest"
)

type RestHandlers interface {
	RequestInfo() rest.RequestInfo
}

func GenerateOpenAPI() (*oa.T, error) {
	outdoc := &oa.T{
		OpenAPI: "3.0.0",
		Info: &oa.Info{
			Title:       "Transitland REST API",
			Description: "Transitland REST API",
			Version:     "1.0.0-oas3",
			Contact: &oa.Contact{
				Email: "hello@transit.land",
			},
		},
	}

	// Add parameter components
	outdoc.Components = &oa.Components{
		Parameters: oa.ParametersMap{},
	}
	for paramName, paramRef := range rest.ParameterComponents {
		outdoc.Components.Parameters[paramName] = paramRef
	}

	// Create PathItem for each handler
	var pathOpts []oa.NewPathsOption
	var handlers = []RestHandlers{
		&rest.FeedRequest{},
		&rest.FeedVersionRequest{},
		&rest.OperatorRequest{},
		&rest.AgencyRequest{},
		&rest.RouteRequest{},
		&rest.TripRequest{},
		&rest.StopRequest{},
	}
	for _, handler := range handlers {
		requestInfo := handler.RequestInfo()
		pathOpts = append(pathOpts, oa.WithPath(requestInfo.Path, requestInfo.PathItem))
	}
	outdoc.Paths = oa.NewPaths(pathOpts...)
	return outdoc, nil
}
