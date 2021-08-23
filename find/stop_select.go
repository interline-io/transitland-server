package find

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

func FindStops(atx sqlx.Ext, limit *int, after *int, ids []int, where *model.StopFilter) (ents []*model.Stop, err error) {
	active := false
	if where != nil && where.FeedVersionSha1 == nil {
		active = true
	}
	q := StopSelect(limit, after, ids, active, where)
	MustSelect(model.DB, q, &ents)
	return ents, nil
}

func StopSelect(limit *int, after *int, ids []int, active bool, where *model.StopFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.Select(
		"gtfs_stops.*",
		"current_feeds.id AS feed_id",
		"current_feeds.onestop_id AS feed_onestop_id",
		"feed_versions.sha1 AS feed_version_sha1",
		"tl_stop_onestop_ids.onestop_id",
	).
		From("gtfs_stops").
		Join("feed_versions ON feed_versions.id = gtfs_stops.feed_version_id").
		Join("current_feeds ON current_feeds.id = feed_versions.feed_id").
		JoinClause(`LEFT JOIN tl_stop_onestop_ids ON tl_stop_onestop_ids.stop_id = gtfs_stops.id`).
		Where(sq.Eq{"current_feeds.deleted_at": nil}).
		OrderBy("gtfs_stops.id")
	distinct := false
	if where != nil {
		if len(where.ServedByRouteTypes) > 0 {
			distinct = true
			qView = qView.
				Join("tl_route_stops on tl_route_stops.stop_id = gtfs_stops.id").
				Join("gtfs_routes on tl_route_stops.route_id = gtfs_routes.id").
				Where(sq.Eq{"gtfs_routes.route_type": where.ServedByRouteTypes})
		}
		if len(where.ServedByOnestopIds) > 0 {
			// Accepts both route and operator Onestop IDs
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
			qView = qView.Join("tl_route_stops tlrs2 on tlrs2.stop_id = gtfs_stops.id")
			if len(routes) > 0 {
				qView = qView.Join("tl_route_onestop_ids on tlrs2.route_id = tl_route_onestop_ids.route_id")
			}
			if len(agencies) > 0 {
				qView = qView.
					Join("gtfs_agencies on gtfs_agencies.id = tlrs2.agency_id").
					Join("current_operators_in_feed coif ON coif.resolved_gtfs_agency_id = gtfs_agencies.agency_id AND coif.feed_id = current_feeds.id")
			}
			if len(routes) > 0 && len(agencies) > 0 {
				qView = qView.Where(sq.Or{sq.Eq{"tl_route_onestop_ids.onestop_id": routes}, sq.Eq{"coif.resolved_onestop_id": agencies}})
			} else if len(routes) > 0 {
				qView = qView.Where(sq.Eq{"tl_route_onestop_ids.onestop_id": routes})
			} else if len(agencies) > 0 {
				qView = qView.Where(sq.Eq{"coif.resolved_onestop_id": agencies})
			}
		}
	}
	if distinct {
		qView = qView.Distinct().Options("on (gtfs_stops.id)")
	}
	if active {
		qView = qView.Join("feed_states on feed_states.feed_version_id = gtfs_stops.feed_version_id")
	}

	// Other filters
	q := sq.StatementBuilder.Select("t.*").FromSelect(qView, "t")
	if len(ids) > 0 {
		q = q.Where(sq.Eq{"t.id": ids})
	}
	if after != nil {
		q = q.Where(sq.Gt{"t.id": *after})
	}
	q = q.Limit(checkLimit(limit))

	if where != nil {
		if where.Search != nil && len(*where.Search) > 1 {
			rank, wc := tsQuery(*where.Search)
			q = q.Column(rank).Where(wc)
		}
		if where.FeedOnestopID != nil {
			q = q.Where(sq.Eq{"feed_onestop_id": *where.FeedOnestopID})
		}
		if where.FeedVersionSha1 != nil {
			q = q.Where(sq.Eq{"feed_version_sha1": *where.FeedVersionSha1})
		}
		if where.OnestopID != nil {
			q = q.Where(sq.Eq{"onestop_id": *where.OnestopID})
		}
		if where.StopID != nil {
			q = q.Where(sq.Eq{"stop_id": *where.StopID})
		}
		if where.Within != nil && where.Within.Valid {
			q = q.Where("ST_Intersects(t.geometry, ?)", where.Within)
		}
		if where.Near != nil {
			radius := checkFloat(&where.Near.Radius, 0, 10_000)
			q = q.Where("ST_DWithin(t.geometry, ST_MakePoint(?,?), ?)", where.Near.Lat, where.Near.Lon, radius)
		}
	}
	return q
}

func PathwaySelect(limit *int, after *int, ids []int, where *model.PathwayFilter) sq.SelectBuilder {
	q := quickSelectOrder("gtfs_pathways", limit, after, ids, "")
	if where != nil {
		if where.PathwayMode != nil {
			q = q.Where(sq.Eq{"pathway_mode": where.PathwayMode})
		}
	}
	return q
}
