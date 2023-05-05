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

	t.Run("GetObjectTuples", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "get" {
				continue
			}
			tkKey := tk.TupleKey
			t.Run(tkKey.String(), func(t *testing.T) {
				tks, err := fgac.GetObjectTuples(context.Background(), tkKey)
				if err != nil {
					t.Error(err)
				}
				expect := strings.Split(tk.Expect, " ")
				var got []string
				for _, v := range tks {
					got = append(got, fmt.Sprintf("%s:%s:%s", v.UserType, v.UserName, v.Relation))
				}
				assert.ElementsMatch(t, expect, got, "usertype:username:relation does not match")

			})
		}
	})

	t.Run("Check", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "check" {
				continue
			}
			for _, checkAction := range tk.Checks {
				tkKey := tk.TupleKey
				expect := true
				if strings.HasPrefix(checkAction, "+") {
					checkAction = strings.TrimPrefix(checkAction, "+")
				} else if strings.HasPrefix(checkAction, "-") {
					expect = false
					checkAction = strings.TrimPrefix(checkAction, "-")
				}
				var err error
				tkKey.Action, err = ActionString(checkAction)
				if err != nil {
					t.Fatal(err)
				}
				t.Run(tkKey.String(), func(t *testing.T) {
					ok, err := fgac.Check(context.Background(), tkKey)
					if err != nil {
						t.Fatal(err)
					}
					if ok && !expect {
						t.Errorf("got %t, expected %t", ok, expect)
					}
					if !ok && expect {
						t.Errorf("got %t, expected %t", ok, expect)
					}
				})
			}

		}
	})

	t.Run("ListObjects", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "list" {
				continue
			}
			for _, checkAction := range tk.Checks {
				tkKey := tk.TupleKey
				tkKey.Action, err = ActionString(checkAction)
				if err != nil {
					t.Fatal(err)
				}
				t.Run(tkKey.String(), func(t *testing.T) {
					objs, err := fgac.ListObjects(context.Background(), tkKey)
					if err != nil {
						t.Fatal(err)
					}
					var gotIds []string
					for _, v := range objs {
						gotIds = append(gotIds, v.ObjectName)
					}
					expIds := strings.Split(tk.Expect, " ")
					assert.ElementsMatch(t, expIds, gotIds, "object ids")
				})
			}
		}
	})

	t.Run("WriteTuple", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "write" {
				continue
			}
			tkKey := tk.TupleKey
			t.Run(tkKey.String(), func(t *testing.T) {
				// Write tuple and check if error was expected
				expectOk := true
				if tk.Expect == "fail" {
					expectOk = false
				}
				err := fgac.WriteTuple(context.Background(), tkKey)
				if err != nil && expectOk {
					t.Errorf("got error %s, expected ok", err.Error())
				}
				if err == nil && !expectOk {
					t.Errorf("no error, expected error")
				}
				// Check was written
				tks, err := fgac.GetObjectTuples(context.Background(), tkKey)
				if err != nil {
					t.Error(err)
				}
				var gotTks []string
				for _, v := range tks {
					gotTks = append(gotTks, fmt.Sprintf("%s:%s:%s", v.UserType, v.UserName, v.Relation))
				}
				checkTk := fmt.Sprintf("%s:%s:%s", tkKey.UserType, tkKey.UserName, tkKey.Relation)
				assert.Contains(t, gotTks, checkTk, "written tuple not found in updated object tuples")
			})
		}
	})

	t.Run("DeleteTuple", func(t *testing.T) {
		for _, tk := range checks {
			if tk.Test != "delete" {
				continue
			}
			tkKey := tk.TupleKey
			t.Run(tkKey.String(), func(t *testing.T) {
				expectOk := true
				if tk.Expect == "fail" {
					expectOk = false
				}
				err := fgac.DeleteTuple(context.Background(), tkKey)
				if err != nil && expectOk {
					t.Errorf("got error %s, expected ok", err.Error())
				}
				if err == nil && !expectOk {
					t.Errorf("no error, expected error")
				}
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
			if tk.Test != "" {
				continue
			}
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

	// Create new store
	resp, _, err := apiClient.OpenFgaApi.CreateStore(context.Background()).Body(openfga.CreateStoreRequest{
		Name: storeName,
	}).Execute()
	if err != nil {
		return "", err
	}
	storeId := resp.GetId()
	log.Infof("created store: %s", storeId)
	apiClient.SetStoreId(storeId)

	// Create model from DSL
	modelId, err := cc.CreateModel(context.Background(), modelFn)
	if err != nil {
		return "", err
	}
	log.Infof("created model: %s", modelId)
	return modelId, nil
}
