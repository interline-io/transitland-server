package resolvers

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/directions"
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
	return For(ctx).LevelsByID.Load(ctx, atoi(obj.LevelID.Val))()
}

func (r *stopResolver) Parent(ctx context.Context, obj *model.Stop) (*model.Stop, error) {
	if !obj.ParentStation.Valid {
		return nil, nil
	}
	return For(ctx).StopsByID.Load(ctx, atoi(obj.ParentStation.Val))()
}

func (r *stopResolver) Children(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Stop, error) {
	return For(ctx).StopsByParentStopID.Load(ctx, model.StopParam{ParentStopID: obj.ID, Limit: limit})()
}

func (r *stopResolver) RouteStops(ctx context.Context, obj *model.Stop, limit *int) ([]*model.RouteStop, error) {
	return For(ctx).RouteStopsByStopID.Load(ctx, model.RouteStopParam{StopID: obj.ID, Limit: limit})()
}

func (r *stopResolver) PathwaysFromStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByFromStopID.Load(ctx, model.PathwayParam{FromStopID: obj.ID, Limit: limit})()
}

func (r *stopResolver) PathwaysToStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByToStopID.Load(ctx, model.PathwayParam{ToStopID: obj.ID, Limit: limit})()
}

func (r *stopResolver) ExternalReference(ctx context.Context, obj *model.Stop) (*model.StopExternalReference, error) {
	return For(ctx).StopExternalReferencesByStopID.Load(ctx, obj.ID)()
}

func (r *stopResolver) Observations(ctx context.Context, obj *model.Stop, limit *int, where *model.StopObservationFilter) ([]*model.StopObservation, error) {
	return For(ctx).StopObservationsByStopID.Load(ctx, model.StopObservationParam{StopID: obj.ID, Where: where, Limit: limit})()
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
	// Further processing of the StopTimeFilter
	if where != nil {
		// Convert where.Next into departure date and time window
		if where.Next != nil {
			loc, ok := r.rtfinder.StopTimezone(obj.ID, obj.StopTimezone)
			if !ok {
				return nil, errors.New("timezone not available for stop")
			}
			serviceDate := time.Now().In(loc)
			if r.cfg.Clock != nil {
				serviceDate = r.cfg.Clock.Now().In(loc)
			}
			st, et := 0, 0
			st = serviceDate.Hour()*3600 + serviceDate.Minute()*60 + serviceDate.Second()
			et = st + *where.Next
			sd2 := tl.Date{Valid: true, Val: serviceDate}
			where.ServiceDate = &sd2
			where.StartTime = &st
			where.EndTime = &et
			where.Next = nil
		}
		// Check if service date is outside the window for this feed version
		if where.ServiceDate != nil && (where.UseServiceWindow != nil && *where.UseServiceWindow) {
			sl, ok := r.fvslCache.Get(obj.FeedVersionID)
			if !ok {
				return nil, errors.New("service level information not available for feed version")
			}
			s := where.ServiceDate.Val
			if s.Before(sl.StartDate) || s.After(sl.EndDate) {
				dow := int(s.Weekday()) - 1
				if dow < 0 {
					dow = 6
				}
				where.ServiceDate.Val = sl.BestWeek.AddDate(0, 0, dow)
				// fmt.Println(
				// 	"service window, requested day:", s, s.Weekday(),
				// 	"window start:", sl.StartDate,
				// 	"window end:", sl.EndDate,
				// 	"best week:", sl.BestWeek, sl.BestWeek.Weekday(),
				// 	"switching to:", where.ServiceDate.Time, where.ServiceDate.Time.Weekday(),
				// )
			}
		}
	}
	//
	sts, err := (For(ctx).StopTimesByStopID.Load(ctx, model.StopTimeParam{
		StopID:        obj.ID,
		FeedVersionID: obj.FeedVersionID,
		Limit:         limit,
		Where:         where,
	})())
	if err != nil {
		return nil, err
	}

	// Add service date used in query, if any
	if where != nil && where.ServiceDate != nil {
		for _, st := range sts {
			st.ServiceDate = tt.NewDate(where.ServiceDate.Val)
		}
	}

	// Merge scheduled stop times with rt stop times
	// TODO: handle StopTimeFilter in RT
	// Handle scheduled trips; these can be matched on trip_id or (route_id,direction_id,...)
	for _, st := range sts {
		ft := model.Trip{}
		ft.FeedVersionID = obj.FeedVersionID
		ft.TripID, _ = r.rtfinder.GetGtfsTripID(atoi(st.TripID)) // TODO!
		if ste, ok := r.rtfinder.FindStopTimeUpdate(&ft, st); ok {
			st.RTStopTimeUpdate = ste
		}
	}
	// Handle added trips; these must specify stop_id in StopTimeUpdates
	for _, rtTrip := range r.rtfinder.GetAddedTripsForStop(obj) {
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
		a := sts[i].DepartureTime.Seconds
		b := sts[j].DepartureTime.Seconds
		return a < b
	})
	return sts, nil
}

func (r *stopResolver) Alerts(ctx context.Context, obj *model.Stop, active *bool, limit *int) ([]*model.Alert, error) {
	rtAlerts := r.rtfinder.FindAlertsForStop(obj, limit, active)
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
	nearbyStops, err := r.finder.FindStops(ctx, limit, nil, nil, nil, &model.StopFilter{Near: &model.PointRadius{Lon: c[0], Lat: c[1], Radius: checkFloat(radius, 0, 10_000)}})
	return nearbyStops, err
}

func checkFloat(v *float64, min float64, max float64) float64 {
	if v == nil || *v < min {
		return min
	} else if *v > max {
		return max
	}
	return *v
}

//////////

type stopExternalReferenceResolver struct {
	*Resolver
}

func (r *stopExternalReferenceResolver) TargetActiveStop(ctx context.Context, obj *model.StopExternalReference) (*model.Stop, error) {
	return For(ctx).TargetStopsByStopID.Load(ctx, obj.ID)()
}
