package resolvers

import (
	"context"
	"fmt"

	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
)

// STOP

type stopResolver struct{ *Resolver }

func (r *stopResolver) FeedVersion(ctx context.Context, obj *model.Stop) (*model.FeedVersion, error) {
	return find.For(ctx).FeedVersionsByID.Load(obj.FeedVersionID)
}

func (r *stopResolver) Level(ctx context.Context, obj *model.Stop) (*model.Level, error) {
	if !obj.LevelID.Valid {
		return nil, nil
	}
	return find.For(ctx).LevelsByID.Load(atoi(obj.LevelID.Key))
}

func (r *stopResolver) Parent(ctx context.Context, obj *model.Stop) (*model.Stop, error) {
	if !obj.ParentStation.Valid {
		return nil, nil
	}
	return find.For(ctx).StopsByID.Load(atoi(obj.ParentStation.Key))
}

func (r *stopResolver) Children(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Stop, error) {
	return find.For(ctx).StopsByParentStopID.Load(model.StopParam{ParentStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) RouteStops(ctx context.Context, obj *model.Stop, limit *int) ([]*model.RouteStop, error) {
	return find.For(ctx).RouteStopsByStopID.Load(model.RouteStopParam{StopID: obj.ID, Limit: limit})
}

func (r *stopResolver) PathwaysFromStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return find.For(ctx).PathwaysByFromStopID.Load(model.PathwayParam{FromStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) PathwaysToStop(ctx context.Context, obj *model.Stop, limit *int) ([]*model.Pathway, error) {
	return find.For(ctx).PathwaysByToStopID.Load(model.PathwayParam{ToStopID: obj.ID, Limit: limit})
}

func (r *stopResolver) StopTimes(ctx context.Context, obj *model.Stop, limit *int, where *model.StopTimeFilter) ([]*model.StopTime, error) {
	sts, err := find.For(ctx).StopTimesByStopID.Load(model.StopTimeParam{FeedVersionID: obj.FeedVersionID, StopID: obj.ID, Limit: limit, Where: where})
	if err != nil {
		return nil, err
	}
	// Merge scheduled stop times with rt stop times
	// TODO: handle StopTimeFilter
	rtcm := r.getConsumerManager()
	// Handle scheduled trips; these can be matched on trip_id or (route_id,direction_id,...)
	for _, st := range sts {
		rtTrip, ok := rtcm.GetTrip2(st.FeedVersionID, atoi(st.TripID))
		if !ok {
			continue
		}
		fmt.Println("FOUND SCHEDULED TRIP:", rtTrip)
	}
	// Handle added trips; these must specify stop_id in StopTimeUpdates
	fmt.Println("looking for added rt for stop:", obj.FeedOnestopID, obj.StopID)
	for _, rtTrip := range rtcm.GetAddedTripsForStop(obj.FeedOnestopID, obj.StopID) {
		fmt.Println("FOUND ADDED RT:", rtTrip)
		for _, stu := range rtTrip.StopTimeUpdate {
			if stu.GetStopId() != obj.StopID {
				continue
			}
			rtst := &model.StopTime{}
			if rtst == nil {
				continue
			}
			rtst.RTTrip = rtTrip
			rtst.FeedVersionID = obj.FeedVersionID
			rtst.StopID = obj.StopID
			rtst.TripID = "0"
			sts = append(sts, rtst)
		}
	}
	return sts, nil
}

func (r *stopResolver) Alerts(ctx context.Context, obj *model.Stop) ([]*model.Alert, error) {
	// TODO
	return nil, nil
}
