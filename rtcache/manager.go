package rtcache

import (
	"fmt"
	"sync"

	"github.com/interline-io/transitland-lib/rt/pb"
)

func GetTopicKey(topic string, t string) string {
	return fmt.Sprintf("rtdata:%s:%s", topic, t)
}

type RTConsumerManager struct {
	cache    Cache
	fetchers map[string]*RTConsumer
	lock     sync.Mutex
}

func NewRTConsumerManager(cache Cache) *RTConsumerManager {
	return &RTConsumerManager{
		cache:    cache,
		fetchers: map[string]*RTConsumer{},
	}
}

func (f *RTConsumerManager) getListener(topicKey string) (*RTConsumer, error) {
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
		a, _ = NewRTConsumer()
		a.feed = topicKey
		a.Start(ch)
		fmt.Printf("manager: '%s' consumer started\n", topicKey)
		f.fetchers[topicKey] = a
	}
	f.lock.Unlock()
	return a, nil
}

func (f *RTConsumerManager) GetTripIDs(topic string) ([]string, error) {
	var ret []string
	a, err := f.getListener(GetTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil, err
	}
	for k := range a.entityByTrip {
		ret = append(ret, k)
	}
	return ret, nil
}

func (f *RTConsumerManager) GetTrip(topic string, tid string) (*pb.TripUpdate, bool) {
	a, err := f.getListener(GetTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil, false
	}
	trip, ok := a.GetTrip(tid)
	return trip, ok
}

// TODO: put this method on consumer and wrap, as with GetTrip
func (f *RTConsumerManager) GetAddedTripsForStop(topic string, sid string) []*pb.TripUpdate {
	a, err := f.getListener(GetTopicKey(topic, "trip_updates"))
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
