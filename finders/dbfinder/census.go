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

func (f *Finder) CensusGeographiesByEntityIDs(ctx context.Context, limit *int, where *model.CensusGeographyFilter, entityType string, entityIds []int) (ents []*model.CensusGeography, err error) {
	err = dbutil.Select(ctx, f.db, censusGeographySelect2(limit, where, entityType, entityIds), &ents)
	return ents, err
}

func (f *Finder) CensusValuesByGeographyIDs(ctx context.Context, limit *int, tableNames []string, keys []string) (ents []*model.CensusValue, err error) {
	// TODO: FIXME
	// return paramGroupQuery(
	// 	params,
	// 	func(p model.CensusValueParam) (string, *model.CensusValueParam, *int) {
	// 		rp := model.CensusValueParam{
	// 			TableNames: p.TableNames,
	// 			Dataset:    p.Dataset,
	// 		}
	// 		return p.Geoid, &rp, p.Limit
	// 	},
	// 	func(keys []string, where *model.CensusValueParam, limit *int) (ents []*model.CensusValue, err error) {
	// err = dbutil.Select(
	// 	ctx,
	// 	f.db,
	// 	censusValueSelect(where, keys),
	// 	&ents,
	// )
	// return ents, err
	// 	},
	// 	func(ent *model.CensusValue) string {
	// 		return ent.Geoid
	// 	},
	// )
	return nil, nil
}

func (f *Finder) CensusSourcesByDatasetIDs(ctx context.Context, limit *int, where *model.CensusSourceFilter, keys []int) (ents []*model.CensusSource, err error) {
	q := censusSourceSelect(limit, nil, nil, where)
	err = dbutil.Select(ctx,
		f.db,
		lateralWrap(
			q,
			"tl_census_datasets",
			"id",
			"tl_census_sources",
			"dataset_id",
			keys,
		),
		&ents,
	)
	return ents, err
}

func (f *Finder) CensusDatasetLayersByDatasetIDs(ctx context.Context, ids []int) ([][]string, []error) {
	var ret [][]string
	var errs []error
	for _, id := range ids {
		var layers []string
		err := dbutil.Select(ctx,
			f.db,
			sq.StatementBuilder.
				Select("tlcg.layer_name").
				Distinct().Options("on (tlcg.layer_name)").
				From("tl_census_datasets tlcd").
				Join("tl_census_sources tlcs on tlcs.dataset_id = tlcd.id").
				Join("tl_census_geographies tlcg on tlcg.source_id = tlcs.id").
				Where(sq.Eq{"tlcd.id": id}),
			&layers,
		)
		if err != nil {
			errs = append(errs, logErr(ctx, err))
			continue
		}
		ret = append(ret, layers)
	}
	return ret, errs
}

func (f *Finder) CensusSourceLayersBySourceIDs(ctx context.Context, ids []int) ([][]string, []error) {
	var ret [][]string
	var errs []error
	for _, id := range ids {
		var layers []string
		err := dbutil.Select(ctx,
			f.db,
			sq.StatementBuilder.
				Select("tlcg.layer_name").
				Distinct().Options("on (tlcg.layer_name)").
				From("tl_census_sources tlcs").
				Join("tl_census_geographies tlcg on tlcg.source_id = tlcs.id").
				Where(sq.Eq{"tlcs.id": id}),
			&layers,
		)
		if err != nil {
			errs = append(errs, logErr(ctx, err))
			continue
		}
		ret = append(ret, layers)
	}
	return ret, errs
}

func (f *Finder) CensusGeographiesByDatasetIDs(ctx context.Context, limit *int, p *model.CensusDatasetGeographyFilter, keys []int) (ents []*model.CensusGeography, err error) {
	q := censusDatasetGeographySelect(limit, p, keys)
	err = dbutil.Select(ctx,
		f.db,
		lateralWrap(
			q,
			"tl_census_datasets",
			"id",
			"tlcd",
			"id",
			keys,
		),
		&ents,
	)
	return ents, err
}

