package authz

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/url"

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
		fgaCtxTuples.TupleKeys = append(fgaCtxTuples.TupleKeys, ToFGATupleKey(ctxTuple))
	}
	body := openfga.CheckRequest{
		AuthorizationModelId: openfga.PtrString(c.ModelID),
		TupleKey:             ToFGATupleKey(tk),
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
	fgatk := ToFGATupleKey(tk)
	body := openfga.ReadRequest{
		TupleKey: &fgatk,
	}

	data, _, err := c.client.OpenFgaApi.Read(ctx).Body(body).Execute()
	if err != nil {
		return nil, err
	}
	var ret []TupleKey
	for _, t := range *data.Tuples {
		ret = append(ret, FromFGATupleKey(t.GetKey()))
	}
	return ret, nil
}

func (c *FGAClient) ReplaceAllRelation(ctx context.Context, tk TupleKey) error {
	if err := tk.Validate(); err != nil {
		log.Error().Err(err).Str("tk", tk.String()).Msg("ReplaceAllRelation")
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("ReplaceAllRelation")

	currentTks, err := c.GetObjectTuples(ctx, NewTupleKey().WithObject(tk.Object.Type, tk.Object.Name))
	if err != nil {
		return err
	}
	var matched []TupleKey
	var notMatched []TupleKey
	for _, checkTk := range currentTks {
		if tk.Relation != checkTk.Relation {
			continue
		}
		if tk.Equals(checkTk) {
			matched = append(matched, checkTk)
		} else {
			notMatched = append(notMatched, checkTk)
		}
	}

	// Write new tuple before deleting others
	if len(matched) == 0 {
		if err := c.WriteTuple(ctx, tk); err != nil {
			return err
		}
	}

	// Delete exsiting tuples
	var errs []error
	for _, delTk := range notMatched {
		if err := c.DeleteTuple(ctx, delTk); err != nil {
			errs = append(errs, err)
		}
	}
	for _, err := range errs {
		log.Trace().Err(err).Str("tk", tk.String()).Msg("ReplaceAllRelation")
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (c *FGAClient) ReplaceTuple(ctx context.Context, tk TupleKey) error {
	if err := tk.Validate(); err != nil {
		log.Error().Err(err).Str("tk", tk.String()).Msg("ReplaceTuple")
		return err
	}
	log.Trace().Str("tk", tk.String()).Msg("ReplaceTuple")

	currentTks, err := c.GetObjectTuples(ctx, NewTupleKey().WithObject(tk.Object.Type, tk.Object.Name))
	if err != nil {
		return err
	}

	// Write new tuple before deleting others, if it exists
	exists := false
	for _, checkTk := range currentTks {
		if tk.Equals(checkTk) {
			exists = true
		}
	}
	if !exists {
		if err := c.WriteTuple(ctx, tk); err != nil {
			return err
		}
	}

	// Delete other tuples
	var errs []error
	for _, k := range currentTks {
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
		Writes:               &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{ToFGATupleKey(tk)}},
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
		Deletes:              &openfga.TupleKeys{TupleKeys: []openfga.TupleKey{ToFGATupleKey(tk)}},
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
	if dslJson, err = ioutil.ReadFile(fn); err != nil {
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
