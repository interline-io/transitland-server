package resolvers

import (
	"context"
	"encoding/json"

	"github.com/interline-io/transitland-server/model"
)

// OPERATOR

type operatorResolver struct{ *Resolver }

func (r *operatorResolver) Cursor(ctx context.Context, obj *model.Operator) (*model.Cursor, error) {
	return &model.Cursor{ID: obj.ID}, nil
}

func (r *operatorResolver) Agencies(ctx context.Context, obj *model.Operator) ([]*model.Agency, error) {
	a := obj.OnestopID.String
	return For(ctx).AgenciesByOnestopID.Load(model.AgencyParam{OnestopID: &a})
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

func (r *operatorResolver) Feeds(ctx context.Context, obj *model.Operator, limit *int, where *model.FeedFilter) ([]*model.Feed, error) {
	return For(ctx).FeedsByOperatorOnestopID.Load(model.FeedParam{OperatorOnestopID: obj.OnestopID.String, Where: where, Limit: limit})
}
