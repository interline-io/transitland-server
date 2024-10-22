package dbfinder

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func CensusGeographySelect(param *model.CensusGeographyParam, entityIds []int) sq.SelectBuilder {
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

func CensusValueSelect(param *model.CensusValueParam, geoids []string) sq.SelectBuilder {
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
	return q
}

func sliceToLower(v []string) []string {
	for i := 0; i < len(v); i++ {
		v[i] = strings.ToLower(v[i])
	}
	return v
}
