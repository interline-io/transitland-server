package dbfinder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/xy"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

////////

type Finder struct {
	Clock      clock.Clock
	db         sqlx.Ext
	adminCache *adminCache
}

func NewFinder(db sqlx.Ext) *Finder {
	return &Finder{db: db}
}

func (f *Finder) DBX() sqlx.Ext {
	return f.db
}

func (f *Finder) LoadAdmins() error {
	log.Trace().Msg("loading admins")
	adminCache := newAdminCache()
	if err := adminCache.LoadAdmins(context.Background(), f.db); err != nil {
		return err
	}
	f.adminCache = adminCache
	return nil
}

func (f *Finder) PermFilter(ctx context.Context) *model.PermFilter {
	permFilter, _ := checkActive(ctx)
	return permFilter
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
	fmt.Println("FIND STOPS")
	log.For(ctx).Info().Msg("find stops")
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
	q := TripSelect(limit, after, ids, active, f.PermFilter(ctx), where)
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
	if err := dbutil.Select(ctx, f.db, OperatorSelect(limit, after, ids, nil, f.PermFilter(ctx), where), &ents); err != nil {
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

func (f *Finder) FindFeedVersionServiceWindow(ctx context.Context, fvid int) (time.Time, time.Time, time.Time, error) {
	type fvslQuery struct {
		FetchedAt    tl.Time
		StartDate    tl.Time
		EndDate      tl.Time
		TotalService tl.Int
	}
	minServiceRatio := 0.75
	startDate := time.Time{}
	endDate := time.Time{}
	bestWeek := time.Time{}

	// Get FVSLs
	q := sq.StatementBuilder.
		Select("fv.fetched_at", "fvsl.start_date", "fvsl.end_date", "monday + tuesday + wednesday + thursday + friday + saturday + sunday as total_service").
		From("feed_version_service_levels fvsl").
		Join("feed_versions fv on fv.id = fvsl.feed_version_id").
		Where(sq.Eq{"route_id": nil}).
		Where(sq.Eq{"fvsl.feed_version_id": fvid}).
		OrderBy("fvsl.start_date").
		Limit(1000)
	var ents []fvslQuery
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return startDate, endDate, bestWeek, logErr(ctx, err)
	}
	if len(ents) == 0 {
		return startDate, endDate, bestWeek, errors.New("no fvsl results")
	}

	var fis []tl.FeedInfo
	fiq := sq.StatementBuilder.Select("*").From("gtfs_feed_infos").Where(sq.Eq{"feed_version_id": fvid}).OrderBy("feed_start_date").Limit(1)
	if err := dbutil.Select(ctx, f.db, fiq, &fis); err != nil {
		return startDate, endDate, bestWeek, logErr(ctx, err)
	}

	// Check if we have feed infos, otherwise calculate based on fetched week or highest service week
	fetched := ents[0].FetchedAt.Val
	if len(fis) > 0 && fis[0].FeedStartDate.Valid && fis[0].FeedEndDate.Valid {
		// fmt.Println("using feed infos")
		startDate = fis[0].FeedStartDate.Val
		endDate = fis[0].FeedEndDate.Val
	} else {
		// Get the week which includes fetched_at date, and the highest service week
		highestIdx := 0
		highestService := -1
		fetchedWeek := -1
		for i, ent := range ents {
			sd := ent.StartDate.Val
			ed := ent.EndDate.Val
			if (sd.Before(fetched) || sd.Equal(fetched)) && (ed.After(fetched) || ed.Equal(fetched)) {
				fetchedWeek = i
			}
			if ent.TotalService.Int() > highestService {
				highestIdx = i
				highestService = ent.TotalService.Int()
			}
		}
		if fetchedWeek < 0 {
			// fmt.Println("fetched week not in fvsls, using highest week:", highestIdx, highestService)
			fetchedWeek = highestIdx
		} else {
			// fmt.Println("using fetched week:", fetchedWeek)
		}
		// If the fetched week has bad service, use highest week
		if float64(ents[fetchedWeek].TotalService.Val)/float64(highestService) < minServiceRatio {
			// fmt.Println("fetched week has poor service ratio, falling back to highest week:", fetchedWeek)
			fetchedWeek = highestIdx
		}

		// Expand window in both directions from chosen week
		startDate = ents[fetchedWeek].StartDate.Val
		endDate = ents[fetchedWeek].EndDate.Val
		for i := fetchedWeek; i < len(ents); i++ {
			ent := ents[i]
			if float64(ent.TotalService.Val)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.StartDate.Val.Before(startDate) {
				startDate = ent.StartDate.Val
			}
			endDate = ent.EndDate.Val
		}
		for i := fetchedWeek - 1; i > 0; i-- {
			ent := ents[i]
			if float64(ent.TotalService.Val)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.EndDate.Val.After(endDate) {
				endDate = ent.EndDate.Val
			}
			startDate = ent.StartDate.Val
		}
	}

	// Check again to find the highest service week in the window
	// This will be used as the typical week for dates outside the window
	// bestWeek must start with a Monday
	bestWeek = ents[0].StartDate.Val
	bestService := ents[0].TotalService.Val
	for _, ent := range ents {
		sd := ent.StartDate.Val
		ed := ent.EndDate.Val
		if (sd.Before(endDate) || sd.Equal(endDate)) && (ed.After(startDate) || ed.Equal(startDate)) {
			if ent.TotalService.Val > bestService {
				bestService = ent.TotalService.Val
				bestWeek = ent.StartDate.Val
			}
		}
	}
	return startDate, endDate, bestWeek, nil
}

// Loaders

func (f *Finder) TripsByID(ctx context.Context, ids []int) (ents []*model.Trip, errs []error) {
	ents, err := f.FindTrips(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Trip) int { return ent.ID }), nil
}

