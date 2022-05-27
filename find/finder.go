package find

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

////////

type DBFinder struct {
	db sqlx.Ext
}

func NewDBFinder(db sqlx.Ext) *DBFinder {
	return &DBFinder{db: db}
}

func (f *DBFinder) DBX() sqlx.Ext {
	return f.db
}

func (f *DBFinder) FindAgencies(ctx context.Context, limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	var ents []*model.Agency
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) {
		active = false
	}
	q := AgencySelect(limit, after, ids, active, where)
	MustSelect(ctx, f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindRoutes(ctx context.Context, limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	var ents []*model.Route
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) {
		active = false
	}
	q := RouteSelect(limit, after, ids, active, where)
	MustSelect(ctx, f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindStops(ctx context.Context, limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	var ents []*model.Stop
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) {
		active = false
	}
	q := StopSelect(limit, after, ids, active, where)
	MustSelect(ctx, f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindTrips(ctx context.Context, limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	var ents []*model.Trip
	active := true
	if len(ids) > 0 || (where != nil && where.FeedVersionSha1 != nil) {
		active = false
	}
	q := TripSelect(limit, after, ids, active, where)
	MustSelect(ctx, f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindFeedVersions(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	var ents []*model.FeedVersion
	MustSelect(ctx, f.db, FeedVersionSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func (f *DBFinder) FindFeeds(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	var ents []*model.Feed
	MustSelect(ctx, f.db, FeedSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func (f *DBFinder) FindOperators(ctx context.Context, limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	var ents []*model.Operator
	MustSelect(ctx, f.db, OperatorSelect(limit, after, ids, nil, where), &ents)
	return ents, nil
}

func (f *DBFinder) RouteStopBuffer(ctx context.Context, param *model.RouteStopBufferParam) ([]*model.RouteStopBuffer, error) {
	if param == nil {
		return nil, nil
	}
	var ents []*model.RouteStopBuffer
	q := RouteStopBufferSelect(*param)
	MustSelect(ctx, f.db, q, &ents)
	return ents, nil
}

// Custom queries

func (f *DBFinder) FindFeedVersionServiceWindow(ctx context.Context, fvid int) (time.Time, time.Time, time.Time, error) {
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
	MustSelect(ctx, f.db, q, &ents)
	if len(ents) == 0 {
		return startDate, endDate, bestWeek, errors.New("no fvsl results")
	}
	var fis []tl.FeedInfo
	fiq := sq.StatementBuilder.Select("*").From("gtfs_feed_infos").Where(sq.Eq{"feed_version_id": fvid}).OrderBy("feed_start_date").Limit(1)
	MustSelect(ctx, f.db, fiq, &fis)

	// Check if we have feed infos, otherwise calculate based on fetched week or highest service week
	fetched := ents[0].FetchedAt.Time
	if len(fis) > 0 {
		// fmt.Println("using feed infos")
		startDate = fis[0].FeedStartDate.Time
		endDate = fis[0].FeedEndDate.Time
	} else {
		// Get the week which includes fetched_at date, and the highest service week
		highestIdx := 0
		highestService := -1
		fetchedWeek := -1
		for i, ent := range ents {
			sd := ent.StartDate.Time
			ed := ent.EndDate.Time
			if (sd.Before(fetched) || sd.Equal(fetched)) && (ed.After(fetched) || ed.Equal(fetched)) {
				fetchedWeek = i
			}
			if ent.TotalService.Int > highestService {
				highestIdx = i
				highestService = ent.TotalService.Int
			}
		}
		if fetchedWeek < 0 {
			// fmt.Println("fetched week not in fvsls, using highest week:", highestIdx, highestService)
			fetchedWeek = highestIdx
		} else {
			// fmt.Println("using fetched week:", fetchedWeek)
		}
		// If the fetched week has bad service, use highest week
		if float64(ents[fetchedWeek].TotalService.Int)/float64(highestService) < minServiceRatio {
			// fmt.Println("fetched week has poor service ratio, falling back to highest week:", fetchedWeek)
			fetchedWeek = highestIdx
		}

		// Expand window in both directions from chosen week
		startDate = ents[fetchedWeek].StartDate.Time
		endDate = ents[fetchedWeek].EndDate.Time
		for i := fetchedWeek; i < len(ents); i++ {
			ent := ents[i]
			if float64(ent.TotalService.Int)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.StartDate.Time.Before(startDate) {
				startDate = ent.StartDate.Time
			}
			endDate = ent.EndDate.Time
		}
		for i := fetchedWeek - 1; i > 0; i-- {
			ent := ents[i]
			if float64(ent.TotalService.Int)/float64(highestService) < minServiceRatio {
				break
			}
			if ent.EndDate.Time.After(endDate) {
				endDate = ent.EndDate.Time
			}
			startDate = ent.StartDate.Time
		}
	}

	// Check again to find the highest service week in the window
	// This will be used as the typical week for dates outside the window
	// bestWeek must start with a Monday
	bestWeek = ents[0].StartDate.Time
	bestService := ents[0].TotalService.Int
	for _, ent := range ents {
		sd := ent.StartDate.Time
		ed := ent.EndDate.Time
		if (sd.Before(endDate) || sd.Equal(endDate)) && (ed.After(startDate) || ed.Equal(startDate)) {
			if ent.TotalService.Int > bestService {
				bestService = ent.TotalService.Int
				bestWeek = ent.StartDate.Time
			}
		}
	}
	return startDate, endDate, bestWeek, nil
}

// Loaders

func (f *DBFinder) TripsByID(ctx context.Context, ids []int) (ents []*model.Trip, errs []error) {
	ents, err := f.FindTrips(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.Trip{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Trip, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

// Simple ID loaders
func (f *DBFinder) LevelsByID(ctx context.Context, ids []int) ([]*model.Level, []error) {
	var ents []*model.Level
	MustSelect(ctx,
		f.db,
		quickSelect("gtfs_levels", nil, nil, ids),
		&ents,
	)
	byid := map[int]*model.Level{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Level, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) CalendarsByID(ctx context.Context, ids []int) ([]*model.Calendar, []error) {
	var ents []*model.Calendar
	MustSelect(ctx,
		f.db,
		quickSelect("gtfs_calendars", nil, nil, ids),
		&ents,
	)
	byid := map[int]*model.Calendar{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Calendar, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) ShapesByID(ctx context.Context, ids []int) ([]*model.Shape, []error) {
	var ents []*model.Shape
	MustSelect(ctx,
		f.db,
		quickSelect("gtfs_shapes", nil, nil, ids),
		&ents,
	)
	byid := map[int]*model.Shape{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Shape, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) FeedVersionsByID(ctx context.Context, ids []int) ([]*model.FeedVersion, []error) {
	ents, err := f.FindFeedVersions(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.FeedVersion{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.FeedVersion, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) FeedsByID(ctx context.Context, ids []int) ([]*model.Feed, []error) {
	ents, err := f.FindFeeds(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.Feed{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Feed, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) AgenciesByID(ctx context.Context, ids []int) ([]*model.Agency, []error) {
	var ents []*model.Agency
	ents, err := f.FindAgencies(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.Agency{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Agency, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) StopsByID(ctx context.Context, ids []int) ([]*model.Stop, []error) {
	ents, err := f.FindStops(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.Stop{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Stop, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) RoutesByID(ctx context.Context, ids []int) ([]*model.Route, []error) {
	ents, err := f.FindRoutes(ctx, nil, nil, ids, nil)
	if err != nil {
		return nil, []error{err}
	}
	byid := map[int]*model.Route{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Route, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) CensusTableByID(ctx context.Context, ids []int) ([]*model.CensusTable, []error) {
	var ents []*model.CensusTable
	MustSelect(ctx,
		f.db,
		quickSelect("tl_census_tables", nil, nil, ids),
		&ents,
	)
	byid := map[int]*model.CensusTable{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.CensusTable, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) FeedVersionGtfsImportsByFeedVersionID(ctx context.Context, ids []int) ([]*model.FeedVersionGtfsImport, []error) {
	var ents []*model.FeedVersionGtfsImport
	MustSelect(ctx,
		f.db,
		quickSelect("feed_version_gtfs_imports", nil, nil, nil).Where(sq.Eq{"feed_version_id": ids}),
		&ents,
	)
	byid := map[int]*model.FeedVersionGtfsImport{}
	for _, ent := range ents {
		byid[ent.FeedVersionID] = ent
	}
	ents2 := make([]*model.FeedVersionGtfsImport, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) FeedStatesByFeedID(ctx context.Context, ids []int) ([]*model.FeedState, []error) {
	var ents []*model.FeedState
	MustSelect(ctx,
		f.db,
		quickSelect("feed_states", nil, nil, nil).Where(sq.Eq{"feed_id": ids}),
		&ents,
	)
	byid := map[int]*model.FeedState{}
	for _, ent := range ents {
		byid[ent.FeedID] = ent
	}
	ents2 := make([]*model.FeedState, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) FeedFetchesStatesByFeedID(ctx context.Context, ids []int) ([]*model.FeedFetch, []error) {
	var ents []*model.FeedFetch
	MustSelect(ctx,
		f.db,
		quickSelect("feed_fetches", nil, nil, nil).Where(sq.Eq{"feed_id": ids}),
		&ents,
	)
	byid := map[int]*model.FeedFetch{}
	for _, ent := range ents {
		byid[ent.FeedID] = ent
	}
	ents2 := make([]*model.FeedFetch, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

func (f *DBFinder) OperatorsByCOIF(ctx context.Context, ids []int) ([]*model.Operator, []error) {
	var ents []*model.Operator
	MustSelect(ctx,
		f.db,
		OperatorSelect(nil, nil, ids, nil, nil),
		&ents,
	)
	byid := map[int]*model.Operator{}
	for _, ent := range ents {
		byid[ent.ID] = ent
	}
	ents2 := make([]*model.Operator, len(ids))
	for i, id := range ids {
		ents2[i] = byid[id]
	}
	return ents2, nil
}

// Param loaders

func (f *DBFinder) OperatorsByFeedID(ctx context.Context, params []model.OperatorParam) ([][]*model.Operator, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedID)
	}
	qents := []*model.Operator{}
	MustSelect(ctx,
		f.db,
		lateralWrap(OperatorSelect(params[0].Limit, nil, nil, ids, params[0].Where), "current_feeds", "id", "feed_id", ids),
		&qents,
	)
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

func marshalParam(param interface{}) string {
	a, _ := json.Marshal(param)
	return string(a)
}

func (f *DBFinder) FeedFetchesByFeedID(ctx context.Context, params []model.FeedFetchParam) ([][]*model.FeedFetch, []error) {
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
		MustSelect(ctx,
			f.db,
			lateralWrap(q, "current_feeds", "id", "feed_id", ids),
			&qents,
		)
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

func (f *DBFinder) FeedsByOperatorOnestopID(ctx context.Context, params []model.FeedParam) ([][]*model.Feed, []error) {
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
	q := FeedSelect(nil, nil, nil, params[0].Where).
		Distinct().Options("on (coif.resolved_onestop_id, t.id)").
		Column("coif.resolved_onestop_id as operator_onestop_id").
		Join("current_operators_in_feed coif on coif.feed_id = t.id").
		Where(sq.Eq{"coif.resolved_onestop_id": osids})
	MustSelect(ctx,
		f.db,
		q,
		&qents,
	)
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

func (f *DBFinder) FrequenciesByTripID(ctx context.Context, params []model.FrequencyParam) ([][]*model.Frequency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.TripID)
	}
	qents := []*model.Frequency{}
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelect("gtfs_frequencies", params[0].Limit, nil, nil), "gtfs_trips", "id", "trip_id", ids),
		&qents,
	)
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

func (f *DBFinder) StopTimesByTripID(ctx context.Context, params []model.StopTimeParam) ([][]*model.StopTime, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	limit := checkLimit(params[0].Limit)
	tpairs := []FVPair{}
	for _, p := range params {
		tpairs = append(tpairs, FVPair{EntityID: p.TripID, FeedVersionID: p.FeedVersionID})
	}
	qents := []*model.StopTime{}
	MustSelect(ctx,
		f.db,
		StopTimeSelect(tpairs, nil, params[0].Where),
		&qents,
	)
	group := map[int][]*model.StopTime{}
	for _, ent := range qents {
		group[atoi(ent.TripID)] = append(group[atoi(ent.TripID)], ent)
	}
	for k, ents := range group {
		if uint64(len(ents)) > limit {
			group[k] = ents[0:limit]
		}
	}
	var ents [][]*model.StopTime
	for _, tp := range tpairs {
		ents = append(ents, group[tp.EntityID])
	}
	return ents, nil
}

func (f *DBFinder) StopTimesByStopID(ctx context.Context, params []model.StopTimeParam) ([][]*model.StopTime, []error) {
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
			panic(err)
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
			MustSelect(ctx,
				f.db,
				StopDeparturesSelect(dg.pairs, p),
				&qents,
			)
		} else {
			// Otherwise get all stop_times for stop
			MustSelect(ctx,
				f.db,
				StopTimeSelect(nil, dg.pairs, p),
				&qents,
			)
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

func (f *DBFinder) RouteStopsByStopID(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.StopID)
	}
	qents := []*model.RouteStop{}
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelectOrder("tl_route_stops", params[0].Limit, nil, nil, "stop_id"), "gtfs_stops", "id", "stop_id", ids),
		&qents,
	)
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

func (f *DBFinder) StopsByRouteID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
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
	qso := StopSelect(params[0].Limit, nil, nil, false, params[0].Where)
	qso = qso.Join("tl_route_stops on tl_route_stops.stop_id = t.id").Where(sq.Eq{"route_id": routeIds}).Column("route_id")
	MustSelect(ctx,
		f.db,
		qso,
		&qents,
	)
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

func (f *DBFinder) RouteStopsByRouteID(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteStop{}
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelectOrder("tl_route_stops", params[0].Limit, nil, nil, "stop_id"), "gtfs_routes", "id", "route_id", ids),
		&qents,
	)
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

func (f *DBFinder) RouteHeadwaysByRouteID(ctx context.Context, params []model.RouteHeadwayParam) ([][]*model.RouteHeadway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteHeadway{}
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelectOrder("tl_route_headways", params[0].Limit, nil, nil, "route_id"), "gtfs_routes", "id", "route_id", ids),
		&qents,
	)
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

func (f *DBFinder) FeedVersionFileInfosByFeedVersionID(ctx context.Context, params []model.FeedVersionFileInfoParam) ([][]*model.FeedVersionFileInfo, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedVersionFileInfo{}
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelectOrder("feed_version_file_infos", params[0].Limit, nil, nil, "id"), "feed_versions", "id", "feed_version_id", ids),
		&qents,
	)
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

func (f *DBFinder) StopsByParentStopID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ParentStopID)
	}
	qents := []*model.Stop{}
	MustSelect(ctx,
		f.db,
		lateralWrap(StopSelect(params[0].Limit, nil, nil, false, params[0].Where), "gtfs_stops", "id", "parent_station", ids),
		&qents,
	)
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

func (f *DBFinder) FeedVersionsByFeedID(ctx context.Context, params []model.FeedVersionParam) ([][]*model.FeedVersion, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedID)
	}
	qents := []*model.FeedVersion{}
	MustSelect(ctx,
		f.db,
		lateralWrap(FeedVersionSelect(params[0].Limit, nil, nil, params[0].Where), "current_feeds", "id", "feed_id", ids),
		&qents,
	)
	group := map[int][]*model.FeedVersion{}
	for _, ent := range qents {
		group[ent.FeedID] = append(group[ent.FeedID], ent)
	}
	var ents [][]*model.FeedVersion
	for _, id := range ids {
		ents = append(ents, group[id])
	}
	return ents, nil
}

func (f *DBFinder) AgencyPlacesByAgencyID(ctx context.Context, params []model.AgencyPlaceParam) ([][]*model.AgencyPlace, []error) {
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
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelectOrder("tl_agency_places", params[0].Limit, nil, nil, "agency_id").Where(sq.GtOrEq{"rank": minRank}), "gtfs_agencies", "id", "agency_id", ids),
		&qents,
	)
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

func (f *DBFinder) RouteGeometriesByRouteID(ctx context.Context, params []model.RouteGeometryParam) ([][]*model.RouteGeometry, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteGeometry{}
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelectOrder("tl_route_geometries", params[0].Limit, nil, nil, "route_id"), "gtfs_routes", "id", "route_id", ids),
		&qents,
	)
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

func (f *DBFinder) TripsByRouteID(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.Trip{}
	MustSelect(ctx,
		f.db,
		lateralWrap(TripSelect(params[0].Limit, nil, nil, false, params[0].Where), "gtfs_routes", "id", "route_id", ids),
		&qents,
	)
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

func (f *DBFinder) RoutesByAgencyID(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.AgencyID)
	}
	qents := []*model.Route{}
	MustSelect(ctx,
		f.db,
		lateralWrap(RouteSelect(params[0].Limit, nil, nil, false, params[0].Where), "gtfs_agencies", "id", "agency_id", ids),
		&qents,
	)
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

func (f *DBFinder) AgenciesByFeedVersionID(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Agency{}
	MustSelect(ctx,
		f.db,
		lateralWrap(AgencySelect(params[0].Limit, nil, nil, false, params[0].Where), "feed_versions", "id", "feed_version_id", ids),
		&qents,
	)
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

func (f *DBFinder) AgenciesByOnestopID(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []string{}
	for _, p := range params {
		ids = append(ids, *p.OnestopID)
	}
	qents := []*model.Agency{}
	MustSelect(ctx,
		f.db,
		AgencySelect(params[0].Limit, nil, nil, true, nil).Where(sq.Eq{"onestop_id": ids}), // active=true
		&qents,
	)
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

func (f *DBFinder) StopsByFeedVersionID(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Stop{}
	MustSelect(ctx,
		f.db,
		lateralWrap(StopSelect(params[0].Limit, nil, nil, false, params[0].Where), "feed_versions", "id", "feed_version_id", ids),
		&qents,
	)
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

func (f *DBFinder) TripsByFeedVersionID(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Trip{}
	MustSelect(ctx,
		f.db,
		lateralWrap(TripSelect(params[0].Limit, nil, nil, false, params[0].Where), "feed_versions", "id", "feed_version_id", ids),
		&qents,
	)
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

func (f *DBFinder) FeedInfosByFeedVersionID(ctx context.Context, params []model.FeedInfoParam) ([][]*model.FeedInfo, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedInfo{}
	MustSelect(ctx,
		f.db,
		lateralWrap(quickSelectOrder("gtfs_feed_infos", params[0].Limit, nil, nil, "id"), "feed_versions", "id", "feed_version_id", ids),
		&qents,
	)
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

func (f *DBFinder) RoutesByFeedVersionID(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Route{}
	MustSelect(ctx,
		f.db,
		lateralWrap(RouteSelect(params[0].Limit, nil, nil, false, params[0].Where), "feed_versions", "id", "feed_version_id", ids),
		&qents,
	)
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

func (f *DBFinder) FeedVersionServiceLevelsByFeedVersionID(ctx context.Context, params []model.FeedVersionServiceLevelParam) ([][]*model.FeedVersionServiceLevel, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedVersionServiceLevel{}
	MustSelect(ctx,
		f.db,
		lateralWrap(FeedVersionServiceLevelSelect(params[0].Limit, nil, nil, params[0].Where), "feed_versions", "id", "feed_version_id", ids),
		&qents,
	)
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

func (f *DBFinder) PathwaysByFromStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FromStopID)
	}
	qents := []*model.Pathway{}
	MustSelect(ctx,
		f.db,
		lateralWrap(PathwaySelect(params[0].Limit, nil, nil, params[0].Where), "gtfs_stops", "id", "from_stop_id", ids),
		&qents,
	)
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

func (f *DBFinder) PathwaysByToStopID(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ToStopID)
	}
	qents := []*model.Pathway{}
	MustSelect(ctx,
		f.db,
		lateralWrap(PathwaySelect(params[0].Limit, nil, nil, params[0].Where), "gtfs_stops", "id", "to_stop_id", ids),
		&qents,
	)
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

func (f *DBFinder) CalendarDatesByServiceID(ctx context.Context, params []model.CalendarDateParam) ([][]*model.CalendarDate, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ServiceID)
	}
	qents := []*model.CalendarDate{}
	MustSelect(ctx,
		f.db,
		quickSelectOrder("gtfs_calendar_dates", nil, nil, nil, "date").Where(sq.Eq{"service_id": ids}),
		&qents,
	)
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

func (f *DBFinder) CensusGeographiesByEntityID(ctx context.Context, params []model.CensusGeographyParam) ([][]*model.CensusGeography, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.EntityID)
	}
	qents := []*model.CensusGeography{}
	MustSelect(ctx,
		f.db,
		CensusGeographySelect(&params[0], ids),
		&qents,
	)
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

func (f *DBFinder) CensusValuesByGeographyID(ctx context.Context, params []model.CensusValueParam) ([][]*model.CensusValue, []error) {
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
	MustSelect(ctx,
		f.db,
		// lateralWrap(CensusValueSelect(&params[0], ids), "tl_census_geographies", "id", "geography_id", ids),
		CensusValueSelect(&params[0], ids),
		&qents,
	)
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
