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

func NewFGAClient(storeId string, modelId string, endpoint string) (*FGAClient, error) {
	ep, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	cfg, err := openfga.NewConfiguration(openfga.Configuration{
		ApiScheme: ep.Scheme,
		ApiHost:   ep.Host,
		StoreId:   storeId,
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
	if err := tk.Validate(); err != nil {
		return false, err
	}
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
	if err := tk.Validate(); err != nil {
		return err
	}
	body := openfga.WriteRequest{
		Writes:               &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{tk.FGATupleKey()}},
		AuthorizationModelId: openfga.PtrString(c.Model),
	}
	_, _, err := c.client.OpenFgaApi.Write(context.Background()).Body(body).Execute()
	return err
}

func (c *FGAClient) DeleteTuple(ctx context.Context, tk TupleKey) error {
	if err := tk.Validate(); err != nil {
		return err
	}
	body := openfga.WriteRequest{
		Deletes:              &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{tk.FGATupleKey()}},
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

func (c *FGAClient) ListObjects(ctx context.Context, tk TupleKey) ([]TupleKey, error) {
	body := openfga.ListObjectsRequest{
		AuthorizationModelId: openfga.PtrString(c.Model),
		User:                 cunsplit(tk.UserType, tk.UserName),
		Relation:             tk.Action.String(),
		Type:                 tk.ObjectType.String(),
	}
	data, _, err := c.client.OpenFgaApi.ListObjects(context.Background()).Body(body).Execute()
	if err != nil {
		return nil, err
	}
	var ret []TupleKey
	for _, v := range data.GetObjects() {
		okey := csplit(v)
		ret = append(ret, TupleKey{
			UserType:   tk.UserType,
			UserName:   tk.UserName,
			ObjectType: okey.Type,
			ObjectName: okey.Name,
			Action:     tk.Action,
		})
	}
	return ret, nil
}

func (c *FGAClient) GetObjectTuples(ctx context.Context, tk TupleKey) ([]TupleKey, error) {
	if err := tk.Validate(); err != nil {
		return nil, err
	}
	fgatk := tk.FGATupleKey()
	body := openfga.ReadRequest{
		TupleKey: &fgatk,
	}
	data, _, err := c.client.OpenFgaApi.Read(ctx).Body(body).Execute()
	if err != nil {
		return nil, err
	}
	var ret []TupleKey
	for _, t := range *data.Tuples {
		ret = append(ret, fromFGATupleKey(t.GetKey()))
	}
	return ret, nil
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
