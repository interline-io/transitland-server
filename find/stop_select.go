package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func StopSelect(limit *int, after *model.Cursor, ids []int, active bool, userCheck *model.UserCheck, where *model.StopFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.Select(
		"gtfs_stops.*",
		"current_feeds.id AS feed_id",
		"current_feeds.onestop_id AS feed_onestop_id",
		"feed_versions.sha1 AS feed_version_sha1",
		"coalesce(tl_stop_onestop_ids.onestop_id, '') as onestop_id",
	).
		From("gtfs_stops").
		Join("feed_versions ON feed_versions.id = gtfs_stops.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		OrderBy("gtfs_stops.feed_version_id,gtfs_stops.id")
	distinct := false

	// Handle previous OnestopIds
	if where != nil {
		// Allow either a single onestop id or multiple
		if where.OnestopID != nil {
			where.OnestopIds = append(where.OnestopIds, *where.OnestopID)
		}
		if len(where.OnestopIds) > 0 {
			q = q.Where(sq.Eq{"tl_stop_onestop_ids.onestop_id": where.OnestopIds})
		}
		if len(where.OnestopIds) > 0 && where.AllowPreviousOnestopIds != nil && *where.AllowPreviousOnestopIds {
			sub := sq.StatementBuilder.
				Select("tl_stop_onestop_ids.onestop_id", "gtfs_stops.stop_id", "feed_versions.feed_id").
				Distinct().Options("on (tl_stop_onestop_ids.onestop_id, gtfs_stops.stop_id, feed_versions.feed_id)").
				From("tl_stop_onestop_ids").
				Join("gtfs_stops on gtfs_stops.id = tl_stop_onestop_ids.stop_id").
				Join("feed_versions on feed_versions.id = gtfs_stops.feed_version_id").
				Where(sq.Eq{"tl_stop_onestop_ids.onestop_id": where.OnestopIds}).
				OrderBy("tl_stop_onestop_ids.onestop_id, gtfs_stops.stop_id, feed_versions.feed_id, feed_versions.id DESC")
			subClause := sub.
				Prefix("LEFT JOIN (").
				Suffix(") tl_stop_onestop_ids on tl_stop_onestop_ids.stop_id = gtfs_stops.stop_id and tl_stop_onestop_ids.feed_id = feed_versions.feed_id")
			q = q.JoinClause(subClause)
		} else {
			q = q.JoinClause(`LEFT JOIN tl_stop_onestop_ids ON tl_stop_onestop_ids.stop_id = gtfs_stops.id`)
		}
	} else {
		q = q.JoinClause(`LEFT JOIN tl_stop_onestop_ids ON tl_stop_onestop_ids.stop_id = gtfs_stops.id`)
	}

	// Handle other clauses
	if where != nil {
		if where.Within != nil && where.Within.Valid {
			q = q.Where("ST_Intersects(gtfs_stops.geometry, ?)", where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 10_000)
			q = q.Where("ST_DWithin(gtfs_stops.geometry, ST_MakePoint(?,?), ?)", where.Near.Lon, where.Near.Lat, radius)
		}
		if where.StopCode != nil {
			q = q.Where(sq.Eq{"gtfs_stops.stop_code": where.StopCode})
		}
		if where.LocationType != nil {
			q = q.Where(sq.Eq{"gtfs_stops.location_type": where.LocationType})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"current_feeds.onestop_id": *where.FeedOnestopID})
		}
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_versions.sha1": *where.FeedVersionSha1})
		}
		if where.StopID != nil {
			q = q.Where(sq.Eq{"gtfs_stops.stop_id": *where.StopID})
		}
		if where.Serviced != nil {
			q = q.JoinClause(`left join lateral (select tlrs.stop_id from tl_route_stops tlrs where tlrs.stop_id = gtfs_stops.id limit 1) scount on true`)
			if *where.Serviced {
				q = q.Where(sq.NotEq{"scount.stop_id": nil})
			} else {
				q = q.Where(sq.Eq{"scount.stop_id": nil})
			}
		}
		// Served by agency ID
		if len(where.AgencyIds) > 0 {
			distinct = true
			q = q.Join("tl_route_stops on tl_route_stops.stop_id = gtfs_stops.id").Where(sq.Eq{"tl_route_stops.agency_id": where.AgencyIds})
		}
		// Accepts both route and operator Onestop IDs
		if len(where.ServedByOnestopIds) > 0 {
			distinct = true
			agencies := []string{}
			routes := []string{}
			for _, osid := range where.ServedByOnestopIds {
				if len(osid) == 0 {
				} else if osid[0] == 'o' {
					agencies = append(agencies, osid)
				} else if osid[0] == 'r' {
					routes = append(routes, osid)
				}
			}
			q = q.Join("tl_route_stops on tl_route_stops.stop_id = gtfs_stops.id")
			if len(routes) > 0 {
				q = q.Join("tl_route_onestop_ids on tl_route_stops.route_id = tl_route_onestop_ids.route_id")
			}
			if len(agencies) > 0 {
				q = q.
					Join("gtfs_agencies on gtfs_agencies.id = tl_route_stops.agency_id").
					Join("current_operators_in_feed coif ON coif.resolved_gtfs_agency_id = gtfs_agencies.agency_id AND coif.feed_id = current_feeds.id")
			}
			if len(routes) > 0 && len(agencies) > 0 {
				q = q.Where(sq.Or{sq.Eq{"tl_route_onestop_ids.onestop_id": routes}, sq.Eq{"coif.resolved_onestop_id": agencies}})
			} else if len(routes) > 0 {
				q = q.Where(sq.Eq{"tl_route_onestop_ids.onestop_id": routes})
			} else if len(agencies) > 0 {
				q = q.Where(sq.Eq{"coif.resolved_onestop_id": agencies})
			}
		}
		// Handle license filtering
		q = licenseFilter(where.License, q)
	}
	if distinct {
		q = q.Distinct().Options("on (gtfs_stops.feed_version_id,gtfs_stops.id)")
	}
	if active {
		q = q.Join("feed_states on feed_states.feed_version_id = gtfs_stops.feed_version_id")
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"gtfs_stops.id": ids})
	}
	if userCheck != nil {
		q = q.Where(sq.Or{sq.Eq{"feed_versions.feed_id": userCheck.AllowedFeeds}, sq.Eq{"feed_versions.id": userCheck.AllowedFeedVersions}})
	}
	if after != nil && after.Valid && after.ID > 0 {
		if after.FeedVersionID == 0 {
			q = q.Where(sq.Expr("(gtfs_stops.feed_version_id, gtfs_stops.id) > (select feed_version_id,id from gtfs_stops where id = ?)", after.ID))
		} else {
			q = q.Where(sq.Expr("(gtfs_stops.feed_version_id, gtfs_stops.id) > (?,?)", after.FeedVersionID, after.ID))
		}
	}

	// Outer query
	qView := sq.StatementBuilder.Select("t.*").FromSelect(q, "t")
	qView = qView.Limit(checkLimit(limit))
	if where != nil {
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsQuery(*where.Search)
			qView = qView.Column(rank).Where(wc)
		}
	}
	return qView
}
