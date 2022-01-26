package resolvers

import (
	"context"
	"errors"
	"fmt"
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
	// Merge scheduled stop times with rt stop times
	// TODO: handle StopTimeFilter in RT
	_, ok := r.rtcm.StopTimezone(obj.ID, obj.StopTimezone)
	if !ok {
		return nil, errors.New("timezone not available for stop 0")
	}
	// Handle scheduled trips; these can be matched on trip_id or (route_id,direction_id,...)
	for _, st := range sts {
		a, aok := r.rtcm.FeedVersionOnestopID(st.FeedVersionID)
		b, bok := r.rtcm.TripGTFSTripID(atoi(st.TripID))
		if !aok || !bok {
			fmt.Println("could not get onestop id or trip gtfs id, skipping")
			continue
		}
		rtTrip, rtok := r.rtcm.GetTrip(a, b)
		if !rtok {
			fmt.Println("no trip for:", a, b)
			continue
		}
		for _, ste := range rtTrip.StopTimeUpdate {
			// Must match on StopSequence
			// TODO: allow matching on stop_id if stop_sequence is not provided
			if int(ste.GetStopSequence()) == st.StopSequence {
				st.RTStopTimeUpdate = ste
				break
			}
		}
	}
	// Handle added trips; these must specify stop_id in StopTimeUpdates
	fmt.Println("looking for added rt for stop:", obj.FeedOnestopID, obj.StopID)
	for _, rtTrip := range r.rtcm.GetAddedTripsForStop(obj.FeedOnestopID, obj.StopID) {
		for _, stu := range rtTrip.StopTimeUpdate {
			if stu.GetStopId() != obj.StopID {
				continue
			}
			rtst := &model.StopTime{}
			rtst.RTStopTimeUpdate = stu
			rtst.FeedVersionID = obj.FeedVersionID
			// create a new StopTime
			rtst.TripID = "0"
			rtst.StopID = obj.StopID
			rtst.StopSequence = int(stu.GetStopSequence())
			sts = append(sts, rtst)
		}
	}
	return sts, nil
}

func (r *stopResolver) Alerts(ctx context.Context, obj *model.Stop) ([]*model.Alert, error) {
	// TODO
	return nil, nil
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