// Simple ID loaders
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
	q := sq.StatementBuilder.Select("*").From("tl_stop_external_references").Where(sq.Eq{"id": ids})
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.StopExternalReference{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.StopExternalReference, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *Finder) StopObservationsByStopID(ctx context.Context, params []model.StopObservationParam) ([][]*model.StopObservation, []error) {
	type wrappedStopObservation struct {
		StopID int
		model.StopObservation
	}
	var ents []*wrappedStopObservation
	var ids []int
	var where *model.StopObservationFilter
	for _, p := range params {
		if where == nil && p.Where != nil {
			where = p.Where
		}
		ids = append(ids, p.StopID)
	}
	// Prepare output
	// Currently Where must not be nil for this query
	// This may not be required in the future.
	if where == nil {
		return retEmpty[[]*model.StopObservation](len(params))
	}
	q := sq.StatementBuilder.Select("gtfs_stops.id as stop_id", "obs.*").
		From("ext_performance_stop_observations obs").
		Join("gtfs_stops on gtfs_stops.stop_id = obs.to_stop_id").
		Where(sq.Eq{"gtfs_stops.id": ids}).
		Limit(100000)
	q = q.Where("obs.feed_version_id = ?", where.FeedVersionID)
	q = q.Where("obs.trip_start_date = ?", where.TripStartDate)
	q = q.Where("obs.source = ?", where.Source)
	// q = q.Where("start_time >= ?", where.StartTime)
	// q = q.Where("end_time <= ?", where.EndTime)
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		// return retError[[]*model.StopObservation](len(params))
		return nil, logExtendErr(ctx, len(ids), err)
	}
	byid := map[int][]*model.StopObservation{}
	for _, ent := range ents {
		ent := ent
		byid[ent.StopID] = append(byid[ent.StopID], &ent.StopObservation)
	}
	ret := make([][]*model.StopObservation, len(ids))
	for i, id := range ids {
		ret[i] = byid[id]
	}
	return ret, nil
}

func (f *Finder) RouteAttributesByRouteID(ctx context.Context, ids []int) ([]*model.RouteAttribute, []error) {
	var ents []*model.RouteAttribute
	q := sq.StatementBuilder.Select("*").From("ext_plus_route_attributes").Where(sq.Eq{"route_id": ids})
	if err := dbutil.Select(ctx, f.db, q, &ents); err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.RouteAttribute{}
	for _, ent := range ents {
		byid[ent.RouteID] = ent
	}
	ents2 := make([]*model.RouteAttribute, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
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

func (f *Finder) FeedVersionGtfsImportsByFeedVersionID(ctx context.Context, ids []int) ([]*model.FeedVersionGtfsImport, []error) {
	var ents []*model.FeedVersionGtfsImport
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("feed_version_gtfs_imports", nil, nil, nil).Where(sq.Eq{"feed_version_id": ids}),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.FeedVersionGtfsImport) int { return ent.FeedVersionID }), nil
}

