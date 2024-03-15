package gql

import (
	"context"
	"errors"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/model"
)

// STOP TIME

type stopTimeResolver struct {
	*Resolver
}

func (r *stopTimeResolver) Stop(ctx context.Context, obj *model.StopTime) (*model.Stop, error) {
	return For(ctx).StopsByID.Load(ctx, atoi(obj.StopID))()
}

func (r *stopTimeResolver) Trip(ctx context.Context, obj *model.StopTime) (*model.Trip, error) {
	if obj.TripID == "0" && obj.RTTripID != "" {
		t := model.Trip{}
		t.FeedVersionID = obj.FeedVersionID
		t.TripID = obj.RTTripID
		a, err := model.ForContext(ctx).RTFinder.MakeTrip(&t)
		return a, err
	}
	return For(ctx).TripsByID.Load(ctx, atoi(obj.TripID))()
}

func (r *stopTimeResolver) Arrival(ctx context.Context, obj *model.StopTime) (*model.StopTimeEvent, error) {
	// lookup timezone
	loc, ok := model.ForContext(ctx).RTFinder.StopTimezone(atoi(obj.StopID), "")
	if !ok {
		return nil, errors.New("timezone not available for stop")
	}
	// create arrival; fallback to RT departure if arrival is not present
	a := model.StopTimeEvent{}
	if obj.RTStopTimeUpdate != nil {
		if obj.RTStopTimeUpdate.Arrival != nil {
			a = fromSte(obj.RTStopTimeUpdate.Arrival, obj.ArrivalTime, loc)
		} else if obj.RTStopTimeUpdate.Departure != nil {
			a = fromSte(obj.RTStopTimeUpdate.Departure, obj.DepartureTime, loc)
		}
	}
	a.StopTimezone = loc.String()
	tt := obj.ArrivalTime
	a.Scheduled = &tt
	return &a, nil
}

func (r *stopTimeResolver) Departure(ctx context.Context, obj *model.StopTime) (*model.StopTimeEvent, error) {
	// lookup timezone
	loc, ok := model.ForContext(ctx).RTFinder.StopTimezone(atoi(obj.StopID), "")
	if !ok {
		return nil, errors.New("timezone not available for stop")
	}
	// create departure; fallback to RT arrival if departure is not present
	a := model.StopTimeEvent{}
	if obj.RTStopTimeUpdate != nil {
		if obj.RTStopTimeUpdate.Departure != nil {
			a = fromSte(obj.RTStopTimeUpdate.Departure, obj.DepartureTime, loc)
		} else if obj.RTStopTimeUpdate.Arrival != nil {
			a = fromSte(obj.RTStopTimeUpdate.Arrival, obj.ArrivalTime, loc)
		}
	}
	a.StopTimezone = loc.String()
	tt := obj.ArrivalTime
	a.Scheduled = &tt
	return &a, nil
}

func fromSte(ste *pb.TripUpdate_StopTimeEvent, sched tl.WideTime, loc *time.Location) model.StopTimeEvent {
	a := model.StopTimeEvent{
		StopTimezone: loc.String(),
	}
	if ste == nil {
		return a
	}
	if ste.Time != nil {
		t := time.Unix(ste.GetTime(), 0).UTC()
		lt := t.In(loc)
		et := tt.NewWideTimeFromSeconds(lt.Hour()*3600 + lt.Minute()*60 + lt.Second())
		ett := t
		a.Estimated = &et
		a.EstimatedUtc = &ett
	} else if ste.Delay != nil && sched.Valid {
		// Create a local adjusted time
		// Note: can't create an EstimatedUtc value because we'd have to guess the local date
		et := tt.NewWideTimeFromSeconds(sched.Seconds + int(*ste.Delay))
		a.Estimated = &et
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
