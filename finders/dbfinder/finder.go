package dbfinder

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/tlxy"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

////////

type Finder struct {
	Clock      clock.Clock
	db         sqlx.Ext
	adminCache *adminCache
	fvslCache  *clock.ServiceWindowCache
}

func NewFinder(db sqlx.Ext) *Finder {
	finder := &Finder{
		db:        db,
		fvslCache: clock.NewServiceWindowCache(db),
	}
	return finder
}

func (f *Finder) DBX() sqlx.Ext {
	return f.db
}

func (f *Finder) LoadAdmins(ctx context.Context) error {
	log.For(ctx).Trace().Msg("loading admins")
	adminCache := newAdminCache()
	if err := adminCache.LoadAdmins(ctx, f.db); err != nil {
		return err
	}
	f.adminCache = adminCache
	return nil
}

func (f *Finder) PermFilter(ctx context.Context) *model.PermFilter {
	return model.PermsForContext(ctx)
}

func (f *Finder) FindAgencies(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	var ents []*model.Agency
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) {
		active = false
	}
	q := AgencySelect(limit, after, ids, active, f.PermFilter(ctx), where)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) FindRoutes(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	var ents []*model.Route
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) {
		active = false
	}
	q := RouteSelect(limit, after, ids, active, f.PermFilter(ctx), where)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) FindStops(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	var ents []*model.Stop
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) {
		active = false
	}
	q := StopSelect(limit, after, ids, active, f.PermFilter(ctx), where)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) FindTrips(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	var ents []*model.Trip
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) || (where != nil && len(where.RouteIds) > 0) {
		active = false
	}
	q := TripSelect(limit, after, ids, active, f.PermFilter(ctx), where, nil)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) FindFeedVersions(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	var ents []*model.FeedVersion
	if err := dbutil.Select(ctx, f.db, FeedVersionSelect(limit, after, ids, f.PermFilter(ctx), where), &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) FindFeeds(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	var ents []*model.Feed
	if err := dbutil.Select(ctx, f.db, FeedSelect(limit, after, ids, f.PermFilter(ctx), where), &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) FindOperators(ctx context.Context, limit *int, after *model.Cursor, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	var ents []*model.Operator
	if err := dbutil.Select(ctx, f.db, OperatorSelect(limit, after, ids, f.PermFilter(ctx), where), &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

func (f *Finder) FindPlaces(ctx context.Context, limit *int, after *model.Cursor, ids []int, level *model.PlaceAggregationLevel, where *model.PlaceFilter) ([]*model.Place, error) {
	var ents []*model.Place
	q := PlaceSelect(limit, after, ids, level, f.PermFilter(ctx), where)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, err
	}
	return ents, nil
}

func (f *Finder) RouteStopBuffer(ctx context.Context, param *model.RouteStopBufferParam) ([]*model.RouteStopBuffer, error) {
	if param == nil {
		return nil, nil
	}
	var ents []*model.RouteStopBuffer
	q := RouteStopBufferSelect(*param)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, logErr(ctx, err)
	}
	return ents, nil
}

// Custom queries

func (f *Finder) FindFeedVersionServiceWindow(ctx context.Context, fvid int) (*model.ServiceWindow, error) {
	a, _, err := f.fvslCache.Get(ctx, fvid)
	if err != nil {
		return nil, err
	}
	// Get local time
	nowLocal := time.Now().In(a.Location)
	if model.ForContext(ctx).Clock != nil {
		nowLocal = model.ForContext(ctx).Clock.Now().In(a.Location)
	}
	// Copy back to model
	ret := &model.ServiceWindow{
		NowLocal:     nowLocal,
		StartDate:    a.StartDate,
		EndDate:      a.EndDate,
		FallbackWeek: a.FallbackWeek,
		Location:     a.Location,
	}
	return ret, err
}

// Simple ID loaders

func (f *Finder) TripsByID(ctx context.Context, ids []int) (ents []*model.Trip, errs []error) {
	ents, err := f.FindTrips(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Trip) int { return ent.ID }), nil
}

func (f *Finder) LevelsByID(ctx context.Context, ids []int) ([]*model.Level, []error) {
	var ents []*model.Level
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("gtfs_levels", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Level) int { return ent.ID }), nil
}

func (f *Finder) PathwaysByID(ctx context.Context, ids []int) ([]*model.Pathway, []error) {
	var ents []*model.Pathway
	err := dbutil.Select(ctx,
		f.db,
		PathwaySelect(nil, nil, ids, nil, nil),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Pathway) int { return ent.ID }), nil
}

func (f *Finder) CalendarsByID(ctx context.Context, ids []int) ([]*model.Calendar, []error) {
	var ents []*model.Calendar
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("gtfs_calendars", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Calendar) int { return ent.ID }), nil
}

func (f *Finder) ShapesByID(ctx context.Context, ids []int) ([]*model.Shape, []error) {
	var ents []*model.Shape
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("gtfs_shapes", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Shape) int { return ent.ID }), nil
}

func (f *Finder) FeedVersionsByID(ctx context.Context, ids []int) ([]*model.FeedVersion, []error) {
	ents, err := f.FindFeedVersions(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.FeedVersion) int { return ent.ID }), nil
}

func (f *Finder) FeedsByID(ctx context.Context, ids []int) ([]*model.Feed, []error) {
	ents, err := f.FindFeeds(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Feed) int { return ent.ID }), nil
}

func (f *Finder) StopExternalReferencesByStopID(ctx context.Context, ids []int) ([]*model.StopExternalReference, []error) {
	var ents []*model.StopExternalReference
	q := sq.StatementBuilder.Select("*").From("tl_stop_external_references").Where(In("id", ids))
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, []error{err}
	}
	return arrangeBy(ids, ents, func(ent *model.StopExternalReference) int { return ent.ID }), nil
}