func (f *Finder) FeedStatesByFeedID(ctx context.Context, ids []int) ([]*model.FeedState, []error) {
	var ents []*model.FeedState
	err := dbutil.Select(ctx,
		f.db,
		quickSelect("feed_states", nil, nil, nil).Where(sq.Eq{"feed_id": ids}),
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
		OperatorSelect(nil, nil, ids, nil, f.PermFilter(ctx), nil),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Operator) int { return ent.ID }), nil
}

func (f *Finder) OperatorsByOnestopID(ctx context.Context, ids []string) ([]*model.Operator, []error) {
	var ents []*model.Operator
	err := dbutil.Select(ctx,
		f.db,
		OperatorsByAgencyID(nil, nil, nil, ids),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Operator) string { return ent.OnestopID.Val }), nil
}

func (f *Finder) OperatorsByAgencyID(ctx context.Context, ids []int) ([]*model.Operator, []error) {
	var ents []*model.Operator
	err := dbutil.Select(ctx,
		f.db,
		OperatorsByAgencyID(nil, nil, ids, nil),
		&ents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeBy(ids, ents, func(ent *model.Operator) int { return ent.AgencyID }), nil
}

// Param loaders

func (f *Finder) OperatorsByFeedID(ctx context.Context, params []model.OperatorParam) ([][]*model.Operator, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedID)
	}
	qents := []*model.Operator{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			OperatorSelect(params[0].Limit, nil, nil, ids, f.PermFilter(ctx), params[0].Where),
			"current_feeds",
			"id",
			"t",
			"feed_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Operator{}
	for _, ent := range qents {
		if v := ent.FeedID; v > 0 {
			group[v] = append(group[v], ent)
		}
	}
	var ents [][]*model.Operator
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) FeedFetchesByFeedID(ctx context.Context, params []model.FeedFetchParam) ([][]*model.FeedFetch, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	// This is horrendous :laughing:
	ents := make([][]*model.FeedFetch, len(params))
	pgroups := map[string][]int{}
	// All in group share same params except FeedID
	for pidx, param := range params {
		param.FeedID = 0 // unset feedid on copy
		k := marshalParam(param)
		pgroups[k] = append(pgroups[k], pidx)
	}
	for _, pidxs := range pgroups {
		var ids []int
		for _, pidx := range pidxs {
			ids = append(ids, params[pidx].FeedID)
		}
		var qents []*model.FeedFetch
		q := sq.StatementBuilder.
			Select("*").
			From("feed_fetches").
			Limit(checkLimit(params[pidxs[0]].Limit)).
			OrderBy("feed_fetches.fetched_at desc")
		if p := params[pidxs[0]].Where; p != nil {
			if p.Success != nil {
				q = q.Where(sq.Eq{"success": *p.Success})
			}
		}
		err := dbutil.Select(ctx,
			f.db,
			lateralWrap(q, "current_feeds", "id", "feed_fetches", "feed_id", ids),
			&qents,
		)
		if err != nil {
			return nil, logExtendErr(ctx, len(params), err)
		}
		group := map[int][]*model.FeedFetch{}
		for _, ent := range qents {
			if v := ent.FeedID; v > 0 {
				group[v] = append(group[v], ent)
			}
		}
		for _, pidx := range pidxs {
			ents[pidx] = group[params[pidx].FeedID]
		}
	}
	return ents, nil
}

