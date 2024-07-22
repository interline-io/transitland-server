package gql

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tlxy"
	"github.com/interline-io/transitland-server/internal/directions"
	"github.com/interline-io/transitland-server/model"
)

// STOP

type stopResolver struct {
	*Resolver
}

func (r *stopResolver) Cursor(ctx context.Context, obj *model.Stop) (*model.Cursor, error) {
	c := model.NewCursor(obj.FeedVersionID, obj.ID)
	return &c, nil
}

func (r *stopResolver) FeedVersion(ctx context.Context, obj *model.Stop) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(ctx, obj.FeedVersionID)()
}

func (r *stopResolver) Level(ctx context.Context, obj *model.Stop) (*model.Level, error) {
	if !obj.LevelID.Valid {
		return nil, nil
	}
	return For(ctx).LevelsByID.Load(ctx, obj.LevelID.Int())()
}

func (r *stopResolver) ChildLevels(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Level, error) {
	return For(ctx).LevelsByParentStationID.Load(ctx, model.LevelParam{ParentStationID: obj.ID, Limit: limit})()
}

func (r *stopResolver) Parent(ctx context.Context, obj *model.Stop) (*model.Stop, error) {
	if !obj.ParentStation.Valid {
		return nil, nil
	}
	return For(ctx).StopsByID.Load(ctx, obj.ParentStation.Int())()
}

func (r *stopResolver) Children(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Stop, error) {
	return For(ctx).StopsByParentStopID.Load(ctx, model.StopParam{ParentStopID: obj.ID, Limit: checkLimit(limit)})()
}

func (r *stopResolver) Place(ctx context.Context, obj *model.Stop) (*model.StopPlace, error) {
	pt := tlxy.Point{Lon: obj.Geometry.X(), Lat: obj.Geometry.Y()}
	return For(ctx).StopPlacesByStopID.Load(ctx, model.StopPlaceParam{ID: obj.ID, Point: pt})()
}

func (r *stopResolver) RouteStops(ctx context.Context, obj *model.Stop, limit *int) ([]*model.RouteStop, error) {
	return For(ctx).RouteStopsByStopID.Load(ctx, model.RouteStopParam{StopID: obj.ID, Limit: checkLimit(limit)})()
}

func (r *stopResolver) PathwaysFromStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByFromStopID.Load(ctx, model.PathwayParam{FromStopID: obj.ID, Limit: checkLimit(limit)})()
}

func (r *stopResolver) PathwaysToStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByToStopID.Load(ctx, model.PathwayParam{ToStopID: obj.ID, Limit: checkLimit(limit)})()
}

func (r *stopResolver) ExternalReference(ctx context.Context, obj *model.Stop) (*model.StopExternalReference, error) {
	return For(ctx).StopExternalReferencesByStopID.Load(ctx, obj.ID)()
}

func (r *stopResolver) Observations(ctx context.Context, obj *model.Stop, limit *int, where *model.StopObservationFilter) ([]*model.StopObservation, error) {
	return For(ctx).StopObservationsByStopID.Load(ctx, model.StopObservationParam{StopID: obj.ID, Where: where, Limit: checkLimit(limit)})()
}

func (r *stopResolver) Departures(ctx context.Context, obj *model.Stop, limit *int, where *model.StopTimeFilter) ([]*model.StopTime, error) {
	if where == nil {
		where = &model.StopTimeFilter{}
	}
	t := true
	where.ExcludeLast = &t
	return r.getStopTimes(ctx, obj, limit, where)
}

func (r *stopResolver) Arrivals(ctx context.Context, obj *model.Stop, limit *int, where *model.StopTimeFilter) ([]*model.StopTime, error) {
	if where == nil {
		where = &model.StopTimeFilter{}
	}
	t := true
	where.ExcludeFirst = &t
	return r.getStopTimes(ctx, obj, limit, where)
}

func (r *stopResolver) StopTimes(ctx context.Context, obj *model.Stop, limit *int, where *model.StopTimeFilter) ([]*model.StopTime, error) {
	return r.getStopTimes(ctx, obj, limit, where)
}

