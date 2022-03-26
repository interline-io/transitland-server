package find

import (
	sq "github.com/Masterminds/squirrel"
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

func (f *DBFinder) FindAgencies(limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	var ents []*model.Agency
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := AgencySelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindRoutes(limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	var ents []*model.Route
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := RouteSelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindStops(limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	var ents []*model.Stop
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := StopSelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindTrips(limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	var ents []*model.Trip
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := TripSelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindFeedVersions(limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	var ents []*model.FeedVersion
	MustSelect(f.db, FeedVersionSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func (f *DBFinder) FindFeeds(limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	var ents []*model.Feed
	MustSelect(f.db, FeedSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func (f *DBFinder) FindOperators(limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	var ents []*model.Operator
	MustSelect(f.db, OperatorSelect(limit, after, ids, nil, where), &ents)
	return ents, nil
}

func (f *DBFinder) RouteStopBuffer(param *model.RouteStopBufferParam) ([]*model.RouteStopBuffer, error) {
	if param == nil {
		return nil, nil
	}
	var ents []*model.RouteStopBuffer
	q := RouteStopBufferSelect(*param)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

// Loaders

func (f *DBFinder) TripsByID(ids []int) (ents []*model.Trip, errs []error) {
	ents, err := f.FindTrips(nil, nil, ids, nil)
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
func (f *DBFinder) LevelsByID(ids []int) ([]*model.Level, []error) {
	var ents []*model.Level
	MustSelect(
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

func (f *DBFinder) CalendarsByID(ids []int) ([]*model.Calendar, []error) {
	var ents []*model.Calendar
	MustSelect(
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

func (f *DBFinder) ShapesByID(ids []int) ([]*model.Shape, []error) {
	var ents []*model.Shape
	MustSelect(
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

func (f *DBFinder) FeedVersionsByID(ids []int) ([]*model.FeedVersion, []error) {
	ents, err := f.FindFeedVersions(nil, nil, ids, nil)
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

func (f *DBFinder) FeedsByID(ids []int) ([]*model.Feed, []error) {
	ents, err := f.FindFeeds(nil, nil, ids, nil)
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

func (f *DBFinder) AgenciesByID(ids []int) ([]*model.Agency, []error) {
	var ents []*model.Agency
	ents, err := f.FindAgencies(nil, nil, ids, nil)
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

func (f *DBFinder) StopsByID(ids []int) ([]*model.Stop, []error) {
	ents, err := f.FindStops(nil, nil, ids, nil)
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

func (f *DBFinder) RoutesByID(ids []int) ([]*model.Route, []error) {
	ents, err := f.FindRoutes(nil, nil, ids, nil)
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

func (f *DBFinder) CensusTableByID(ids []int) ([]*model.CensusTable, []error) {
	var ents []*model.CensusTable
	MustSelect(
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

func (f *DBFinder) FeedVersionGtfsImportsByFeedVersionID(ids []int) ([]*model.FeedVersionGtfsImport, []error) {
	var ents []*model.FeedVersionGtfsImport
	MustSelect(
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

func (f *DBFinder) FeedStatesByFeedID(ids []int) ([]*model.FeedState, []error) {
	var ents []*model.FeedState
	MustSelect(
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

func (f *DBFinder) OperatorsByCOIF(ids []int) ([]*model.Operator, []error) {
	var ents []*model.Operator
	MustSelect(
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

func (f *DBFinder) OperatorsByFeedID(params []model.OperatorParam) ([][]*model.Operator, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedID)
	}
	qents := []*model.Operator{}
	MustSelect(
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

func (f *DBFinder) FrequenciesByTripID(params []model.FrequencyParam) ([][]*model.Frequency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.TripID)
	}
	qents := []*model.Frequency{}
	MustSelect(
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

func (f *DBFinder) StopTimesByTripID(params []model.StopTimeParam) ([][]*model.StopTime, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	limit := checkLimit(params[0].Limit)
	tpairs := []FVPair{}
	for _, p := range params {
		tpairs = append(tpairs, FVPair{EntityID: p.TripID, FeedVersionID: p.FeedVersionID})
	}
	qents := []*model.StopTime{}
	MustSelect(
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

func (f *DBFinder) StopTimesByStopID(params []model.StopTimeParam) ([][]*model.StopTime, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	limit := checkLimit(params[0].Limit)
	tzgroups := map[string][]FVPair{}
	group := map[int][]*model.StopTime{}
	for _, p := range params {
		s := ""
		if p.StopTimezone != nil {
			s = p.StopTimezone.String()
		}
		tzgroups[s] = append(tzgroups[s], FVPair{EntityID: p.StopID, FeedVersionID: p.FeedVersionID})
	}
	for tzloc, tzpairs := range tzgroups {
		qents := []*model.StopTime{}
		if p := params[0].Where; p != nil && (p.ServiceDate != nil || p.Next != nil) {
			p.Timezone = &tzloc
			MustSelect(
				f.db,
				StopDeparturesSelect(tzpairs, p),
				&qents,
			)
		} else {
			// Otherwise get all stop_times for stop
			MustSelect(
				f.db,
				StopTimeSelect(nil, tzpairs, params[0].Where),
				&qents,
			)
		}
		for _, ent := range qents {
			group[atoi(ent.StopID)] = append(group[atoi(ent.StopID)], ent)
		}
		for k, ents := range group {
			if uint64(len(ents)) > limit {
				group[k] = ents[0:limit]
			}
		}
	}
	var ents [][]*model.StopTime
	for _, sp := range params {
		ents = append(ents, group[sp.StopID])
	}
	return ents, nil
}

func (f *DBFinder) RouteStopsByStopID(params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.StopID)
	}
	qents := []*model.RouteStop{}
	MustSelect(
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

func (f *DBFinder) StopsByRouteID(params []model.StopParam) ([][]*model.Stop, []error) {
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
	MustSelect(
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

func (f *DBFinder) RouteStopsByRouteID(params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteStop{}
	MustSelect(
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

func (f *DBFinder) RouteHeadwaysByRouteID(params []model.RouteHeadwayParam) ([][]*model.RouteHeadway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteHeadway{}
	MustSelect(
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

func (f *DBFinder) FeedVersionFileInfosByFeedVersionID(params []model.FeedVersionFileInfoParam) ([][]*model.FeedVersionFileInfo, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedVersionFileInfo{}
	MustSelect(
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

func (f *DBFinder) StopsByParentStopID(params []model.StopParam) ([][]*model.Stop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ParentStopID)
	}
	qents := []*model.Stop{}
	MustSelect(
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

func (f *DBFinder) FeedVersionsByFeedID(params []model.FeedVersionParam) ([][]*model.FeedVersion, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedID)
	}
	qents := []*model.FeedVersion{}
	MustSelect(
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

func (f *DBFinder) AgencyPlacesByAgencyID(params []model.AgencyPlaceParam) ([][]*model.AgencyPlace, []error) {
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
	MustSelect(
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

func (f *DBFinder) RouteGeometriesByRouteID(params []model.RouteGeometryParam) ([][]*model.RouteGeometry, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.RouteGeometry{}
	MustSelect(
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

func (f *DBFinder) TripsByRouteID(params []model.TripParam) ([][]*model.Trip, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.RouteID)
	}
	qents := []*model.Trip{}
	MustSelect(
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

func (f *DBFinder) RoutesByAgencyID(params []model.RouteParam) ([][]*model.Route, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.AgencyID)
	}
	qents := []*model.Route{}
	MustSelect(
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

func (f *DBFinder) AgenciesByFeedVersionID(params []model.AgencyParam) ([][]*model.Agency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Agency{}
	MustSelect(
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

func (f *DBFinder) AgenciesByOnestopID(params []model.AgencyParam) ([][]*model.Agency, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []string{}
	for _, p := range params {
		ids = append(ids, *p.OnestopID)
	}
	qents := []*model.Agency{}
	MustSelect(
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

func (f *DBFinder) StopsByFeedVersionID(params []model.StopParam) ([][]*model.Stop, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Stop{}
	MustSelect(
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

func (f *DBFinder) TripsByFeedVersionID(params []model.TripParam) ([][]*model.Trip, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Trip{}
	MustSelect(
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

func (f *DBFinder) FeedInfosByFeedVersionID(params []model.FeedInfoParam) ([][]*model.FeedInfo, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedInfo{}
	MustSelect(
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

func (f *DBFinder) RoutesByFeedVersionID(params []model.RouteParam) ([][]*model.Route, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.Route{}
	MustSelect(
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

func (f *DBFinder) FeedVersionServiceLevelsByFeedVersionID(params []model.FeedVersionServiceLevelParam) ([][]*model.FeedVersionServiceLevel, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FeedVersionID)
	}
	qents := []*model.FeedVersionServiceLevel{}
	MustSelect(
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

func (f *DBFinder) PathwaysByFromStopID(params []model.PathwayParam) ([][]*model.Pathway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.FromStopID)
	}
	qents := []*model.Pathway{}
	MustSelect(
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

func (f *DBFinder) PathwaysByToStopID(params []model.PathwayParam) ([][]*model.Pathway, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ToStopID)
	}
	qents := []*model.Pathway{}
	MustSelect(
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

func (f *DBFinder) CalendarDatesByServiceID(params []model.CalendarDateParam) ([][]*model.CalendarDate, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.ServiceID)
	}
	qents := []*model.CalendarDate{}
	MustSelect(
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

func (f *DBFinder) CensusGeographiesByEntityID(params []model.CensusGeographyParam) ([][]*model.CensusGeography, []error) {
	if len(params) == 0 {
		return nil, nil
	}
	ids := []int{}
	for _, p := range params {
		ids = append(ids, p.EntityID)
	}
	qents := []*model.CensusGeography{}
	MustSelect(
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

func (f *DBFinder) CensusValuesByGeographyID(params []model.CensusValueParam) ([][]*model.CensusValue, []error) {
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
	MustSelect(
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
