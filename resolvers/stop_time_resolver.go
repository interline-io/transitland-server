package resolvers

import (
	"context"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/find"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/rt"
)

// STOP TIME

type stopTimeResolver struct{ *Resolver }

func (r *stopTimeResolver) Stop(ctx context.Context, obj *model.StopTime) (*model.Stop, error) {
	return find.For(ctx).StopsByID.Load(atoi(obj.StopID))
}

func (r *stopTimeResolver) Trip(ctx context.Context, obj *model.StopTime) (*model.Trip, error) {
	return find.For(ctx).TripsByID.Load(atoi(obj.TripID))
}

func (r *stopTimeResolver) Rt(ctx context.Context, obj *model.StopTime) (*model.StopTimeUpdate, error) {
	stop, err := r.Stop(ctx, obj)
	if stop == nil || err != nil {
		panic(err)
	}
	trip, err := r.Trip(ctx, obj)
	if trip == nil || err != nil {
		panic(err)
	}
	sid := stop.StopID
	tid := trip.TripID
	msg, ok := rt.MC.Get(stop.FeedOnestopID, "trip_updates")
	if !ok {
		return nil, nil
	}
	found := false
	rent := model.StopTimeUpdate{}
	for _, fent := range msg.Entity {
		v := fent.TripUpdate
		if v == nil {
			continue
		}
		t := v.GetTrip()
		if t.GetTripId() != tid {
			continue
		}
		for _, st := range v.StopTimeUpdate {
			if st.GetStopId() == sid {
				arv := st.Arrival
				if arv != nil {
					found = true
					rent.ArrivalTime = steToWt(arv)
				}
				dep := st.Departure
				if dep == nil {
					dep = arv
				}
				if dep != nil {
					found = true
					rent.DepartureTime = steToWt(dep)
				}
			}
		}
	}
	if !found {
		return nil, nil
	}
	return &rent, nil
}

func steToWt(st *pb.TripUpdate_StopTimeEvent) *tl.WideTime {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	wtt := time.Unix(*st.Time, 0).In(loc)
	wt := tl.NewWideTimeFromSeconds(wtt.Hour()*3600 + wtt.Minute()*60 + wtt.Second())
	return &wt
}
