package authn

import (
	"context"
	"fmt"
	"os"
	"testing"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/credentials"
)

func TestFGA(t *testing.T) {
	configuration, err := openfga.NewConfiguration(openfga.Configuration{
		// ApiScheme: os.Getenv("FGA_API_SCHEME"), // Optional. Can be "http" or "https". Defaults to "https"
		ApiHost: os.Getenv("FGA_API_HOST"), // required, define without the scheme (e.g. api.openfga.example instead of https://api.openfga.example)
		StoreId: os.Getenv("FGA_STORE_ID"),
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: os.Getenv("FGA_BEARER_TOKEN"), // will be passed as the "Authorization: Bearer ${ApiToken}" request header
			},
		},
	})

	if err != nil {
		t.Fatal(err)
		// .. Handle error
	}

	apiClient := openfga.NewAPIClient(configuration)

	body := openfga.ListObjectsRequest{
		User:     "user:ian",
		Relation: "can_view",
		Type:     "feed",
	}

	data, response, err := apiClient.OpenFgaApi.ListObjects(context.Background()).Body(body).Execute()
	_ = data
	_ = response
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("data:", data.GetObjects())
	fmt.Println("resp:", response)

	// data = { "objects": ["document:otherdoc", "document:planning"] }
}
