package gql

import (
	"context"

	"github.com/interline-io/transitland-server/internal/meters"
	"github.com/interline-io/transitland-server/model"
)

// query root

type queryResolver struct{ *Resolver }

func (r *queryResolver) Agencies(ctx context.Context, limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	addMetric(ctx, "agencies")
	return r.finder.FindAgencies(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Routes(ctx context.Context, limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	addMetric(ctx, "routes")
	return r.finder.FindRoutes(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Stops(ctx context.Context, limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	addMetric(ctx, "stops")
	return r.finder.FindStops(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Trips(ctx context.Context, limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	addMetric(ctx, "trips")
	return r.finder.FindTrips(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) FeedVersions(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	addMetric(ctx, "feedVersions")
	return r.finder.FindFeedVersions(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Feeds(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	addMetric(ctx, "feeds")
	return r.finder.FindFeeds(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Operators(ctx context.Context, limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	addMetric(ctx, "operators")
	return r.finder.FindOperators(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Places(ctx context.Context, limit *int, after *int, level *model.PlaceAggregationLevel, where *model.PlaceFilter) ([]*model.Place, error) {
	return r.finder.FindPlaces(ctx, checkLimit(limit), checkCursor(after), nil, level, where)
}

func addMetric(ctx context.Context, resolverName string) {
	if apiMeter := meters.ForContext(ctx); apiMeter != nil {
		apiMeter.AddDimension("graphql", "resolver", resolverName)
	}
}

func checkCursor(after *int) *model.Cursor {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	return cursor
}
