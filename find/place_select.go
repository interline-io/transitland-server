package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func PlaceSelect(limit *int, after *model.Cursor, ids []int, level int, where *model.PlaceFilter) sq.SelectBuilder {
	var groupKeys []string
	switch level {
	case 0: // COUNTRY
		groupKeys = []string{"adm0name"}
	case 1: // COUNTRY/STATE
		groupKeys = []string{"adm0name", "adm1name"}
	case 2: // COUNTRY/STATE/CITY
		groupKeys = []string{"adm0name", "adm1name", "name"}
	case 3: // COUNTRY/CITY
		groupKeys = []string{"adm0name", "name"}
	case 4: // STATE/CITY
		groupKeys = []string{"adm1name", "name"}
	case 5: // CITY
		groupKeys = []string{"name"}
	}
	// TODO: is it necessary to check for deleted feeds? Or just deleted operators?
	q := sq.StatementBuilder.
		Select(groupKeys...).
		Columns("json_agg(distinct coif.resolved_onestop_id) as operator_onestop_ids").
		From("feed_states fs").
		// Join("current_feeds cf on cf.id = fs.feed_id").
		Join("gtfs_agencies a on a.feed_version_id = fs.feed_version_id").
		Join("tl_agency_places tlap on tlap.agency_id = a.id").
		Join("current_operators_in_feed coif on coif.feed_id = fs.feed_id and coif.resolved_gtfs_agency_id = a.agency_id").
		LeftJoin("current_operators co on co.id = coif.operator_id").
		Where(sq.Eq{"co.deleted_at": nil}).
		// Where(sq.Eq{"cf.deleted_at": nil}).
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
	return q
}