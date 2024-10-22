package gql

import (
	"context"
	"strings"

	"github.com/interline-io/transitland-server/model"
)

////////////////////////// CENSUS RESOLVERS

type censusGeographyResolver struct{ *Resolver }

func (r *censusGeographyResolver) Values(ctx context.Context, obj *model.CensusGeography, tableNames []string, limit *int) (ents []*model.CensusValue, err error) {
	// dataloader cant easily pass []string
	return For(ctx).CensusValuesByGeographyID.Load(ctx, model.CensusValueParam{TableNames: strings.Join(tableNames, ","), Limit: limit, Geoid: *obj.Geoid})()
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