func (r *stopResolver) getStopTimes(ctx context.Context, obj *model.Stop, limit *int, where *model.StopTimeFilter) ([]*model.StopTime, error) {
	// We need timezone information to do anything with stop times
	loc, ok := model.ForContext(ctx).RTFinder.StopTimezone(obj.ID, obj.StopTimezone)
	if !ok {
		return nil, errors.New("timezone not available for stop")
	}

	// Local times
	nowLocal := time.Now().In(loc)
	if model.ForContext(ctx).Clock != nil {
		nowLocal = model.ForContext(ctx).Clock.Now().In(loc)
	}

	// Pre-processing
	// Convert Start, End to StartTime, EndTime
	if where != nil {
		if where.Start != nil {
			where.StartTime = ptr(where.Start.Seconds)
			where.Start = nil
		}
		if where.End != nil {
			where.EndTime = ptr(where.End.Seconds)
			where.End = nil
		}
	}

	// Further processing of the StopTimeFilter
	if where != nil {
		// Set ServiceDate to local timezone
		// ServiceDate is a strict GTFS calendar date
		if where.ServiceDate != nil {
			where.ServiceDate = tzTruncate(where.ServiceDate.Val, loc)
		}

		// Set Date to local timezone
		if where.Date != nil {
			where.Date = tzTruncate(where.Date.Val, loc)
		}

		// Convert relative date
		if where.RelativeDate != nil {
			s, err := tt.RelativeDate(nowLocal, kebabize(string(*where.RelativeDate)))
			if err != nil {
				return nil, err
			}
			where.Date = tzTruncate(s, loc)
		}

		// Convert where.Next into departure date and time window
		if where.Next != nil {
			if where.Date == nil {
				where.Date = tzTruncate(nowLocal, loc)
			}
			st := nowLocal.Hour()*3600 + nowLocal.Minute()*60 + nowLocal.Second()
			where.StartTime = ptr(st)
			where.EndTime = ptr(st + *where.Next)
		}

		// Map date into service window
		if nilOr(where.UseServiceWindow, false) {
			serviceLevels, ok := r.fvslCache.Get(ctx, obj.FeedVersionID)
			if !ok {
				return nil, errors.New("service level information not available for feed version")
			}
			// Check if date is outside window
			if where.Date != nil {
				s := where.Date.Val
				if s.Before(serviceLevels.StartDate) || s.After(serviceLevels.EndDate) {
					dow := int(s.Weekday()) - 1
					if dow < 0 {
						dow = 6
					}
					where.Date = tzTruncate(serviceLevels.BestWeek.AddDate(0, 0, dow), loc)
				}
			}
			// Repeat for ServiceDate
			if where.ServiceDate != nil {
				s := where.ServiceDate.Val
				if s.Before(serviceLevels.StartDate) || s.After(serviceLevels.EndDate) {
					dow := int(s.Weekday()) - 1
					if dow < 0 {
						dow = 6
					}
					where.ServiceDate = tzTruncate(serviceLevels.BestWeek.AddDate(0, 0, dow), loc)
				}
			}
		}
	}

	// Crossing day boundaries...
	var whereGroups []*model.StopTimeFilter
	if where != nil && where.Date != nil {
		date := where.Date
		dayStart := 0
		dayEnd := 24 * 60 * 60
		dayEndMax := 100 * 60 * 60
		whereStartTime := dayStart
		if where.StartTime != nil {
			whereStartTime = *where.StartTime
		}
		whereEndTime := dayEnd
		if where.EndTime != nil {
			whereEndTime = *where.EndTime
		}
		lookBehind := 6 * 3600
		// if where.ServiceDateLookbehind != nil {
		// 	lookBehind = where.ServiceDateLookbehind.Seconds
		// }
		// Query previous day
		if whereStartTime < lookBehind {
			whereCopy := *where
			whereCopy.ServiceDate = ptr(tt.NewDate(date.Val.AddDate(0, 0, -1)))
			whereCopy.StartTime = ptr(dayEnd + whereStartTime)
			whereCopy.EndTime = ptr(dayEndMax)
			whereGroups = append(whereGroups, &whereCopy)
		}
		// Query requested day, clamped to 0 - 24h
		whereCopy := *where
		whereCopy.ServiceDate = ptr(tt.NewDate(date.Val))
		whereCopy.StartTime = ptr(max(dayStart, whereStartTime))
		whereCopy.EndTime = ptr(whereEndTime)
		whereGroups = append(whereGroups, &whereCopy)
		// Query next day
		if whereEndTime > dayEnd {
			whereCopy := *where
			whereCopy.ServiceDate = ptr(tt.NewDate(date.Val.AddDate(0, 0, 1)))
			whereCopy.StartTime = ptr(dayStart)
			whereCopy.EndTime = ptr(whereEndTime - dayEnd)
			whereGroups = append(whereGroups, &whereCopy)
		}
	}

	// Default
	if len(whereGroups) == 0 {
		whereGroups = append(whereGroups, where)
	}

	// Query for each day group
	var sts []*model.StopTime
	for _, w := range whereGroups {
		ents, err := (For(ctx).StopTimesByStopID.Load(ctx, model.StopTimeParam{
			StopID:        obj.ID,
			FeedVersionID: obj.FeedVersionID,
			Limit:         checkLimit(limit),
			Where:         w,
		})())
		if err != nil {
			return nil, err
		}
		// Set service date and calendar date; move calendar date one day forward if > midnight
		if w != nil && w.ServiceDate != nil {
			for _, ent := range ents {
				ent.ServiceDate = tt.NewDate(w.ServiceDate.Val)
				if ent.ArrivalTime.Seconds < 24*60*60 {
					ent.Date = tt.NewDate(w.ServiceDate.Val)
				} else {
					ent.Date = tt.NewDate(w.ServiceDate.Val.AddDate(0, 0, 1))
				}
			}
		}
		sts = append(sts, ents...)
	}

	// Merge scheduled stop times with rt stop times
	// TODO: handle StopTimeFilter in RT
	// Handle scheduled trips; these can be matched on trip_id or (route_id,direction_id,...)
	for _, st := range sts {
		ft := model.Trip{}
		ft.FeedVersionID = obj.FeedVersionID
		ft.TripID, _ = model.ForContext(ctx).RTFinder.GetGtfsTripID(atoi(st.TripID)) // TODO!
		if ste, ok := model.ForContext(ctx).RTFinder.FindStopTimeUpdate(&ft, st); ok {
			st.RTStopTimeUpdate = ste
		}
	}
	// Handle added trips; these must specify stop_id in StopTimeUpdates
	for _, rtTrip := range model.ForContext(ctx).RTFinder.GetAddedTripsForStop(obj) {
		for _, stu := range rtTrip.StopTimeUpdate {
			if stu.GetStopId() != obj.StopID {
				continue
			}
			// create a new StopTime
			rtst := &model.StopTime{}
			rtst.RTTripID = rtTrip.Trip.GetTripId()
			rtst.RTStopTimeUpdate = stu
			rtst.FeedVersionID = obj.FeedVersionID
			rtst.TripID = "0"
			rtst.StopID = strconv.Itoa(obj.ID)
			rtst.StopSequence = int(stu.GetStopSequence())
			sts = append(sts, rtst)
		}
	}
	// Sort by scheduled departure time.
	// TODO: Sort by rt departure time? Requires full StopTime Resolver for timezones, processing, etc.
	sort.Slice(sts, func(i, j int) bool {
		sta := sts[i]
		stb := sts[j]
		a := int(sta.ServiceDate.Val.Unix()) + sta.DepartureTime.Seconds
		b := int(stb.ServiceDate.Val.Unix()) + stb.DepartureTime.Seconds
		return a < b
	})
	return sts, nil
}

