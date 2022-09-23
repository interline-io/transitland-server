package resolvers

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

func (r *queryResolver) Bikes(ctx context.Context, where model.GbfsBikeRequest) ([]*model.GbfsFreeBikeStatus, error) {
	if where.Near == nil {
		return nil, nil
	}
	return r.gbfsFinder.FindBikes(ctx, *where.Near)
}

func (r *queryResolver) Docks(ctx context.Context, where model.GbfsBikeRequest) ([]*model.GbfsStationInformation, error) {
	if where.Near == nil {
		return nil, nil
	}
	return nil, nil
	// return r.gbfsFinder.FindDocks(ctx, *where.Near)
}
