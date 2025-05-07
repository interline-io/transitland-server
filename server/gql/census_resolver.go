package gql

import (
	"context"
	"fmt"
	"strings"

	"github.com/interline-io/transitland-server/model"
)

////////////////////////// CENSUS RESOLVERS

type censusDatasetResolver struct{ *Resolver }

func (r *censusDatasetResolver) Geographies(ctx context.Context, obj *model.CensusDataset, limit *int) (ents []*model.CensusGeography, err error) {
	fmt.Println("CensusDatasetResolver.Geographies")
	return nil, nil
}

func (r *censusDatasetResolver) Sources(ctx context.Context, obj *model.CensusDataset) (ents []*model.CensusSource, err error) {
	fmt.Println("CensusDatasetResolver.Sources")
	return nil, nil
}

func (r *censusDatasetResolver) Tables(ctx context.Context, obj *model.CensusDataset, limit *int) (ents []*model.CensusTable, err error) {
	fmt.Println("CensusDatasetResolver.Tables")
	return nil, nil
}

type censusGeographyResolver struct{ *Resolver }

func (r *censusGeographyResolver) Values(ctx context.Context, obj *model.CensusGeography, tableNames []string, datasetName *string, limit *int) (ents []*model.CensusValue, err error) {
	// dataloader cant easily pass []string
	return For(ctx).CensusValuesByGeographyID.Load(ctx, model.CensusValueParam{Dataset: datasetName, TableNames: strings.Join(tableNames, ","), Limit: limit, Geoid: *obj.Geoid})()
}

type censusValueResolver struct{ *Resolver }

func (r *censusValueResolver) Table(ctx context.Context, obj *model.CensusValue) (*model.CensusTable, error) {
	return For(ctx).CensusTableByID.Load(ctx, obj.TableID)()
}

type censusTableResolver struct{ *Resolver }

func (r *censusTableResolver) Fields(ctx context.Context, obj *model.CensusTable) ([]*model.CensusField, error) {
	return For(ctx).CensusFieldsByTableID.Load(ctx, model.CensusFieldParam{TableID: obj.ID})()
}

// add geography resolvers to agency, route, stop

func (r *agencyResolver) CensusGeographies(ctx context.Context, obj *model.Agency, limit *int, where *model.CensusGeographyFilter) (ents []*model.CensusGeography, err error) {
	return For(ctx).CensusGeographiesByEntityID.Load(ctx, model.CensusGeographyParam{
		EntityType: "agency",
		EntityID:   obj.ID,
		Limit:      limit,
		Where:      where,
	})()
}

func (r *routeResolver) CensusGeographies(ctx context.Context, obj *model.Route, limit *int, where *model.CensusGeographyFilter) (ents []*model.CensusGeography, err error) {
	return For(ctx).CensusGeographiesByEntityID.Load(ctx, model.CensusGeographyParam{
		EntityType: "route",
		EntityID:   obj.ID,
		Limit:      limit,
		Where:      where,
	})()
}

func (r *stopResolver) CensusGeographies(ctx context.Context, obj *model.Stop, limit *int, where *model.CensusGeographyFilter) (ents []*model.CensusGeography, err error) {
	return For(ctx).CensusGeographiesByEntityID.Load(ctx, model.CensusGeographyParam{
		EntityType: "stop",
		EntityID:   obj.ID,
		Limit:      limit,
		Where:      where,
	})()
}
