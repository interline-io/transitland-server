package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func AgencySelect(limit *int, after *int, ids []int, active bool, where *model.AgencyFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.
		Select(
			"gtfs_agencies.*",
			"tl_agency_geometries.geometry",
			"feed_versions.sha1 AS feed_version_sha1",
			"current_feeds.onestop_id AS feed_onestop_id",
			"coalesce (coif.resolved_onestop_id, '') as onestop_id",
			"coif.id as coif_id",
		).
		From("gtfs_agencies").
		Join("feed_versions ON feed_versions.id = gtfs_agencies.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		JoinClause("left join tl_agency_geometries ON tl_agency_geometries.agency_id = gtfs_agencies.id").
		JoinClause("left join current_operators_in_feed coif ON coif.feed_id = current_feeds.id AND coif.resolved_gtfs_agency_id = gtfs_agencies.agency_id").
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		OrderBy("gtfs_agencies.id")
	if active {
		qView = qView.Join("feed_states on feed_states.feed_version_id = gtfs_agencies.feed_version_id")
	}

	q := sq.StatementBuilder.Select("t.*").FromSelect(qView, "t").Limit(checkLimit(limit))
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	if where != nil {
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsQuery(*where.Search)
			q = q.Column(rank).Where(wc)
		}
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_version_sha1": *where.FeedVersionSha1})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"feed_onestop_id": *where.FeedOnestopID})
		}
		if where.AgencyID != nil {
			q = q.Where(sq.Eq{"agency_id": *where.AgencyID})
		}
		if where.AgencyName != nil {
			q = q.Where(sq.Eq{"agency_name": *where.AgencyName})
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"onestop_id": *where.OnestopID})
		}
		if where.Within != nil && where.Within.Valid {
			q = q.Where("ST_Intersects(t.geometry, ?)", where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 10_000)
			q = q.Where("ST_DWithin(t.geometry, ST_MakePoint(?,?), ?)", where.Near.Lon, where.Near.Lat, radius)
		}
	}
	return q
}

func AgencyPlaceSelect(limit *int, after *int, ids []int, where *model.AgencyPlaceFilter) sq.SelectBuilder {
	q := quickSelectOrder("tl_agency_places", limit, after, ids, "rank desc")
	if where != nil {
		// if where.Search != nil && len(*where.Search) > 1 {
		// 	q = q.Where(tsQuery(*where.Search))
		// }
		if where.MinRank != nil {
			q = q.Where(sq.GtOrEq{"rank": where.MinRank})
		}
	}
	return q
}
