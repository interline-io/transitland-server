package rest

import (
	"fmt"
	"regexp"
	"strings"

	oa "github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/server/gql"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func queryToOAResponses(queryString string) *oa.Responses {
	// Load schema
	schema := gqlout.NewExecutableSchema(gqlout.Config{Resolvers: &gql.Resolver{}})

	// Prepare document
	query, err := gqlparser.LoadQuery(schema.Schema(), queryString)
	if err != nil {
		panic(err)
	}

	///////////
	responseObj := oa.SchemaRef{Value: &oa.Schema{
		Title:      "data",
		Properties: oa.Schemas{},
	}}
	for _, op := range query.Operations {
		for _, sel := range op.SelectionSet {
			queryRecurse(sel, responseObj.Value.Properties, 0)
		}
	}
	desc := "ok"
	res := oa.WithStatus(200, &oa.ResponseRef{Value: &oa.Response{
		Description: &desc,
		Content:     oa.NewContentWithSchemaRef(&responseObj, []string{"application/json"}),
	}})
	ret := oa.NewResponses(res)
	return ret
}

var gqlScalarToOASchema = map[string]oa.Schema{
	"Time": {
		Type:    oa.NewStringSchema().Type,
		Format:  "datetime",
		Example: "2019-11-15T00:45:55.409906",
	},
	"Int": {
		Type: oa.NewIntegerSchema().Type,
	},
	"Float": {
		Type: oa.NewFloat64Schema().Type,
	},
	"String": {
		Type: oa.NewStringSchema().Type,
	},
	"Boolean": {
		Type: oa.NewBoolSchema().Type,
	},
	"ID": {
		Type: oa.NewInt64Schema().Type,
	},
	"Counts":   {},
	"Tags":     {},
	"Geometry": {},
	"Date": {
		Type:    oa.NewStringSchema().Type,
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

func queryRecurse(recurseValue any, parentSchema oa.Schemas, level int) {
	schema := &oa.Schema{
		Properties: oa.Schemas{},
	}
	namedType := ""
	if frag, ok := recurseValue.(*ast.FragmentSpread); ok {
		for _, sel := range frag.Definition.SelectionSet {
			queryRecurse(sel, schema.Properties, level)
		}
		schema.Title = frag.Name
		schema.Description = frag.ObjectDefinition.Description
	} else if field, ok := recurseValue.(*ast.Field); ok {
		for _, sel := range field.SelectionSet {
			queryRecurse(sel, schema.Properties, level+1)
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
	if scalarType, ok := gqlScalarToOASchema[namedType]; ok {
		schema.Type = scalarType.Type
		schema.Format = scalarType.Format
		schema.Example = scalarType.Example
	} else {
		schema.Type = oa.NewObjectSchema().Type
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
		schema.ExternalDocs = &oa.ExternalDocs{URL: doc.URL, Description: doc.Text}
	}
	for _, e := range parsed.Enum {
		schema.Enum = append(schema.Enum, e)
	}

	// Add to parent
	parentSchema[schema.Title] = oa.NewSchemaRef("", schema)
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
