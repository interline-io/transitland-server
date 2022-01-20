package resolvers

import (
	"context"
	"time"

	"github.com/interline-io/transitland-server/model"
)

// STOP

type stopResolver struct{ *Resolver }

func (r *stopResolver) FeedVersion(ctx context.Context, obj *model.Stop) (*model.FeedVersion, error) {
	return For(ctx).FeedVersionsByID.Load(obj.FeedVersionID)
}

func (r *stopResolver) Level(ctx context.Context, obj *model.Stop) (*model.Level, error) {
	if !obj.LevelID.Valid {
		return nil, nil
	}
	return For(ctx).LevelsByID.Load(atoi(obj.LevelID.Key))
}

func (r *stopResolver) Parent(ctx context.Context, obj *model.Stop) (*model.Stop, error) {
	if !obj.ParentStation.Valid {
		return nil, nil
	}
	return For(ctx).StopsByID.Load(atoi(obj.ParentStation.Key))
}

func (r *stopResolver) Children(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Stop, error) {
	return For(ctx).StopsByParentStopID.Load(model.StopParam{ParentStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) RouteStops(ctx context.Context, obj *model.Stop, limit *int) ([]*model.RouteStop, error) {
	return For(ctx).RouteStopsByStopID.Load(model.RouteStopParam{StopID: obj.ID, Limit: limit})
}

func (r *stopResolver) PathwaysFromStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByFromStopID.Load(model.PathwayParam{FromStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) PathwaysToStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return For(ctx).PathwaysByToStopID.Load(model.PathwayParam{ToStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) StopTimes(ctx context.Context, obj *model.Stop, limit *int, where *model.StopTimeFilter) ([]*model.StopTime, error) {
	sts, err := For(ctx).StopTimesByStopID.Load(model.StopTimeParam{FeedVersionID: obj.FeedVersionID, StopID: obj.ID, Limit: limit, Where: where})
	if err != nil {
		return nil, err
	}
	return sts, nil
}

func (r *stopResolver) Directions(ctx context.Context, obj *model.Stop, from *model.WaypointInput, to *model.WaypointInput, mode *model.StepMode, departAt *time.Time) (*model.Directions, error) {
	oc := obj.Coordinates()
	swp := model.Waypoint{
		Lon:  oc[0],
		Lat:  oc[1],
		Name: &obj.StopName,
	}
	p := directionsRequest{}
	if from != nil {
		p.Origin = model.Waypoint{
			Lon: from.Lon,
			Lat: from.Lat,
		}
		p.Destination = swp
	} else if to != nil {
		p.Origin = swp
		p.Destination = model.Waypoint{
			Lon: to.Lon,
			Lat: to.Lat,
		}
	}
	ret, err := demoValhallaRequest(p)
	return ret, err
}
