package rtcache

import (
	"fmt"
	"sync"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/jmoiron/sqlx"
)

func getTopicKey(topic string, t string) string {
	return fmt.Sprintf("rtdata:%s:%s", topic, t)
}

type RTFinder struct {
	cache    Cache
	lookup   *lookupCache
	fetchers map[string]*rtConsumer
	lock     sync.Mutex
}

func NewRTFinder(cache Cache, db sqlx.Ext) *RTFinder {
	return &RTFinder{
		cache:    cache,
		lookup:   newLookupCache(db),
		fetchers: map[string]*rtConsumer{},
	}
}

func (f *RTFinder) FeedVersionOnestopID(id int) (string, bool) {
	return f.lookup.FeedVersionOnestopID(id)
}

func (f *RTFinder) StopTimezone(id int, known string) (*time.Location, bool) {
	return f.lookup.StopTimezone(id, known)
}

func (f *RTFinder) TripGTFSTripID(id int) (string, bool) {
	return f.lookup.TripGTFSTripID(id)
}

func (f *RTFinder) GetTripIDs(topic string) ([]string, error) {
	var ret []string
	a, err := f.getListener(getTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil, err
	}
	for k := range a.entityByTrip {
		ret = append(ret, k)
	}
	return ret, nil
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

// TODO: put this method on consumer and wrap, as with GetTrip
func (f *RTFinder) GetAddedTripsForStop(topic string, sid string) []*pb.TripUpdate {
	a, err := f.getListener(getTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil
	}
	var ret []*pb.TripUpdate
	for _, trip := range a.entityByTrip {
		if trip.Trip.ScheduleRelationship != pb.TripDescriptor_ADDED.Enum() {
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
			fmt.Printf("manager: '%s' failed to create listener\n", topicKey)
			return nil, err
		}
		fmt.Printf("manager: '%s' listener created\n", topicKey)
		a, _ = newRTConsumer()
		a.feed = topicKey
		a.Start(ch)
		fmt.Printf("manager: '%s' consumer started\n", topicKey)
		f.fetchers[topicKey] = a
	}
	f.lock.Unlock()
	return a, nil
}
