package gql

import (
	"context"
	"strings"

	"github.com/interline-io/transitland-server/model"
)

////////////////////////// CENSUS RESOLVERS

type censusDatasetResolver struct{ *Resolver }

func (r *censusDatasetResolver) Geographies(ctx context.Context, obj *model.CensusDataset, limit *int, where *model.CensusDatasetGeographyFilter) (ents []*model.CensusGeography, err error) {
	return LoaderFor(ctx).CensusGeographiesByDatasetIDs.Load(ctx, model.CensusDatasetGeographyParam{DatasetID: obj.ID, Limit: limit, Where: where})()
}

func (r *censusDatasetResolver) Sources(ctx context.Context, obj *model.CensusDataset, limit *int, where *model.CensusSourceFilter) (ents []*model.CensusSource, err error) {
	return LoaderFor(ctx).CensusSourcesByDatasetIDs.Load(ctx, model.CensusSourceParam{DatasetID: obj.ID, Limit: limit, Where: where})()
}

func (r *censusDatasetResolver) Tables(ctx context.Context, obj *model.CensusDataset, limit *int, where *model.CensusTableFilter) (ents []*model.CensusTable, err error) {
	return nil, nil
}

func (r *censusDatasetResolver) Layers(ctx context.Context, obj *model.CensusDataset) (ret []string, err error) {
	return LoaderFor(ctx).CensusDatasetLayersByDatasetIDs.Load(ctx, obj.ID)()
}

type censusSourceResolver struct{ *Resolver }

func (r *censusSourceResolver) Layers(ctx context.Context, obj *model.CensusSource) (ret []string, err error) {
	return LoaderFor(ctx).CensusSourceLayersBySourceIDs.Load(ctx, obj.ID)()
}

func (r *censusSourceResolver) Geographies(ctx context.Context, obj *model.CensusSource, _ *int) (ret []*model.CensusGeography, err error) {
	return nil, nil
}

type censusGeographyResolver struct{ *Resolver }

func (r *censusGeographyResolver) Values(ctx context.Context, obj *model.CensusGeography, tableNames []string, datasetName *string, limit *int) (ents []*model.CensusValue, err error) {
	// dataloader cant easily pass []string
	return LoaderFor(ctx).CensusValuesByGeographyIDs.Load(ctx, model.CensusValueParam{Dataset: datasetName, TableNames: strings.Join(tableNames, ","), Limit: limit, Geoid: *obj.Geoid})()
}

type censusValueResolver struct{ *Resolver }

func (r *censusValueResolver) Table(ctx context.Context, obj *model.CensusValue) (*model.CensusTable, error) {
	return LoaderFor(ctx).CensusTableByIDs.Load(ctx, obj.TableID)()
}

type censusTableResolver struct{ *Resolver }

func (r *censusTableResolver) Fields(ctx context.Context, obj *model.CensusTable) ([]*model.CensusField, error) {
	return LoaderFor(ctx).CensusFieldsByTableIDs.Load(ctx, model.CensusFieldParam{TableID: obj.ID})()
}

// add geography resolvers to agency, route, stop

func (r *agencyResolver) CensusGeographies(ctx context.Context, obj *model.Agency, limit *int, where *model.CensusGeographyFilter) (ents []*model.CensusGeography, err error) {
	return LoaderFor(ctx).CensusGeographiesByEntityIDs.Load(ctx, model.CensusGeographyParam{
		EntityType: "agency",
		EntityID:   obj.ID,
		Limit:      limit,
		Where:      where,
	})()
}

func (r *routeResolver) CensusGeographies(ctx context.Context, obj *model.Route, limit *int, where *model.CensusGeographyFilter) (ents []*model.CensusGeography, err error) {
	return LoaderFor(ctx).CensusGeographiesByEntityIDs.Load(ctx, model.CensusGeographyParam{
		EntityType: "route",
		EntityID:   obj.ID,
		Limit:      limit,
		Where:      where,
	})()
}

func (r *stopResolver) CensusGeographies(ctx context.Context, obj *model.Stop, limit *int, where *model.CensusGeographyFilter) (ents []*model.CensusGeography, err error) {
	return LoaderFor(ctx).CensusGeographiesByEntityIDs.Load(ctx, model.CensusGeographyParam{
		EntityType: "stop",
		EntityID:   obj.ID,
		Limit:      limit,
		Where:      where,
	})()
}
