package rest

import (
	oa "github.com/getkin/kin-openapi/openapi3"
)

type RestHandlers interface {
	RequestInfo() RequestInfo
}

func GenerateOpenAPI(restPrefix string, opts ...SchemaOption) (*oa.T, error) {
	// Apply options
	config := &SchemaConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Determine server URL based on RestPrefix
	serverURL := "https://transit.land/api/v2/rest"
	if restPrefix != "" {
		serverURL = restPrefix
	}

	outdoc := &oa.T{
		OpenAPI: "3.0.0",
		Info: &oa.Info{
			Title:       "Transitland REST API",
			Description: "Transitland REST API - Access transit data including feeds, agencies, routes, stops, operators, and real-time departures",
			Version:     "2.0.0",
			Contact: &oa.Contact{
				Email: "info@interline.io",
			},
		},
		Servers: []*oa.Server{
			{
				URL:         serverURL,
				Description: "Transitland REST API",
			},
		},
	}

	// Add parameter components
	outdoc.Components = &oa.Components{
		Parameters: oa.ParametersMap{},
	}
	for paramName, paramRef := range ParameterComponents {
		outdoc.Components.Parameters[paramName] = paramRef
	}

	// Apply custom components if provided
	if config.Components != nil {
		if config.Components.SecuritySchemes != nil {
			outdoc.Components.SecuritySchemes = config.Components.SecuritySchemes
		}
		// Could add other component types here (schemas, responses, etc.)
	}

	// Create PathItem for each handler
	var pathOpts []oa.NewPathsOption
	var handlers = []RestHandlers{
		&FeedRequest{},
		&FeedVersionRequest{},
		&OperatorRequest{},
		&AgencyRequest{},
		&RouteRequest{},
		&TripRequest{},
		&StopRequest{},
		&StopDepartureRequest{},
		// Individual resource handlers
		&AgencyKeyRequest{},
		&RouteKeyRequest{},
		&TripEntityRequest{},
		&StopEntityRequest{},
		&FeedDownloadLatestFeedVersionRequest{},
		&FeedVersionDownloadRequest{},
		&FeedDownloadRtRequest{},
	}
	for _, handler := range handlers {
		requestInfo := handler.RequestInfo()
		oaResponse, err := queryToOAResponses(requestInfo.Get.Query)
		if err != nil {
			return outdoc, err
		}
		getOp := requestInfo.Get.Operation
		getOp.Responses = oaResponse
		getOp.Description = requestInfo.Description

		// Apply custom security if provided
		if config.GlobalSecurity != nil {
			getOp.Security = config.GlobalSecurity
		}

		pathItem := &oa.PathItem{Get: getOp}
		pathOpts = append(pathOpts, oa.WithPath(requestInfo.Path, pathItem))
	}
	outdoc.Paths = oa.NewPaths(pathOpts...)
	return outdoc, nil
}

// queryToOAResponses converts a GraphQL query to OpenAPI responses
func queryToOAResponses(_ string) (*oa.Responses, error) {
	// Create responses with proper error handling
	description := "Successful response"
	responses := oa.NewResponses()
	responses.Set("200", &oa.ResponseRef{
		Value: &oa.Response{
			Description: &description,
			Content: oa.NewContentWithJSONSchema(&oa.Schema{
				Type: &oa.Types{"object"},
			}),
		},
	})

	// Add common error responses
	badRequestDesc := "Bad request - invalid parameters"
	responses.Set("400", &oa.ResponseRef{
		Value: &oa.Response{
			Description: &badRequestDesc,
			Content: oa.NewContentWithJSONSchema(&oa.Schema{
				Type: &oa.Types{"object"},
			}),
		},
	})

	serverErrorDesc := "Internal server error"
	responses.Set("500", &oa.ResponseRef{
		Value: &oa.Response{
			Description: &serverErrorDesc,
			Content: oa.NewContentWithJSONSchema(&oa.Schema{
				Type: &oa.Types{"object"},
			}),
		},
	})

	// Add explicit default response to avoid empty description
	defaultDesc := "Unexpected error"
	responses.Set("default", &oa.ResponseRef{
		Value: &oa.Response{
			Description: &defaultDesc,
			Content: oa.NewContentWithJSONSchema(&oa.Schema{
				Type: &oa.Types{"object"},
			}),
		},
	})

	return responses, nil
}

// SchemaConfig holds configuration options for schema generation
type SchemaConfig struct {
	Components     *oa.Components
	GlobalSecurity *oa.SecurityRequirements
}

// SchemaOption is a function that modifies schema configuration
type SchemaOption func(*SchemaConfig)

// WithComponents adds custom components to the schema
func WithComponents(components *oa.Components) SchemaOption {
	return func(config *SchemaConfig) {
		config.Components = components
	}
}

// WithSecurity adds global security requirements to all operations
func WithSecurity(security *oa.SecurityRequirements) SchemaOption {
	return func(config *SchemaConfig) {
		config.GlobalSecurity = security
	}
}
