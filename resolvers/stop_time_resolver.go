package resolvers

import (
	"context"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
)

// STOP TIME

type stopTimeResolver struct{ *Resolver }

func (r *stopTimeResolver) Stop(ctx context.Context, obj *model.StopTime) (*model.Stop, error) {
	return find.For(ctx).StopsByID.Load(atoi(obj.StopID))
}

func (r *stopTimeResolver) Trip(ctx context.Context, obj *model.StopTime) (*model.Trip, error) {
	return find.For(ctx).TripsByID.Load(atoi(obj.TripID))
}

func (r *stopTimeResolver) Rt(ctx context.Context, obj *model.StopTime) (*model.StopTimeUpdate, error) {
	rt := model.StopTimeUpdate{}
	if obj.ArrivalTime.Valid {
		rt.ArrivalTime = tl.NewWideTimeFromSeconds(obj.ArrivalTime.Seconds + 32)
	}
	if obj.DepartureTime.Valid {
		rt.DepartureTime = tl.NewWideTimeFromSeconds(obj.DepartureTime.Seconds + 32)
	}
	return &rt, nil
}