func (f *Finder) StopObservationsByStopID(ctx context.Context, params []model.StopObservationParam) ([][]*model.StopObservation, []error) {
	type qent struct {
		StopID int
		model.StopObservation
	}
	qentGroups, err := paramGroupQuery(
		params,
		func(p model.StopObservationParam) (int, *model.StopObservationFilter, *int) {
			return p.StopID, p.Where, p.Limit
		},
		func(keys []int, where *model.StopObservationFilter, limit *int) (ents []*qent, err error) {
			// Prepare output
			q := sq.StatementBuilder.Select("gtfs_stops.id as stop_id", "obs.*").
				From("ext_performance_stop_observations obs").
				Join("gtfs_stops on gtfs_stops.stop_id = obs.to_stop_id").
				Where(In("gtfs_stops.id", keys)).
				Limit(100000)
			if where != nil {
				q = q.Where("obs.feed_version_id = ?", where.FeedVersionID)
				q = q.Where("obs.trip_start_date = ?", where.TripStartDate)
				q = q.Where("obs.source = ?", where.Source)
				// q = q.Where("start_time >= ?", where.StartTime)
				// q = q.Where("end_time <= ?", where.EndTime)
			}
			err = dbutil.Select(ctx, f.db, q, &ents)
			return ents, err
		},
		func(ent *qent) int {
			return ent.StopID
		},
	)
	return convertEnts(qentGroups, func(a *qent) *model.StopObservation { return &a.StopObservation }), err
}

func (f *Finder) RouteAttributesByRouteID(ctx context.Context, ids []int) ([]*model.RouteAttribute, []error) {
	var ents []*model.RouteAttribute
	q := sq.StatementBuilder.Select("*").From("ext_plus_route_attributes").Where(In("route_id", ids))
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, []error{err}
	}
	return arrangeBy(ids, ents, func(ent *model.RouteAttribute) int { return ent.RouteID }), nil
}

func (f *Finder) SegmentsByFeedVersionID(ctx context.Context, params []model.SegmentParam) ([][]*model.Segment, []error) {
	return paramGroupQuery(
		params,
		func(p model.SegmentParam) (int, string, *int) {
			return p.FeedVersionID, p.Layer, p.Limit
		},
		func(keys []int, layer string, limit *int) (ents []*model.Segment, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_segments", limit, nil, nil),
					"feed_versions",
					"id",
					"tl_segments",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Segment) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) SegmentsByID(ctx context.Context, ids []int) ([]*model.Segment, []error) {
	var ents []*model.Segment
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("tl_segments", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Segment) int { return ent.ID }), nil
}

func (f *Finder) SegmentsByRouteID(ctx context.Context, params []model.SegmentParam) ([][]*model.Segment, []error) {
	type qent struct {
		RouteID int
		model.Segment
	}
	qentGroups, err := paramGroupQuery(
		params,
		func(p model.SegmentParam) (int, *model.SegmentFilter, *int) {
			return p.RouteID, p.Where, p.Limit
		},
		func(keys []int, where *model.SegmentFilter, limit *int) (ents []*qent, err error) {
			q := sq.Select("s.id", "s.way_id", "s.geometry", "s.route_id").
				From("gtfs_routes").
				JoinClause(
					`join lateral (select distinct on (tl_segments.id, tl_segment_patterns.route_id) tl_segments.id, tl_segments.way_id, tl_segments.geometry, tl_segment_patterns.route_id from tl_segments join tl_segment_patterns on tl_segment_patterns.segment_id = tl_segments.id where tl_segment_patterns.route_id = gtfs_routes.id limit ?) s on true`,
					checkLimit(limit),
				).
				Where(In("gtfs_routes.id", keys))
			err = dbutil.Select(ctx,
				f.db,
				q,
				&ents,
			)
			return ents, err
		},
		func(ent *qent) int {
			return ent.RouteID
		},
	)
	return convertEnts(qentGroups, func(a *qent) *model.Segment { return &a.Segment }), err
}

func (f *Finder) SegmentPatternsByRouteID(ctx context.Context, params []model.SegmentPatternParam) ([][]*model.SegmentPattern, []error) {
	return paramGroupQuery(
		params,
		func(p model.SegmentPatternParam) (int, *model.SegmentPatternFilter, *int) {
			return p.RouteID, p.Where, p.Limit
		},
		func(keys []int, where *model.SegmentPatternFilter, limit *int) (ents []*model.SegmentPattern, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_segment_patterns", limit, nil, nil),
					"gtfs_routes",
					"id",
					"tl_segment_patterns",
					"route_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.SegmentPattern) int {
			return ent.RouteID
		},
	)
}

func (f *Finder) SegmentPatternsBySegmentID(ctx context.Context, params []model.SegmentPatternParam) ([][]*model.SegmentPattern, []error) {
	return paramGroupQuery(
		params,
		func(p model.SegmentPatternParam) (int, *model.SegmentPatternFilter, *int) {
			return p.SegmentID, p.Where, p.Limit
		},
		func(keys []int, where *model.SegmentPatternFilter, limit *int) (ents []*model.SegmentPattern, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_segment_patterns", limit, nil, nil),
					"tl_segments",
					"id",
					"tl_segment_patterns",
					"segment_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.SegmentPattern) int {
			return ent.SegmentID
		},
	)
}

func (f *Finder) AgenciesByID(ctx context.Context, ids []int) ([]*model.Agency, []error) {
	var ents []*model.Agency
	ents, err := f.FindAgencies(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Agency) int { return ent.ID }), nil

}

func (f *Finder) StopsByID(ctx context.Context, ids []int) ([]*model.Stop, []error) {
	ents, err := f.FindStops(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Stop) int { return ent.ID }), nil
}

func (f *Finder) RoutesByID(ctx context.Context, ids []int) ([]*model.Route, []error) {
	ents, err := f.FindRoutes(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Route) int { return ent.ID }), nil
}

func (f *Finder) CensusTableByID(ctx context.Context, ids []int) ([]*model.CensusTable, []error) {
	var ents []*model.CensusTable
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("tl_census_tables", nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.CensusTable) int { return ent.ID }), nil
}

func (f *Finder) FeedVersionGtfsImportByFeedVersionID(ctx context.Context, ids []int) ([]*model.FeedVersionGtfsImport, []error) {
	var ents []*model.FeedVersionGtfsImport
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("feed_version_gtfs_imports", nil, nil, nil).Where(In("feed_version_id", ids)),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.FeedVersionGtfsImport) int { return ent.FeedVersionID }), nil
}

func (f *Finder) FeedVersionServiceWindowByFeedVersionID(ctx context.Context, ids []int) ([]*model.FeedVersionServiceWindow, []error) {
	var ents []*model.FeedVersionServiceWindow
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("feed_version_service_windows", nil, nil, nil).Where(In("feed_version_id", ids)),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.FeedVersionServiceWindow) int { return ent.FeedVersionID }), nil
}

