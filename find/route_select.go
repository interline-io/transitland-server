package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

func FindRoutes(atx sqlx.Ext, limit *int, after *int, ids []int, where *model.RouteFilter) (ents []*model.Route, err error) {
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := RouteSelect(limit, after, ids, active, where)
	MustSelect(model.DB, q, &ents)
	return ents, nil
}

func RouteSelect(limit *int, after *int, ids []int, active bool, where *model.RouteFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.Select(
		"gtfs_routes.*",
		"tlrg.combined_geometry as geometry",
		"tlrg.generated AS geometry_generated",
		"current_feeds.id AS feed_id",
		"current_feeds.onestop_id AS feed_onestop_id",
		"feed_versions.sha1 AS feed_version_sha1",
		"tl_route_onestop_ids.onestop_id",
	).
		From("gtfs_routes").
		Join("feed_versions ON feed_versions.id = gtfs_routes.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		JoinClause("LEFT JOIN tl_route_onestop_ids ON tl_route_onestop_ids.route_id = gtfs_routes.id").
		JoinClause(`LEFT JOIN tl_route_geometries tlrg ON tlrg.route_id = gtfs_routes.id`).
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		OrderBy("gtfs_routes.id")
	if active {
		qView = qView.Join("feed_states on feed_states.feed_version_id = gtfs_routes.feed_version_id")
	}
	if where != nil {
		if where.Within != nil && where.Within.Valid {
			qView = qView.JoinClause(`JOIN (
				SELECT DISTINCT ON (tlrs.route_id) tlrs.route_id FROM gtfs_stops
				JOIN tl_route_stops tlrs ON gtfs_stops.id = tlrs.stop_id
				WHERE ST_Intersects(gtfs_stops.geometry, ?)
			) tlrs on tlrs.route_id = gtfs_routes.id`, where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 10_000)
			qView = qView.JoinClause(`JOIN (
				SELECT DISTINCT ON (tlrs.route_id) tlrs.route_id FROM gtfs_stops
				JOIN tl_route_stops tlrs ON gtfs_stops.id = tlrs.stop_id
				WHERE ST_DWithin(gtfs_stops.geometry, ST_MakePoint(?,?), ?)
			) tlrs on tlrs.route_id = gtfs_routes.id`, where.Near.Lon, where.Near.Lat, radius)
		}
		if where.OperatorOnestopID != nil {
			qView = qView.
				Join("gtfs_agencies ON gtfs_agencies.id = gtfs_routes.agency_id").
				JoinClause("LEFT JOIN current_operators_in_feed coif ON coif.feed_id = feed_versions.feed_id AND coif.resolved_gtfs_agency_id = gtfs_agencies.agency_id").
				Where(sq.Eq{"coif.resolved_onestop_id": *where.OperatorOnestopID})
		}
	}
	q := sq.StatementBuilder.Select("t.*").FromSelect(qView, "t")
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	q = q.Limit(checkLimit(limit))
	if where != nil {
		if where.Search != nil && len(*where.Search) > 0 {
			rank, wc := tsQuery(*where.Search)
			q = q.Column(rank).Where(wc)
		}
		if len(where.AgencyIds) > 0 {
			q = q.Where(sq.Eq{"agency_id": where.AgencyIds})
		}
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_version_sha1": *where.FeedVersionSha1})
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"feed_onestop_id": *where.FeedOnestopID})
		}
		if where.RouteID != nil {
			q = q.Where(sq.Eq{"route_id": *where.RouteID})
		}
		if where.OnestopID != nil {
			where.OnestopIds = append(where.OnestopIds, *where.OnestopID)
		}
		if len(where.OnestopIds) > 0 {
			q = q.Where(sq.Eq{"onestop_id": where.OnestopIds})
		}
		if where.RouteType != nil {
			q = q.Where(sq.Eq{"route_type": where.RouteType})
		}
	}
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
