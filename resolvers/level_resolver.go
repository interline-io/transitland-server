package resolvers

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

type levelResolver struct {
	*Resolver
}

func (r *levelResolver) Stops(ctx context.Context, obj *model.Level) ([]*model.Stop, error) {
	return For(ctx).StopsByLevelID.Load(ctx, model.StopParam{LevelID: obj.ID})()
}