func (f *Finder) CensusFieldsByTableIDs(ctx context.Context, limit *int, keys []int) (ents []*model.CensusField, err error) {
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

func censusSourceSelect(limit *int, after *model.Cursor, ids []int, where *model.CensusSourceFilter) sq.SelectBuilder {
	q := quickSelectOrder("tl_census_sources", limit, after, ids, "id")
	if where != nil {
		if where.SourceName != nil {
			q = q.Where(sq.Eq{"source_name": *where.SourceName})
		}
	}
	return q
}

func censusDatasetGeographySelect(limit *int, where *model.CensusDatasetGeographyFilter, datasetIds []int) sq.SelectBuilder {
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
		"tlcs.id as source_id",
		"tlcd.dataset_name",
		"tlcd.id as dataset_id",
	}

	// A normal query..
	q := sq.StatementBuilder.
		Select(cols...).
		From("tl_census_geographies tlcg").
		Join("tl_census_sources tlcs on tlcs.id = tlcg.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlcs.dataset_id").
		Limit(checkLimit(limit))

	if where != nil {
		if where.Bbox != nil {
			q = q.Where("ST_Intersects(tlcg.geometry, ST_MakeEnvelope(?,?,?,?,4326))", where.Bbox.MinLon, where.Bbox.MinLat, where.Bbox.MaxLon, where.Bbox.MaxLat)
		}
		if where.Within != nil && where.Within.Valid {
			q = q.Where("ST_Intersects(tlcg.geometry, ?)", where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 1_000_000)
			q = q.Where("ST_DWithin(tlcg.geometry, ST_MakePoint(?,?), ?)", where.Near.Lon, where.Near.Lat, radius)
		}
	}

	// Check layer, dataset
	if where != nil {
		if where.Layer != nil {
			q = q.Where(sq.Eq{"tlcg.layer_name": where.Layer})
		}
		if where.Search != nil {
			q = q.Where(sq.ILike{"tlcg.name": fmt.Sprintf("%%%s%%", *where.Search)})
		}
	}
	return q
}

func censusGeographySelect(param *model.CensusGeographyParam, entityIds []int) sq.SelectBuilder {
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
		"tlcs.id as source_id",
		"tlcd.dataset_name",
		"tlcd.id as dataset_id",
	}

	// A normal query..
	q := sq.StatementBuilder.
		Select(cols...).
		From("tl_census_geographies tlcg").
		Join("tl_census_sources tlcs on tlcs.id = tlcg.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlcs.dataset_id").
		Limit(checkLimit(param.Limit))

	if len(entityIds) > 0 {
		// Handle aggregation by entity type
		q = q.Join("gtfs_stops ON ST_DWithin(tlcg.geometry, gtfs_stops.geometry, ?)", radius)
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
	}

	// Check layer, dataset
	if where := param.Where; where != nil {
		if where.Layer != nil {
			q = q.Where(sq.Eq{"tlcg.layer_name": where.Layer})
		}
		if where.Dataset != nil {
			q = q.Where(sq.Eq{"tlcd.dataset_name": where.Dataset})
		}
		if where.Search != nil {
			q = q.Where(sq.ILike{"tlcg.name": fmt.Sprintf("%%%s%%", *where.Search)})
		}
	}
	return q
}

func censusGeographySelect2(limit *int, where *model.CensusGeographyFilter, entityType string, entityIds []int) sq.SelectBuilder {
	// Get search radius
	radius := 0.0
	if where != nil {
		radius = checkFloat(where.Radius, 0, 2000.0)
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
		"tlcs.id as source_id",
		"tlcd.dataset_name",
		"tlcd.id as dataset_id",
	}

	// A normal query..
	q := sq.StatementBuilder.
		Select(cols...).
		From("tl_census_geographies tlcg").
		Join("tl_census_sources tlcs on tlcs.id = tlcg.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlcs.dataset_id").
		Limit(checkLimit(limit))

	if len(entityIds) > 0 {
		// Handle aggregation by entity type
		q = q.Join("gtfs_stops ON ST_DWithin(tlcg.geometry, gtfs_stops.geometry, ?)", radius)
		if entityType == "route" {
			q = q.Column("tl_route_stops.route_id as match_entity_id")
			q = q.Join("tl_route_stops ON tl_route_stops.stop_id = gtfs_stops.id")
			q = q.Distinct().Options("on (tl_route_stops.route_id,tlcg.id)").Where(In("tl_route_stops.route_id", entityIds)).OrderBy("tl_route_stops.route_id,tlcg.id")
		} else if entityType == "agency" {
			q = q.Column("tl_route_stops.agency_id as match_entity_id")
			q = q.Join("tl_route_stops ON tl_route_stops.stop_id = gtfs_stops.id")
			q = q.Distinct().Options("on (tl_route_stops.stop_id,tlcg.id)").Where(In("tl_route_stops.agency_id", entityIds)).OrderBy("tl_route_stops.agency_id,tlcg.id")
		} else if entityType == "stop" {
			q = q.Column("gtfs_stops.id as match_entity_id")
			q = q.Where(In("gtfs_stops.id", entityIds)).OrderBy("gtfs_stops.id,tlcg.id")
		}
	}

	// Check layer, dataset
	if where != nil {
		if where.Layer != nil {
			q = q.Where(sq.Eq{"tlcg.layer_name": where.Layer})
		}
		if where.Dataset != nil {
			q = q.Where(sq.Eq{"tlcd.dataset_name": where.Dataset})
		}
		if where.Search != nil {
			q = q.Where(sq.ILike{"tlcg.name": fmt.Sprintf("%%%s%%", *where.Search)})
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
