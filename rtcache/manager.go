package rtcache

import (
	"fmt"
	"sync"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/jmoiron/sqlx"
)

type RTConsumerManager struct {
	cache         Cache
	fetchers      map[string]*RTConsumer
	lock          sync.Mutex
	db            sqlx.Ext
	fvidFeedCache *rtEntityIDCache
	stopIdCache   *rtEntityIDCache
	tripIdCache   *rtEntityIDCache
}

func NewRTConsumerManager(cache Cache, db sqlx.Ext) *RTConsumerManager {
	return &RTConsumerManager{
		cache:         cache,
		fetchers:      map[string]*RTConsumer{},
		db:            db,
		fvidFeedCache: newRTEntityIDCache(),
		stopIdCache:   newRTEntityIDCache(),
		tripIdCache:   newRTEntityIDCache(),
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
	a, err := f.getListener(fmt.Sprintf("rtdata:%s:trip_updates", topic))
	if err != nil {
		return nil, err
	}
	for k := range a.entityByTrip {
		ret = append(ret, k)
	}
	return ret, nil
}

func (f *RTConsumerManager) GetTrip2(fvid int, id int) (*pb.TripUpdate, bool) {
	return f.GetTrip(f.getFvidTopic(fvid), f.getTripId(id))
}

func (f *RTConsumerManager) GetTrip(topic string, tid string) (*pb.TripUpdate, bool) {
	a, err := f.getListener(fmt.Sprintf("rtdata:%s:trip_updates", topic))
	if err != nil {
		return nil, false
	}
	trip, ok := a.GetTrip(tid)
	return trip, ok
}

// TODO: put this method on consumer and wrap, as with GetTrip
func (f *RTConsumerManager) GetAddedTripsForStop(topic string, sid string) []*pb.TripUpdate {
	a, err := f.getListener(fmt.Sprintf("rtdata:%s:trip_updates", topic))
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

// caching

func (f *RTConsumerManager) getFvidTopic(id int) string {
	if a, ok := f.fvidFeedCache.Get(id); ok {
		return a
	}
	eid := ""
	err := sqlx.Get(
		f.db,
		&eid,
		"select current_feeds.onestop_id from feed_versions join current_feeds on current_feeds.id = feed_versions.feed_id where feed_versions.id = $1",
		id,
	)
	_ = err
	f.fvidFeedCache.Set(id, eid)
	return eid
}

func (f *RTConsumerManager) getStopId(id int) string {
	if a, ok := f.stopIdCache.Get(id); ok {
		return a
	}
	eid := ""
	err := sqlx.Get(f.db, &eid, "select stop_id from gtfs_stops where id = $1", id)
	_ = err
	f.stopIdCache.Set(id, eid)
	return eid
}

func (f *RTConsumerManager) getTripId(id int) string {
	if a, ok := f.tripIdCache.Get(id); ok {
		return a
	}
	eid := ""
	err := sqlx.Get(f.db, &eid, "select trip_id from gtfs_trips where id = $1", id)
	_ = err
	f.tripIdCache.Set(id, eid)
	return eid
}

/////
type rtEntityIDCache struct {
	lock   sync.Mutex
	values map[int]string
}

func (c *rtEntityIDCache) Get(key int) (string, bool) {
	c.lock.Lock()
	a, ok := c.values[key]
	c.lock.Unlock()
	return a, ok
}

func (c *rtEntityIDCache) Set(key int, value string) {
	c.lock.Lock()
	c.values[key] = value
	c.lock.Unlock()
}

func newRTEntityIDCache() *rtEntityIDCache {
	return &rtEntityIDCache{
		values: map[int]string{},
	}
}