func (f *Finder) FeedsByOperatorOnestopID(ctx context.Context, params []model.FeedParam) ([][]*model.Feed, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	osids := []string{}
	for _, p := range params {
		osids = append(osids, p.OperatorOnestopID)
	}
	type ffeed struct {
		OperatorOnestopID string
		model.Feed
	}
	var qents []*ffeed
	q := FeedSelect(nil, nil, nil, f.PermFilter(ctx), params[0].Where).
		Distinct().Options("on (coif.resolved_onestop_id, current_feeds.id)").
		Column("coif.resolved_onestop_id as operator_onestop_id").
		Join("current_operators_in_feed coif on coif.feed_id = current_feeds.id").
		Where(sq.Eq{"coif.resolved_onestop_id": osids})
	err := dbutil.Select(ctx,
		f.db,
		q,
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[string][]*model.Feed{}
	for i := 0; i < len(qents); i++ {
		ent := qents[i]
		group[ent.OperatorOnestopID] = append(group[ent.OperatorOnestopID], &ent.Feed)
	}
	limit := checkLimit(params[0].Limit)
	for k, ents := range group {
		if uint64(len(ents)) > limit {
			group[k] = ents[0:limit]
		}
	}
	var ents [][]*model.Feed
	for _, osid := range osids {
		ents = append(ents, group[osid])
	}
	return ents, nil
}

func (f *Finder) FrequenciesByTripID(ctx context.Context, params []model.FrequencyParam) ([][]*model.Frequency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.TripID)
	}
	qents := []*model.Frequency{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			quickSelect("gtfs_frequencies", params[0].Limit, nil, nil),
			"gtfs_trips",
			"id",
			"gtfs_frequencies",
			"trip_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Frequency{}
	for _, ent := range qents {
		group[atoi(ent.TripID)] = append(group[atoi(ent.TripID)], ent)
	}
	var ents [][]*model.Frequency
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) StopTimesByTripID(ctx context.Context, params []model.TripStopTimeParam) ([][]*model.StopTime, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	var pitems []paramItem[FVPair, *model.TripStopTimeFilter]
	for _, p := range params {
		pitem := paramItem[FVPair, *model.TripStopTimeFilter]{
			Key:   FVPair{EntityID: p.TripID, FeedVersionID: p.FeedVersionID},
			Where: p.Where,
			Limit: p.Limit,
		}
		pitems = append(pitems, pitem)
	}
	pitemGroups, err := paramsByGroup(pitems)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	ret := make([][]*model.StopTime, len(params))
	for _, group := range pitemGroups {
		qents := []*model.StopTime{}
		if err := dbutil.Select(ctx,
			f.db,
			StopTimeSelect(group.Keys, nil, f.PermFilter(ctx), group.Where),
			&qents,
		); err != nil {
			return nil, logExtendErr(ctx, len(params), err)
		}
		grouped := groupBy(group.Keys, qents, checkLimit(group.Limit), func(ent *model.StopTime) FVPair {
			return FVPair{EntityID: atoi(ent.TripID), FeedVersionID: ent.FeedVersionID}
		})
		for i := 0; i < len(group.Keys); i++ {
			ret[group.Index[i]] = grouped[i]
		}
	}
	return ret, nil
}

func (f *Finder) StopTimesByStopID(ctx context.Context, params []model.StopTimeParam) ([][]*model.StopTime, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	// Group each param into a query group
	// only exported fields are included in key
	type dGroup struct {
		Where *model.StopTimeFilter
		Limit *int
		pairs []FVPair
		idx   []int
	}
	dGroups := map[string]*dGroup{}
	for i, p := range params {
		// somewhat ugly, use json representation for grouping
		dg := &dGroup{Where: p.Where, Limit: p.Limit}
		key, err := json.Marshal(dg)
		if err != nil {
			return nil, logExtendErr(ctx, len(params), err)
		}
		if a, ok := dGroups[string(key)]; ok {
			dg = a
		} else {
			dGroups[string(key)] = dg
		}
		dg.pairs = append(dg.pairs, FVPair{EntityID: p.StopID, FeedVersionID: p.FeedVersionID})
		dg.idx = append(dg.idx, i) // original input position
	}
	ents := make([][]*model.StopTime, len(params))
	for _, dg := range dGroups {
		// group results by stop
		group := map[int][]*model.StopTime{}
		limit := checkLimit(dg.Limit)
		qents := []*model.StopTime{}
		p := dg.Where
		if p != nil && p.ServiceDate != nil {
			// Get stops on a specified day
			err := dbutil.Select(ctx,
				f.db,
				StopDeparturesSelect(dg.pairs, f.PermFilter(ctx), p),
				&qents,
			)
			if err != nil {
				return nil, logExtendErr(ctx, len(params), err)
			}
		} else {
			// Otherwise get all stop_times for stop
			err := dbutil.Select(ctx,
				f.db,
				StopTimeSelect(nil, dg.pairs, f.PermFilter(ctx), nil),
				&qents,
			)
			if err != nil {
				return nil, logExtendErr(ctx, len(params), err)
			}
		}
		for _, ent := range qents {
			group[atoi(ent.StopID)] = append(group[atoi(ent.StopID)], ent)
		}
		for i := 0; i < len(dg.pairs); i++ {
			pair := dg.pairs[i]
			idx := dg.idx[i]
			g := group[pair.EntityID]
			if uint64(len(g)) > limit {
				g = g[0:limit]
			}
			ents[idx] = g
		}
	}
	return ents, nil
}

