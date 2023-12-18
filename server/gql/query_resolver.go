package gql

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-mw/meters"
	"github.com/interline-io/transitland-server/model"
)

const MAX_RADIUS = 100_000

// query root

type queryResolver struct{ *Resolver }

func (r *queryResolver) Me(ctx context.Context) (*model.Me, error) {
	me, err := r.frs.Checker.Me(ctx, &authz.MeRequest{})
	if err != nil {
		return nil, err
	}
	gme := model.Me{
		ID:    me.User.Id,
		Email: &me.User.Email,
		Name:  &me.User.Name,
		Roles: me.Roles,
	}
	gme.ExternalData = map[string]any{}
	for k, v := range me.ExternalData {
		gme.ExternalData[k] = v
	}
	return &gme, nil
}

func (r *queryResolver) Agencies(ctx context.Context, limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	addMetric(ctx, "agencies")
	if where != nil {
		if where.Near != nil && where.Near.Radius > MAX_RADIUS {
			return nil, errors.New("radius too large")
		}
		if where.Bbox != nil && !checkBbox(where.Bbox, MAX_RADIUS*MAX_RADIUS) {
			return nil, errors.New("bbox too large")
		}
	}
	return r.frs.Finder.FindAgencies(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Routes(ctx context.Context, limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	addMetric(ctx, "routes")
	if where != nil {
		if where.Near != nil && where.Near.Radius > MAX_RADIUS {
			return nil, errors.New("radius too large")
		}
		if where.Bbox != nil && !checkBbox(where.Bbox, MAX_RADIUS*MAX_RADIUS) {
			return nil, errors.New("bbox too large")
		}
	}
	return r.frs.Finder.FindRoutes(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Stops(ctx context.Context, limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	addMetric(ctx, "stops")
	if where != nil {
		if where.Near != nil && where.Near.Radius > MAX_RADIUS {
			return nil, errors.New("radius too large")
		}
		if where.Bbox != nil && !checkBbox(where.Bbox, MAX_RADIUS*MAX_RADIUS) {
			return nil, errors.New("bbox too large")
		}
	}
	return r.frs.Finder.FindStops(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Trips(ctx context.Context, limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	addMetric(ctx, "trips")
	return r.frs.Finder.FindTrips(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) FeedVersions(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	addMetric(ctx, "feedVersions")
	if where != nil {
		if where.Near != nil && where.Near.Radius > MAX_RADIUS {
			return nil, errors.New("radius too large")
		}
		if where.Bbox != nil && !checkBbox(where.Bbox, MAX_RADIUS*MAX_RADIUS) {
			return nil, errors.New("bbox too large")
		}
	}
	return r.frs.Finder.FindFeedVersions(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Feeds(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	addMetric(ctx, "feeds")
	if where != nil {
		if where.Near != nil && where.Near.Radius > MAX_RADIUS {
			return nil, errors.New("radius too large")
		}
		if where.Bbox != nil && !checkBbox(where.Bbox, MAX_RADIUS*MAX_RADIUS) {
			return nil, errors.New("bbox too large")
		}
	}
	return r.frs.Finder.FindFeeds(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Operators(ctx context.Context, limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	addMetric(ctx, "operators")
	if where != nil {
		if where.Near != nil && where.Near.Radius > MAX_RADIUS {
			return nil, errors.New("radius too large")
		}
		if where.Bbox != nil && !checkBbox(where.Bbox, MAX_RADIUS*MAX_RADIUS) {
			return nil, errors.New("bbox too large")
		}
	}
	return r.frs.Finder.FindOperators(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Places(ctx context.Context, limit *int, after *int, level *model.PlaceAggregationLevel, where *model.PlaceFilter) ([]*model.Place, error) {
	return r.frs.Finder.FindPlaces(ctx, checkLimit(limit), checkCursor(after), nil, level, where)
}

func addMetric(ctx context.Context, resolverName string) {
	if apiMeter := meters.ForContext(ctx); apiMeter != nil {
		apiMeter.AddDimension("graphql", "resolver", resolverName)
	}
}
