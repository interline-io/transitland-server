package rest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenapiLoader(t *testing.T) {
	newSchemaText := func(sc *openapi3.Schema) string {
		enumText := "nil"
		if len(sc.Enum) > 0 {
			var a = []string{}
			for _, b := range sc.Enum {
				a = append(a, fmt.Sprintf(`"%v"`, b))
			}
			enumText = fmt.Sprintf(`[]any{%s}`, strings.Join(a, ","))
		}
		stype := "string"
		if a := sc.Type.Slice(); len(a) > 0 {
			stype = a[0]
		}
		return fmt.Sprintf(`newSchemaRef("%s", %s)`, stype, enumText)
	}

	schema, err := openapi3.NewLoader().LoadFromFile("../../docs/openapi/rest.json")
	if err != nil {
		t.Fatal(err)
	}

	var comps = []string{}
	for k, c := range schema.Components.Parameters {
		val := c.Value
		sval := val.Schema.Value
		fmt.Println("sval.Type:", sval.Type)
		schemaText := newSchemaText(sval)
		paramText := fmt.Sprintf(`
		"%s": &ParameterRef{Value: &Parameter{
			In: "%s",
			Description: "%s",
			Schema: %s,
		}}`,
			k,
			val.In,
			val.Description,
			schemaText,
		)
		comps = append(comps, paramText)
	}
	fmt.Printf(`var ComponentReferences = openapi3.ParametersMap{%s}`, strings.Join(comps, ","))

	for _, c := range schema.Paths.InMatchingOrder() {
		path := schema.Paths.Find(c)
		fmt.Println("c:", c, "path:", path)
		fmt.Println(
			path.Summary,
			path.Description,
		)
		for _, param := range path.Get.Parameters {
			val := param.Value
			fmt.Println(
				"param:",
				val.Name, "ref:", param.Ref,
				"examples:", val.Examples,
			)
		}

	}

}