func (f *Finder) RouteStopsByStopID(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.StopID)
	}
	qents := []*model.RouteStop{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			quickSelectOrder("tl_route_stops", params[0].Limit, nil, nil, "stop_id"),
			"gtfs_stops",
			"id",
			"tl_route_stops",
			"stop_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.RouteStop{}
	for _, ent := range qents {
		group[ent.StopID] = append(group[ent.StopID], ent)
	}
	var ents [][]*model.RouteStop
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) StopsByRouteID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	type qEnt struct {
		RouteID int
		model.Stop
	}
	if len(params) == 0 {
		return nil, nil
	}
	routeIds := []int{}
	for _, p := range params {
		routeIds = append(routeIds, p.RouteID)
	}
	qents := []*qEnt{}
	qso := StopSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where)
	qso = qso.Join("tl_route_stops on tl_route_stops.stop_id = t.id").Where(sq.Eq{"route_id": routeIds}).Column("route_id")
	err := dbutil.Select(ctx,
		f.db,
		qso,
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Stop{}
	for _, ent := range qents {
		group[ent.RouteID] = append(group[ent.RouteID], &ent.Stop)
	}
	var ents [][]*model.Stop
	for _, id := range routeIds {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) RouteStopsByRouteID(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteStop{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			quickSelectOrder("tl_route_stops", params[0].Limit, nil, nil, "stop_id"),
			"gtfs_routes",
			"id",
			"tl_route_stops",
			"route_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.RouteStop{}
	for _, ent := range qents {
		group[ent.RouteID] = append(group[ent.RouteID], ent)
	}
	var ents [][]*model.RouteStop
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) RouteHeadwaysByRouteID(ctx context.Context, params []model.RouteHeadwayParam) ([][]*model.RouteHeadway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteHeadway{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			quickSelectOrder("tl_route_headways", params[0].Limit, nil, nil, "route_id"),
			"gtfs_routes",
			"id",
			"tl_route_headways",
			"route_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.RouteHeadway{}
	for _, ent := range qents {
		group[ent.RouteID] = append(group[ent.RouteID], ent)
	}
	var ents [][]*model.RouteHeadway
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) RouteStopPatternsByRouteID(ctx context.Context, params []model.RouteStopPatternParam) ([][]*model.RouteStopPattern, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	var qents []*model.RouteStopPattern
	q := sq.StatementBuilder.
		Select("route_id", "direction_id", "stop_pattern_id", "count(*) as count").
		From("gtfs_trips").
		Where(sq.Eq{"route_id": ids}).
		GroupBy("route_id,direction_id,stop_pattern_id").
		OrderBy("route_id,count desc").
		Limit(1000)
	err := dbutil.Select(ctx,
		f.db,
		q,
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.RouteStopPattern{}
	for _, ent := range qents {
		group[ent.RouteID] = append(group[ent.RouteID], ent)
	}
	var ents [][]*model.RouteStopPattern
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) FeedVersionFileInfosByFeedVersionID(ctx context.Context, params []model.FeedVersionFileInfoParam) ([][]*model.FeedVersionFileInfo, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedVersionFileInfo{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			quickSelectOrder("feed_version_file_infos", params[0].Limit, nil, nil, "feed_version_id"),
			"feed_versions",
			"id",
			"feed_version_file_infos",
			"feed_version_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.FeedVersionFileInfo{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.FeedVersionFileInfo
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) StopsByParentStopID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ParentStopID)
	}
	qents := []*model.Stop{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			StopSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where),
			"gtfs_stops",
			"id",
			"gtfs_stops",
			"parent_station",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Stop{}
	for _, ent := range qents {
		group[ent.ParentStation.Int()] = append(group[ent.ParentStation.Int()], ent)
	}
	var ents [][]*model.Stop
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
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
		Where(sq.Eq{"tlse.id": ids})
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
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedID)
	}
	qents := []*model.FeedVersion{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			FeedVersionSelect(params[0].Limit, nil, nil, f.PermFilter(ctx), params[0].Where),
			"current_feeds",
			"id",
			"feed_versions",
			"feed_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	return groupBy(ids, qents, checkLimit(params[0].Limit), func(ent *model.FeedVersion) int { return ent.FeedID }), nil
}

func (f *Finder) AgencyPlacesByAgencyID(ctx context.Context, params []model.AgencyPlaceParam) ([][]*model.AgencyPlace, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	minRank := 0.0
	for _, p := range params {
		ids = append(ids, p.AgencyID)
		if p.Where != nil && p.Where.MinRank != nil {
			minRank = *p.Where.MinRank
		}
	}
	qents := []*model.AgencyPlace{}

	q := sq.StatementBuilder.Select(
		"tl_agency_places.agency_id",
		"tl_agency_places.rank",
		"tl_agency_places.name",
		"tl_agency_places.adm0name",
		"tl_agency_places.adm1name",
		"ne_admin.iso_a2 as adm0iso",
		"ne_admin.iso_3166_2 as adm1iso",
	).
		From("tl_agency_places").
		Join("ne_10m_admin_1_states_provinces ne_admin on ne_admin.name = tl_agency_places.adm1name and ne_admin.admin = tl_agency_places.adm0name").
		Where(sq.GtOrEq{"rank": minRank})

	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			q,
			"gtfs_agencies",
			"id",
			"tl_agency_places",
			"agency_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.AgencyPlace{}
	for _, ent := range qents {
		group[ent.AgencyID] = append(group[ent.AgencyID], ent)
	}
	var ents [][]*model.AgencyPlace
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) RouteGeometriesByRouteID(ctx context.Context, params []model.RouteGeometryParam) ([][]*model.RouteGeometry, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteGeometry{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			quickSelectOrder("tl_route_geometries", params[0].Limit, nil, nil, "route_id"),
			"gtfs_routes",
			"id",
			"tl_route_geometries",
			"route_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.RouteGeometry{}
	for _, ent := range qents {
		group[ent.RouteID] = append(group[ent.RouteID], ent)
	}
	var ents [][]*model.RouteGeometry
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) TripsByRouteID(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.Trip{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			TripSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where),
			"gtfs_routes",
			"id",
			"gtfs_trips",
			"route_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Trip{}
	for _, ent := range qents {
		group[atoi(ent.RouteID)] = append(group[atoi(ent.RouteID)], ent)
	}
	var ents [][]*model.Trip
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) RoutesByAgencyID(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.AgencyID)
	}
	qents := []*model.Route{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			RouteSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where),
			"gtfs_agencies",
			"id",
			"gtfs_routes",
			"agency_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Route{}
	for _, ent := range qents {
		group[atoi(ent.AgencyID)] = append(group[atoi(ent.AgencyID)], ent)
	}
	var ents [][]*model.Route
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) AgenciesByFeedVersionID(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Agency{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			AgencySelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where),
			"feed_versions",
			"id",
			"gtfs_agencies",
			"feed_version_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Agency{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.Agency
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) AgenciesByOnestopID(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []string{}
	for _, p := range params {
		if p.OnestopID != nil {
			ids = append(ids, *p.OnestopID)
		}
	}
	qents := []*model.Agency{}
	err := dbutil.Select(ctx,
		f.db,
		AgencySelect(params[0].Limit, nil, nil, true, f.PermFilter(ctx), nil).Where(sq.Eq{"coif.resolved_onestop_id": ids}), // active=true
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[string][]*model.Agency{}
	for _, ent := range qents {
		group[ent.OnestopID] = append(group[ent.OnestopID], ent)
	}
	var ents [][]*model.Agency
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) StopsByFeedVersionID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Stop{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(StopSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where), "feed_versions", "id", "gtfs_stops", "feed_version_id", ids),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Stop{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.Stop
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) StopsByLevelID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Stop{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			StopSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where),
			"gtfs_levels",
			"id",
			"gtfs_stops",
			"level_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Stop{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.Stop
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) TripsByFeedVersionID(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Trip{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			TripSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where),
			"feed_versions",
			"id",
			"gtfs_trips",
			"feed_version_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Trip{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.Trip
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) FeedInfosByFeedVersionID(ctx context.Context, params []model.FeedInfoParam) ([][]*model.FeedInfo, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedInfo{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			quickSelectOrder("gtfs_feed_infos", params[0].Limit, nil, nil, "feed_version_id"),
			"feed_versions",
			"id",
			"gtfs_feed_infos",
			"feed_version_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.FeedInfo{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.FeedInfo
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) RoutesByFeedVersionID(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Route{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			RouteSelect(params[0].Limit, nil, nil, false, f.PermFilter(ctx), params[0].Where),
			"feed_versions",
			"id",
			"gtfs_routes",
			"feed_version_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Route{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.Route
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) FeedVersionServiceLevelsByFeedVersionID(ctx context.Context, params []model.FeedVersionServiceLevelParam) ([][]*model.FeedVersionServiceLevel, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedVersionServiceLevel{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			FeedVersionServiceLevelSelect(params[0].Limit, nil, nil, f.PermFilter(ctx), params[0].Where),
			"feed_versions",
			"id",
			"feed_version_service_levels",
			"feed_version_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.FeedVersionServiceLevel{}
	for _, ent := range qents {
		group[ent.FeedVersionID] = append(group[ent.FeedVersionID], ent)
	}
	var ents [][]*model.FeedVersionServiceLevel
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) PathwaysByFromStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FromStopID)
	}
	qents := []*model.Pathway{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			PathwaySelect(params[0].Limit, nil, nil, f.PermFilter(ctx), params[0].Where),
			"gtfs_stops",
			"id",
			"gtfs_pathways",
			"from_stop_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Pathway{}
	for _, ent := range qents {
		group[atoi(ent.FromStopID)] = append(group[atoi(ent.FromStopID)], ent)
	}
	var ents [][]*model.Pathway
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) PathwaysByToStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ToStopID)
	}
	qents := []*model.Pathway{}
	err := dbutil.Select(ctx,
		f.db,
		lateralWrap(
			PathwaySelect(params[0].Limit, nil, nil, f.PermFilter(ctx), params[0].Where),
			"gtfs_stops",
			"id",
			"gtfs_pathways",
			"to_stop_id",
			ids,
		),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.Pathway{}
	for _, ent := range qents {
		group[atoi(ent.ToStopID)] = append(group[atoi(ent.ToStopID)], ent)
	}
	var ents [][]*model.Pathway
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) CalendarDatesByServiceID(ctx context.Context, params []model.CalendarDateParam) ([][]*model.CalendarDate, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ServiceID)
	}
	qents := []*model.CalendarDate{}
	err := dbutil.Select(ctx,
		f.db,
		quickSelectOrder("gtfs_calendar_dates", nil, nil, nil, "date").Where(sq.Eq{"service_id": ids}),
		&qents,
	)
	if err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.CalendarDate{}
	for _, ent := range qents {
		group[atoi(ent.ServiceID)] = append(group[atoi(ent.ServiceID)], ent)
	}
	var ents [][]*model.CalendarDate
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
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
		lk := map[int]xy.Point{}
		for _, ent := range ents {
			lk[ent.ID] = xy.Point{Lon: ent.Geometry.X(), Lat: ent.Geometry.Y()}
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
		Where(sq.Eq{"gtfs_stops.id": ids})
	if err := dbutil.Select(ctx, f.db, detailedQuery, &ents); err != nil {
		return nil, logExtendErr(ctx, len(ids), err)
	}
	return arrangeMap(ids, ents, func(ent result) (int, *model.StopPlace) { return ent.StopID, &ent.StopPlace }), nil
}

