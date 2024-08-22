package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/vektah/gqlparser/v2"
)

func TestGen(t *testing.T) {
	type restQuery struct {
		Operation   string
		Path        string
		Summary     string
		Description string
		Doc         string
		QueryParams map[string]string
	}
	restQueries := map[string]restQuery{
		"stops": {
			Operation:   "stops",
			Path:        "/stops",
			Doc:         "stop_request.gql",
			QueryParams: map[string]string{"stop_key": "Stop lookup key..."},
		},
	}

	// Load schema
	schema := gqlout.NewExecutableSchema(gqlout.Config{Resolvers: &gql.Resolver{}})

	// Prepare document
	root := openapi3.T{}
	root.Paths = openapi3.NewPaths()
	root.Components = &openapi3.Components{}
	root.Components.Schemas = openapi3.Schemas{}

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

		///////////
		pathOp := openapi3.NewOperation()
		pathOp.Summary = k
		pathOp.Description = k
		pathOp.Parameters = openapi3.NewParameters()
		for k, v := range rq.QueryParams {
			pathOp.Parameters = append(pathOp.Parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					In:          "query",
					Name:        k,
					Description: v,
					Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: openapi3.NewStringSchema().Type}},
				},
			})
		}
		root.Paths.Set(rq.Path, &openapi3.PathItem{
			Summary:     rq.Summary,
			Description: rq.Description,
			Get:         pathOp,
		})

	}

	jj, _ := json.MarshalIndent(root, "", "  ")
	fmt.Println(string(jj))
}