func (r *stopResolver) Alerts(ctx context.Context, obj *model.Stop, active *bool, limit *int) ([]*model.Alert, error) {
	rtAlerts := model.ForContext(ctx).RTFinder.FindAlertsForStop(obj, checkLimit(limit), active)
	return rtAlerts, nil
}

func (r *stopResolver) Directions(ctx context.Context, obj *model.Stop, from *model.WaypointInput, to *model.WaypointInput, mode *model.StepMode, departAt *time.Time) (*model.Directions, error) {
	oc := obj.Coordinates()
	swp := &model.WaypointInput{
		Lon:  oc[0],
		Lat:  oc[1],
		Name: &obj.StopName,
	}
	p := model.DirectionRequest{}
	if from != nil {
		p.From = from
		p.To = swp
	} else if to != nil {
		p.From = swp
		p.To = to
	}
	if mode != nil {
		p.Mode = *mode
	}
	return directions.HandleRequest("", p)
}

func (r *stopResolver) NearbyStops(ctx context.Context, obj *model.Stop, limit *int, radius *float64) ([]*model.Stop, error) {
	c := obj.Coordinates()
	nearbyStops, err := model.ForContext(ctx).Finder.FindStops(ctx, checkLimit(limit), nil, nil, &model.StopFilter{Near: &model.PointRadius{Lon: c[0], Lat: c[1], Radius: checkFloat(radius, 0, MAX_RADIUS)}})
	return nearbyStops, err
}

//////////

type stopExternalReferenceResolver struct {
	*Resolver
}

func (r *stopExternalReferenceResolver) TargetActiveStop(ctx context.Context, obj *model.StopExternalReference) (*model.Stop, error) {
	return For(ctx).TargetStopsByStopID.Load(ctx, obj.ID)()
}
