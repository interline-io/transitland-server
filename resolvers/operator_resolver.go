package resolvers

import (
	"context"
	"encoding/json"

	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
)

// OPERATOR

type operatorResolver struct{ *Resolver }

func (r *operatorResolver) Agencies(ctx context.Context, obj *model.Operator) ([]*model.Agency, error) {
	a := obj.OnestopID.String
	return find.For(ctx).AgenciesByOnestopID.Load(model.AgencyParam{OnestopID: &a})
}

func (r *operatorResolver) Tags(ctx context.Context, obj *model.Operator) (interface{}, error) {
	return obj.Tags, nil
}

func (r *operatorResolver) AssociatedFeeds(ctx context.Context, obj *model.Operator) (interface{}, error) {
	return json.Marshal(obj.AssociatedFeeds)
}
