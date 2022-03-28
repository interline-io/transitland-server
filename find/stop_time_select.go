package find

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/model"
	"github.com/lib/pq"
)

type FVPair struct {
	EntityID      int
	FeedVersionID int
}

func StopTimeSelect(tpairs []FVPair, spairs []FVPair, where *model.StopTimeFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.Select(
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
		Join("gtfs_trips t2 ON t2.trip_id::text = gtfs_trips.journey_pattern_id AND gtfs_trips.feed_version_id = t2.feed_version_id").
		Join("gtfs_stop_times sts ON sts.trip_id = t2.id").
		OrderBy("sts.stop_sequence, sts.arrival_time")
	if len(tpairs) > 0 {
		eids, fvids := pairKeys(tpairs)
		qView = qView.Where(sq.Eq{"gtfs_trips.id": eids, "sts.feed_version_id": fvids, "gtfs_trips.feed_version_id": fvids})
	}
	if len(spairs) > 0 {
		eids, fvids := pairKeys(spairs)
		qView = qView.Where(sq.Eq{"sts.stop_id": eids, "sts.feed_version_id": fvids})
	}
	return qView
}

func StopDeparturesSelect(spairs []FVPair, nowtime clock.Clock, tz string, where *model.StopTimeFilter) sq.SelectBuilder {
	serviceDate := time.Now()
	if nowtime != nil {
		serviceDate = nowtime.Now()
	}
	if where != nil && where.Next != nil {
		// Require a valid timezone
		if loc, err := time.LoadLocation(tz); err == nil {
			serviceDate = serviceDate.In(loc)
			st, et := 0, 0
			st = serviceDate.Hour()*3600 + serviceDate.Minute()*60 + serviceDate.Second()
			et = st + *where.Next
			where.StartTime = &st
			where.EndTime = &et
		}
	}
	if where != nil && where.ServiceDate != nil {
		serviceDate = where.ServiceDate.Time
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
		Join("gtfs_trips t2 ON t2.trip_id::text = gtfs_trips.journey_pattern_id AND gtfs_trips.feed_version_id = t2.feed_version_id").
		Join("gtfs_stop_times sts ON sts.trip_id = t2.id").
		JoinClause(`inner join (
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
				AND id NOT IN (
					SELECT service_id 
					FROM gtfs_calendar_dates 
					WHERE service_id = gtfs_calendars.id AND date = ? AND exception_type = 2 AND feed_version_id = ANY(?)
				)
			UNION
			SELect
				service_id as id
			FROM
				gtfs_calendar_dates
			WHERE
				date = ?
				AND exception_type = 1
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
		OrderBy("sts.arrival_time asc")
	if where != nil {
		if len(where.RouteOnestopIds) > 0 {
			q = q.
				Join("gtfs_routes on gtfs_routes.id = gtfs_trips.route_id").
				Join("tl_route_onestop_ids on tl_route_onestop_ids.route_id = gtfs_routes.id").
				Where(sq.Eq{"tl_route_onestop_ids.onestop_id": where.RouteOnestopIds})
		}
		if where.StartTime != nil {
			q = q.Where(sq.GtOrEq{"sts.departure_time + gtfs_trips.journey_pattern_offset": where.StartTime})
		}
		if where.EndTime != nil {
			q = q.Where(sq.LtOrEq{"sts.departure_time + gtfs_trips.journey_pattern_offset": where.EndTime})
		}
	}
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
