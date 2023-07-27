package dbfinder

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/lib/pq"
)

type FVPair struct {
	EntityID      int
	FeedVersionID int
}

func StopTimeSelect(tpairs []FVPair, spairs []FVPair, permFilter *model.PermFilter, where *model.TripStopTimeFilter) sq.SelectBuilder {
	q := sq.StatementBuilder.Select(
		"gtfs_trips.journey_pattern_id",
		"gtfs_trips.journey_pattern_offset",
		"gtfs_trips.id AS trip_id",
		"gtfs_trips.feed_version_id",
		"sts.stop_id",
		"sts.arrival_time + gtfs_trips.journey_pattern_offset AS arrival_time",
		"sts.departure_time + gtfs_trips.journey_pattern_offset AS departure_time",
		"sts.stop_sequence",
		"sts.shape_dist_traveled",
		"sts.pickup_type",
		"sts.drop_off_type",
		"sts.timepoint",
		"sts.interpolated",
		"sts.stop_headsign",
		"sts.continuous_pickup",
		"sts.continuous_drop_off",
	).
		From("gtfs_trips").
		Join("feed_versions on feed_versions.id = gtfs_trips.feed_version_id").
		Join("current_feeds on current_feeds.id = feed_versions.feed_id").
		Join("gtfs_trips t2 ON t2.trip_id::text = gtfs_trips.journey_pattern_id AND gtfs_trips.feed_version_id = t2.feed_version_id").
		Join("gtfs_stop_times sts ON sts.trip_id = t2.id AND sts.feed_version_id = t2.feed_version_id").
		OrderBy("sts.stop_sequence, sts.arrival_time")

	if where != nil {
		if where.Start != nil {
			q = q.Where(sq.GtOrEq{"sts.departure_time + gtfs_trips.journey_pattern_offset": where.Start.Seconds})
		}
		if where.End != nil {
			q = q.Where(sq.LtOrEq{"sts.arrival_time + gtfs_trips.journey_pattern_offset": where.End.Seconds})
		}
	}
	if len(tpairs) > 0 {
		eids, fvids := pairKeys(tpairs)
		q = q.Where(sq.Eq{"gtfs_trips.id": eids, "sts.feed_version_id": fvids, "gtfs_trips.feed_version_id": fvids})
	}
	if len(spairs) > 0 {
		eids, fvids := pairKeys(spairs)
		q = q.Where(sq.Eq{"sts.stop_id": eids, "sts.feed_version_id": fvids})
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

func StopDeparturesSelect(spairs []FVPair, permFilter *model.PermFilter, where *model.StopTimeFilter) sq.SelectBuilder {
	// Where must already be set for local service date and timezone
	serviceDate := time.Now()
	if where != nil && where.ServiceDate != nil {
		serviceDate = where.ServiceDate.Val
	}
	sids, fvids := pairKeys(spairs)
	pqfvids := pq.Array(fvids)
	q := sq.StatementBuilder.Select(
		"gtfs_trips.journey_pattern_id",
		"gtfs_trips.journey_pattern_offset",
		"gtfs_trips.id AS trip_id",
		"gtfs_trips.feed_version_id",
		"sts.stop_id",
		"sts.arrival_time + gtfs_trips.journey_pattern_offset AS arrival_time",
		"sts.departure_time + gtfs_trips.journey_pattern_offset AS departure_time",
		"sts.stop_sequence",
		"sts.shape_dist_traveled",
		"sts.pickup_type",
		"sts.drop_off_type",
		"sts.timepoint",
		"sts.interpolated",
		"sts.stop_headsign",
		"sts.continuous_pickup",
		"sts.continuous_drop_off",
	).
		From("gtfs_trips").
		Join("feed_versions on feed_versions.id = gtfs_trips.feed_version_id").
		Join("current_feeds on current_feeds.id = feed_versions.feed_id").
		Join("gtfs_trips t2 ON t2.trip_id::text = gtfs_trips.journey_pattern_id AND gtfs_trips.feed_version_id = t2.feed_version_id").
		Join("gtfs_stop_times sts ON sts.trip_id = t2.id and sts.feed_version_id = t2.feed_version_id").
		JoinClause(`join lateral (
			select 
				min(stop_sequence), 
				max(stop_sequence) max 
			from gtfs_stop_times sts2 
			where 
				sts2.trip_id = t2.id 
				AND sts2.feed_version_id = t2.feed_version_id
			) trip_stop_sequence on true`).
		JoinClause(`join (
			SELECT
				id
			FROM
				gtfs_calendars
			WHERE
				start_date <= ?
				AND end_date >= ?
				AND (CASE EXTRACT(isodow FROM ?::date)
					WHEN 1 THEN monday = 1
					WHEN 2 THEN tuesday = 1
					WHEN 3 THEN wednesday = 1
					WHEN 4 THEN thursday = 1
					WHEN 5 THEN friday = 1
					WHEN 6 THEN saturday = 1
					WHEN 7 THEN sunday = 1
				END)
				AND feed_version_id = ANY(?)
			UNION
			SELect
				service_id as id
			FROM
				gtfs_calendar_dates
			WHERE
				date = ?
				AND exception_type = 1
				AND feed_version_id = ANY(?)
			EXCEPT
			SELECT service_id as id
			FROM gtfs_calendar_dates 
			WHERE 
				date = ? 
				AND exception_type = 2 
				AND feed_version_id = ANY(?)			
			) gc on gc.id = gtfs_trips.service_id`,
			serviceDate,
			serviceDate,
			serviceDate,
			pqfvids,
			serviceDate,
			pqfvids,
			serviceDate,
			pqfvids).
		Where(sq.Eq{"sts.stop_id": sids, "sts.feed_version_id": fvids}).
		OrderBy("departure_time") // base + offset

	if where != nil {
		if where.ExcludeFirst != nil && *where.ExcludeFirst {
			q = q.Where("sts.stop_sequence > trip_stop_sequence.min")
		}
		if where.ExcludeLast != nil && *where.ExcludeLast {
			q = q.Where("sts.stop_sequence < trip_stop_sequence.max")
		}
		if len(where.RouteOnestopIds) > 0 {
			if where.AllowPreviousRouteOnestopIds != nil && *where.AllowPreviousRouteOnestopIds {
				// Find a way to make this simpler, perhaps handle elsewhere
				sub := sq.StatementBuilder.
					Select("gtfs_routes.route_id", "feed_versions.feed_id").
					Distinct().Options("on (gtfs_routes.route_id, feed_versions.feed_id)").
					From("tl_route_onestop_ids").
					Join("gtfs_routes on gtfs_routes.id = tl_route_onestop_ids.route_id").
					Join("feed_versions on feed_versions.id = gtfs_routes.feed_version_id").
					Where(sq.Eq{"tl_route_onestop_ids.onestop_id": where.RouteOnestopIds}).
					OrderBy("gtfs_routes.route_id, feed_versions.feed_id, feed_versions.id DESC")
				subClause := sub.
					Prefix("JOIN (").
					Suffix(") tl_route_onestop_ids on tl_route_onestop_ids.route_id = gtfs_routes.route_id and tl_route_onestop_ids.feed_id = feed_versions.feed_id")
				q = q.
					Join("gtfs_routes on gtfs_routes.id = gtfs_trips.route_id").
					JoinClause(subClause)
			} else {
				q = q.
					Join("tl_route_onestop_ids on tl_route_onestop_ids.route_id = gtfs_trips.route_id").
					Where(sq.Eq{"tl_route_onestop_ids.onestop_id": where.RouteOnestopIds})

			}
		}
		if where.Start != nil && where.Start.Valid {
			where.StartTime = &where.Start.Seconds
		}
		if where.End != nil && where.End.Valid {
			where.EndTime = &where.End.Seconds
		}
		if where.StartTime != nil {
			q = q.Where(sq.GtOrEq{"sts.departure_time + gtfs_trips.journey_pattern_offset": *where.StartTime})
		}
		if where.EndTime != nil {
			q = q.Where(sq.LtOrEq{"sts.departure_time + gtfs_trips.journey_pattern_offset": *where.EndTime})
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

func pairKeys(spairs []FVPair) ([]int, []int) {
	eids := map[int]bool{}
	fvids := map[int]bool{}
	for _, v := range spairs {
		eids[v.EntityID] = true
		fvids[v.FeedVersionID] = true
	}
	var ueids []int
	for k := range eids {
		ueids = append(ueids, k)
	}
	var ufvids []int
	for k := range fvids {
		ufvids = append(ufvids, k)
	}
	return ueids, ufvids
}
