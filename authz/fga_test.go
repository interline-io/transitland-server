package authz

import (
	"context"
	"fmt"
	"testing"

	"github.com/interline-io/transitland-lib/log"
	openfga "github.com/openfga/go-sdk"
)

func TestFGAClient(t *testing.T) {
	cfg := newTestConfig()
	cfg.FGAEndpoint = "http://localhost:8090"
	cfg.FGATestModelPath = "../test/authz/tls.model"
	cfg.FGATestTuplesPath = "../test/authz/tls.csv"
	fgac, err := newTestFGAClient(t, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Test assertions
	checks, err := LoadTuples("../test/authz/assert.csv")
	if err != nil {
		t.Fatal(err)
	}
	for _, tk := range checks {
		ok, err := fgac.Check(context.Background(), tk)
		if err != nil {
			t.Fatal(err)
		}
		if ok != tk.Assert {
			t.Errorf("%s/%s/%s/%s got %t, expected %t", tk.UserName, tk.ObjectType, tk.ObjectName, tk.Relation, ok, tk.Assert)
		}
	}

	// List objects
	checkusers := []string{"admin", "ian", "drew", "kapeel"}
	rels := []string{"can_view", "can_edit"}
	for _, user := range checkusers {
		for _, rel := range rels {
			tk := TupleKey{ObjectType: "feed", Relation: "can_view", UserType: "user", UserName: user}
			if objs, err := fgac.ListObjects(context.Background(), tk); err != nil {
				t.Fatal(err)
			} else {
				t.Logf("user %s: %s: %v", user, rel, objs)
			}
		}
	}
}

func newTestFGAClient(t testing.TB, cfg AuthzConfig) (*FGAClient, error) {
	fgac, err := NewFGAClient(cfg.FGAStoreID, cfg.FGAModelID, cfg.FGAEndpoint)
	if err != nil {
		return nil, err
	}
	if cfg.FGATestModelPath != "" {
		modelId, err := createTestStoreAndModel(fgac, "test", cfg.FGATestModelPath, true)
		if err != nil {
			return nil, err
		}
		fgac.Model = modelId
		t.Log("using FGA model:", modelId)
	}
	if cfg.FGATestTuplesPath != "" {
		tkeys, err := LoadTuples(cfg.FGATestTuplesPath)
		if err != nil {
			return nil, err
		}
		for _, tk := range tkeys {
			if err := fgac.WriteTuple(context.Background(), tk); err != nil {
				return nil, err
			}
		}
	}
	return fgac, nil
}

func createTestStoreAndModel(cc *FGAClient, storeName string, modelFn string, deleteExisting bool) (string, error) {
	// Configure API client
	apiClient := cc.client

	// Find store
	storeId := ""
	if stores, _, err := apiClient.OpenFgaApi.ListStores(context.Background()).Execute(); err != nil {
		return "", err
	} else {
		for _, store := range stores.GetStores() {
			if store.GetName() == storeName {
				storeId = store.GetId()
			}
		}
	}

	// Delete existing store
	if storeId != "" && deleteExisting {
		log.Tracef("deleting existing store: %s", storeId)
		apiClient.SetStoreId(storeId)
		if _, err := apiClient.OpenFgaApi.DeleteStore(context.Background()).Execute(); err != nil {
			return "", err
		}
		storeId = ""
	}

	// Create new store
	if storeId == "" {
		resp, _, err := apiClient.OpenFgaApi.CreateStore(context.Background()).Body(openfga.CreateStoreRequest{
			Name: storeName,
		}).Execute()
		if err != nil {
			return "", err
		}
		storeId = resp.GetId()
		log.Tracef("created store: %s", storeId)
		apiClient.SetStoreId(storeId)
	}

	fmt.Println("store:", storeId)

	// Create model from DSL
	modelId, err := cc.CreateModel(context.Background(), modelFn)
	if err != nil {
		return "", err
	}
	log.Tracef("created model: %s", modelId)
	return modelId, nil
}
