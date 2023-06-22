package gql

import (
	"context"
	"time"

	"github.com/interline-io/transitland-server/model"
)

// TRIP

type tripResolver struct{ *Resolver }

func (r *tripResolver) Cursor(ctx context.Context, obj *model.Trip) (*model.Cursor, error) {
	c := model.NewCursor(obj.FeedVersionID, obj.ID)
	return &c, nil
}

func (r *tripResolver) Route(ctx context.Context, obj *model.Trip) (*model.Route, error) {
	return For(ctx).RoutesByID.Load(ctx, atoi(obj.RouteID))()
}

func (r *tripResolver) FeedVersion(ctx context.Context, obj *model.Trip) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(ctx, obj.FeedVersionID)()
}

func (r *tripResolver) Shape(ctx context.Context, obj *model.Trip) (*model.Shape, error) {
	if !obj.ShapeID.Valid {
		return nil, nil
	}
	return For(ctx).ShapesByID.Load(ctx, obj.ShapeID.Int())()
}

func (r *tripResolver) Calendar(ctx context.Context, obj *model.Trip) (*model.Calendar, error) {
	return For(ctx).CalendarsByID.Load(ctx, atoi(obj.ServiceID))()
}

func (r *tripResolver) StopTimes(ctx context.Context, obj *model.Trip, limit *int, where *model.TripStopTimeFilter) ([]*model.StopTime, error) {
	return For(ctx).StopTimesByTripID.Load(ctx, model.TripStopTimeParam{FeedVersionID: obj.FeedVersionID, TripID: obj.ID, Limit: limit, Where: where})()
}

func (r *tripResolver) Frequencies(ctx context.Context, obj *model.Trip, limit *int) ([]*model.Frequency, error) {
	return For(ctx).FrequenciesByTripID.Load(ctx, model.FrequencyParam{TripID: obj.ID, Limit: limit})()
}

func (r *tripResolver) ScheduleRelationship(ctx context.Context, obj *model.Trip) (*model.ScheduleRelationship, error) {
	msr := model.ScheduleRelationshipScheduled
	if rtt := r.rtfinder.FindTrip(obj); rtt != nil {
		sr := rtt.GetTrip().GetScheduleRelationship().String()
		switch sr {
		case "SCHEDULED":
			msr = model.ScheduleRelationshipScheduled
		case "ADDED":
			msr = model.ScheduleRelationshipAdded
		case "CANCELED":
			msr = model.ScheduleRelationshipCanceled
		case "UNSCHEDULED":
			msr = model.ScheduleRelationshipUnscheduled
		default:
			return nil, nil
		}
	}
	return &msr, nil
}

func (r *tripResolver) Timestamp(ctx context.Context, obj *model.Trip) (*time.Time, error) {
	if rtt := r.rtfinder.FindTrip(obj); rtt != nil {
		t := time.Unix(int64(rtt.GetTimestamp()), 0).In(time.UTC)
		return &t, nil
	}
	return nil, nil
}

func (r *tripResolver) Alerts(ctx context.Context, obj *model.Trip, active *bool, limit *int) ([]*model.Alert, error) {
	rtAlerts := r.rtfinder.FindAlertsForTrip(obj, limit, active)
	return rtAlerts, nil
}
