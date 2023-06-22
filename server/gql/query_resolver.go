package gql

import (
	"context"

	"github.com/interline-io/transitland-server/authz"
	"github.com/interline-io/transitland-server/internal/meters"
	"github.com/interline-io/transitland-server/model"
)

// query root

type queryResolver struct{ *Resolver }

func (r *queryResolver) Agencies(ctx context.Context, limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	addMetric(ctx, "agencies")
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindAgencies(ctx, limit, checkCursor(after), ids, ca, where)
}

func (r *queryResolver) Routes(ctx context.Context, limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	addMetric(ctx, "routes")
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindRoutes(ctx, limit, checkCursor(after), ids, ca, where)
}

func (r *queryResolver) Stops(ctx context.Context, limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	addMetric(ctx, "stops")
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindStops(ctx, limit, checkCursor(after), ids, ca, where)
}

func (r *queryResolver) Trips(ctx context.Context, limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	addMetric(ctx, "trips")
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindTrips(ctx, limit, checkCursor(after), ids, ca, where)
}

func (r *queryResolver) FeedVersions(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	addMetric(ctx, "feedVersions")
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindFeedVersions(ctx, limit, checkCursor(after), ids, ca, where)
}

func (r *queryResolver) Feeds(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	addMetric(ctx, "feeds")
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindFeeds(ctx, limit, checkCursor(after), ids, ca, where)
}

func (r *queryResolver) Operators(ctx context.Context, limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	addMetric(ctx, "operators")
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindOperators(ctx, limit, checkCursor(after), ids, ca, where)
}

func (r *queryResolver) Places(ctx context.Context, limit *int, after *int, level *model.PlaceAggregationLevel, where *model.PlaceFilter) ([]*model.Place, error) {
	ca, err := checkActive(ctx, nil, r.authzChecker)
	if err != nil {
		return nil, err
	}
	return r.finder.FindPlaces(ctx, limit, checkCursor(after), nil, level, ca, where)
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

func checkActive(ctx context.Context, ids []int, checker *authz.Checker) (*model.PermFilter, error) {
	if checker == nil {
		return nil, nil
	}
	active := &model.PermFilter{}
	if a, err := checker.CheckGlobalAdmin(ctx); err != nil {
		return nil, err
	} else if a {
		return nil, nil
	}
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
	// fmt.Println("active allowed feeds:", active.AllowedFeeds, "fvs:", active.AllowedFeedVersions)
	return active, nil
}
