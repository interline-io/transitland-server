package gql

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// FEED

type feedResolver struct{ *Resolver }

func (r *feedResolver) Cursor(ctx context.Context, obj *model.Feed) (*model.Cursor, error) {
	c := model.NewCursor(0, obj.ID)
	return &c, nil
}

func (r *feedResolver) FeedState(ctx context.Context, obj *model.Feed) (*model.FeedState, error) {
	return LoaderFor(ctx).FeedStatesByFeedID.Load(ctx, obj.ID)()
}

func (r *feedResolver) FeedVersions(ctx context.Context, obj *model.Feed, limit *int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	return LoaderFor(ctx).FeedVersionsByFeedID.Load(ctx, model.FeedVersionParam{FeedID: obj.ID, Limit: checkLimit(limit), Where: where})()
}

func (r *feedResolver) License(ctx context.Context, obj *model.Feed) (*model.FeedLicense, error) {
	return &model.FeedLicense{FeedLicense: obj.License}, nil
}

func (r *feedResolver) Languages(ctx context.Context, obj *model.Feed) ([]string, error) {
	return obj.Languages, nil
}

func (r *feedResolver) Urls(ctx context.Context, obj *model.Feed) (*model.FeedUrls, error) {
	return &model.FeedUrls{FeedUrls: obj.URLs}, nil
}

func (r *feedResolver) AssociatedOperators(ctx context.Context, obj *model.Feed) ([]*model.Operator, error) {
	return LoaderFor(ctx).OperatorsByFeedID.Load(ctx, model.OperatorParam{FeedID: obj.ID})()
}

func (r *feedResolver) Authorization(ctx context.Context, obj *model.Feed) (*model.FeedAuthorization, error) {
	return &model.FeedAuthorization{FeedAuthorization: obj.Authorization}, nil
}

func (r *feedResolver) FeedFetches(ctx context.Context, obj *model.Feed, limit *int, where *model.FeedFetchFilter) ([]*model.FeedFetch, error) {
	return LoaderFor(ctx).FeedFetchesByFeedIDs.Load(ctx, model.FeedFetchParam{FeedID: obj.ID, Limit: checkLimit(limit), Where: where})()
}

func (r *feedResolver) Spec(ctx context.Context, obj *model.Feed) (*model.FeedSpecTypes, error) {
	var s model.FeedSpecTypes
	s2 := s.FromDBString(obj.Spec)
	return s2, nil
}

// FEED STATE

type feedStateResolver struct{ *Resolver }
