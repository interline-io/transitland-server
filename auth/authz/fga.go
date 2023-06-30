package authz

import (
	"context"
	"encoding/json"
	"fmt"
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

func (c *FGAClient) SetExclusiveRelation(ctx context.Context, tk TupleKey) error {
	return c.replaceTuple(ctx, tk, false, tk.Relation)

}

func (c *FGAClient) SetExclusiveSubjectRelation(ctx context.Context, tk TupleKey, checkRelations ...Relation) error {
	return c.replaceTuple(ctx, tk, true, checkRelations...)
}

func (c *FGAClient) replaceTuple(ctx context.Context, tk TupleKey, checkSubjectEqual bool, checkRelations ...Relation) error {
	if err := tk.Validate(); err != nil {
		log.Error().Err(err).Str("tk", tk.String()).Msg("replaceTuple")
		return err
	}
	relTypeOk := false
	for _, checkRel := range checkRelations {
		if tk.Relation == checkRel {
			relTypeOk = true
		}
	}
	if !relTypeOk {
		return fmt.Errorf("unknown relation %s for types %s and %s", tk.Relation.String(), tk.Subject.Type.String(), tk.Object.Type.String())
	}
	log.Trace().Str("tk", tk.String()).Msg("replaceTuple")

	currentTks, err := c.GetObjectTuples(ctx, NewTupleKey().WithObject(tk.Object.Type, tk.Object.Name))
	if err != nil {
		return err
	}

	var matchTks []TupleKey
	var delTks []TupleKey
	for _, checkTk := range currentTks {
		relMatch := false
		for _, r := range checkRelations {
			if checkTk.Relation == r {
				relMatch = true
			}
		}
		if !relMatch {
			continue
		}
		if checkSubjectEqual && !checkTk.Subject.Equals(tk.Subject) {
			continue
		}
		if checkTk.Equals(tk) {
			matchTks = append(matchTks, checkTk)
		} else {
			delTks = append(delTks, checkTk)
		}
	}

	// Write new tuple before deleting others
	if len(matchTks) == 0 {
		if err := c.WriteTuple(ctx, tk); err != nil {
			return err
		}
	}

	// Delete exsiting tuples
	var errs []error
	for _, delTk := range delTks {
		if err := c.DeleteTuple(ctx, delTk); err != nil {
			errs = append(errs, err)
		}
	}
	for _, err := range errs {
		log.Trace().Err(err).Str("tk", tk.String()).Msg("replaceTuple")
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
