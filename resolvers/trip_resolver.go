package resolvers

import (
	"context"
	"time"

	"github.com/interline-io/transitland-server/model"
)

// TRIP

type tripResolver struct{ *Resolver }

func (r *tripResolver) Route(ctx context.Context, obj *model.Trip) (*model.Route, error) {
	return For(ctx).RoutesByID.Load(atoi(obj.RouteID))
}

func (r *tripResolver) FeedVersion(ctx context.Context, obj *model.Trip) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(obj.FeedVersionID)
}

func (r *tripResolver) Shape(ctx context.Context, obj *model.Trip) (*model.Shape, error) {
	if !obj.ShapeID.Valid {
		return nil, nil
	}
	return For(ctx).ShapesByID.Load(obj.ShapeID.Int())
}

func (r *tripResolver) Calendar(ctx context.Context, obj *model.Trip) (*model.Calendar, error) {
	return For(ctx).CalendarsByID.Load(atoi(obj.ServiceID))
}

func (r *tripResolver) StopTimes(ctx context.Context, obj *model.Trip, limit *int) ([]*model.StopTime, error) {
	return For(ctx).StopTimesByTripID.Load(model.StopTimeParam{FeedVersionID: obj.FeedVersionID, TripID: obj.ID, Limit: limit})
}

func (r *tripResolver) Frequencies(ctx context.Context, obj *model.Trip, limit *int) ([]*model.Frequency, error) {
	return For(ctx).FrequenciesByTripID.Load(model.FrequencyParam{TripID: obj.ID, Limit: limit})
}

func (r *tripResolver) ScheduleRelationship(ctx context.Context, obj *model.Trip) (*model.ScheduleRelationship, error) {
	return nil, nil
}

func (r *tripResolver) Timestamp(ctx context.Context, obj *model.Trip) (*time.Time, error) {
	return nil, nil
}