func (f *Finder) FeedStatesByFeedID(ctx context.Context, ids []int) ([]*model.FeedState, []error) {
	var ents []*model.FeedState
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("feed_states", nil, nil, nil).Where(In("feed_id", ids)),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.FeedState) int { return ent.FeedID }), nil
}

func (f *Finder) OperatorsByCOIF(ctx context.Context, ids []int) ([]*model.Operator, []error) {
	var ents []*model.Operator
	err := dbutil.Select(ctx,
		f.db,
		OperatorSelect(nil, nil, ids, f.PermFilter(ctx), nil),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Operator) int { return ent.ID }), nil
}

func (f *Finder) OperatorsByAgencyID(ctx context.Context, ids []int) ([]*model.Operator, []error) {
	var ents []*model.Operator
	err := dbutil.Select(ctx,
		f.db,
		OperatorsByAgencyID(nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Operator) int { return ent.AgencyID }), nil
}

// Param loaders

func (f *Finder) LevelsByParentStationID(ctx context.Context, params []model.LevelParam) ([][]*model.Level, []error) {
	return paramGroupQuery(
		params,
		func(p model.LevelParam) (int, bool, *int) {
			return p.ParentStationID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.Level, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("gtfs_levels", limit, nil, nil),
					"gtfs_stops",
					"id",
					"gtfs_levels",
					"parent_station",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Level) int {
			return ent.ParentStation.Int()
		},
	)
}

func (f *Finder) OperatorsByFeedID(ctx context.Context, params []model.OperatorParam) ([][]*model.Operator, []error) {
	return paramGroupQuery(
		params,
		func(p model.OperatorParam) (int, *model.OperatorFilter, *int) {
			return p.FeedID, p.Where, p.Limit
		},
		func(keys []int, where *model.OperatorFilter, limit *int) (ents []*model.Operator, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					OperatorSelectBase(true, nil),
					"current_feeds",
					"id",
					"coif",
					"feed_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Operator) int {
			return ent.FeedID
		},
	)
}

func (f *Finder) FeedFetchesByFeedID(ctx context.Context, params []model.FeedFetchParam) ([][]*model.FeedFetch, []error) {
	return paramGroupQuery(
		params,
		func(p model.FeedFetchParam) (int, *model.FeedFetchFilter, *int) {
			return p.FeedID, p.Where, p.Limit
		},
		func(keys []int, where *model.FeedFetchFilter, limit *int) (ents []*model.FeedFetch, err error) {
			q := sq.StatementBuilder.
				Select("*").
				From("feed_fetches").
				Limit(checkLimit(limit)).
				OrderBy("feed_fetches.fetched_at desc")
			if where != nil {
				if where.Success != nil {
					q = q.Where(sq.Eq{"success": *where.Success})
				}
			}
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(q, "current_feeds", "id", "feed_fetches", "feed_id", keys),
				&ents,
			)
			return ents, err
		},
		func(ent *model.FeedFetch) int {
			return ent.FeedID
		},
	)
}

func (f *Finder) FeedsByOperatorOnestopID(ctx context.Context, params []model.FeedParam) ([][]*model.Feed, []error) {
	type qent struct {
		OperatorOnestopID string
		model.Feed
	}
	qentGroups, err := paramGroupQuery(
		params,
		func(p model.FeedParam) (string, *model.FeedFilter, *int) {
			return p.OperatorOnestopID, p.Where, p.Limit
		},
		func(keys []string, where *model.FeedFilter, limit *int) (ents []*qent, err error) {
			q := FeedSelect(nil, nil, nil, f.PermFilter(ctx), where).
				Distinct().Options("on (coif.resolved_onestop_id, current_feeds.id)").
				Column("coif.resolved_onestop_id as operator_onestop_id").
				Join("current_operators_in_feed coif on coif.feed_id = current_feeds.id").
				Where(In("coif.resolved_onestop_id", keys))
			err = dbutil.Select(ctx,
				f.db,
				q,
				&ents,
			)
			return ents, err
		},
		func(ent *qent) string {
			return ent.OperatorOnestopID
		},
	)
	return convertEnts(qentGroups, func(a *qent) *model.Feed { return &a.Feed }), err
}

func (f *Finder) FrequenciesByTripID(ctx context.Context, params []model.FrequencyParam) ([][]*model.Frequency, []error) {
	return paramGroupQuery(
		params,
		func(p model.FrequencyParam) (int, bool, *int) {
			return p.TripID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.Frequency, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("gtfs_frequencies", limit, nil, nil),
					"gtfs_trips",
					"id",
					"gtfs_frequencies",
					"trip_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Frequency) int {
			return ent.TripID.Int()
		},
	)
}

func (f *Finder) StopTimesByTripID(ctx context.Context, params []model.TripStopTimeParam) ([][]*model.StopTime, []error) {
	return paramGroupQuery(
		params,
		func(p model.TripStopTimeParam) (FVPair, *model.TripStopTimeFilter, *int) {
			a := FVPair{FeedVersionID: p.FeedVersionID, EntityID: p.TripID}
			return a, p.Where, p.Limit
		},
		func(keys []FVPair, where *model.TripStopTimeFilter, limit *int) (ents []*model.StopTime, err error) {
			err = dbutil.Select(ctx,
				f.db,
				StopTimeSelect(keys, nil, where),
				&ents,
			)
			return ents, err
		},
		func(ent *model.StopTime) FVPair {
			return FVPair{FeedVersionID: ent.FeedVersionID, EntityID: ent.TripID.Int()}
		},
	)
}

