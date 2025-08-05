//go:generate go run . rest.json
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

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
	var handlers = rest.RestHandlersList
	for _, handler := range handlers {
		requestInfo := handler.RequestInfo()
		oaResponse, err := queryToOAResponses(requestInfo.Get.Query)
		if err != nil {
			return outdoc, err
		}
		getOp := requestInfo.Get.Operation
		getOp.Responses = oaResponse
		getOp.Description = requestInfo.Description
		pathItem := &oa.PathItem{Get: getOp}
		pathOpts = append(pathOpts, oa.WithPath(requestInfo.Path, pathItem))
	}
	outdoc.Paths = oa.NewPaths(pathOpts...)
	return outdoc, nil
}

func main() {
	ctx := context.Background()
	args := os.Args
	if len(args) != 2 {
		exit(errors.New("output file required"))
	}
	outfile := args[1]

	// Generate OpenAPI schema
	outdoc, err := GenerateOpenAPI()
	if err != nil {
		exit(err)
	}

	// Validate output
	jj, err := json.MarshalIndent(outdoc, "", "  ")
	if err != nil {
		exit(err)
	}

	schema, err := oa.NewLoader().LoadFromData(jj)
	if err != nil {
		exit(err)
	}
	var validationOpts []oa.ValidationOption
	if err := schema.Validate(ctx, validationOpts...); err != nil {
		exit(err)
	}

	// After validation, write to file
	outf, err := os.Create(outfile)
	if err != nil {
		exit(err)
	}
	outf.Write(jj)
}

func exit(err error) {
	fmt.Println("Error: ", err.Error())
	os.Exit(1)
}
