package resolvers

import (
	"context"
	"errors"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/model"
)

// STOP TIME

type stopTimeResolver struct {
	*Resolver
}

func (r *stopTimeResolver) Stop(ctx context.Context, obj *model.StopTime) (*model.Stop, error) {
	return For(ctx).StopsByID.Load(atoi(obj.StopID))
}

func (r *stopTimeResolver) Trip(ctx context.Context, obj *model.StopTime) (*model.Trip, error) {
	return For(ctx).TripsByID.Load(atoi(obj.TripID))
}

func (r *stopTimeResolver) Arrival(ctx context.Context, obj *model.StopTime) (*model.StopTimeEvent, error) {
	// lookup timezone
	loc, ok := r.rtcm.StopTimezone(atoi(obj.StopID), "")
	if !ok {
		return nil, errors.New("timezone not available for stop 1")
	}
	// create departure
	a := model.StopTimeEvent{}
	if obj.RTStopTimeUpdate != nil && obj.RTStopTimeUpdate.Arrival != nil {
		a = fromSte(obj.RTStopTimeUpdate.Arrival, loc)
	}
	a.StopTimezone = loc.String()
	a.Scheduled = obj.ArrivalTime
	return &a, nil
}

func (r *stopTimeResolver) Departure(ctx context.Context, obj *model.StopTime) (*model.StopTimeEvent, error) {
	// lookup timezone
	loc, ok := r.rtcm.StopTimezone(atoi(obj.StopID), "")
	if !ok {
		return nil, errors.New("timezone not available for stop 2")
	}
	// create departure
	a := model.StopTimeEvent{}
	if obj.RTStopTimeUpdate != nil && obj.RTStopTimeUpdate.Departure != nil {
		a = fromSte(obj.RTStopTimeUpdate.Departure, loc)
	}
	a.StopTimezone = loc.String()
	a.Scheduled = obj.ArrivalTime
	return &a, nil
}

func fromSte(ste *pb.TripUpdate_StopTimeEvent, loc *time.Location) model.StopTimeEvent {
	if loc == nil {
		panic("loc is nil")
	}
	a := model.StopTimeEvent{
		StopTimezone: loc.String(),
	}
	if ste == nil {
		return a
	}
	if ste.Time != nil {
		t := time.Unix(ste.GetTime(), 0).UTC()
		lt := t.In(loc)
		a.Estimated = tl.NewWideTimeFromSeconds(lt.Hour()*3600 + lt.Minute()*60 + lt.Second())
		a.EstimatedUtc = tl.NewOTime(t)
	}
	if ste.Delay != nil {
		v := int(ste.GetDelay())
		a.Delay = &v
	}
	if ste.Uncertainty != nil {
		v := int(ste.GetUncertainty())
		a.Uncertainty = &v
	}
	return a
}