func (f *Finder) StopTimesByStopID(ctx context.Context, params []model.StopTimeParam) ([][]*model.StopTime, []error) {
	// We need to split by feed version id to extract service window
	// Fields must be public
	type fvParamGroup struct {
		FeedVersionID int
		Where         *model.StopTimeFilter
	}
	return paramGroupQuery(
		params,
		func(p model.StopTimeParam) (FVPair, fvParamGroup, *int) {
			a := FVPair{FeedVersionID: p.FeedVersionID, EntityID: p.StopID}
			w := fvParamGroup{FeedVersionID: p.FeedVersionID, Where: p.Where}
			return a, w, p.Limit
		},
		func(keys []FVPair, fvwhere fvParamGroup, limit *int) (ents []*model.StopTime, err error) {
			fvsw, err := f.FindFeedVersionServiceWindow(ctx, fvwhere.FeedVersionID)
			if err != nil {
				return nil, err
			}
			// Run separate queries for each possible service day
			for _, w := range StopTimeFilterExpand(fvwhere.Where, fvsw) {
				var serviceDate *tt.Date
				if w != nil && w.ServiceDate != nil {
					serviceDate = w.ServiceDate
				}
				var sts []*model.StopTime
				var q sq.SelectBuilder
				if serviceDate != nil {
					// Get stops on a specified day
					q = StopDeparturesSelect(keys, w)
				} else {
					// Otherwise get all stop_times for stop
					q = StopTimeSelect(nil, keys, nil)
				}
				// Run query
				if err := dbutil.Select(ctx,
					f.db,
					q,
					&sts,
				); err != nil {
					return nil, err
				}
				// Set service date based on StopTimeFilter, and adjust calendar date if needed
				if serviceDate != nil {
					for _, ent := range sts {
						ent.ServiceDate.Set(serviceDate.Val)
						if ent.ArrivalTime.Val > 24*60*60 {
							ent.Date.Set(serviceDate.Val.AddDate(0, 0, 1))
						} else {
							ent.Date.Set(serviceDate.Val)
						}
					}
				}
				ents = append(ents, sts...)
			}
			return ents, err
		},
		func(ent *model.StopTime) FVPair {
			return FVPair{FeedVersionID: ent.FeedVersionID, EntityID: ent.StopID.Int()}
		},
	)
}

func (f *Finder) RouteStopsByStopID(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	return paramGroupQuery(
		params,
		func(p model.RouteStopParam) (int, bool, *int) {
			return p.StopID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.RouteStop, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("tl_route_stops", limit, nil, nil, "stop_id"),
					"gtfs_stops",
					"id",
					"tl_route_stops",
					"stop_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.RouteStop) int {
			return ent.StopID
		},
	)
}

func (f *Finder) StopsByRouteID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	type qent struct {
		RouteID int
		model.Stop
	}
	qentGroups, err := paramGroupQuery(
		params,
		func(p model.StopParam) (int, *model.StopFilter, *int) {
			return p.RouteID, p.Where, p.Limit
		},
		func(keys []int, where *model.StopFilter, limit *int) (ents []*qent, err error) {
			qso := StopSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), where)
			qso = qso.Join("tl_route_stops on tl_route_stops.stop_id = gtfs_stops.id").Where(In("route_id", keys)).Column("route_id")
			err = dbutil.Select(ctx,
				f.db,
				qso,
				&ents,
			)
			return ents, err
		},
		func(ent *qent) int {
			return ent.RouteID
		},
	)
	return convertEnts(qentGroups, func(a *qent) *model.Stop { return &a.Stop }), err
}

func (f *Finder) RouteStopsByRouteID(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	return paramGroupQuery(
		params,
		func(p model.RouteStopParam) (int, bool, *int) {
			return p.RouteID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.RouteStop, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("tl_route_stops", limit, nil, nil, "stop_id"),
					"gtfs_routes",
					"id",
					"tl_route_stops",
					"route_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.RouteStop) int {
			return ent.RouteID
		},
	)
}

func (f *Finder) RouteHeadwaysByRouteID(ctx context.Context, params []model.RouteHeadwayParam) ([][]*model.RouteHeadway, []error) {
	return paramGroupQuery(
		params,
		func(p model.RouteHeadwayParam) (int, bool, *int) {
			return p.RouteID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.RouteHeadway, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("tl_route_headways", limit, nil, nil, "route_id"),
					"gtfs_routes",
					"id",
					"tl_route_headways",
					"route_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.RouteHeadway) int {
			return ent.RouteID
		},
	)
}

func (f *Finder) RouteStopPatternsByRouteID(ctx context.Context, params []model.RouteStopPatternParam) ([][]*model.RouteStopPattern, []error) {
	// TODO: Add limit option in resolver
	return paramGroupQuery(
		params,
		func(p model.RouteStopPatternParam) (int, bool, *int) {
			return p.RouteID, false, nil
		},
		func(keys []int, where bool, limit *int) (ents []*model.RouteStopPattern, err error) {
			q := sq.StatementBuilder.
				Select("route_id", "direction_id", "stop_pattern_id", "count(*) as count").
				From("gtfs_trips").
				Where(In("route_id", keys)).
				GroupBy("route_id,direction_id,stop_pattern_id").
				OrderBy("route_id,count desc").
				Limit(1000)
			err = dbutil.Select(ctx,
				f.db,
				q,
				&ents,
			)
			return ents, err
		},
		func(ent *model.RouteStopPattern) int {
			return ent.RouteID
		},
	)
}

func (f *Finder) FeedVersionFileInfosByFeedVersionID(ctx context.Context, params []model.FeedVersionFileInfoParam) ([][]*model.FeedVersionFileInfo, []error) {
	return paramGroupQuery(
		params,
		func(p model.FeedVersionFileInfoParam) (int, bool, *int) {
			return p.FeedVersionID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.FeedVersionFileInfo, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("feed_version_file_infos", limit, nil, nil, "feed_version_id"),
					"feed_versions",
					"id",
					"feed_version_file_infos",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.FeedVersionFileInfo) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) StopsByParentStopID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	return paramGroupQuery(
		params,
		func(p model.StopParam) (int, *model.StopFilter, *int) {
			return p.ParentStopID, p.Where, p.Limit
		},
		func(keys []int, where *model.StopFilter, limit *int) (ents []*model.Stop, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					StopSelect(limit, nil, nil, false, f.PermFilter(ctx), where),
					"gtfs_stops",
					"id",
					"gtfs_stops",
					"parent_station",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Stop) int {
			return ent.ParentStation.Int()
		},
	)
}

