package openapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

var scalarTypes = map[string]openapi3.Schema{
	"Time": {
		Type:    openapi3.NewStringSchema().Type,
		Format:  "datetime",
		Example: "2019-11-15T00:45:55.409906",
	},
	"Int": {
		Type: openapi3.NewIntegerSchema().Type,
	},
	"Float": {
		Type: openapi3.NewFloat64Schema().Type,
	},
	"String": {
		Type: openapi3.NewStringSchema().Type,
	},
	"Boolean": {
		Type: openapi3.NewBoolSchema().Type,
	},
	// "ID":      {},
	// "Counts":     {},
	// "Tags":       {},
	// "Geometry":   {},
	// "Date":       {},
	// "Point":      {},
	// "LineString": {},
	// "Seconds":    {},
	// "Polygon":    {},
	// "Map":        {},
	// "Any":        {},
	// "Upload":     {},
	// "Key":        {},
	// "Bool":       {},
	// "Strings":    {},
}

func TestGen(t *testing.T) {
	q, _ := ioutil.ReadFile("../../server/rest/stop_request.gql")
	c := gqlout.Config{Resolvers: &gql.Resolver{}}
	schema := gqlout.NewExecutableSchema(c)
	_ = schema
	// schema.Schema().Query
	query, err := gqlparser.LoadQuery(schema.Schema(), string(q))
	if err != nil {
		panic(err)
	}
	fmt.Println("query:", query)
	root := openapi3.Schemas{}
	for _, op := range query.Operations {
		fmt.Printf("\top: %s : %#v\n", op.Name, op)
		for _, sel := range op.SelectionSet {
			selRecurse(sel, root, 0)
		}
	}
	jj, _ := json.Marshal(root)
	fmt.Println(string(jj))
}

func selRecurse(v any, p openapi3.Schemas, level int) {
	if frag, ok := v.(*ast.FragmentSpread); ok {
		x := openapi3.Schema{}
		x.Title = frag.Name
		x.Properties = openapi3.Schemas{}
		x.Type = openapi3.NewStringSchema().Type
		x.Description = frag.ObjectDefinition.Description
		fmt.Println(strings.Repeat(" ", level*4), "frag:", frag.Name)
		for _, r := range frag.Definition.SelectionSet {
			selRecurse(r, x.Properties, level)
		}
		p[frag.Name] = openapi3.NewSchemaRef("", &x)
	}
	if sel, ok := v.(*ast.Field); ok {
		x := openapi3.Schema{}
		x.Title = sel.Name
		x.Properties = openapi3.Schemas{}
		x.Description = sel.Definition.Description
		x.Nullable = !sel.Definition.Type.NonNull
		if scalarType, ok := scalarTypes[sel.Definition.Type.NamedType]; ok {
			x.Type = scalarType.Type
			x.Format = scalarType.Format
			x.Example = scalarType.Example
		}
		fmt.Printf("%s %s\n", strings.Repeat(" ", level*4), sel.Name)
		for _, r := range sel.SelectionSet {
			selRecurse(r, x.Properties, level+1)
		}
		p[sel.Name] = openapi3.NewSchemaRef("", &x)
	}
}
