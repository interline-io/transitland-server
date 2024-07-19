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
	// create departure; fallback to RT departure if arrival is not present
	var ste *pb.TripUpdate_StopTimeEvent
	if obj.RTStopTimeUpdate != nil {
		if t := obj.RTStopTimeUpdate.Arrival; t != nil {
			ste = t
		} else if t := obj.RTStopTimeUpdate.Departure; t != nil {
			ste = t
		}
	}
	return fromSte(ste, obj.DepartureTime, obj.ServiceDate, loc), nil
}

func (r *stopTimeResolver) Departure(ctx context.Context, obj *model.StopTime) (*model.StopTimeEvent, error) {
	// lookup timezone
	loc, ok := model.ForContext(ctx).RTFinder.StopTimezone(atoi(obj.StopID), "")
	if !ok {
		return nil, errors.New("timezone not available for stop")
	}
	// create departure; fallback to RT arrival if departure is not present
	var ste *pb.TripUpdate_StopTimeEvent
	if obj.RTStopTimeUpdate != nil {
		if t := obj.RTStopTimeUpdate.Departure; t != nil {
			ste = t
		} else if t := obj.RTStopTimeUpdate.Arrival; t != nil {
			ste = t
		}
	}
	return fromSte(ste, obj.DepartureTime, obj.ServiceDate, loc), nil
}

func fromSte(ste *pb.TripUpdate_StopTimeEvent, sched tl.WideTime, serviceDate tl.Date, loc *time.Location) *model.StopTimeEvent {
	a := model.StopTimeEvent{
		StopTimezone: loc.String(),
		Scheduled:    &sched,
	}

	// Nothing else to do without timezone
	if loc == nil {
		return &a
	}

	// Apply local timezone
	// Hours, minutes, seconds in local scheduled time
	sd := serviceDate.Val
	h, m, s := sched.HMS()
	schedLocal := time.Date(sd.Year(), sd.Month(), sd.Day(), h, m, s, 0, loc)
	schedUtc := schedLocal.In(time.UTC)
	if serviceDate.Valid {
		a.ScheduledLocal = &schedLocal
		a.ScheduledUtc = &schedUtc
	}

	// Apply StopTimeEvent
	if ste != nil {
		if ste.Time != nil {
			// TODO: Should serviceDate override this, regardless?
			estUtc := time.Unix(ste.GetTime(), 0).UTC()
			estLocal := estUtc.In(loc)
			est := tt.NewWideTimeFromSeconds(estLocal.Hour()*3600 + estLocal.Minute()*60 + estLocal.Second())
			a.Estimated = &est
			a.EstimatedUtc = &estUtc
			a.EstimatedLocal = &estLocal
		} else if ste.Delay != nil && sched.Valid {
			// Create a local adjusted time
			est := tt.NewWideTimeFromSeconds(sched.Seconds + int(*ste.Delay))
			estUtc := schedUtc.Add(time.Second * time.Duration(est.Seconds))
			estLocal := estUtc.In(loc)
			a.Estimated = &est
			if serviceDate.Valid {
				a.EstimatedUtc = &estUtc
				a.EstimatedLocal = &estLocal
			}
		}
		if ste.Delay != nil {
			v := int(ste.GetDelay())
			a.Delay = &v
		}
		if ste.Uncertainty != nil {
			v := int(ste.GetUncertainty())
			a.Uncertainty = &v
		}
	}
	return &a
}
