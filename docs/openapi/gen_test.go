package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	oa "github.com/getkin/kin-openapi/openapi3"
)

func TestGenerateOpenAPI(t *testing.T) {
	outdoc, err := GenerateOpenAPI()
	if err != nil {
		t.Fatal(err)
	}
	// Write output
	jj, _ := json.MarshalIndent(outdoc, "", "  ")
	fmt.Println(string(jj))

	// Validate output
	schema, err := oa.NewLoader().LoadFromData(jj)
	if err != nil {
		t.Fatal(err)
	}
	var validationOpts []oa.ValidationOption
	if err := schema.Validate(context.Background(), validationOpts...); err != nil {
		t.Fatal(err)
	}

}
