package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func PathwaySelect(limit *int, after *model.Cursor, ids []int, permFilter *model.PermFilter, where *model.PathwayFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.
		Select(
			"gtfs_pathways.id",
			"gtfs_pathways.feed_version_id",
			"gtfs_pathways.pathway_id",
			"gtfs_pathways.from_stop_id",
			"gtfs_pathways.to_stop_id",
			"gtfs_pathways.pathway_mode",
			"gtfs_pathways.is_bidirectional",
			"gtfs_pathways.length",
			"gtfs_pathways.traversal_time",
			"gtfs_pathways.stair_count",
			"gtfs_pathways.max_slope",
			"gtfs_pathways.min_width",
			"gtfs_pathways.signposted_as",
			"gtfs_pathways.reverse_signposted_as",
		).
		From("gtfs_pathways").
		Join("feed_versions on feed_versions.id = gtfs_pathways.feed_version_id").
		Join("current_feeds on current_feeds.id = feed_versions.feed_id").
		Limit(checkLimit(limit)).
		OrderBy("gtfs_pathways.id")

	if where != nil {
		if where.PathwayMode != nil {
			q = q.Where(sq.Eq{"pathway_mode": where.PathwayMode})
		}
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"gtfs_pathways.id": ids})
	}
	if after != nil && after.Valid && after.ID > 0 {
		q = q.Where(sq.Gt{"gtfs_pathways.id": after.ID})
	}

	// Handle permissions
	q = q.
		Join("feed_states fsp on fsp.feed_id = current_feeds.id").
		Where(sq.Or{
			sq.Expr("fsp.public = true"),
			sq.Eq{"true": permFilter.IsGlobalAdmin()},
			sq.Eq{"fsp.feed_id": permFilter.GetAllowedFeeds()},
			sq.Eq{"feed_versions.id": permFilter.GetAllowedFeedVersions()},
		})

	return q
}
