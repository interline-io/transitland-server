package authn

import (
	"context"
	"strings"
	"testing"

	openfga "github.com/openfga/go-sdk"
)

func TestFGA(t *testing.T) {
	cc, err := createTestClient(t, "test", true)
	if err != nil {
		t.Fatal(err)
	}

	// Add test tuples
	tkeys, err := LoadTuples("../test/authn/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	for _, tk := range tkeys {
		if err := cc.WriteTuple(context.Background(), tk); err != nil {
			t.Fatal(err)
		}
	}

	// Test assertions
	checks, err := LoadTuples("../test/authn/assert.csv")
	if err != nil {
		t.Fatal(err)
	}
	for _, tk := range checks {
		ok, err := cc.Check(context.Background(), tk)
		if err != nil {
			t.Fatal(err)
		}
		if ok != tk.Assert {
			t.Errorf("%s/%s/%s got %t, expected %t", tk.User, tk.Object, tk.Relation, ok, tk.Assert)
		}
	}

	// List objects
	checkusers := []string{"admin", "ian", "drew", "kapeel"}
	rels := []string{"can_view", "can_edit"}
	for _, user := range checkusers {
		for _, rel := range rels {
			if objs, err := cc.ListObjects(context.Background(), TupleKey{User: "user:" + user, Object: "feed", Relation: rel}); err != nil {
				t.Fatal(err)
			} else {
				t.Logf("user %s: %s: %s", user, rel, strings.Join(objs, " "))
			}
		}
	}
}

func createTestClient(t *testing.T, storeName string, deleteExisting bool) (*Client, error) {
	// Configure API client
	cfg, err := openfga.NewConfiguration(openfga.Configuration{
		ApiScheme: "http",
		ApiHost:   "localhost:8080",
	})
	if err != nil {
		t.Fatal(err)
	}
	apiClient := openfga.NewAPIClient(cfg)

	// Find store
	storeId := ""
	if stores, _, err := apiClient.OpenFgaApi.ListStores(context.Background()).Execute(); err != nil {
		t.Fatal(err)
	} else {
		for _, store := range stores.GetStores() {
			if store.GetName() == storeName {
				storeId = store.GetId()
			}
		}
	}

	// Delete existing store
	if storeId != "" && deleteExisting {
		t.Log("deleting existing store:", storeId)
		apiClient.SetStoreId(storeId)
		if _, err := apiClient.OpenFgaApi.DeleteStore(context.Background()).Execute(); err != nil {
			t.Fatal(err)
		}
		storeId = ""
	}

	// Create new store
	if storeId == "" {
		resp, _, err := apiClient.OpenFgaApi.CreateStore(context.Background()).Body(openfga.CreateStoreRequest{
			Name: storeName,
		}).Execute()
		if err != nil {
			t.Fatal(err)
		}
		storeId = resp.GetId()
		t.Log("created store:", storeId)
		apiClient.SetStoreId(storeId)
	}

	cc := Client{Model: "", client: apiClient}

	// Create model from DSL
	modelId, err := cc.CreateModel(context.Background(), "../test/authn/test.model")
	if err != nil {
		return nil, err
	}
	cc.Model = modelId

	return &cc, nil
}
