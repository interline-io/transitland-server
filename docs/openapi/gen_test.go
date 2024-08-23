package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/interline-io/transitland-server/server/rest"
	"github.com/vektah/gqlparser/v2"
)

type RestQuery interface {
	RequestInfo() rest.RequestInfo
}

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
		elem := reflect.TypeOf(rq.Handler).Elem()
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			// fieldType := field.Type
			jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
			queryParam := openapi3.Parameter{
				In:          "query",
				Name:        jsonTag,
				Description: field.Name,
				// Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: openapi3.NewStringSchema().Type}},
			}
			pathOp.Parameters = append(pathOp.Parameters, &openapi3.ParameterRef{Value: &queryParam})
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