func (f *Finder) TargetStopsByStopID(ctx context.Context, ids []int) ([]*model.Stop, []error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// TODO: this is moderately cursed
	type qlookup struct {
		SourceID int
		*model.Stop
	}
	var qents []*qlookup
	q := sq.
		Select("t.*", "tlse.id as source_id").
		FromSelect(StopSelect(nil, nil, nil, true, f.PermFilter(ctx), nil), "t").
		Join("tl_stop_external_references tlse on tlse.target_feed_onestop_id = t.feed_onestop_id and tlse.target_stop_id = t.stop_id").
		Where(In("tlse.id", ids))
	if err := dbutil.Select(ctx,
		f.db,
		q,
		&qents,
	); err != nil {
		return nil, logExtendErr(ctx, 0, err)
	}
	group := map[int]*model.Stop{}
	for _, ent := range qents {
		group[ent.SourceID] = ent.Stop
	}
	var ents []*model.Stop
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) FeedVersionsByFeedID(ctx context.Context, params []model.FeedVersionParam) ([][]*model.FeedVersion, []error) {
	return paramGroupQuery(
		params,
		func(p model.FeedVersionParam) (int, *model.FeedVersionFilter, *int) {
			return p.FeedID, p.Where, p.Limit
		},
		func(keys []int, where *model.FeedVersionFilter, limit *int) ([]*model.FeedVersion, error) {
			var ents []*model.FeedVersion
			err := dbutil.Select(ctx,
				f.db,
				lateralWrap(
					FeedVersionSelect(limit, nil, nil, f.PermFilter(ctx), where),
					"current_feeds",
					"id",
					"feed_versions",
					"feed_id",
					keys,
				),
				&ents,
			)
			if err != nil {
				return nil, err
			}
			return ents, nil
		},
		func(ent *model.FeedVersion) int {
			return ent.FeedID
		},
	)
}

