package rtcache

import (
	"sync"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/jmoiron/sqlx"
)

type RTFinder struct {
	cache    Cache
	fetchers map[string]*rtConsumer
	lock     sync.Mutex
	*lookupCache
}

func NewRTFinder(cache Cache, db sqlx.Ext) *RTFinder {
	return &RTFinder{
		cache:       cache,
		lookupCache: newLookupCache(db),
		fetchers:    map[string]*rtConsumer{},
	}
}

func (f *RTFinder) AddData(topic string, data []byte) error {
	return f.cache.AddData(topic, data)
}

func (f *RTFinder) GetTrip(topic string, tid string) (*pb.TripUpdate, bool) {
	a, err := f.getListener(getTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil, false
	}
	trip, ok := a.GetTrip(tid)
	return trip, ok
}

func (f *RTFinder) FindStopTimeUpdate(topic string, tid string, sid string, seq int) (*pb.TripUpdate_StopTimeUpdate, bool) {
	rtTrip, rtok := f.GetTrip(topic, tid)
	if !rtok {
		return nil, false
	}
	for _, ste := range rtTrip.StopTimeUpdate {
		// Must match on StopSequence
		// TODO: allow matching on stop_id if stop_sequence is not provided
		if int(ste.GetStopSequence()) == seq {
			return ste, true
		}
	}
	return nil, false
}

// TODO: put this method on consumer and wrap, as with GetTrip
func (f *RTFinder) GetAddedTripsForStop(topic string, sid string) []*pb.TripUpdate {
	a, err := f.getListener(getTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil
	}
	// TODO: index more efficiently
	var ret []*pb.TripUpdate
	for _, trip := range a.entityByTrip {
		if trip.Trip.GetScheduleRelationship() != pb.TripDescriptor_ADDED {
			continue
		}
		for _, ste := range trip.StopTimeUpdate {
			if ste.GetStopId() == sid {
				ret = append(ret, trip)
				break // continue to next trip
			}
		}
	}
	return ret
}

func (f *RTFinder) getListener(topicKey string) (*rtConsumer, error) {
	f.lock.Lock()
	a, ok := f.fetchers[topicKey]
	if !ok {
		ch, err := f.cache.Listen(topicKey)
		// Failed to create listener
		if err != nil {
			// fmt.Printf("manager: '%s' failed to create listener\n", topicKey)
			return nil, err
		}
		// fmt.Printf("manager: '%s' listener created\n", topicKey)
		a, _ = newRTConsumer()
		a.feed = topicKey
		a.Start(ch)
		// fmt.Printf("manager: '%s' consumer started\n", topicKey)
		f.fetchers[topicKey] = a
	}
	f.lock.Unlock()
	return a, nil
}
