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

func GenerateOpenAPI() (*oa.T, error) {
	type RestHandlers interface {
		RequestInfo() RequestInfo
	}

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
		Schemas:    oa.Schemas{},
	}
	for paramName, paramRef := range ParameterComponents {
		outdoc.Components.Parameters[paramName] = paramRef
	}

	// Add type components
	schema := gqlout.NewExecutableSchema(gqlout.Config{Resolvers: &gql.Resolver{}})
	for schemaName, schemaTypeFn := range gqlScalarToOASchema {
		outdoc.Components.Schemas[schemaName] = oa.NewSchemaRef("", schemaTypeFn())
	}
	for _, schemaType := range schema.Schema().Types {
		fmt.Println("SCHEMA TYPE:", schemaType.Name)
		if strings.HasPrefix(schemaType.Name, "_") {
			continue
		}
		if schemaType.BuiltIn {
			continue
		}
		if schemaType.Kind == ast.Scalar {
			continue
		}
		if schemaType.Kind == ast.InputObject {
			continue
		}
		fmt.Println("\t\tok")
		gqlRecurse(schemaType, outdoc.Components.Schemas, 0, 0)
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
	}
	for _, handler := range handlers {
		requestInfo := handler.RequestInfo()
		pathOpts = append(pathOpts, oa.WithPath(requestInfo.Path, requestInfo.PathItem))
	}
	outdoc.Paths = oa.NewPaths(pathOpts...)
	return outdoc, nil
}

