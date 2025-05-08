package dbfinder

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-server/model"
)

func (f *Finder) FindCensusDatasets(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.CensusDatasetFilter) ([]*model.CensusDataset, error) {
	var ents []*model.CensusDataset
	q := censusDatasetSelect(limit, after, ids, where)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) CensusTableByID(ctx context.Context, ids []int) ([]*model.CensusTable, []error) {
	var ents []*model.CensusTable
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("tl_census_tables", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.CensusTable) int { return ent.ID }), nil
}

func (f *Finder) CensusGeographiesByEntityID(ctx context.Context, params []model.CensusGeographyParam) ([][]*model.CensusGeography, []error) {
	return paramGroupQuery(
		params,
		func(p model.CensusGeographyParam) (int, *model.CensusGeographyParam, *int) {
			rp := model.CensusGeographyParam{
				EntityType: p.EntityType,
				Limit:      p.Limit,
				Where:      p.Where,
			}
			return p.EntityID, &rp, p.Limit
		},
		func(keys []int, where *model.CensusGeographyParam, limit *int) (ents []*model.CensusGeography, err error) {
			err = dbutil.Select(ctx, f.db, censusGeographySelect(where, keys), &ents)
			return ents, err
		},
		func(ent *model.CensusGeography) int {
			return ent.MatchEntityID
		},
	)
}

func (f *Finder) CensusValuesByGeographyID(ctx context.Context, params []model.CensusValueParam) ([][]*model.CensusValue, []error) {
	return paramGroupQuery(
		params,
		func(p model.CensusValueParam) (string, *model.CensusValueParam, *int) {
			rp := model.CensusValueParam{
				TableNames: p.TableNames,
				Dataset:    p.Dataset,
			}
			return p.Geoid, &rp, p.Limit
		},
		func(keys []string, where *model.CensusValueParam, limit *int) (ents []*model.CensusValue, err error) {
			err = dbutil.Select(
				ctx,
				f.db,
				censusValueSelect(where, keys),
				&ents,
			)
			return ents, err
		},
		func(ent *model.CensusValue) string {
			return ent.Geoid
		},
	)
}

func (f *Finder) CensusSourcesByDatasetID(ctx context.Context, params []model.CensusSourceParam) ([][]*model.CensusSource, []error) {
	return paramGroupQuery(
		params,
		func(p model.CensusSourceParam) (int, *model.CensusSourceParam, *int) {
			return p.DatasetID, nil, p.Limit
		},
		func(keys []int, where *model.CensusSourceParam, limit *int) (ents []*model.CensusSource, err error) {
			fmt.Println("KEYS:", keys)
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("tl_census_sources", limit, nil, nil, "id"),
					"tl_census_datasets",
					"id",
					"tl_census_sources",
					"dataset_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.CensusSource) int {
			return ent.DatasetID
		},
	)
}

func (f *Finder) CensusFieldsByTableID(ctx context.Context, params []model.CensusFieldParam) ([][]*model.CensusField, []error) {
	return paramGroupQuery(
		params,
		func(p model.CensusFieldParam) (int, *model.CensusFieldParam, *int) {
			return p.TableID, nil, p.Limit
		},
		func(keys []int, where *model.CensusFieldParam, limit *int) (ents []*model.CensusField, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("tl_census_fields", limit, nil, nil, "id"),
					"tl_census_tables",
					"id",
					"tl_census_fields",
					"table_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.CensusField) int {
			return ent.TableID
		},
	)
}

func censusDatasetSelect(_ *int, _ *model.Cursor, _ []int, where *model.CensusDatasetFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select("*").
		From("tl_census_datasets")
	if where != nil {
		if where.DatasetName != nil {
			q = q.Where(sq.Eq{"dataset_name": *where.DatasetName})
		}
		if where.Search != nil {
			q = q.Where(sq.Like{"dataset_name": fmt.Sprintf("%%%s%%", *where.Search)})
		}
	}
	return q
}

func censusGeographySelect(param *model.CensusGeographyParam, entityIds []int) sq.SelectBuilder {
	if param.EntityID > 0 {
		entityIds = append(entityIds, param.EntityID)
	}

	// Get search radius
	radius := 0.0
	if param.Where != nil {
		radius = checkFloat(param.Where.Radius, 0, 2000.0)
	}

	// Include matched entity column
	cols := []string{
		"tlcg.id",
		"tlcg.geometry",
		"tlcg.layer_name",
		"tlcg.geoid",
		"tlcg.name",
		"tlcg.aland",
		"tlcg.awater",
		"tlcs.source_name",
		"tlcd.dataset_name",
	}

	// A normal query..
	q := sq.StatementBuilder.
		Select(cols...).
		From("tl_census_geographies tlcg").
		Join("tl_census_sources tlcs on tlcs.id = tlcg.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlcs.dataset_id").
		Join("gtfs_stops ON ST_DWithin(tlcg.geometry, gtfs_stops.geometry, ?)", radius).
		Limit(checkLimit(param.Limit))

	// Handle aggregation by entity type
	if param.EntityType == "route" {
		q = q.Column("tl_route_stops.route_id as match_entity_id")
		q = q.Join("tl_route_stops ON tl_route_stops.stop_id = gtfs_stops.id")
		q = q.Distinct().Options("on (tl_route_stops.route_id,tlcg.id)").Where(In("tl_route_stops.route_id", entityIds)).OrderBy("tl_route_stops.route_id,tlcg.id")
	} else if param.EntityType == "agency" {
		q = q.Column("tl_route_stops.agency_id as match_entity_id")
		q = q.Join("tl_route_stops ON tl_route_stops.stop_id = gtfs_stops.id")
		q = q.Distinct().Options("on (tl_route_stops.stop_id,tlcg.id)").Where(In("tl_route_stops.agency_id", entityIds)).OrderBy("tl_route_stops.agency_id,tlcg.id")
	} else if param.EntityType == "stop" {
		q = q.Column("gtfs_stops.id as match_entity_id")
		q = q.Where(In("gtfs_stops.id", entityIds)).OrderBy("gtfs_stops.id,tlcg.id")
	}

	// Check layer, dataset
	if where := param.Where; where != nil {
		if where.Layer != nil {
			q = q.Where(sq.Eq{"tlcg.layer_name": where.Layer})
		}
		if where.Dataset != nil {
			q = q.Where(sq.Eq{"tlcd.dataset_name": where.Dataset})
		}
	}
	return q
}

func censusValueSelect(param *model.CensusValueParam, geoids []string) sq.SelectBuilder {
	tnames := sliceToLower(strings.Split(param.TableNames, ","))
	q := sq.StatementBuilder.
		Select(
			"tlcv.table_values as values",
			"tlcv.geoid",
			"tlcv.table_id",
			"tlcs.source_name",
			"tlcd.dataset_name",
		).
		From("tl_census_values tlcv").
		Limit(checkLimit(param.Limit)).
		Join("tl_census_tables tlct ON tlct.id = tlcv.table_id").
		Join("tl_census_sources tlcs on tlcs.id = tlcv.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlct.dataset_id").
		Where(sq.Eq{"tlcv.geoid": geoids}).
		Where(sq.Eq{"tlct.table_name": tnames}).
		OrderBy("tlcv.table_id")
	if param.Dataset != nil {
		q = q.Where(sq.Eq{"tlcd.dataset_name": param.Dataset})
	}

	return q
}
