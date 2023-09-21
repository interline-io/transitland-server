package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func StopSelect(limit *int, after *model.Cursor, ids []int, active bool, permFilter *model.PermFilter, where *model.StopFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.Select(
		"gtfs_stops.id",
		"gtfs_stops.feed_version_id",
		"gtfs_stops.stop_id",
		"gtfs_stops.stop_code",
		"gtfs_stops.stop_desc",
		"gtfs_stops.stop_name",
		"gtfs_stops.stop_timezone",
		"gtfs_stops.stop_url",
		"gtfs_stops.location_type",
		"gtfs_stops.wheelchair_boarding",
		"gtfs_stops.zone_id",
		"gtfs_stops.platform_code",
		"gtfs_stops.tts_stop_name",
		"gtfs_stops.geometry",
		"gtfs_stops.level_id",
		"gtfs_stops.parent_station",
		"gtfs_stops.area_id",
		"current_feeds.id AS feed_id",
		"current_feeds.onestop_id AS feed_onestop_id",
		"feed_versions.sha1 AS feed_version_sha1",
		"coalesce(tl_stop_onestop_ids.onestop_id, '') as onestop_id",
	).
		From("gtfs_stops").
		Join("feed_versions ON feed_versions.id = gtfs_stops.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		OrderBy("gtfs_stops.feed_version_id,gtfs_stops.id").
		Limit(checkLimit(limit))
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
			q = q.JoinClause(`LEFT JOIN tl_stop_onestop_ids ON tl_stop_onestop_ids.stop_id = gtfs_stops.id and tl_stop_onestop_ids.feed_version_id = gtfs_stops.feed_version_id`)
		}
	} else {
		q = q.JoinClause(`LEFT JOIN tl_stop_onestop_ids ON tl_stop_onestop_ids.stop_id = gtfs_stops.id and tl_stop_onestop_ids.feed_version_id = gtfs_stops.feed_version_id`)
	}

	// Handle other clauses
	if where != nil {
		if where.Bbox != nil {
			q = q.Where("ST_Intersects(gtfs_stops.geometry, ST_MakeEnvelope(?,?,?,?,4326))", where.Bbox.MinLon, where.Bbox.MinLat, where.Bbox.MaxLon, where.Bbox.MaxLat)
		}
		if where.Within != nil && where.Within.Valid {
			q = q.Where("ST_Intersects(gtfs_stops.geometry, ?)", where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 1_000_000)
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
			q = q.Where("feed_versions.id = (select id from feed_versions where sha1 = ? limit 1)", *where.FeedVersionSha1)
		}
		if where.StopID != nil {
			q = q.Where(sq.Eq{"gtfs_stops.stop_id": *where.StopID})
		}
		if where.Serviced != nil {
			q = q.JoinClause(`left join lateral (select tlrs_serviced.stop_id from tl_route_stops tlrs_serviced where tlrs_serviced.stop_id = gtfs_stops.id limit 1) scount on true`)
			if *where.Serviced {
				q = q.Where(sq.NotEq{"scount.stop_id": nil})
			} else {
				q = q.Where(sq.Eq{"scount.stop_id": nil})
			}
		}

		// Served by agency ID
		if len(where.AgencyIds) > 0 {
			distinct = true
			q = q.Join("tl_route_stops tlrs_agencies on tlrs_agencies.stop_id = gtfs_stops.id").Where(sq.Eq{"tlrs_agencies.agency_id": where.AgencyIds})
		}

		// Served by route type
		if where.ServedByRouteType != nil {
			q = q.JoinClause(`join lateral (select tlrs_rt.stop_id from tl_route_stops tlrs_rt join gtfs_routes on gtfs_routes.id = tlrs_rt.route_id where tlrs_rt.stop_id = gtfs_stops.id and gtfs_routes.route_type = ? limit 1) rt on true`, *where.ServedByRouteType)
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
			q = q.Join("tl_route_stops tlrs_routes on tlrs_routes.stop_id = gtfs_stops.id")
			if len(routes) > 0 {
				q = q.Join("tl_route_onestop_ids on tlrs_routes.route_id = tl_route_onestop_ids.route_id")
			}
			if len(agencies) > 0 {
				q = q.
					Join("gtfs_agencies on gtfs_agencies.id = tlrs_routes.agency_id").
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

		// Text search
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsTableQuery("gtfs_stops", *where.Search)
			q = q.Column(rank).Where(wc)
		}
	}

	if distinct {
		q = q.Distinct().Options("on (gtfs_stops.feed_version_id,gtfs_stops.id)")
	}
	if active {
		// in (select feed_version_id) from feed_states -- helps the query planner skip fv's that dont contribute to result
		q = q.
			Join("feed_states on feed_states.feed_version_id = gtfs_stops.feed_version_id").
			Where(sq.Expr("feed_versions.id in (select feed_version_id from feed_states)"))
	}
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"gtfs_stops.id": ids})
	}

	// Handle cursor
	if after != nil && after.Valid && after.ID > 0 {
		// first check helps improve query performance
		if after.FeedVersionID == 0 {
			q = q.
				Where(sq.Expr("gtfs_stops.feed_version_id >= (select feed_version_id from gtfs_stops where id = ?)", after.ID)).
				Where(sq.Expr("(gtfs_stops.feed_version_id, gtfs_stops.id) > (select feed_version_id,id from gtfs_stops where id = ?)", after.ID))
		} else {
			q = q.
				Where(sq.Expr("gtfs_stops.feed_version_id >= ?", after.FeedVersionID)).
				Where(sq.Expr("(gtfs_stops.feed_version_id, gtfs_stops.id) > (?,?)", after.FeedVersionID, after.ID))
		}
	}

	// Handle permissions
	q = q.
		Join("feed_states fsp on fsp.feed_id = current_feeds.id").
		Where(sq.Or{
			sq.Expr("fsp.public = true"),
			sq.Eq{"fsp.feed_id": permFilter.GetAllowedFeeds()},
			sq.Eq{"feed_versions.id": permFilter.GetAllowedFeedVersions()},
		})

	return q
}
