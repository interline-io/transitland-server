package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/model"
)

func FeedVersionSelect(limit *int, after *model.Cursor, ids []int, where *model.FeedVersionFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select("t.*, tl_feed_version_geometries.geometry").
		From("feed_versions t").
		Join("current_feeds cf on cf.id = t.feed_id").Where(sq.Eq{"cf.deleted_at": nil}).
		JoinClause("left join tl_feed_version_geometries on tl_feed_version_geometries.feed_version_id = t.id").
		Limit(checkLimit(limit)).
		OrderBy("t.fetched_at desc, t.id desc")
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil && after.Valid && after.ID > 0 {
		q = q.Where(sq.Expr("(t.fetched_at,t.id) < (select fetched_at,id from feed_versions where id = ?)", after.ID))
	}
	if where != nil {
		if where.Sha1 != nil {
			q = q.Where(sq.Eq{"sha1": *where.Sha1})
		}
		if len(where.FeedIds) > 0 {
			q = q.Where(sq.Eq{"feed_id": where.FeedIds})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"cf.onestop_id": *where.FeedOnestopID})
		}
		// Import import status
		// Similar logic to FeedSelect
		if where.ImportStatus != nil {
			// in_progress must be false to check success and vice-versa
			var checkSuccess bool
			var checkInProgress bool
			switch v := *where.ImportStatus; v {
			case model.ImportStatusSuccess:
				checkSuccess = true
				checkInProgress = false
			case model.ImportStatusInProgress:
				checkSuccess = false
				checkInProgress = true
			case model.ImportStatusError:
				checkSuccess = false
				checkInProgress = false
			default:
				log.Error().Str("value", v.String()).Msg("unknown imnport status enum")
			}
			q = q.Join(`feed_version_gtfs_imports fvgi on fvgi.feed_version_id = t.id`).
				Where(sq.Eq{"fvgi.success": checkSuccess, "fvgi.in_progress": checkInProgress})
		}
	}
	return q
}

func FeedVersionServiceLevelSelect(limit *int, after *model.Cursor, ids []int, where *model.FeedVersionServiceLevelFilter) sq.SelectBuilder {
	q := quickSelectOrder("feed_version_service_levels", limit, after, nil, "")
	if where == nil {
		where = &model.FeedVersionServiceLevelFilter{}
	}
	q = q.Where(sq.Eq{"route_id": nil})
	if where.StartDate != nil {
		q = q.Where(sq.GtOrEq{"start_date": where.StartDate})
	}
	if where.EndDate != nil {
		q = q.Where(sq.LtOrEq{"end_date": where.EndDate})
	}
	return q
}
