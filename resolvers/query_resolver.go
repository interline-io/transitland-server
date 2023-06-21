package resolvers

import (
	"context"
	"fmt"

	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/authz"
	"github.com/interline-io/transitland-server/internal/meters"
	"github.com/interline-io/transitland-server/model"
)

// query root

type queryResolver struct{ *Resolver }

func (r *queryResolver) Agencies(ctx context.Context, limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	addMetric(ctx, "agencies")
	return r.finder.FindAgencies(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Routes(ctx context.Context, limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	addMetric(ctx, "routes")
	return r.finder.FindRoutes(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Stops(ctx context.Context, limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	addMetric(ctx, "stops")
	return r.finder.FindStops(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Trips(ctx context.Context, limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	addMetric(ctx, "trips")
	return r.finder.FindTrips(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) FeedVersions(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	addMetric(ctx, "feedVersions")
	return r.finder.FindFeedVersions(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Feeds(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	addMetric(ctx, "feeds")
	return r.finder.FindFeeds(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Operators(ctx context.Context, limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	addMetric(ctx, "operators")
	return r.finder.FindOperators(ctx, limit, cursor, ids, where)
}

func (r *queryResolver) Places(ctx context.Context, limit *int, after *int, level *model.PlaceAggregationLevel, where *model.PlaceFilter) ([]*model.Place, error) {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	return r.finder.FindPlaces(ctx, limit, cursor, nil, level, where)
}

func addMetric(ctx context.Context, resolverName string) {
	if apiMeter := meters.ForContext(ctx); apiMeter != nil {
		apiMeter.AddDimension("graphql", "resolver", resolverName)
	}
}

func checkActive(ctx context.Context, ids []int, checker *authz.Checker) (*model.ActiveCheck, error) {
	active := &model.ActiveCheck{IDs: ids}
	user := auth.ForContext(ctx)
	if user == nil {
		return active, nil
	}
	if checker != nil {
		active.CheckAllowed = true
		okFeeds, err := checker.FeedList(ctx, &authz.FeedListRequest{})
		if err != nil {
			return nil, err
		}
		for _, feed := range okFeeds.Feeds {
			active.AllowedFeeds = append(active.AllowedFeeds, int(feed.Id))
		}
		okFvids, err := checker.FeedVersionList(ctx, &authz.FeedVersionListRequest{})
		if err != nil {
			return nil, err
		}
		for _, fv := range okFvids.FeedVersions {
			active.AllowedFeedVersions = append(active.AllowedFeedVersions, int(fv.Id))
		}
		fmt.Println("active allowed feeds:", active.AllowedFeeds, "fvs:", active.AllowedFeedVersions)
	}
	return active, nil
}