func (f *Finder) CensusGeographiesByEntityID(ctx context.Context, params []model.CensusGeographyParam) ([][]*model.CensusGeography, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.EntityID)
	}
	qents := []*model.CensusGeography{}
	if err := dbutil.Select(ctx, f.db, CensusGeographySelect(&params[0], ids), &qents); err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.CensusGeography{}
	for _, ent := range qents {
		group[ent.MatchEntityID] = append(group[ent.MatchEntityID], ent)
	}
	var ents [][]*model.CensusGeography
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *Finder) CensusValuesByGeographyID(ctx context.Context, params []model.CensusValueParam) ([][]*model.CensusValue, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.GeographyID)
	}
	a := 1000
	params[0].Limit = &a // only a single result allowed
	qents := []*model.CensusValue{}
	if err := dbutil.Select(ctx, f.db, CensusValueSelect(&params[0], ids), &qents); err != nil {
		return nil, logExtendErr(ctx, len(params), err)
	}
	group := map[int][]*model.CensusValue{}
	for _, ent := range qents {
		group[ent.GeographyID] = append(group[ent.GeographyID], ent)
	}
	var ents [][]*model.CensusValue
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func logErr(ctx context.Context, err error) error {
	if ctx.Err() == context.Canceled {
		return nil
	}
	log.Error().Err(err).Msg("query failed")
	return errors.New("database error")
}

