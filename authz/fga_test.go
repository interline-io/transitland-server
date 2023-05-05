package authz

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/interline-io/transitland-lib/log"
	openfga "github.com/openfga/go-sdk"
	"github.com/stretchr/testify/assert"
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
	checks, err := LoadTuples("../test/authz/tls.csv")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("check", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "check" {
				continue
			}
			for _, checkAction := range tk.Checks {
				tk2 := tk
				tk2.TupleKey.Action, _ = ActionString(checkAction)
				t.Run(tk.String(), func(t *testing.T) {
					ok, err := fgac.Check(context.Background(), tk.TupleKey)
					if err != nil {
						t.Fatal(err)
					}
					if ok && tk.Expect != "true" {
						t.Errorf("got %t, expected %s", ok, tk.Expect)
					}
					if !ok && tk.Expect != "false" {
						t.Errorf("got %t, expected %s", ok, tk.Expect)
					}
				})

			}

		}
	})

	t.Run("list", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "list" {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				objs, err := fgac.ListObjects(context.Background(), tk.TupleKey)
				if err != nil {
					t.Fatal(err)
				}
				var gotIds []string
				for _, v := range objs {
					gotIds = append(gotIds, v.ObjectName)
				}
				expIds := strings.Split(tk.Expect, "-")
				assert.ElementsMatch(t, expIds, gotIds, "object ids")
			})
		}
	})
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
	}
	if cfg.FGATestTuplesPath != "" {
		tkeys, err := LoadTuples(cfg.FGATestTuplesPath)
		if err != nil {
			return nil, err
		}
		for _, tk := range tkeys {
			if !tk.Relation.IsARelation() {
				continue
			}
			if err := fgac.WriteTuple(context.Background(), tk.TupleKey); err != nil {
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
