package resolvers

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/directions"
	"github.com/interline-io/transitland-server/model"
)

// STOP

type stopResolver struct {
	*Resolver
}

func (r *stopResolver) FeedVersion(ctx context.Context, obj *model.Stop) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(obj.FeedVersionID)
}

func (r *stopResolver) Level(ctx context.Context, obj *model.Stop) (*model.Level, error) {
	if !obj.LevelID.Valid {
		return nil, nil
	}
	return For(ctx).LevelsByID.Load(atoi(obj.LevelID.Key))
}

func (r *stopResolver) Parent(ctx context.Context, obj *model.Stop) (*model.Stop, error) {
	if !obj.ParentStation.Valid {
		return nil, nil
	}
	return For(ctx).StopsByID.Load(atoi(obj.ParentStation.Key))
}

func (r *stopResolver) Children(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Stop, error) {
	return For(ctx).StopsByParentStopID.Load(model.StopParam{ParentStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) RouteStops(ctx context.Context, obj *model.Stop, limit *int) ([]*model.RouteStop, error) {
	return For(ctx).RouteStopsByStopID.Load(model.RouteStopParam{StopID: obj.ID, Limit: limit})
}

func (r *stopResolver) PathwaysFromStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByFromStopID.Load(model.PathwayParam{FromStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) PathwaysToStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByToStopID.Load(model.PathwayParam{ToStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) StopTimes(ctx context.Context, obj *model.Stop, limit *int, where *model.StopTimeFilter) ([]*model.StopTime, error) {
	// Further processing of the StopTimeFilter
	if where != nil {
		// Convert where.Next into departure date and time window
		if where.Next != nil {
			loc, ok := r.rtfinder.StopTimezone(obj.ID, obj.StopTimezone)
			if !ok {
				return nil, errors.New("timezone not available for stop")
			}
			serviceDate := r.cfg.Clock.Now().In(loc)
			st, et := 0, 0
			st = serviceDate.Hour()*3600 + serviceDate.Minute()*60 + serviceDate.Second()
			et = st + *where.Next
			sd2 := tl.Date{Valid: true, Time: serviceDate}
			where.ServiceDate = &sd2
			where.StartTime = &st
			where.EndTime = &et
			where.Next = nil
		}
		// Check if service date is outside the window for this feed version
		if where.ServiceDate != nil && (where.UseExactDate == nil || !*where.UseExactDate) {
			sl, ok := r.fvslCache.Get(obj.FeedVersionID)
			if !ok {
				return nil, errors.New("service level information not available for feed version")
			}
			s := where.ServiceDate.Time
			if s.Before(sl.StartDate) || s.After(sl.EndDate) {
				dow := int(s.Weekday()) - 1
				if dow < 0 {
					dow = 6
				}
				where.ServiceDate.Time = sl.BestWeek.AddDate(0, 0, dow)
				// fmt.Println(
				// 	"requested day:", s, s.Weekday(),
				// 	"window start:", sl.StartDate,
				// 	"window end:", sl.EndDate,
				// 	"best week:", sl.BestWeek, sl.BestWeek.Weekday(),
				// 	"switching to:", where.ServiceDate.Time, where.ServiceDate.Time.Weekday(),
				// )
			}
		}
	}
	//
	sts, err := For(ctx).StopTimesByStopID.Load(model.StopTimeParam{
		StopID:        obj.ID,
		FeedVersionID: obj.FeedVersionID,
		Limit:         limit,
		Where:         where,
	})
	if err != nil {
		return nil, err
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

func (r *stopResolver) Alerts(ctx context.Context, obj *model.Stop) ([]*model.Alert, error) {
	rtAlerts := r.rtfinder.FindAlertsForStop(obj)
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
