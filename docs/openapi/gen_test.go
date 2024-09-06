package openapi

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	oa "github.com/getkin/kin-openapi/openapi3"
	"github.com/interline-io/transitland-server/server/rest"
)

func TestGenerateOpenAPI(t *testing.T) {
	outdoc, err := rest.GenerateOpenAPI()
	if err != nil {
		t.Fatal(err)
	}

	// Write output
	jj, _ := json.MarshalIndent(outdoc, "", "  ")
	f, err := os.Create("rest.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(jj)
	f.Close()

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
