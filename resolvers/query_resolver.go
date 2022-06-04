package resolvers

import (
	"context"

	"github.com/interline-io/transitland-server/model"
)

// query root

type queryResolver struct{ *Resolver }

func (r *queryResolver) Agencies(ctx context.Context, limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	var cursor *model.Cursor
	if after != nil {
		cursor = &model.Cursor{ID: *after}
	}
	return r.finder.FindAgencies(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Routes(ctx context.Context, limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	var cursor *model.Cursor
	if after != nil {
		cursor = &model.Cursor{ID: *after}
	}
	return r.finder.FindRoutes(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Stops(ctx context.Context, limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	var cursor *model.Cursor
	if after != nil {
		cursor = &model.Cursor{ID: *after}
	}
	return r.finder.FindStops(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Trips(ctx context.Context, limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	var cursor *model.Cursor
	if after != nil {
		cursor = &model.Cursor{ID: *after}
	}
	return r.finder.FindTrips(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) FeedVersions(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	var cursor *model.Cursor
	if after != nil {
		cursor = &model.Cursor{ID: *after}
	}
	return r.finder.FindFeedVersions(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Feeds(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	var cursor *model.Cursor
	if after != nil {
		cursor = &model.Cursor{ID: *after}
	}
	return r.finder.FindFeeds(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Operators(ctx context.Context, limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	var cursor *model.Cursor
	if after != nil {
		cursor = &model.Cursor{ID: *after}
	}
	return r.finder.FindOperators(ctx, limit, cursor, ids, where)
}
