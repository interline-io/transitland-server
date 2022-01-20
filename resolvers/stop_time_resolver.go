package resolvers

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// STOP TIME

type stopTimeResolver struct{ *Resolver }

func (r *stopTimeResolver) Stop(ctx context.Context, obj *model.StopTime) (*model.Stop, error) {
	return For(ctx).StopsByID.Load(atoi(obj.StopID))
}

func (r *stopTimeResolver) Trip(ctx context.Context, obj *model.StopTime) (*model.Trip, error) {
	return For(ctx).TripsByID.Load(atoi(obj.TripID))
}

func (r *stopTimeResolver) Rt(ctx context.Context, obj *model.StopTime) (*model.StopTimeUpdate, error) {
	// TODO
	return nil, nil
}
