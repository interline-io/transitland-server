package dbfinder

import (
	"context"
	"fmt"

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

func (f *Finder) CensusTableByIDs(ctx context.Context, ids []int) ([]*model.CensusTable, []error) {
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

func (f *Finder) CensusGeographiesByEntityIDs(ctx context.Context, limit *int, where *model.CensusGeographyFilter, entityType string, entityIds []int) ([][]*model.CensusGeography, error) {
	var ents []*model.CensusGeography
	err := dbutil.Select(ctx, f.db, censusGeographySelect(limit, where, entityType, entityIds), &ents)
	return arrangeGroup(entityIds, ents, func(ent *model.CensusGeography) int { return ent.MatchEntityID }), err
}

func (f *Finder) CensusValuesByGeographyIDs(ctx context.Context, limit *int, tableNames []string, keys []string) ([][]*model.CensusValue, error) {
	var ents []*model.CensusValue
	err := dbutil.Select(
		ctx,
		f.db,
		censusValueSelect(limit, "", tableNames, keys),
		&ents,
	)
	return arrangeGroup(keys, ents, func(ent *model.CensusValue) string { return ent.Geoid }), err
}

func (f *Finder) CensusSourcesByDatasetIDs(ctx context.Context, limit *int, where *model.CensusSourceFilter, keys []int) ([][]*model.CensusSource, error) {
	q := censusSourceSelect(limit, nil, nil, where)
	var ents []*model.CensusSource
	err := dbutil.Select(ctx,
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
	return arrangeGroup(keys, ents, func(ent *model.CensusSource) int { return ent.DatasetID }), err
}

func (f *Finder) CensusDatasetLayersByDatasetIDs(ctx context.Context, keys []int) ([][]*model.CensusLayer, []error) {
	var ents []*model.CensusLayer
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			sq.StatementBuilder.Select("*").From("tl_census_layers"),
			"tl_census_datasets",
			"id",
			"tl_census_layers",
			"dataset_id",
			keys,
		),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(keys), err)
	}
	return arrangeGroup(keys, ents, func(ent *model.CensusLayer) int { return ent.DatasetID }), nil
}

func (f *Finder) CensusSourceLayersBySourceIDs(ctx context.Context, keys []int) ([][]*model.CensusLayer, []error) {
	type qent struct {
		SourceID int
		model.CensusLayer
	}
	var ents []*qent
	q := sq.StatementBuilder.
		Select("tlcg.source_id", "tlcl.*").
		Distinct().Options("on (tlcl.id)").
		From("tl_census_geographies tlcg").
		Join("tl_census_layers tlcl on tlcl.id = tlcg.layer_id").
		Where(sq.Eq{"tlcg.source_id": keys})
	err := dbutil.Select(ctx,
		f.db,
		q,
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(keys), err)
	}
	grouped := arrangeGroup(keys, ents, func(ent *qent) int { return ent.SourceID })
	var ret [][]*model.CensusLayer
	for _, group := range grouped {
		var g []*model.CensusLayer
		for _, ent := range group {
			g = append(g, &ent.CensusLayer)
		}
		ret = append(ret, g)
	}
	return ret, nil
}

