package gql

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-mw/auth/authn"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-server/model"
)

const MAX_RADIUS = 100_000

// query root

type queryResolver struct{ *Resolver }

func (r *queryResolver) Me(ctx context.Context) (*model.Me, error) {
	cfg := model.ForContext(ctx)
	me := model.Me{}
	me.ExternalData = tt.NewMap(map[string]any{})
	if checker := cfg.Checker; checker != nil {
		// Use checker if available
		cm, err := checker.Me(ctx, &authz.MeRequest{})
		if err != nil {
			return nil, err
		}
		me.ID = cm.User.Id
		me.Email = &cm.User.Email
		me.Name = &cm.User.Name
		me.Roles = cm.Roles
		for k, v := range cm.ExternalData {
			me.ExternalData.Val[k] = v
		}
	} else if user := authn.ForContext(ctx); user != nil {
		// Fallback to user context
		um := user.Email()
		un := user.Name()
		me.ID = user.ID()
		me.Name = &un
		me.Email = &um
		me.Roles = user.Roles()
	} else {
		return nil, errors.New("no user")
	}
	return &me, nil
}

func (r *queryResolver) Agencies(ctx context.Context, limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	addMetric(ctx, "agencies")
	if where != nil {
		if err := checkGeo(where.Near, where.Bbox); err != nil {
			return nil, err
		}
	}
	return model.ForContext(ctx).Finder.FindAgencies(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Routes(ctx context.Context, limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	addMetric(ctx, "routes")
	if where != nil {
		if err := checkGeo(where.Near, where.Bbox); err != nil {
			return nil, err
		}
	}
	return model.ForContext(ctx).Finder.FindRoutes(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Stops(ctx context.Context, limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	addMetric(ctx, "stops")
	if where != nil {
		if err := checkGeo(where.Near, where.Bbox); err != nil {
			return nil, err
		}
	}
	return model.ForContext(ctx).Finder.FindStops(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Trips(ctx context.Context, limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	addMetric(ctx, "trips")
	return model.ForContext(ctx).Finder.FindTrips(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) FeedVersions(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	addMetric(ctx, "feedVersions")
	if where != nil {
		if err := checkGeo(where.Near, where.Bbox); err != nil {
			return nil, err
		}
	}
	return model.ForContext(ctx).Finder.FindFeedVersions(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Feeds(ctx context.Context, limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	addMetric(ctx, "feeds")
	if where != nil {
		if err := checkGeo(where.Near, where.Bbox); err != nil {
			return nil, err
		}
	}
	return model.ForContext(ctx).Finder.FindFeeds(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Operators(ctx context.Context, limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	addMetric(ctx, "operators")
	if where != nil {
		if err := checkGeo(where.Near, where.Bbox); err != nil {
			return nil, err
		}
	}
	return model.ForContext(ctx).Finder.FindOperators(ctx, checkLimit(limit), checkCursor(after), ids, where)
}

func (r *queryResolver) Places(ctx context.Context, limit *int, after *int, level *model.PlaceAggregationLevel, where *model.PlaceFilter) ([]*model.Place, error) {
	return model.ForContext(ctx).Finder.FindPlaces(ctx, checkLimit(limit), checkCursor(after), nil, level, where)
}
