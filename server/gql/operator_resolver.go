package gql

import (
	"context"
	"encoding/json"

	"github.com/interline-io/transitland-server/model"
)

// OPERATOR

type operatorResolver struct{ *Resolver }

func (r *operatorResolver) Cursor(ctx context.Context, obj *model.Operator) (*model.Cursor, error) {
	c := model.NewCursor(0, obj.ID)
	return &c, nil
}

func (r *operatorResolver) Agencies(ctx context.Context, obj *model.Operator) ([]*model.Agency, error) {
	return LoaderFor(ctx).AgenciesByOnestopID.Load(ctx, model.AgencyParam{OnestopID: &obj.OnestopID.Val})()
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
	return LoaderFor(ctx).FeedsByOperatorOnestopID.Load(ctx, model.FeedParam{OperatorOnestopID: obj.OnestopID.Val, Where: where, Limit: checkLimit(limit)})()
}
