package authz

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os/exec"
	"strings"

	"github.com/interline-io/transitland-lib/log"
	openfga "github.com/openfga/go-sdk"
)

type FGAClient struct {
	ModelID string
	client  *openfga.APIClient
}

func NewFGAClient(endpoint string, storeId string, modelId string) (*FGAClient, error) {
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
		ModelID: modelId,
		client:  apiClient,
	}, nil
}

func (c *FGAClient) Check(ctx context.Context, tk TupleKey, ctxTuples ...TupleKey) (bool, error) {
	if err := tk.Validate(); err != nil {
		return false, err
	}
	var fgaCtxTuples openfga.ContextualTupleKeys
	for _, ctxTuple := range ctxTuples {
		fgaCtxTuples.TupleKeys = append(fgaCtxTuples.TupleKeys, ctxTuple.FGATupleKey())
	}
	body := openfga.CheckRequest{
		AuthorizationModelId: openfga.PtrString(c.ModelID),
		TupleKey:             tk.FGATupleKey(),
		ContextualTuples:     &fgaCtxTuples,
	}
	data, _, err := c.client.OpenFgaApi.Check(context.Background()).Body(body).Execute()
	if err != nil {
		return false, err
	}
	return data.GetAllowed(), nil
}

func (c *FGAClient) ListObjects(ctx context.Context, tk TupleKey) ([]TupleKey, error) {
	body := openfga.ListObjectsRequest{
		AuthorizationModelId: openfga.PtrString(c.ModelID),
		User:                 tk.Subject.String(),
		Relation:             tk.ActionOrRelation(),
		Type:                 tk.Object.Type.String(),
	}
	data, _, err := c.client.OpenFgaApi.ListObjects(context.Background()).Body(body).Execute()
	if err != nil {
		return nil, err
	}
	var ret []TupleKey
	for _, v := range data.GetObjects() {
		ret = append(ret, TupleKey{
			Subject: NewEntityKey(tk.Subject.Type, tk.Subject.Name),
			Object:  NewEntityKeySplit(v),
			Action:  tk.Action,
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

func (c *FGAClient) ReplaceTuple(ctx context.Context, tk TupleKey) error {
	if err := tk.Validate(); err != nil {
		log.Error().Err(err).Str("tk", tk.String()).Msg("ReplaceTuple")
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("ReplaceTuple")

	// Write new tuple before deleting others
	var errs []error
	if err := c.WriteTuple(ctx, tk); err != nil {
		errs = append(errs, err)
	}

	// Delete other tuples
	delKeys, err := c.GetObjectTuples(ctx, NewTupleKey().WithObject(tk.Object.Type, tk.Object.Name))
	if err != nil {
		errs = append(errs, err)
	}
	for _, k := range delKeys {
		if k.Subject.Type == tk.Subject.Type && k.Subject.Name == tk.Subject.Name && k.Relation != tk.Relation {
			if err := c.DeleteTuple(ctx, k); err != nil {
				errs = append(errs, err)
			}
		}
	}
	for _, err := range errs {
		log.Trace().Err(err).Str("tk", tk.String()).Msg("ReplaceTuple")
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (c *FGAClient) WriteTuple(ctx context.Context, tk TupleKey) error {
	if err := tk.Validate(); err != nil {
		log.Error().Str("tk", tk.String()).Msg("WriteTuple")
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("WriteTuple")
	body := openfga.WriteRequest{
		Writes:               &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{tk.FGATupleKey()}},
		AuthorizationModelId: openfga.PtrString(c.ModelID),
	}
	_, _, err := c.client.OpenFgaApi.Write(context.Background()).Body(body).Execute()
	return err
}

func (c *FGAClient) DeleteTuple(ctx context.Context, tk TupleKey) error {
	if err := tk.Validate(); err != nil {
		log.Error().Err(err).Str("tk", tk.String()).Msg("DeleteTuple")
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("DeleteTuple")
	body := openfga.WriteRequest{
		Deletes:              &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{tk.FGATupleKey()}},
		AuthorizationModelId: openfga.PtrString(c.ModelID),
	}
	_, _, err := c.client.OpenFgaApi.Write(context.Background()).Body(body).Execute()
	return err
}

func (c *FGAClient) CreateStore(ctx context.Context, storeName string) (string, error) {
	// Create new store
	resp, _, err := c.client.OpenFgaApi.CreateStore(context.Background()).Body(openfga.CreateStoreRequest{
		Name: storeName,
	}).Execute()
	if err != nil {
		return "", err
	}
	storeId := resp.GetId()
	log.Info().Msgf("created store: %s", storeId)
	c.client.SetStoreId(storeId)
	return storeId, nil
}

func (c *FGAClient) CreateModel(ctx context.Context, fn string) (string, error) {
	// Create new model
	var dslJson []byte
	var err error
	if strings.HasSuffix(fn, ".json") {
		if dslJson, err = ioutil.ReadFile(fn); err != nil {
			return "", err
		}
	} else {
		if dslJson, err = dslToJson(fn); err != nil {
			return "", err
		}
	}
	if err != nil {
		return "", err
	}
	var body openfga.WriteAuthorizationModelRequest
	if err := json.Unmarshal(dslJson, &body); err != nil {
		return "", err
	}
	modelId := ""
	if resp, _, err := c.client.OpenFgaApi.WriteAuthorizationModel(context.Background()).Body(body).Execute(); err != nil {
		return "", err
	} else {
		modelId = resp.GetAuthorizationModelId()
	}
	log.Info().Msgf("created model: %s", modelId)
	return modelId, nil
}

func dslToJson(fn string) ([]byte, error) {
	args := []string{
		"@openfga/syntax-transformer",
		"transform",
		"--from", "dsl",
		"--inputFile", fn,
	}
	cmd := exec.Command("npx", args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return b, err
	}
	return b, nil
}