func logExtendErr(ctx context.Context, size int, err error) []error {
	errs := make([]error, size)
	if ctx.Err() == context.Canceled {
		return errs
	}
	log.Error().Err(err).Msg("query failed")
	for i := 0; i < len(errs); i++ {
		errs[i] = errors.New("database error")
	}
	return errs
}

func marshalParam(param interface{}) string {
	a, _ := json.Marshal(param)
	return string(a)
}

func retEmpty[T any](size int) ([]T, []error) {
	ret := make([]T, size)
	return ret, nil
}

func groupBy[K comparable, T any](keys []K, ents []T, limit uint64, cb func(T) K) [][]T {
	bykey := map[K][]T{}
	for _, ent := range ents {
		key := cb(ent)
		bykey[key] = append(bykey[key], ent)
	}
	ret := make([][]T, len(keys))
	for idx, key := range keys {
		gi := bykey[key]
		if uint64(len(gi)) <= limit {
			ret[idx] = gi
		} else {
			ret[idx] = gi[0:limit]
		}
	}
	return ret
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

type paramItem[K comparable, M any] struct {
	Limit *int
	Key   K
	Where M
}

type paramGroup[K comparable, M any] struct {
	Index []int
	Keys  []K
	Limit *int
	Where M
}

func paramsByGroup[K comparable, M any](items []paramItem[K, M]) ([]paramGroup[K, M], error) {
	// JSON representation of paramItem.Where is used for grouping.
	// This might not be the best way to do it, but it's convenient.
	groups := map[string]paramGroup[K, M]{}
	for i, item := range items {
		// Include the limit in the string key representation
		j, err := json.Marshal(paramItem[K, M]{Where: item.Where, Limit: item.Limit})
		if err != nil {
			return nil, err
		}
		jstr := string(j)
		a, ok := groups[jstr]
		if !ok {
			a = paramGroup[K, M]{Where: item.Where, Limit: item.Limit}
		}
		a.Keys = append(a.Keys, item.Key)
		a.Index = append(a.Index, i)
		groups[jstr] = a
	}
	var ret []paramGroup[K, M]
	for _, v := range groups {
		ret = append(ret, v)
	}
	return ret, nil
}

type canCheckGlobalAdmin interface {
	CheckGlobalAdmin(context.Context) (bool, error)
}

func checkActive(ctx context.Context) (*model.PermFilter, error) {
	checker := model.ForContext(ctx).Checker
	active := &model.PermFilter{}
	if checker == nil {
		return active, nil
	}

	// TODO: Make this part of actual checker interface
	if c, ok := checker.(canCheckGlobalAdmin); ok {
		if a, err := c.CheckGlobalAdmin(ctx); err != nil {
			return nil, err
		} else if a {
			return nil, nil
		}
	}

	okFeeds, err := checker.FeedList(ctx, &authz.FeedListRequest{})
	if err != nil {
		return nil, err
	}
	for _, feed := range okFeeds.Feeds {
		active.AllowedFeeds = append(active.AllowedFeeds, int(feed.Id))
	}
	okFvids, err := checker.FeedVersionList(ctx, &authz.FeedVersionListRequest{})
	if err != nil {
		return nil, err
	}
	for _, fv := range okFvids.FeedVersions {
		active.AllowedFeedVersions = append(active.AllowedFeedVersions, int(fv.Id))
	}
	// fmt.Println("active allowed feeds:", active.AllowedFeeds, "fvs:", active.AllowedFeedVersions)
	return active, nil
}
