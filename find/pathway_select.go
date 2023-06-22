package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func PathwaySelect(limit *int, after *model.Cursor, ids []int, userCheck *model.UserCheck, where *model.PathwayFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select("t.*").
		From("gtfs_pathways t").
		Limit(checkLimit(limit)).
		OrderBy("t.id")

	if where != nil {
		if where.PathwayMode != nil {
			q = q.Where(sq.Eq{"pathway_mode": where.PathwayMode})
		}
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil && after.Valid && after.ID > 0 {
		q = q.Where(sq.Gt{"t.id": after.ID})
	}
	if userCheck != nil {
		q = q.Where(sq.Or{sq.Eq{"feed_versions.feed_id": userCheck.AllowedFeeds}, sq.Eq{"feed_versions.id": userCheck.AllowedFeedVersions}})
	}
	return q
}
