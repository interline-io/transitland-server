package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func PlaceSelect(limit *int, after *model.Cursor, ids []int, level *model.PlaceAggregationLevel, permFilter *model.PermFilter, where *model.PlaceFilter) sq.SelectBuilder {
	var groupKeys []string
	groupKeys = []string{"adm0name"}
	if level != nil {
		switch *level {
		case model.PlaceAggregationLevelAdm0:
			groupKeys = []string{"adm0name"}
		case model.PlaceAggregationLevelAdm0Adm1:
			groupKeys = []string{"adm0name", "adm1name"}
		case model.PlaceAggregationLevelAdm0Adm1City:
			groupKeys = []string{"adm0name", "adm1name", "name"}
		case model.PlaceAggregationLevelAdm0City:
			groupKeys = []string{"adm0name", "name"}
		case model.PlaceAggregationLevelAdm1City:
			groupKeys = []string{"adm1name", "name"}
		case model.PlaceAggregationLevelCity:
			groupKeys = []string{"name"}
		}
	}
	q := sq.StatementBuilder.
		Select(groupKeys...).
		Columns("json_agg(distinct tlap.agency_id) as agency_ids").
		From("feed_states fs").
		Join("tl_agency_places tlap on tlap.feed_version_id = fs.feed_version_id").
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
	if permFilter != nil {
		q = q.Where(sq.Or{sq.Eq{"fs.feed_id": permFilter.AllowedFeeds}, sq.Eq{"fs.feed_version_id": permFilter.AllowedFeedVersions}})
	}
	return q
}
