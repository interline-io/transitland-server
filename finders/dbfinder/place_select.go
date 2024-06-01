package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func PlaceSelect(limit *int, after *model.Cursor, ids []int, level *model.PlaceAggregationLevel, permFilter *model.PermFilter, where *model.PlaceFilter) sq.SelectBuilder {
	// PlaceSelect is limited to active feed versions
	var groupKeys []string
	var selKeys []string
	// Yucky mapping
	selKeys = []string{"adm0name as adm0_name"}
	groupKeys = []string{"adm0name"}
	if level != nil {
		switch *level {
		case model.PlaceAggregationLevelAdm0:
			groupKeys = []string{"adm0name"}
		case model.PlaceAggregationLevelAdm0Adm1:
			selKeys = []string{"adm0name as adm0_name", "adm1name as adm1_name"}
			groupKeys = []string{"adm0name", "adm1name"}
		case model.PlaceAggregationLevelAdm0Adm1City:
			selKeys = []string{"adm0name as adm0_name", "adm1name as adm1_name", "name as city_name"}
			groupKeys = []string{"adm0name", "adm1name", "name"}
		case model.PlaceAggregationLevelAdm0City:
			selKeys = []string{"adm0name as adm0_name", "name as city_name"}
			groupKeys = []string{"adm0name", "name"}
		case model.PlaceAggregationLevelAdm1City:
			selKeys = []string{"adm1name as adm1_name"}
			groupKeys = []string{"adm1name", "name"}
		case model.PlaceAggregationLevelCity:
			selKeys = []string{"name as city_name"}
			groupKeys = []string{"name"}
		}
	}
	q := sq.StatementBuilder.
		Select(selKeys...).
		Columns("json_agg(distinct tlap.agency_id) as agency_ids").
		From("feed_states").
		Join("tl_agency_places tlap on tlap.feed_version_id = feed_states.feed_version_id").
		GroupBy(groupKeys...)

	if where != nil {
		if where.Adm0Name != nil {
			q = q.Where(sq.Eq{"adm0name": where.Adm0Name})
		}
		if where.Adm1Name != nil {
			q = q.Where(sq.Eq{"adm1name": where.Adm1Name})
		}
		if where.CityName != nil {
			q = q.Where(sq.Eq{"name": where.CityName})
		}
	}

	// Handle permissions
	q = pfJoinCheck(q, "feed_states.feed_id", "feed_states.feed_version_id", permFilter)
	return q
}
