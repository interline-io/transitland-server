package rest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func queryToResponses(queryString string) *openapi3.Responses {
	// Load schema
	schema := gqlout.NewExecutableSchema(gqlout.Config{Resolvers: &gql.Resolver{}})

	// Prepare document
	query, err := gqlparser.LoadQuery(schema.Schema(), queryString)
	if err != nil {
		panic(err)
	}

	///////////
	responseObj := openapi3.SchemaRef{Value: &openapi3.Schema{
		Title:      "data",
		Properties: openapi3.Schemas{},
	}}
	for _, op := range query.Operations {
		for _, sel := range op.SelectionSet {
			selRecurse(sel, responseObj.Value.Properties, 0)
		}
	}
	desc := "ok"
	res := openapi3.WithStatus(200, &openapi3.ResponseRef{Value: &openapi3.Response{
		Description: &desc,
		Content:     openapi3.NewContentWithSchemaRef(&responseObj, []string{"application/json"}),
	}})
	ret := openapi3.NewResponses(res)
	return ret
}

var gqlScalarTypes = map[string]openapi3.Schema{
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
	"ID": {
		Type: openapi3.NewInt64Schema().Type,
	},
	"Counts":   {},
	"Tags":     {},
	"Geometry": {},
	"Date": {
		Type:    openapi3.NewStringSchema().Type,
		Format:  "date",
		Example: "2019-11-15",
	},
	"Point":      {},
	"LineString": {},
	"Seconds":    {},
	"Polygon":    {},
	"Map":        {},
	"Any":        {},
	"Upload":     {},
	"Key":        {},
	"Bool":       {},
	"Strings":    {},
}

type ParsedUrl struct {
	Text string
	URL  string
}

type ParsedDocstring struct {
	Text         string
	Type         string
	ExternalDocs []ParsedUrl
	Examples     []string
	Enum         []string
}

var reLinks = regexp.MustCompile(`(\[(?P<text>.+)\]\((?P<url>.+)\))`)
var reAnno = regexp.MustCompile(`(\[(?P<annotype>.+):(?P<value>.+)\])`)

func ParseDocstring(v string) ParsedDocstring {
	ret := ParsedDocstring{}
	for _, matchGroup := range parseGroups(reLinks, v) {
		text := matchGroup["text"]
		url := matchGroup["url"]
		ret.ExternalDocs = append(ret.ExternalDocs, ParsedUrl{URL: url, Text: text})
	}
	for _, matchGroup := range parseGroups(reAnno, v) {
		annotype := matchGroup["annotype"]
		value := strings.TrimSpace(matchGroup["value"])
		switch annotype {
		case "example":
			ret.Examples = append(ret.Examples, value)
		case "see":
			ret.ExternalDocs = append(ret.ExternalDocs, ParsedUrl{URL: value})
		case "enum":
			for _, e := range strings.Split(value, ",") {
				ret.Enum = append(ret.Enum, strings.TrimSpace(e))
			}
		}
	}
	ret.Text = strings.TrimSpace(reAnno.ReplaceAllString(v, ""))
	return ret
}

func selRecurse(recurseValue any, parentSchema openapi3.Schemas, level int) {
	schema := &openapi3.Schema{
		Properties: openapi3.Schemas{},
	}
	namedType := ""
	if frag, ok := recurseValue.(*ast.FragmentSpread); ok {
		for _, sel := range frag.Definition.SelectionSet {
			selRecurse(sel, schema.Properties, level)
		}
		schema.Title = frag.Name
		schema.Description = frag.ObjectDefinition.Description
	} else if field, ok := recurseValue.(*ast.Field); ok {
		for _, sel := range field.SelectionSet {
			selRecurse(sel, schema.Properties, level+1)
		}
		schema.Title = field.Name
		schema.Description = field.Definition.Description
		schema.Nullable = !field.Definition.Type.NonNull
		namedType = field.Definition.Type.NamedType
	} else {
		return
	}

	fmt.Printf("%s %s\n", strings.Repeat(" ", level*4), schema.Title)

	// Scalar types
	if scalarType, ok := gqlScalarTypes[namedType]; ok {
		schema.Type = scalarType.Type
		schema.Format = scalarType.Format
		schema.Example = scalarType.Example
	} else {
		schema.Type = openapi3.NewObjectSchema().Type
	}

	// Parse docstring
	parsed := ParseDocstring(schema.Description)
	if parsed.Text != "" {
		schema.Description = parsed.Text
	}
	for _, example := range parsed.Examples {
		schema.Example = example
	}
	for _, doc := range parsed.ExternalDocs {
		schema.ExternalDocs = &openapi3.ExternalDocs{URL: doc.URL, Description: doc.Text}
	}
	for _, e := range parsed.Enum {
		schema.Enum = append(schema.Enum, e)
	}

	// Add to parent
	parentSchema[schema.Title] = openapi3.NewSchemaRef("", schema)
}

func parseGroups(re *regexp.Regexp, v string) []map[string]string {
	var ret []map[string]string
	for _, match := range re.FindAllStringSubmatch(v, -1) {
		group := map[string]string{}
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				group[name] = match[i]
			}
		}
		ret = append(ret, group)
	}
	return ret
}
