package resolvers

import (
	"context"
	"errors"
	"time"

	"github.com/interline-io/transitland-server/directions"
	"github.com/interline-io/transitland-server/model"
)

// STOP

type stopResolver struct{ *Resolver }

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
	sts, err := For(ctx).StopTimesByStopID.Load(model.StopTimeParam{FeedVersionID: obj.FeedVersionID, StopID: obj.ID, Limit: limit, Where: where})
	if err != nil {
		return nil, err
	}
	_, ok := r.rtcm.StopTimezone(obj.ID, obj.StopTimezone)
	if !ok {
		return nil, errors.New("timezone not available for stop")
	}

	// Merge scheduled stop times with rt stop times
	// TODO: handle StopTimeFilter in RT
	// Handle scheduled trips; these can be matched on trip_id or (route_id,direction_id,...)
	for _, st := range sts {
		ft := model.Trip{}
		ft.FeedVersionID = obj.FeedVersionID
		ft.TripID, _ = r.rtcm.GetGtfsTripID(atoi(st.TripID)) // TODO!
		if ste, ok := r.rtcm.FindStopTimeUpdate(&ft, st); ok {
			st.RTStopTimeUpdate = ste
		}
	}
	// Handle added trips; these must specify stop_id in StopTimeUpdates
	for _, rtTrip := range r.rtcm.GetAddedTripsForStop(obj) {
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
			rtst.StopID = obj.StopID
			rtst.StopSequence = int(stu.GetStopSequence())
			sts = append(sts, rtst)
		}
	}
	return sts, nil
}

func (r *stopResolver) Alerts(ctx context.Context, obj *model.Stop) ([]*model.Alert, error) {
	rtAlerts := r.rtcm.FindAlertsForStop(obj)
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
	return directions.HandleRequest("", p)
}