func (f *Finder) ValidationReportsByFeedVersionID(ctx context.Context, params []model.ValidationReportParam) ([][]*model.ValidationReport, []error) {
	return paramGroupQuery(
		params,
		func(p model.ValidationReportParam) (int, *model.ValidationReportFilter, *int) {
			return p.FeedVersionID, p.Where, p.Limit
		},
		func(keys []int, where *model.ValidationReportFilter, limit *int) ([]*model.ValidationReport, error) {
			q := sq.StatementBuilder.
				Select("*").
				From("tl_validation_reports").
				Limit(checkLimit(limit)).
				OrderBy("tl_validation_reports.created_at desc, tl_validation_reports.id desc")
			if where != nil {
				if len(where.ReportIds) > 0 {
					q = q.Where(In("tl_validation_reports.id", where.ReportIds))
				}
				if where.Success != nil {
					q = q.Where(sq.Eq{"success": where.Success})
				}
				if where.Validator != nil {
					q = q.Where(sq.Eq{"validator": where.Validator})
				}
				if where.ValidatorVersion != nil {
					q = q.Where(sq.Eq{"validator_version": where.ValidatorVersion})
				}
				if where.IncludesRt != nil {
					q = q.Where(sq.Eq{"includes_rt": where.IncludesRt})
				}
				if where.IncludesStatic != nil {
					q = q.Where(sq.Eq{"includes_static": where.IncludesStatic})
				}
			}
			var ents []*model.ValidationReport
			err := dbutil.Select(ctx,
				f.db,
				lateralWrap(
					q,
					"feed_versions",
					"id",
					"tl_validation_reports",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.ValidationReport) int { return ent.FeedVersionID },
	)
}

func (f *Finder) ValidationReportErrorGroupsByValidationReportID(ctx context.Context, params []model.ValidationReportErrorGroupParam) ([][]*model.ValidationReportErrorGroup, []error) {
	return paramGroupQuery(
		params,
		func(p model.ValidationReportErrorGroupParam) (int, bool, *int) {
			return p.ValidationReportID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) ([]*model.ValidationReportErrorGroup, error) {
			var ents []*model.ValidationReportErrorGroup
			err := dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_validation_report_error_groups", limit, nil, nil),
					"tl_validation_reports",
					"id",
					"tl_validation_report_error_groups",
					"validation_report_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.ValidationReportErrorGroup) int { return ent.ValidationReportID },
	)
}

func (f *Finder) ValidationReportErrorExemplarsByValidationReportErrorGroupID(ctx context.Context, params []model.ValidationReportErrorExemplarParam) ([][]*model.ValidationReportError, []error) {
	return paramGroupQuery(
		params,
		func(p model.ValidationReportErrorExemplarParam) (int, bool, *int) {
			return p.ValidationReportGroupID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) ([]*model.ValidationReportError, error) {
			var ents []*model.ValidationReportError
			err := dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelect("tl_validation_report_error_exemplars", limit, nil, nil),
					"tl_validation_report_error_groups",
					"id",
					"tl_validation_report_error_exemplars",
					"validation_report_error_group_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.ValidationReportError) int { return ent.ValidationReportErrorGroupID },
	)
}

func (f *Finder) AgencyPlacesByAgencyID(ctx context.Context, params []model.AgencyPlaceParam) ([][]*model.AgencyPlace, []error) {
	return paramGroupQuery(
		params,
		func(p model.AgencyPlaceParam) (int, *model.AgencyPlaceFilter, *int) {
			return p.AgencyID, p.Where, p.Limit
		},
		func(keys []int, where *model.AgencyPlaceFilter, limit *int) (ents []*model.AgencyPlace, err error) {
			q := sq.StatementBuilder.Select(
				"tl_agency_places.agency_id",
				"tl_agency_places.rank",
				"tl_agency_places.name as city_name",
				"tl_agency_places.adm0name as adm0_name",
				"tl_agency_places.adm1name as adm1_name",
				"ne_admin.iso_a2 as adm0_iso",
				"ne_admin.iso_3166_2 as adm1_iso",
			).
				From("tl_agency_places").
				Join("ne_10m_admin_1_states_provinces ne_admin on ne_admin.name = tl_agency_places.adm1name and ne_admin.admin = tl_agency_places.adm0name")

			if where != nil {
				if where.MinRank != nil {
					q = q.Where(sq.GtOrEq{"rank": where.MinRank})
				}
			}
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					q,
					"gtfs_agencies",
					"id",
					"tl_agency_places",
					"agency_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.AgencyPlace) int {
			return ent.AgencyID
		},
	)
}

func (f *Finder) RouteGeometriesByRouteID(ctx context.Context, params []model.RouteGeometryParam) ([][]*model.RouteGeometry, []error) {
	return paramGroupQuery(
		params,
		func(p model.RouteGeometryParam) (int, bool, *int) {
			return p.RouteID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.RouteGeometry, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("tl_route_geometries", limit, nil, nil, "route_id"),
					"gtfs_routes",
					"id",
					"tl_route_geometries",
					"route_id",
					keys,
				),
				&ents,
			)

			return ents, err
		},
		func(ent *model.RouteGeometry) int {
			return ent.RouteID
		},
	)
}

func (f *Finder) TripsByRouteID(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
	// We need to split by feed version id to extract service window
	// Fields must be public
	type fvParamGroup struct {
		FeedVersionID int
		Where         *model.TripFilter
	}
	return paramGroupQuery(
		params,
		func(p model.TripParam) (int, fvParamGroup, *int) {
			return p.RouteID, fvParamGroup{FeedVersionID: p.FeedVersionID, Where: p.Where}, p.Limit
		},
		func(keys []int, fvwhere fvParamGroup, limit *int) (ents []*model.Trip, err error) {
			fvsw, err := f.FindFeedVersionServiceWindow(ctx, fvwhere.FeedVersionID)
			if err != nil {
				return nil, err
			}
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					TripSelect(limit, nil, nil, false, f.PermFilter(ctx), fvwhere.Where, fvsw),
					"gtfs_routes",
					"id",
					"gtfs_trips",
					"route_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Trip) int {
			return ent.RouteID.Int()
		},
	)
}

func (f *Finder) RoutesByAgencyID(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
	return paramGroupQuery(
		params,
		func(p model.RouteParam) (int, *model.RouteFilter, *int) {
			return p.AgencyID, p.Where, p.Limit
		},
		func(keys []int, where *model.RouteFilter, limit *int) (ents []*model.Route, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					RouteSelect(limit, nil, nil, false, f.PermFilter(ctx), where),
					"gtfs_agencies",
					"id",
					"gtfs_routes",
					"agency_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Route) int {
			return ent.AgencyID.Int()
		},
	)
}

func (f *Finder) AgenciesByFeedVersionID(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
	return paramGroupQuery(
		params,
		func(p model.AgencyParam) (int, *model.AgencyFilter, *int) {
			return p.FeedVersionID, p.Where, p.Limit
		},
		func(keys []int, where *model.AgencyFilter, limit *int) (ents []*model.Agency, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					AgencySelect(limit, nil, nil, false, f.PermFilter(ctx), where),
					"feed_versions",
					"id",
					"gtfs_agencies",
					"feed_version_id",
					keys,
				),
				&ents,
			)

			return ents, err
		},
		func(ent *model.Agency) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) AgenciesByOnestopID(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
	return paramGroupQuery(
		params,
		func(p model.AgencyParam) (string, *model.AgencyFilter, *int) {
			a := ""
			if p.OnestopID != nil {
				a = *p.OnestopID
			}
			return a, p.Where, p.Limit
		},
		func(keys []string, where *model.AgencyFilter, limit *int) (ents []*model.Agency, err error) {
			err = dbutil.Select(ctx,
				f.db,
				AgencySelect(limit, nil, nil, true, f.PermFilter(ctx), nil).Where(In("coif.resolved_onestop_id", keys)),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Agency) string {
			return ent.OnestopID
		},
	)
}

func (f *Finder) StopsByFeedVersionID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	return paramGroupQuery(
		params,
		func(p model.StopParam) (int, *model.StopFilter, *int) {
			return p.FeedVersionID, p.Where, p.Limit
		},
		func(keys []int, where *model.StopFilter, limit *int) (ents []*model.Stop, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					StopSelect(limit, nil, nil, false, f.PermFilter(ctx), where),
					"feed_versions",
					"id",
					"gtfs_stops",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Stop) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) StopsByLevelID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	return paramGroupQuery(
		params,
		func(p model.StopParam) (int, *model.StopFilter, *int) {
			return p.LevelID, p.Where, p.Limit
		},
		func(keys []int, where *model.StopFilter, limit *int) (ents []*model.Stop, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					StopSelect(limit, nil, nil, false, f.PermFilter(ctx), where),
					"gtfs_levels",
					"id",
					"gtfs_stops",
					"level_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Stop) int {
			return ent.LevelID.Int()
		},
	)
}

func (f *Finder) TripsByFeedVersionID(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
	// We need to split by feed version id to extract service window
	// Fields must be public
	type fvParamGroup struct {
		FeedVersionID int
		Where         *model.TripFilter
	}
	return paramGroupQuery(
		params,
		func(p model.TripParam) (int, fvParamGroup, *int) {
			return p.FeedVersionID, fvParamGroup{FeedVersionID: p.FeedVersionID, Where: p.Where}, p.Limit
		},
		func(keys []int, fvwhere fvParamGroup, limit *int) (ents []*model.Trip, err error) {
			fvsw, err := f.FindFeedVersionServiceWindow(ctx, fvwhere.FeedVersionID)
			if err != nil {
				return nil, err
			}
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					TripSelect(limit, nil, nil, false, f.PermFilter(ctx), fvwhere.Where, fvsw),
					"feed_versions",
					"id",
					"gtfs_trips",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Trip) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) FeedInfosByFeedVersionID(ctx context.Context, params []model.FeedInfoParam) ([][]*model.FeedInfo, []error) {
	return paramGroupQuery(
		params,
		func(p model.FeedInfoParam) (int, bool, *int) {
			return p.FeedVersionID, false, p.Limit
		},
		func(keys []int, where bool, limit *int) (ents []*model.FeedInfo, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("gtfs_feed_infos", limit, nil, nil, "feed_version_id"),
					"feed_versions",
					"id",
					"gtfs_feed_infos",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.FeedInfo) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) RoutesByFeedVersionID(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
	return paramGroupQuery(
		params,
		func(p model.RouteParam) (int, *model.RouteFilter, *int) {
			return p.FeedVersionID, p.Where, p.Limit
		},
		func(keys []int, where *model.RouteFilter, limit *int) (ents []*model.Route, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					RouteSelect(limit, nil, nil, false, f.PermFilter(ctx), where),
					"feed_versions",
					"id",
					"gtfs_routes",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Route) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) FeedVersionServiceLevelsByFeedVersionID(ctx context.Context, params []model.FeedVersionServiceLevelParam) ([][]*model.FeedVersionServiceLevel, []error) {
	return paramGroupQuery(
		params,
		func(p model.FeedVersionServiceLevelParam) (int, *model.FeedVersionServiceLevelFilter, *int) {
			return p.FeedVersionID, p.Where, p.Limit
		},
		func(keys []int, where *model.FeedVersionServiceLevelFilter, limit *int) (ents []*model.FeedVersionServiceLevel, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					FeedVersionServiceLevelSelect(limit, nil, nil, f.PermFilter(ctx), where),
					"feed_versions",
					"id",
					"feed_version_service_levels",
					"feed_version_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.FeedVersionServiceLevel) int {
			return ent.FeedVersionID
		},
	)
}

func (f *Finder) PathwaysByFromStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	return paramGroupQuery(
		params,
		func(p model.PathwayParam) (int, *model.PathwayFilter, *int) {
			return p.FromStopID, p.Where, p.Limit
		},
		func(keys []int, where *model.PathwayFilter, limit *int) (ents []*model.Pathway, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					PathwaySelect(limit, nil, nil, f.PermFilter(ctx), where),
					"gtfs_stops",
					"id",
					"gtfs_pathways",
					"from_stop_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Pathway) int {
			return ent.FromStopID.Int()
		},
	)
}

func (f *Finder) PathwaysByToStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	return paramGroupQuery(
		params,
		func(p model.PathwayParam) (int, *model.PathwayFilter, *int) {
			return p.ToStopID, p.Where, p.Limit
		},
		func(keys []int, where *model.PathwayFilter, limit *int) (ents []*model.Pathway, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					PathwaySelect(limit, nil, nil, f.PermFilter(ctx), where),
					"gtfs_stops",
					"id",
					"gtfs_pathways",
					"to_stop_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.Pathway) int {
			return ent.ToStopID.Int()
		},
	)
}

func (f *Finder) CalendarDatesByServiceID(ctx context.Context, params []model.CalendarDateParam) ([][]*model.CalendarDate, []error) {
	return paramGroupQuery(
		params,
		func(p model.CalendarDateParam) (int, *model.CalendarDateFilter, *int) {
			return p.ServiceID, p.Where, p.Limit
		},
		func(keys []int, where *model.CalendarDateFilter, limit *int) (ents []*model.CalendarDate, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("gtfs_calendar_dates", limit, nil, nil, "date").Where(In("service_id", keys)),
					"gtfs_calendars",
					"id",
					"gtfs_calendar_dates",
					"service_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.CalendarDate) int {
			return ent.ServiceID.Int()
		},
	)
}

func (f *Finder) FeedVersionGeometryByID(ctx context.Context, ids []int) ([]*tt.Polygon, []error) {
	if len(ids) == 0 {
		return nil, nil
	}
	qents := []*FeedVersionGeometry{}
	if err := dbutil.Select(ctx, f.db, FeedVersionGeometrySelect(ids), &qents); err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	group := map[int]*tt.Polygon{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = ent.Geometry
	}
	ents := make([]*tt.Polygon, len(ids))
	for i, id := range ids {
		ents[i] = group[id]
	}
	return ents, nil
}

func (f *Finder) StopPlacesByStopID(ctx context.Context, params []model.StopPlaceParam) ([]*model.StopPlace, []error) {
	// TODO: Move to paramGroupQuery
	if f.adminCache == nil {
		return f.stopPlacesByStopIdFallback(ctx, params)
	}

	// Lookup any geometries that were not passed in
	var geomLookup []int
	for _, param := range params {
		if param.Point.Lon == 0 && param.Point.Lat == 0 {
			geomLookup = append(geomLookup, param.ID)
		}
	}
	if len(geomLookup) > 0 {
		var ents []struct {
			ID       int
			Geometry tt.Point
		}
		q := sq.Select("gtfs_stops.id", "gtfs_stops.geometry").From("gtfs_stops").Where(sq.Eq{"id": geomLookup})
		if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
			return nil, logExtendErr(ctx, len(params), err)
		}
		lk := map[int]tlxy.Point{}
		for _, ent := range ents {
			lk[ent.ID] = tlxy.Point{Lon: ent.Geometry.X(), Lat: ent.Geometry.Y()}
		}
		for i := 0; i < len(params); i++ {
			if pt, ok := lk[params[i].ID]; ok {
				params[i].Point = pt
			}
		}
	}

	// Lookup stop places using adminCache
	a := map[int]*model.StopPlace{}
	for _, param := range params {
		if admin, ok := f.adminCache.Check(param.Point); ok {
			a[param.ID] = &model.StopPlace{
				Adm0Name: &admin.Adm0Name,
				Adm1Name: &admin.Adm1Name,
				Adm0Iso:  &admin.Adm0Iso,
				Adm1Iso:  &admin.Adm1Iso,
			}
		}
	}
	ret := make([]*model.StopPlace, len(params))
	for idx, param := range params {
		ret[idx] = a[param.ID]
	}
	return ret, nil
}

func (f *Finder) stopPlacesByStopIdFallback(ctx context.Context, params []model.StopPlaceParam) ([]*model.StopPlace, []error) {
	// Fallback without adminCache
	var ids []int
	for _, param := range params {
		ids = append(ids, param.ID)
	}
	type result struct {
		StopID int
		model.StopPlace
	}
	var ents []result
	detailedQuery := sq.Select("gtfs_stops.id as stop_id", "ne.name as adm1_name", "ne.admin as adm0_name").
		From("ne_10m_admin_1_states_provinces ne").
		Join("gtfs_stops on ST_Intersects(gtfs_stops.geometry, ne.geometry)").
		Where(In("gtfs_stops.id", ids))
	if err := dbutil.Select(ctx, f.db, detailedQuery, &ents); err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeMap(ids, ents, func(ent result) (int, *model.StopPlace) { return ent.StopID, &ent.StopPlace }), nil
}

func (f *Finder) CensusGeographiesByEntityID(ctx context.Context, params []model.CensusGeographyParam) ([][]*model.CensusGeography, []error) {
	return paramGroupQuery(
		params,
		func(p model.CensusGeographyParam) (int, *model.CensusGeographyParam, *int) {
			rp := model.CensusGeographyParam{
				EntityType: p.EntityType,
				Limit:      p.Limit,
				Where:      p.Where,
			}
			return p.EntityID, &rp, p.Limit
		},
		func(keys []int, where *model.CensusGeographyParam, limit *int) (ents []*model.CensusGeography, err error) {
			err = dbutil.Select(ctx, f.db, CensusGeographySelect(where, keys), &ents)
			return ents, err
		},
		func(ent *model.CensusGeography) int {
			return ent.MatchEntityID
		},
	)
}

func (f *Finder) CensusValuesByGeographyID(ctx context.Context, params []model.CensusValueParam) ([][]*model.CensusValue, []error) {
	return paramGroupQuery(
		params,
		func(p model.CensusValueParam) (string, *model.CensusValueParam, *int) {
			rp := model.CensusValueParam{
				TableNames: p.TableNames,
				Dataset:    p.Dataset,
			}
			return p.Geoid, &rp, p.Limit
		},
		func(keys []string, where *model.CensusValueParam, limit *int) (ents []*model.CensusValue, err error) {
			err = dbutil.Select(
				ctx,
				f.db,
				CensusValueSelect(where, keys),
				&ents,
			)
			return ents, err
		},
		func(ent *model.CensusValue) string {
			return ent.Geoid
		},
	)
}

func (f *Finder) CensusFieldsByTableID(ctx context.Context, params []model.CensusFieldParam) ([][]*model.CensusField, []error) {
	return paramGroupQuery(
		params,
		func(p model.CensusFieldParam) (int, *model.CensusFieldParam, *int) {
			return p.TableID, nil, p.Limit
		},
		func(keys []int, where *model.CensusFieldParam, limit *int) (ents []*model.CensusField, err error) {
			err = dbutil.Select(ctx,
				f.db,
				lateralWrap(
					quickSelectOrder("tl_census_fields", limit, nil, nil, "id"),
					"tl_census_tables",
					"id",
					"tl_census_fields",
					"table_id",
					keys,
				),
				&ents,
			)
			return ents, err
		},
		func(ent *model.CensusField) int {
			return ent.TableID
		},
	)
}

// MUTATIONS

// Helpers

func logErr(ctx context.Context, err error) error {
	if ctx.Err() == context.Canceled {
		return nil
	}
	log.For(ctx).Error().Err(err).Msg("query failed")
	return errors.New("database error")
}

func logExtendErr(ctx context.Context, size int, err error) []error {
	errs := make([]error, size)
	if ctx.Err() == context.Canceled {
		return errs
	}
	log.For(ctx).Error().Err(err).Msg("query failed")
	for i := 0; i < len(errs); i++ {
		errs[i] = errors.New("database error")
	}
	return errs
}

func arrangeBy[K comparable, T any](keys []K, ents []T, cb func(T) K) []T {
	bykey := map[K]T{}
	for _, ent := range ents {
		bykey[cb(ent)] = ent
	}
	ret := make([]T, len(keys))
	for idx, key := range keys {
		ret[idx] = bykey[key]
	}
	return ret
}

// func arrangeByDebug[K comparable, T any](keys []K, ents []T, cb func(T) K) []T {
// 	bykey := map[K]T{}
// 	for _, ent := range ents {
// 		bykey[cb(ent)] = ent
// 	}
// 	ret := make([]T, len(keys))
// 	for idx, key := range keys {
// 		a, ok := bykey[key]
// 		if ok {
// 			ret[idx] = a
// 		} else {
// 			log.For(ctx).Error().Any("keys", keys).Any("key", key).Int("idx", idx).Any("ents", ents).Msg("no value for key")
// 			debug.PrintStack()
// 		}
// 	}
// 	return ret
// }

func arrangeMap[K comparable, T any, O any](keys []K, ents []T, cb func(T) (K, O)) []O {
	bykey := map[K]O{}
	for _, ent := range ents {
		k, o := cb(ent)
		bykey[k] = o
	}
	ret := make([]O, len(keys))
	for idx, key := range keys {
		ret[idx] = bykey[key]
	}
	return ret
}

// Multiple param sets

func paramGroupQuery[
	K comparable,
	P any,
	W any,
	R any,
](
	params []P,
	paramFunc func(P) (K, W, *int),
	queryFunc func([]K, W, *int) ([]*R, error),
	keyFunc func(*R) K,
) ([][]*R, []error) {
	// Create return value
	ret := make([][]*R, len(params))
	errs := make([]error, len(params))

	// Group params by JSON representation
	type paramGroupItem[K comparable, M any] struct {
		Limit *int
		Where M
	}
	type paramGroup[K comparable, M any] struct {
		Index []int
		Keys  []K
		Limit *int
		Where M
	}
	paramGroups := map[string]paramGroup[K, W]{}
	for i, param := range params {
		// Get values from supplied func
		key, where, limit := paramFunc(param)

		// Convert to paramGroupItem
		item := paramGroupItem[K, W]{
			Limit: limit,
			Where: where,
		}

		// Use the JSON representation of Where and Limit as the key
		jj, err := json.Marshal(paramGroupItem[K, W]{Where: item.Where, Limit: item.Limit})
		if err != nil {
			// TODO: log and expand error
			errs[i] = err
			continue
		}
		paramGroupKey := string(jj)

		// Add index and key
		a, ok := paramGroups[paramGroupKey]
		if !ok {
			a = paramGroup[K, W]{Where: item.Where, Limit: item.Limit}
		}
		a.Index = append(a.Index, i)
		a.Keys = append(a.Keys, key)
		paramGroups[paramGroupKey] = a
	}

	// Process each param group
	for _, pgroup := range paramGroups {
		// Run query function
		ents, err := queryFunc(pgroup.Keys, pgroup.Where, pgroup.Limit)

		// Group using keyFunc and merge into output
		limit := checkLimit(pgroup.Limit)
		bykey := map[K][]*R{}
		for _, ent := range ents {
			key := keyFunc(ent)
			bykey[key] = append(bykey[key], ent)
		}
		for keyidx, key := range pgroup.Keys {
			idx := pgroup.Index[keyidx]
			gi := bykey[key]
			if err != nil {
				errs[idx] = err
			}
			if uint64(len(gi)) <= limit {
				ret[idx] = gi
			} else {
				ret[idx] = gi[0:limit]
			}
		}
	}
	return ret, errs
}

func convertEnts[A any, B any](vals [][]A, fn func(a A) B) [][]B {
	ret := make([][]B, len(vals))
	for i, x := range vals {
		ret[i] = make([]B, len(x))
		for j, y := range x {
			ret[i][j] = fn(y)
		}
	}
	return ret
}
