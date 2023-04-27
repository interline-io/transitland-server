package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	openfga "github.com/openfga/go-sdk"
)

type Client struct {
	Model  string
	client *openfga.APIClient
}

func (c *Client) Check(ctx context.Context, tk TupleKey) (bool, error) {
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

func (c *Client) WriteTuple(ctx context.Context, tk TupleKey) error {
	body := openfga.WriteRequest{
		Writes:               &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{tk.FGATupleKey()}},
		AuthorizationModelId: openfga.PtrString(c.Model),
	}
	_, _, err := c.client.OpenFgaApi.Write(context.Background()).Body(body).Execute()
	return err
}

func (c *Client) CreateModel(ctx context.Context, fn string) (string, error) {
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

func (c *Client) ListObjects(ctx context.Context, tk TupleKey) ([]string, error) {
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
	// var ret []string
	// for _, v := range data.GetObjects() {
	// 	ret = append(ret, v.GetId())
	// }
	// return ret, nil
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
