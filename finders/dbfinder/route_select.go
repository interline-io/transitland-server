package dbfinder

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
)

func RouteSelect(limit *int, after *model.Cursor, ids []int, active bool, permFilter *model.PermFilter, where *model.RouteFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.Select(
		"gtfs_routes.id",
		"gtfs_routes.feed_version_id",
		"gtfs_routes.agency_id",
		"gtfs_routes.route_id",
		"gtfs_routes.route_short_name",
		"gtfs_routes.route_long_name",
		"gtfs_routes.route_color",
		"gtfs_routes.route_desc",
		"gtfs_routes.route_type",
		"gtfs_routes.route_url",
		"gtfs_routes.route_text_color",
		"gtfs_routes.route_sort_order",
		"gtfs_routes.network_id",
		"gtfs_routes.as_route",
		"gtfs_routes.continuous_pickup",
		"gtfs_routes.continuous_drop_off",
		"current_feeds.id AS feed_id",
		"current_feeds.onestop_id AS feed_onestop_id",
		"feed_versions.sha1 AS feed_version_sha1",
		"coalesce(feed_version_route_onestop_ids.onestop_id, '') as onestop_id",
	).
		From("gtfs_routes").
		Join("feed_versions ON feed_versions.id = gtfs_routes.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		OrderBy("gtfs_routes.feed_version_id,gtfs_routes.id").
		Limit(checkLimit(limit))

	// Handle previous OnestopIds
	if where != nil {
		if where.OnestopID != nil {
			where.OnestopIds = append(where.OnestopIds, *where.OnestopID)
		}
		if len(where.OnestopIds) > 0 {
			q = q.Where(In("feed_version_route_onestop_ids.onestop_id", where.OnestopIds))
		}
		if len(where.OnestopIds) > 0 && where.AllowPreviousOnestopIds != nil && *where.AllowPreviousOnestopIds {
			sub := sq.StatementBuilder.
				Select("feed_version_route_onestop_ids.onestop_id", "feed_version_route_onestop_ids.entity_id", "feed_versions.feed_id").
				Distinct().Options("on (feed_version_route_onestop_ids.onestop_id, feed_version_route_onestop_ids.entity_id, feed_versions.feed_id)").
				From("feed_version_route_onestop_ids").
				Join("feed_versions on feed_versions.id = feed_version_route_onestop_ids.feed_version_id").
				Where(In("feed_version_route_onestop_ids.onestop_id", where.OnestopIds)).
				OrderBy("feed_version_route_onestop_ids.onestop_id, feed_version_route_onestop_ids.entity_id, feed_versions.feed_id, feed_versions.id DESC")
			subClause := sub.
				Prefix("LEFT JOIN (").
				Suffix(") feed_version_route_onestop_ids on feed_version_route_onestop_ids.entity_id = gtfs_routes.route_id and feed_version_route_onestop_ids.feed_id = feed_versions.feed_id")
			q = q.JoinClause(subClause)
		} else {
			q = q.JoinClause(`LEFT JOIN feed_version_route_onestop_ids ON feed_version_route_onestop_ids.entity_id = gtfs_routes.route_id and feed_version_route_onestop_ids.feed_version_id = gtfs_routes.feed_version_id`)
		}
	} else {
		q = q.JoinClause(`LEFT JOIN feed_version_route_onestop_ids ON feed_version_route_onestop_ids.entity_id = gtfs_routes.route_id and feed_version_route_onestop_ids.feed_version_id = gtfs_routes.feed_version_id`)
	}

	if where != nil {
		if len(where.AgencyIds) > 0 {
			q = q.Where(In("gtfs_routes.agency_id", where.AgencyIds))
		}
		if where.RouteID != nil {
			q = q.Where(sq.Eq{"gtfs_routes.route_id": *where.RouteID})
		}
		if where.RouteType != nil {
			where.RouteTypes = append(where.RouteTypes, *where.RouteType)
		}
		if len(where.RouteTypes) > 0 {
			q = q.Where(sq.Eq{"gtfs_routes.route_type": where.RouteTypes})
		}
		if where.FeedVersionSha1 != nil {
			q = q.Where("feed_versions.id = (select id from feed_versions where sha1 = ? limit 1)", *where.FeedVersionSha1)
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"current_feeds.onestop_id": *where.FeedOnestopID})
		}
		if where.Serviced != nil {
			q = q.JoinClause(`left join lateral (select tlrs.route_id from tl_route_stops tlrs where tlrs.route_id = gtfs_routes.id limit 1) scount on true`)
			if *where.Serviced {
				q = q.Where(sq.NotEq{"scount.route_id": nil})
			} else {
				q = q.Where(sq.Eq{"scount.route_id": nil})
			}
		}
		if where.Bbox != nil {
			q = q.JoinClause(`JOIN (
				SELECT DISTINCT ON (tlrs.route_id) tlrs.route_id FROM gtfs_stops
				JOIN tl_route_stops tlrs ON gtfs_stops.id = tlrs.stop_id
				WHERE ST_Intersects(gtfs_stops.geometry, ST_MakeEnvelope(?,?,?,?,4326))
			) tlrs_bbox on tlrs_bbox.route_id = gtfs_routes.id`, where.Bbox.MinLon, where.Bbox.MinLat, where.Bbox.MaxLon, where.Bbox.MaxLat)
		}
		if where.Within != nil && where.Within.Valid {
			q = q.JoinClause(`JOIN (
				SELECT DISTINCT ON (tlrs.route_id) tlrs.route_id FROM gtfs_stops
				JOIN tl_route_stops tlrs ON gtfs_stops.id = tlrs.stop_id
				WHERE ST_Intersects(gtfs_stops.geometry, ?)
			) tlrs_within on tlrs_within.route_id = gtfs_routes.id`, where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 1_000_000)
			q = q.JoinClause(`JOIN (
				SELECT DISTINCT ON (tlrs.route_id) tlrs.route_id FROM gtfs_stops
				JOIN tl_route_stops tlrs ON gtfs_stops.id = tlrs.stop_id
				WHERE ST_DWithin(gtfs_stops.geometry, ST_MakePoint(?,?), ?)
			) tlrs_near on tlrs_near.route_id = gtfs_routes.id`, where.Near.Lon, where.Near.Lat, radius)
		}
		if where.OperatorOnestopID != nil {
			q = q.
				Join("gtfs_agencies ON gtfs_agencies.id = gtfs_routes.agency_id").
				JoinClause("LEFT JOIN current_operators_in_feed coif ON coif.feed_id = feed_versions.feed_id AND coif.resolved_gtfs_agency_id = gtfs_agencies.agency_id").
				Where(sq.Eq{"coif.resolved_onestop_id": *where.OperatorOnestopID})
		}
		// Handle license filtering
		q = licenseFilter(where.License, q)

		// Text search
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsTableQuery("gtfs_routes", *where.Search)
			q = q.Column(rank).Where(wc)
		}
	}
	if active {
		q = q.Join("feed_states on feed_states.feed_version_id = gtfs_routes.feed_version_id")
	}
	if len(ids) > 0 {
		q = q.Where(In("gtfs_routes.id", ids))
	}

	// Handle cursor
	if after != nil && after.Valid && after.ID > 0 {
		// first check helps improve query performance
		if after.FeedVersionID == 0 {
			q = q.
				Where(sq.Expr("gtfs_routes.feed_version_id >= (select feed_version_id from gtfs_routes where id = ?)", after.ID)).
				Where(sq.Expr("(gtfs_routes.feed_version_id, gtfs_routes.id) > (select feed_version_id,id from gtfs_routes where id = ?)", after.ID))
		} else {
			q = q.
				Where(sq.Expr("gtfs_routes.feed_version_id >= ?", after.FeedVersionID)).
				Where(sq.Expr("(gtfs_routes.feed_version_id, gtfs_routes.id) > (?,?)", after.FeedVersionID, after.ID))
		}
	}

	// Handle permissions
	q = pfJoinCheckFv(q, permFilter)
	return q
}

func RouteStopBufferSelect(param model.RouteStopBufferParam) sq.SelectBuilder {
	r := checkFloat(param.Radius, 0, 2000.0)
	q := sq.StatementBuilder.
		Select(
			"ST_Collect(gtfs_stops.geometry::geometry)::geography AS stop_points",
			"ST_ConvexHull(ST_Collect(gtfs_stops.geometry::geometry))::geography AS stop_convexhull",
		).
		Column(sq.Expr("ST_Buffer(ST_Collect(gtfs_stops.geometry::geometry)::geography, ?, 4)::geography AS stop_buffer", r)). // column expr
		From("gtfs_stops").
		InnerJoin("tl_route_stops on tl_route_stops.stop_id = gtfs_stops.id").
		Where(sq.Eq{"tl_route_stops.route_id": param.EntityID})
	return q
}