func generateOpenAPIResponses(queryString string) *oa.Responses {
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
			gqlRecurse(sel, responseObj.Value.Properties, 0, 0)
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

var gqlScalarToOASchema = map[string]func() *oa.Schema{
	"String":  func() *oa.Schema { return oa.NewStringSchema() },
	"Int":     func() *oa.Schema { return oa.NewIntegerSchema() },
	"Float":   func() *oa.Schema { return oa.NewFloat64Schema() },
	"Boolean": func() *oa.Schema { return oa.NewBoolSchema() },
	"Bool":    func() *oa.Schema { return oa.NewObjectSchema() },
	"ID":      func() *oa.Schema { return oa.NewInt64Schema() },
	"Time": func() *oa.Schema {
		return &oa.Schema{
			Type:    oa.NewStringSchema().Type,
			Format:  "date-time",
			Example: "2019-11-15T00:45:55.409906",
		}
	},
	"Date": func() *oa.Schema {
		return &oa.Schema{
			Type:    oa.NewStringSchema().Type,
			Format:  "date",
			Example: "2019-11-15",
		}
	},
	"Seconds": func() *oa.Schema {
		return oa.NewStringSchema().WithFormat("seconds")
	},
	"Tags": func() *oa.Schema {
		return oa.NewObjectSchema().WithAdditionalProperties(oa.NewStringSchema())
	},
	"Counts": func() *oa.Schema {
		return oa.NewObjectSchema().WithAdditionalProperties(oa.NewIntegerSchema())
	},
	"Strings": func() *oa.Schema {
		return oa.NewArraySchema().WithItems(oa.NewStringSchema())
	},
	"Geometry": func() *oa.Schema {
		return oa.NewObjectSchema().
			WithProperty("type", oa.NewStringSchema().WithEnum("Point", "LineString", "Polygon", "MultiLineString", "MultiPolygon"))
	},
	"Point": func() *oa.Schema {
		return oa.NewObjectSchema().
			WithProperty("type", oa.NewStringSchema().WithEnum("Point")).
			WithProperty("coordinates", oa.NewArraySchema().WithItems(oa.NewFloat64Schema()))
	},
	"LineString": func() *oa.Schema {
		return oa.NewObjectSchema().
			WithProperty("type", oa.NewStringSchema().WithEnum("LineString")).
			WithProperty("coordinates", oa.NewArraySchema().WithItems(oa.NewArraySchema().WithItems(oa.NewFloat64Schema())))
	},
	"Polygon": func() *oa.Schema {
		return oa.NewObjectSchema().
			WithProperty("type", oa.NewStringSchema().WithEnum("Polygon")).
			WithProperty("coordinates", oa.NewArraySchema().WithItems(oa.NewArraySchema().WithItems(oa.NewArraySchema().WithItems(oa.NewFloat64Schema()))))

	},
	"Map": func() *oa.Schema {
		return oa.NewObjectSchema().WithAnyAdditionalProperties()
	},
	"Any": func() *oa.Schema {
		return oa.NewAnyOfSchema(
			oa.NewStringSchema(),
			oa.NewIntegerSchema(),
			oa.NewFloat64Schema(),
			oa.NewBoolSchema(),
			oa.NewBytesSchema(),
			oa.NewObjectSchema(),
		)
	},
}

func baseType(t *ast.Type) (string, bool) {
	if t.NamedType == "" {
		return strings.Replace(t.Elem.NamedType, "!", "", 1), true
	}
	return strings.Replace(t.NamedType, "!", "", 1), false
}

func gqlRecurse(recurseValue any, parentSchema oa.Schemas, level int, order int) {
	indent := strings.Repeat(" ", level*4)
	schema := &oa.Schema{
		Properties: oa.Schemas{},
		Extensions: map[string]any{},
	}
	sref := &oa.SchemaRef{
		Value:      schema,
		Extensions: map[string]any{"x-order": order},
	}
	switch field := recurseValue.(type) {
	case *ast.Field:
		fmt.Printf("%s %s %T\n", indent, field.Alias, field)
		schemaSetType(sref, field.Definition.Type, false)
		childProps := schema.Properties
		if field.Definition.Type.NamedType == "" {
			childProps = schema.Items.Value.Properties
		}
		for i, sel := range field.SelectionSet {
			gqlRecurse(sel, childProps, level+1, i)
		}
		schema.Title = field.Alias
		schema.Description = field.Definition.Description
		schema.Nullable = !field.Definition.Type.NonNull
	case *ast.FieldDefinition:
		fmt.Printf("%s %s %T\n", indent, field.Name, field)
		schemaSetType(sref, field.Type, true)
		schema.Title = field.Name
		schema.Description = field.Description
	case *ast.Definition:
		fmt.Printf("%s %s %T\n", indent, field.Name, field)
		for i, sel := range field.Fields {
			gqlRecurse(sel, schema.Properties, level+1, i)
		}
		schema.Title = field.Name
		schema.Description = field.Description
	case *ast.FragmentSpread:
		fmt.Printf("%s %s %T\n", indent, field.Name, field)
		for i, sel := range field.Definition.SelectionSet {
			gqlRecurse(sel, schema.Properties, level, i)
		}
		schema.Title = field.Name
		schema.Description = field.ObjectDefinition.Description
	default:
		fmt.Printf("%s unknown recurse type: %T\n", indent, recurseValue)
		return
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
	parentSchema[schema.Title] = sref
}

func schemaSetType(sref *oa.SchemaRef, t *ast.Type, allowRef bool) {
	namedType, isArray := baseType(t)
	if sref.Extensions == nil {
		sref.Extensions = map[string]any{}
	}
	sref.Extensions["x-graphql-type"] = t.String()
	if isArray {
		sref.Value.Type = &oa.Types{"array"}

		itemSref := &oa.SchemaRef{Value: &oa.Schema{
			Properties: oa.Schemas{},
			Extensions: map[string]any{},
		}}
		schemaSetType(itemSref, t.Elem, allowRef)
		sref.Value.Items = itemSref
	}
	if scalarTypeFn, ok := gqlScalarToOASchema[namedType]; ok {
		scalarType := scalarTypeFn()
		if sref.Extensions == nil {
			sref.Extensions = map[string]any{}
		}
		sref.Extensions["x-graphql-type"] = t.String()
		sref.Value.Type = scalarType.Type
		sref.Value.Format = scalarType.Format
		sref.Value.Example = scalarType.Example
		if scalarType.Items != nil {
			sref.Value.Items = scalarType.Items
		}
	} else if allowRef && !isArray && !strings.HasPrefix(namedType, "_") {
		sref.Ref = fmt.Sprintf("#/components/schemas/%s", namedType)
	}
}

////////

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
