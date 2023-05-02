package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"

	openfga "github.com/openfga/go-sdk"
)

type FGAClient struct {
	Model  string
	client *openfga.APIClient
}

func NewFGAClient(modelId string, endpoint string) (*FGAClient, error) {
	ep, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	cfg, err := openfga.NewConfiguration(openfga.Configuration{
		ApiScheme: ep.Scheme,
		ApiHost:   ep.Host,
	})
	if err != nil {
		return nil, err
	}
	apiClient := openfga.NewAPIClient(cfg)
	return &FGAClient{
		Model:  modelId,
		client: apiClient,
	}, nil
}

func (c *FGAClient) Check(ctx context.Context, tk TupleKey) (bool, error) {
	body := openfga.CheckRequest{
		AuthorizationModelId: openfga.PtrString(c.Model),
		TupleKey:             tk.FGATupleKey(),
	}
	data, _, err := c.client.OpenFgaApi.Check(context.Background()).Body(body).Execute()
	if err != nil {
		return false, err
	}
	return data.GetAllowed(), nil
}

func (c *FGAClient) WriteTuple(ctx context.Context, tk TupleKey) error {
	body := openfga.WriteRequest{
		Writes:               &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{tk.FGATupleKey()}},
		AuthorizationModelId: openfga.PtrString(c.Model),
	}
	_, _, err := c.client.OpenFgaApi.Write(context.Background()).Body(body).Execute()
	return err
}

func (c *FGAClient) CreateModel(ctx context.Context, fn string) (string, error) {
	dslJson, err := dslToJson(fn)
	if err != nil {
		return "", err
	}
	var body openfga.WriteAuthorizationModelRequest
	if err := json.Unmarshal([]byte(dslJson), &body); err != nil {
		return "", err
	}
	modelId := ""
	if resp, _, err := c.client.OpenFgaApi.WriteAuthorizationModel(context.Background()).Body(body).Execute(); err != nil {
		return "", err
	} else {
		modelId = resp.GetAuthorizationModelId()
	}
	return modelId, nil
}

func (c *FGAClient) ListObjects(ctx context.Context, tk TupleKey) ([]string, error) {
	body := openfga.ListObjectsRequest{
		AuthorizationModelId: openfga.PtrString(c.Model),
		User:                 tk.User,
		Relation:             tk.Relation,
		Type:                 tk.Object,
	}
	data, _, err := c.client.OpenFgaApi.ListObjects(context.Background()).Body(body).Execute()
	if err != nil {
		return nil, err
	}
	return data.GetObjects(), nil
}

func dslToJson(fn string) (string, error) {
	args := []string{
		"@openfga/syntax-transformer",
		"transform",
		"--from", "dsl",
		"--inputFile", fn,
	}
	cmd := exec.Command("npx", args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		return string(b), err
	}
	return string(b), nil
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
		// t.Log("deleting existing store:", storeId)
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
		// t.Log("created store:", storeId)
		apiClient.SetStoreId(storeId)
	}

	// Create model from DSL
	modelId, err := cc.CreateModel(context.Background(), modelFn)
	if err != nil {
		return "", err
	}

	return modelId, nil
}
