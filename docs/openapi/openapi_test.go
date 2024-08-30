package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	oa "github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/interline-io/transitland-server/server/rest"
	"github.com/vektah/gqlparser/v2"
)

// type RestQuery interface {
// 	RequestInfo() rest.RequestInfo
// }

func TestGen(t *testing.T) {
	type restQuery struct {
		Operation   string
		Path        string
		Summary     string
		Description string
		Doc         string
		Handler     any
	}
	restQueries := map[string]restQuery{
		"stops": {
			Operation: "stops",
			Path:      "/stops",
			Doc:       "stop_request.gql",
			Handler:   &rest.StopRequest{},
		},
	}

	// Load schema
	schema := gqlout.NewExecutableSchema(gqlout.Config{Resolvers: &gql.Resolver{}})

	// Prepare document
	root := oa.T{}
	root.Paths = oa.NewPaths()
	root.Components = &oa.Components{}
	root.Components.Schemas = oa.Schemas{}

	for k, rq := range restQueries {
		q, _ := os.ReadFile(filepath.Join("..", "..", "server", "rest", rq.Doc))
		query, err := gqlparser.LoadQuery(schema.Schema(), string(q))
		if err != nil {
			panic(err)
		}

		///////////
		for _, op := range query.Operations {
			for _, sel := range op.SelectionSet {
				selRecurse(sel, root.Components.Schemas, 0)
			}
		}
		_ = k
	}
	jj, _ := json.MarshalIndent(root, "", "  ")
	fmt.Println(string(jj))
}

func TestValidateSchema(t *testing.T) {
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
	outdoc.Components = &oa.Components{
		Parameters: oa.ParametersMap{},
	}
	for paramName, paramRef := range rest.ParameterComponents {
		outdoc.Components.Parameters[paramName] = paramRef
	}
	var pathOpts []oa.NewPathsOption
	for pathName, pathItem := range rest.PathItems {
		desc := "ok"
		res := openapi3.WithStatus(200, &oa.ResponseRef{Value: &oa.Response{
			Description: &desc,
		}})
		pathItem.Get.Responses = openapi3.NewResponses(res)
		pathOpts = append(pathOpts, oa.WithPath(pathName, pathItem))
	}
	outdoc.Paths = oa.NewPaths(pathOpts...)

	jj, _ := json.MarshalIndent(outdoc, "", "  ")

	out, _ := os.Create("rest-out.json")
	out.Write(jj)
	out.Close()

	schema, err := oa.NewLoader().LoadFromFile("./rest-out.json")
	if err != nil {
		t.Fatal(err)
	}
	var validationOpts []oa.ValidationOption
	if err := schema.Validate(context.Background(), validationOpts...); err != nil {
		t.Fatal(err)
	}
}
