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

func (r *stopTimeResolver) ScheduleRelationship(ctx context.Context, obj *model.StopTime) (*model.ScheduleRelationship, error) {
	stu := obj.RTStopTimeUpdate
	// Use StopTimeUpdate ScheduleRelationship value if explicitly provided
	// if stu != nil && stu.StopTimeUpdate != nil && stu.StopTimeUpdate.ScheduleRelationship != nil {
	// 	return convertScheduleRelationship(stu.StopTimeUpdate.ScheduleRelationship.String()), nil
	// }
	// Otherwise, try defaulting to TripUpdate ScheduleRelationship value
	if stu != nil && stu.TripUpdate != nil && stu.TripUpdate.Trip != nil && stu.TripUpdate.Trip.ScheduleRelationship != nil {
		return convertScheduleRelationship(stu.TripUpdate.Trip.ScheduleRelationship.String()), nil
	}
	// Otherwise, if ANY RT data is present (e.g. a propagated delay), default to SCHEDULED
	if stu != nil && stu.StopTimeUpdate != nil {
		return ptr(model.ScheduleRelationshipScheduled), nil
	}
	// Otherwise, default to STATIC
	return ptr(model.ScheduleRelationshipStatic), nil
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
	// Lookup timezone
	loc, ok := model.ForContext(ctx).RTFinder.StopTimezone(atoi(obj.StopID), "")
	if !ok {
		return nil, errors.New("timezone not available for stop")
	}
	// Create arrival; fallback to RT departure if arrival is not present
	var ste *pb.TripUpdate_StopTimeEvent
	var delay *int32
	if rtStu := obj.RTStopTimeUpdate; rtStu != nil {
		delay = rtStu.LastDelay
		if stu := rtStu.StopTimeUpdate; stu == nil {
		} else if stu.Arrival != nil {
			ste = stu.Arrival
		} else if stu.Departure != nil {
			ste = stu.Departure
		}
	}
	return fromSte(ste, delay, obj.DepartureTime, obj.ServiceDate, loc), nil
}

func (r *stopTimeResolver) Departure(ctx context.Context, obj *model.StopTime) (*model.StopTimeEvent, error) {
	// Lookup timezone
	loc, ok := model.ForContext(ctx).RTFinder.StopTimezone(atoi(obj.StopID), "")
	if !ok {
		return nil, errors.New("timezone not available for stop")
	}
	// Create departure; fallback to RT arrival if departure is not present
	var ste *pb.TripUpdate_StopTimeEvent
	var delay *int32
	if rtStu := obj.RTStopTimeUpdate; rtStu != nil {
		delay = rtStu.LastDelay
		if stu := rtStu.StopTimeUpdate; stu == nil {
		} else if stu.Departure != nil {
			ste = stu.Departure
		} else if stu.Arrival != nil {
			ste = stu.Arrival
		}
	}
	return fromSte(ste, delay, obj.DepartureTime, obj.ServiceDate, loc), nil
}

func fromSte(ste *pb.TripUpdate_StopTimeEvent, lastDelay *int32, sched tl.WideTime, serviceDate tl.Date, loc *time.Location) *model.StopTimeEvent {
	a := model.StopTimeEvent{
		StopTimezone: loc.String(),
		Scheduled:    &sched,
	}

	// Nothing else to do without timezone or valid schedule
	if loc == nil {
		return &a
	}

	// Apply local timezone
	// Hours, minutes, seconds in local scheduled time
	sd := serviceDate.Val
	h, m, s := sched.HMS()
	schedLocal := time.Date(sd.Year(), sd.Month(), sd.Day(), h, m, s, 0, loc)
	schedUtc := schedLocal.In(time.UTC)
	if serviceDate.Valid && sched.Valid {
		a.ScheduledUtc = &schedUtc
		a.ScheduledUnix = ptr(int(schedUtc.Unix()))
		a.ScheduledLocal = &schedLocal
	}

	// Check to apply lastDelay
	if ste == nil && lastDelay != nil {
		// Create a time based on propagated delay
		est := tt.NewWideTimeFromSeconds(sched.Seconds + int(*lastDelay))
		estUtc := schedUtc.Add(time.Second * time.Duration(int(*lastDelay)))
		estLocal := estUtc.In(loc)
		a.Estimated = ptr(est)
		if serviceDate.Valid {
			a.EstimatedUtc = ptr(estUtc)
			a.EstimatedUnix = ptr(int(estUtc.Unix()))
			a.EstimatedLocal = ptr(estLocal)
		}
	}

	// No ste, nothing else to do
	if ste == nil {
		return &a
	}

	// Apply StopTimeEvent
	if ste.Time != nil {
		// Set est based on rt time
		// TODO: Should serviceDate override this, regardless?
		estUtc := time.Unix(ste.GetTime(), 0).UTC()
		estLocal := estUtc.In(loc)
		est := tt.NewWideTimeFromSeconds(estLocal.Hour()*3600 + estLocal.Minute()*60 + estLocal.Second())
		a.TimeUtc = &estUtc // raw RT
		a.Estimated = ptr(est)
		a.EstimatedUtc = ptr(estUtc)
		a.EstimatedUnix = ptr(int(estUtc.Unix()))
		a.EstimatedLocal = ptr(estLocal)
	} else if ste.Delay != nil && sched.Valid {
		// Create a time based on STE delay
		est := tt.NewWideTimeFromSeconds(sched.Seconds + int(*ste.Delay))
		estUtc := schedUtc.Add(time.Second * time.Duration(int(*ste.Delay)))
		estLocal := estUtc.In(loc)
		a.Estimated = ptr(est)
		if serviceDate.Valid {
			a.EstimatedUtc = ptr(estUtc)
			a.EstimatedUnix = ptr(int(estUtc.Unix()))
			a.EstimatedLocal = ptr(estLocal)
		}
	} else {
		// Could not est time
	}
	// Only pass through actual delay
	if ste.Delay != nil {
		a.Delay = ptr(int(ste.GetDelay()))
	}
	if ste.Uncertainty != nil {
		a.Uncertainty = ptr(int(ste.GetUncertainty()))
	}
	return &a
}
