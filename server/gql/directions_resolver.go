package gql

import (
	"context"

	"github.com/interline-io/transitland-server/finders/directions"
	"github.com/interline-io/transitland-server/model"
)

type directionsResolver struct{ *Resolver }

// Note: where is not a pointer
func (r *directionsResolver) Directions(ctx context.Context, where model.DirectionRequest) (*model.Directions, error) {
	return directions.HandleRequest(ctx, "", where)
}
