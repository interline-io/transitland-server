package openapi

import (
	"context"
	"encoding/json"
	"os"
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

	// Validate output
	schema, err := oa.NewLoader().LoadFromData(jj)
	if err != nil {
		t.Fatal(err)
	}
	var validationOpts []oa.ValidationOption
	if err := schema.Validate(context.Background(), validationOpts...); err != nil {
		t.Fatal(err)
	}
	outf, err := os.Create("rest.json")
	if err != nil {
		t.Fatal(err)
	}
	outf.Write(jj)
}
