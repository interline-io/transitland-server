package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func PathwaySelect(limit *int, after *model.Cursor, ids []int, permFilter *model.PermFilter, where *model.PathwayFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select("t.*").
		From("gtfs_pathways t").
		Join("feed_versions on feed_versions.id = t.feed_version_id").
		Join("current_feeds on current_feeds.id = feed_versions.feed_id").
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

	// Handle permissions
	q = q.
		Join("feed_states fsp on fsp.feed_id = current_feeds.id").
		Where(sq.Or{
			sq.Expr("fsp.public = true"),
			sq.Eq{"feed_versions.feed_id": permFilter.GetAllowedFeeds()},
			sq.Eq{"feed_versions.id": permFilter.GetAllowedFeedVersions()},
		})

	return q
}