func (f *Finder) CensusGeographiesByDatasetIDs(ctx context.Context, limit *int, p *model.CensusDatasetGeographyFilter, keys []int) ([][]*model.CensusGeography, error) {
	var ents []*model.CensusGeography
	q := censusDatasetGeographySelect(limit, p)
	err := dbutil.Select(ctx,
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
	return arrangeGroup(keys, ents, func(ent *model.CensusGeography) int { return ent.DatasetID }), err
}

func (f *Finder) CensusFieldsByTableIDs(ctx context.Context, limit *int, keys []int) ([][]*model.CensusField, error) {
	var ents []*model.CensusField
	err := dbutil.Select(ctx,
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
	return arrangeGroup(keys, ents, func(ent *model.CensusField) int { return ent.TableID }), err
}

func censusDatasetSelect(_ *int, _ *model.Cursor, _ []int, where *model.CensusDatasetFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select("*").
		From("tl_census_datasets")
	if where != nil {
		if where.Name != nil {
			q = q.Where(sq.Eq{"name": *where.Name})
		}
		if where.Search != nil {
			q = q.Where(sq.Like{"name": fmt.Sprintf("%%%s%%", *where.Search)})
		}
	}
	return q
}

func censusSourceSelect(limit *int, after *model.Cursor, ids []int, where *model.CensusSourceFilter) sq.SelectBuilder {
	q := quickSelectOrder("tl_census_sources", limit, after, ids, "id")
	if where != nil {
		if where.Name != nil {
			q = q.Where(sq.Eq{"name": *where.Name})
		}
	}
	return q
}

func censusDatasetGeographySelect(limit *int, where *model.CensusDatasetGeographyFilter) sq.SelectBuilder {
	// Include matched entity column
	cols := []string{
		"tlcg.id",
		"tlcg.geometry",
		"tlcl.name as layer_name",
		"tlcg.geoid",
		"tlcg.name",
		"tlcg.aland",
		"tlcg.awater",
		"tlcg.adm0_name",
		"tlcg.adm1_name",
		"tlcg.adm0_iso",
		"tlcg.adm1_iso",
		"tlcs.name as source_name",
		"tlcs.id as source_id",
		"tlcd.name as dataset_name",
		"tlcd.id as dataset_id",
	}

	orderBy := sq.Expr("tlcg.id")

	// A normal query..
	q := sq.StatementBuilder.
		Select(cols...).
		From("tl_census_geographies tlcg").
		Join("tl_census_sources tlcs on tlcs.id = tlcg.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlcs.dataset_id").
		Join("tl_census_layers tlcl on tlcl.id = tlcg.layer_id").
		Limit(checkLimit(limit))

	if where != nil && where.Location != nil {
		loc := where.Location
		if loc.Bbox != nil {
			q = q.Where("ST_Intersects(tlcg.geometry, ST_MakeEnvelope(?,?,?,?,4326))", loc.Bbox.MinLon, loc.Bbox.MinLat, loc.Bbox.MaxLon, loc.Bbox.MaxLat)
		}
		if loc.Within != nil && loc.Within.Valid {
			q = q.Where("ST_Intersects(tlcg.geometry, ?)", loc.Within)
		}
		if loc.Near != nil {
			radius := checkFloat(&loc.Near.Radius, 0, 1_000_000)
			q = q.Where("ST_DWithin(tlcg.geometry, ST_MakePoint(?,?), ?)", loc.Near.Lon, loc.Near.Lat, radius)
			orderBy = sq.Expr("ST_Distance(tlcg.geometry, ST_MakePoint(?,?))", loc.Near.Lon, loc.Near.Lat)
		}
		if loc.Focus != nil {
			orderBy = sq.Expr("ST_Distance(tlcg.geometry, ST_MakePoint(?,?))", loc.Focus.Lon, loc.Focus.Lat)
		}
	}

	// Check layer, dataset
	if where != nil {
		if where.Layer != nil {
			q = q.Where(sq.Eq{"tlcl.name": where.Layer})
		}
		if where.Search != nil {
			q = q.Where(sq.ILike{"tlcg.name": fmt.Sprintf("%%%s%%", *where.Search)})
		}
		if len(where.Ids) > 0 {
			q = q.Where(sq.Eq{"tlcg.id": where.Ids})
		}
	}

	q = q.OrderByClause(orderBy)
	return q
}

func censusGeographySelect(limit *int, where *model.CensusGeographyFilter, entityType string, entityIds []int) sq.SelectBuilder {
	// Get search radius
	radius := 0.0
	if where != nil {
		radius = checkFloat(where.Radius, 0, 2000.0)
	}

	// Include matched entity column
	cols := []string{
		"tlcg.id",
		"tlcg.geometry",
		"tlcl.name as layer_name",
		"tlcg.geoid",
		"tlcg.name",
		"tlcg.aland",
		"tlcg.awater",
		"tlcs.name as source_name",
		"tlcs.id as source_id",
		"tlcd.name as dataset_name",
		"tlcd.id as dataset_id",
	}

	// A normal query..
	q := sq.StatementBuilder.
		Select(cols...).
		From("tl_census_geographies tlcg").
		Join("tl_census_sources tlcs on tlcs.id = tlcg.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlcs.dataset_id").
		Join("tl_census_layers tlcl on tlcl.id = tlcg.layer_id").
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
			q = q.Where(sq.Eq{"tlcd.name": where.Dataset})
		}
		if where.Search != nil {
			q = q.Where(sq.ILike{"tlcg.name": fmt.Sprintf("%%%s%%", *where.Search)})
		}
	}
	return q
}

func censusValueSelect(limit *int, datasetName string, tnames []string, geoids []string) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select(
			"tlcv.table_values as values",
			"tlcv.geoid",
			"tlcv.table_id",
			"tlcs.name as source_name",
			"tlcd.name as dataset_name",
		).
		From("tl_census_values tlcv").
		Limit(checkLimit(limit)).
		Join("tl_census_tables tlct ON tlct.id = tlcv.table_id").
		Join("tl_census_sources tlcs on tlcs.id = tlcv.source_id").
		Join("tl_census_datasets tlcd on tlcd.id = tlct.dataset_id").
		Where(sq.Eq{"tlcv.geoid": geoids}).
		Where(sq.Eq{"tlct.table_name": tnames}).
		OrderBy("tlcv.table_id")
	if datasetName != "" {
		q = q.Where(sq.Eq{"tlcd.name": datasetName})
	}
	return q
}
