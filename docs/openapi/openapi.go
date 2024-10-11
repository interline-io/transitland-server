package main

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

func queryToOAResponses(queryString string) (*oa.Responses, error) {
	// Load schema
	schema := gqlout.NewExecutableSchema(gqlout.Config{Resolvers: &gql.Resolver{}})
	gs := schema.Schema()

	// Prepare document
	query, err := gqlparser.LoadQuery(gs, queryString)
	if err != nil {
		return nil, err
	}

	///////////
	responseObj := oa.SchemaRef{Value: &oa.Schema{
		Title:      "data",
		Properties: oa.Schemas{},
	}}
	for _, op := range query.Operations {
		for selOrder, sel := range op.SelectionSet {
			queryRecurse(gs, sel, responseObj.Value.Properties, 0, selOrder)
		}
	}
	desc := "ok"
	res := oa.WithStatus(200, &oa.ResponseRef{Value: &oa.Response{
		Description: &desc,
		Content:     oa.NewContentWithSchemaRef(&responseObj, []string{"application/json"}),
	}})
	ret := oa.NewResponses(res)
	return ret, nil
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
	"Counts": {
		Type: oa.NewObjectSchema().Type,
	},
	"Tags": {
		Type: oa.NewObjectSchema().Type,
	},
	"Date": {
		Type:    oa.NewStringSchema().Type,
		Format:  "date",
		Example: "2019-11-15",
	},
	"Seconds": {
		Type:    oa.NewStringSchema().Type,
		Format:  "hms",
		Example: "15:21:04",
	},
	"Map": {
		Type: oa.NewObjectSchema().Type,
	},
	"Bool": {
		Type: oa.NewBoolSchema().Type,
	},
	"Strings": {
		Type: oa.NewArraySchema().Type,
	},
	"Any":        {},
	"Upload":     {},
	"Key":        {},
	"Polygon":    {},
	"Geometry":   {},
	"Point":      {},
	"LineString": {},
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
	Hide         bool
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
		case "hide":
			ret.Hide = true
		}
	}
	ret.Text = strings.TrimSpace(reAnno.ReplaceAllString(v, ""))
	return ret
}

func queryRecurse(gs *ast.Schema, recurseValue any, parentSchema oa.Schemas, level int, order int) int {
	schema := &oa.Schema{
		Properties: oa.Schemas{},
		Extensions: map[string]any{},
	}
	gqlType := ""
	namedType := ""
	isArray := false
	if field, ok := recurseValue.(*ast.Field); ok {
		if field.Comment != nil {
			for _, c := range field.Comment.List {
				pd := ParseDocstring(c.Value)
				if pd.Hide {
					return order
				}
			}
		}
		schema.Title = field.Name
		schema.Description = field.Definition.Description
		schema.Nullable = !field.Definition.Type.NonNull
		namedType = field.Definition.Type.NamedType
		gqlType = field.Definition.Type.NamedType
		if field.Definition.Type.Elem != nil {
			gqlType = field.Definition.Type.Elem.Name()

		}
		if gst, ok := gs.Types[field.Definition.Type.String()]; ok {
			for _, ev := range gst.EnumValues {
				schema.Enum = append(schema.Enum, ev.Name)
			}
		}
		if strings.HasPrefix(field.Definition.Type.String(), "[") {
			isArray = true
		}
		for _, sel := range field.SelectionSet {
			order = queryRecurse(gs, sel, schema.Properties, level+1, order+1)
		}
	} else if frag, ok := recurseValue.(*ast.FragmentSpread); ok {
		for _, sel := range frag.Definition.SelectionSet {
			// Ugly hack to put fragments at the end of the selection set
			order = queryRecurse(gs, sel, parentSchema, level, order+1)
		}
		return order
	} else {
		return order
	}

	fmt.Printf("%s %s (%s : %s : order %d)\n", strings.Repeat(" ", level*4), schema.Title, namedType, gqlType, order)
	order += 1
	schema.Extensions["x-order"] = order

	// Scalar types
	if scalarType, ok := gqlScalarToOASchema[namedType]; ok {
		schema.Type = scalarType.Type
		schema.Format = scalarType.Format
		schema.Example = scalarType.Example
	} else {
		schema.Type = oa.NewObjectSchema().Type
		if gqlType != "" {
			schema.Extensions["x-graphql-type"] = gqlType
		}
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

	if isArray {
		innerSchema := &oa.Schema{
			Properties: schema.Properties,
			Type:       schema.Type,
			Extensions: schema.Extensions,
		}
		outerSchema := &oa.Schema{
			Title:        schema.Title,
			Description:  schema.Description,
			Nullable:     schema.Nullable,
			Type:         oa.NewArraySchema().Type,
			ExternalDocs: schema.ExternalDocs,
			Enum:         schema.Enum,
			Items:        oa.NewSchemaRef("", innerSchema),
			Extensions:   schema.Extensions,
		}
		schema = outerSchema
	}

	// Add to parent
	parentSchema[schema.Title] = oa.NewSchemaRef("", schema)
	return order
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
