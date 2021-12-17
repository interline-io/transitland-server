package find

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/lib/pq"
)

func StopTimeSelect(limit *int, after *int, ids []int, tripids []int, stopids []int, where *model.StopTimeFilter) sq.SelectBuilder {
	qView := sq.StatementBuilder.Select(
		"gtfs_trips.journey_pattern_id",
		"gtfs_trips.journey_pattern_offset",
		"gtfs_trips.id AS trip_id",
		"gtfs_trips.feed_version_id",
		"st.stop_id",
		"st.arrival_time + gtfs_trips.journey_pattern_offset AS arrival_time",
		"st.departure_time + gtfs_trips.journey_pattern_offset AS departure_time",
		"st.stop_sequence",
		"st.shape_dist_traveled",
		"st.pickup_type",
		"st.drop_off_type",
		"st.timepoint",
		"st.interpolated",
		"st.stop_headsign",
	).
		From("gtfs_trips").
		Join("gtfs_trips t2 ON t2.trip_id::text = gtfs_trips.journey_pattern_id AND gtfs_trips.feed_version_id = t2.feed_version_id").
		Join("gtfs_stop_times st ON st.trip_id = t2.id").
		Limit(checkLimit(limit)).
		OrderBy("gtfs_trips.id asc, stop_sequence asc")
	if len(tripids) > 0 {
		qView = qView.Where(sq.Eq{"gtfs_trips.id": tripids})
	}
	if len(stopids) > 0 {
		qView = qView.Where(sq.Eq{"st.stop_id": stopids})
	}
	// Still need to wrap for lateral joins
	q := sq.StatementBuilder.Select("t.*").FromSelect(qView, "t")
	return q
}

type StopFVPair struct {
	FeedVersionID int
	StopID        int
}

func StopDeparturesSelect2(limit *int, after *int, stopids []StopFVPair, where *model.StopTimeFilter) sq.SelectBuilder {
	var sids []int
	fvidMap := map[int]bool{}
	for _, p := range stopids {
		sids = append(sids, p.StopID)
		fvidMap[p.FeedVersionID] = true
	}
	var fvids []int
	for k := range fvidMap {
		fvids = append(fvids, k)
	}
	q := sq.StatementBuilder.
		Select("sts.*").From("gtfs_stop_times sts").
		Where(sq.Eq{"stop_id": sids, "feed_version_id": fvids})
	return q
}

func StopDeparturesSelect(limit *int, after *int, stopids []StopFVPair, where *model.StopTimeFilter) sq.SelectBuilder {
	serviceDate := time.Now()
	if where != nil && where.Next != nil && where.Timezone != nil {
		// Require a valid timezone
		if loc, err := time.LoadLocation(*where.Timezone); err == nil {
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
	var sids []int
	fvidMap := map[int]bool{}
	for _, p := range stopids {
		sids = append(sids, p.StopID)
		fvidMap[p.FeedVersionID] = true
	}
	var fvids []int
	for k := range fvidMap {
		fvids = append(fvids, k)
	}
	pqfvids := pq.Array(fvids)
	// TODO: support journey patterns properly
	q := sq.StatementBuilder.Select("gtfs_stop_times.*").
		From("gtfs_stop_times").
		Join("gtfs_trips on gtfs_trips.id = gtfs_stop_times.trip_id").
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
		Where(sq.Eq{"gtfs_stop_times.stop_id": sids, "gtfs_stop_times.feed_version_id": fvids})
	if where != nil {
		if where.StartTime != nil {
			q = q.Where(sq.GtOrEq{"gtfs_stop_times.departure_time": where.StartTime})
		}
		if where.EndTime != nil {
			q = q.Where(sq.LtOrEq{"gtfs_stop_times.departure_time": where.EndTime})
		}
	}
	return q
}
