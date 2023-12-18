package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

func (r *queryResolver) Bikes(ctx context.Context, limit *int, where *model.GbfsBikeRequest) ([]*model.GbfsFreeBikeStatus, error) {
	return r.frs.GbfsFinder.FindBikes(ctx, checkLimit(limit), where)
}

func (r *queryResolver) Docks(ctx context.Context, limit *int, where *model.GbfsDockRequest) ([]*model.GbfsStationInformation, error) {
	return r.frs.GbfsFinder.FindDocks(ctx, checkLimit(limit), where)
}
