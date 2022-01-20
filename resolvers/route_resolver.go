package resolvers

import (
	"context"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/model"
)

// ROUTE

type routeResolver struct{ *Resolver }

func (r *routeResolver) Geometries(ctx context.Context, obj *model.Route, limit *int) ([]*model.RouteGeometry, error) {
	return For(ctx).RouteGeometriesByRouteID.Load(model.RouteGeometryParam{RouteID: obj.ID, Limit: limit})
}

func (r *routeResolver) Trips(ctx context.Context, obj *model.Route, limit *int, where *model.TripFilter) ([]*model.Trip, error) {
	return For(ctx).TripsByRouteID.Load(model.TripParam{RouteID: obj.ID, Limit: limit, Where: where})
}

func (r *routeResolver) Agency(ctx context.Context, obj *model.Route) (*model.Agency, error) {
	return For(ctx).AgenciesByID.Load(atoi(obj.AgencyID))
}

func (r *routeResolver) FeedVersion(ctx context.Context, obj *model.Route) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(obj.FeedVersionID)
}

func (r *routeResolver) Stops(ctx context.Context, obj *model.Route, limit *int, where *model.StopFilter) ([]*model.Stop, error) {
	return For(ctx).StopsByRouteID.Load(model.StopParam{RouteID: obj.ID, Limit: limit, Where: where})
}

func (r *routeResolver) RouteStops(ctx context.Context, obj *model.Route, limit *int) ([]*model.RouteStop, error) {
	return For(ctx).RouteStopsByRouteID.Load(model.RouteStopParam{RouteID: obj.ID, Limit: limit})
}

func (r *routeResolver) Headways(ctx context.Context, obj *model.Route, limit *int) ([]*model.RouteHeadway, error) {
	return For(ctx).RouteHeadwaysByRouteID.Load(model.RouteHeadwayParam{RouteID: obj.ID, Limit: limit})
}

func (r *routeResolver) RouteStopBuffer(ctx context.Context, obj *model.Route, radius *float64) (*model.RouteStopBuffer, error) {
	// TODO: remove n+1 (which is tricky, what if multiple radius specified in different parts of query)
	p := model.RouteStopBufferParam{Radius: radius, EntityID: obj.ID}
	ents, err := r.finder.RouteStopBuffer(&p)
	if err != nil {
		return nil, err
	}
	if len(ents) > 0 {
		return ents[0], nil
	}
	return nil, nil
}

func (r *routeResolver) Alerts(ctx context.Context, obj *model.Route) ([]*model.Alert, error) {
	// TODO
	return nil, nil
}

// ROUTE HEADWAYS

type routeHeadwayResolver struct{ *Resolver }

func (r *routeHeadwayResolver) Stop(ctx context.Context, obj *model.RouteHeadway) (*model.Stop, error) {
	return For(ctx).StopsByID.Load(obj.SelectedStopID)
}

func (r *routeHeadwayResolver) Departures(ctx context.Context, obj *model.RouteHeadway) ([]*tl.WideTime, error) {
	var ret []*tl.WideTime
	for _, v := range obj.Departures.Ints {
		w := tl.NewWideTimeFromSeconds(v)
		ret = append(ret, &w)
	}
	return ret, nil
}

// ROUTE STOP

type routeStopResolver struct{ *Resolver }

func (r *routeStopResolver) Route(ctx context.Context, obj *model.RouteStop) (*model.Route, error) {
	return For(ctx).RoutesByID.Load(obj.RouteID)
}

func (r *routeStopResolver) Stop(ctx context.Context, obj *model.RouteStop) (*model.Stop, error) {
	return For(ctx).StopsByID.Load(obj.StopID)
}

func (r *routeStopResolver) Agency(ctx context.Context, obj *model.RouteStop) (*model.Agency, error) {
	return For(ctx).AgenciesByID.Load(obj.AgencyID)
}
