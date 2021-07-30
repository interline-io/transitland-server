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
	a, err := json.Marshal(obj.AssociatedFeeds)
	return json.RawMessage(a), err
}

func (r *operatorResolver) Generated(ctx context.Context, obj *model.Operator) (bool, error) {
	if obj.Generated {
		return true, nil
	}
	return false, nil
}
