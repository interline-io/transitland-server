package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

func (r *queryResolver) Vehicles(ctx context.Context, limit *int, where *model.VehiclePositionRequest) ([]*model.VehiclePosition, error) {
	return r.rtfinder.FindVehiclePositions(ctx, limit, where)
}
