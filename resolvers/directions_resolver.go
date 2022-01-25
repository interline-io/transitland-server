package resolvers

import (
	"context"

	"github.com/interline-io/transitland-server/directions"
	"github.com/interline-io/transitland-server/model"
)

type directionsResolver struct{ *Resolver }

// Note: where is not a pointer
func (r *directionsResolver) Directions(ctx context.Context, where model.DirectionRequest) (*model.Directions, error) {
	return directions.HandleRequest("", where)
}
