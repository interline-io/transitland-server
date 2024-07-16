package rtfinder

import (
	"sync"
	"time"

	"github.com/interline-io/transitland-server/internal/tzcache"
	"github.com/jmoiron/sqlx"
)

type skey struct {
	fvid int
	eid  string
}

type lookupCache struct {
	db              sqlx.Ext
	fvidSourceCache map[int][]string
	fvidFeedCache   map[int]string
	gtfsTripIdCache map[int]string
	gtfsStopIdCache map[int]string
	routeIdCache    map[skey]int
	tzc             *tzcache.Cache
	lock            sync.Mutex
}

func newLookupCache(db sqlx.Ext) *lookupCache {
	return &lookupCache{
		db:              db,
		tzc:             tzcache.NewCache(db),
		fvidSourceCache: map[int][]string{},
		fvidFeedCache:   map[int]string{},
		gtfsTripIdCache: map[int]string{},
		gtfsStopIdCache: map[int]string{},
		routeIdCache:    map[skey]int{},
	}
}

func (f *lookupCache) GetRouteID(fvid int, tid string) (int, bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	sk := skey{fvid, tid}
	if a, ok := f.routeIdCache[sk]; ok {
		return a, ok
	}
	eid := 0
	err := sqlx.Get(f.db, &eid, "select id from gtfs_routes where feed_version_id = $1 and route_id = $2", fvid, tid)
	f.routeIdCache[sk] = eid
	return eid, err == nil
}

func (f *lookupCache) GetGtfsTripID(id int) (string, bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if a, ok := f.gtfsTripIdCache[id]; ok {
		return a, ok
	}
	q := `select trip_id from gtfs_trips where id = $1 limit 1`
	eid := ""
	err := sqlx.Get(f.db, &eid, q, id)
	f.gtfsTripIdCache[id] = eid
	return eid, err == nil
}

func (f *lookupCache) GetGtfsStopID(id int) (string, bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if a, ok := f.gtfsStopIdCache[id]; ok {
		return a, ok
	}
	q := `select stop_id from gtfs_stops where id = $1 limit 1`
	eid := ""
	err := sqlx.Get(f.db, &eid, q, id)
	f.gtfsStopIdCache[id] = eid
	return eid, err == nil
}

func (f *lookupCache) GetFeedVersionRTFeeds(id int) ([]string, bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if a, ok := f.fvidSourceCache[id]; ok {
		return a, ok
	}
	q := `
	select 
		distinct on(cf.onestop_id)
		cf.onestop_id 
	from feed_versions fv 
	join current_operators_in_feed coif on coif.feed_id = fv.feed_id 
	join current_operators_in_feed coif2 on coif2.resolved_onestop_id = coif.resolved_onestop_id 
	join current_feeds cf on coif2.feed_id = cf.id
	where fv.id = $1 
	order by cf.onestop_id
	`
	var eid []string
	err := sqlx.Select(
		f.db,
		&eid,
		q,
		id,
	)
	f.fvidSourceCache[id] = eid
	if err != nil {
		return nil, false
	}
	return eid, true
}

// Lookup time.Location by name
func (f *lookupCache) Location(tz string) (*time.Location, bool) {
	return f.tzc.Location(tz)
}

// StopTimezone looks up the timezone for a stop
func (f *lookupCache) StopTimezone(id int, known string) (*time.Location, bool) {
	return f.tzc.StopTimezone(id, known)
}

// AgencyTimezone looks up the timezone for an agency
func (f *lookupCache) Agencyimezone(id int, known string) (*time.Location, bool) {
	return f.tzc.AgencyTimezone(id, known)
}
