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
	selKeys = []string{"tlap.adm0name as adm0_name"}
	groupKeys = []string{"tlap.adm0name"}
	if level != nil {
		switch *level {
		case model.PlaceAggregationLevelAdm0:
			groupKeys = []string{"tlap.adm0name"}
		case model.PlaceAggregationLevelAdm0Adm1:
			selKeys = []string{"tlap.adm0name as adm0_name", "tlap.adm1name as adm1_name"}
			groupKeys = []string{"tlap.adm0name", "tlap.adm1name"}
		case model.PlaceAggregationLevelAdm0Adm1City:
			selKeys = []string{"tlap.adm0name as adm0_name", "tlap.adm1name as adm1_name", "tlap.name as city_name"}
			groupKeys = []string{"tlap.adm0name", "tlap.adm1name", "tlap.name"}
		case model.PlaceAggregationLevelAdm0City:
			selKeys = []string{"tlap.adm0name as adm0_name", "tlap.name as city_name"}
			groupKeys = []string{"tlap.adm0name", "tlap.name"}
		case model.PlaceAggregationLevelAdm1City:
			selKeys = []string{"tlap.adm1name as adm1_name"}
			groupKeys = []string{"tlap.adm1name", "tlap.name"}
		case model.PlaceAggregationLevelCity:
			selKeys = []string{"tlap.name as city_name"}
			groupKeys = []string{"tlap.name"}
		}
	}
	q := sq.StatementBuilder.
		Select(selKeys...).
		Columns("json_agg(distinct tlap.agency_id) as agency_ids").
		From("feed_states").
		Join("tl_agency_places tlap on tlap.feed_version_id = feed_states.feed_version_id").
		Join("feed_versions on feed_versions.id = feed_states.feed_version_id").
		Join("current_feeds on current_feeds.id = feed_states.feed_id").
		GroupBy(groupKeys...)

	if where != nil {
		if where.Adm0Name != nil {
			q = q.Where(sq.Eq{"tlap.adm0name": where.Adm0Name})
		}
		if where.Adm1Name != nil {
			q = q.Where(sq.Eq{"tlap.adm1name": where.Adm1Name})
		}
		if where.CityName != nil {
			q = q.Where(sq.Eq{"tlap.name": where.CityName})
		}
	}

	// Handle permissions
	q = pfJoinCheckFv(q, permFilter)
	return q
}
