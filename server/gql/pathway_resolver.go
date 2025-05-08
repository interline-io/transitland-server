package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// PATHWAYS

type pathwayResolver struct{ *Resolver }

func (r *pathwayResolver) FromStop(ctx context.Context, obj *model.Pathway) (*model.Stop, error) {
	return LoaderFor(ctx).StopsByID.Load(ctx, obj.FromStopID.Int())()
}

func (r *pathwayResolver) ToStop(ctx context.Context, obj *model.Pathway) (*model.Stop, error) {
	return LoaderFor(ctx).StopsByID.Load(ctx, obj.ToStopID.Int())()
}
