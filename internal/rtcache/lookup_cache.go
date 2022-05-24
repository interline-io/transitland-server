package rtcache

import (
	"sync"
	"time"

	"github.com/interline-io/transitland-lib/log"
	"github.com/jmoiron/sqlx"
)

type lookupCache struct {
	db              sqlx.Ext
	fvidFeedCache   simpleCache
	gtfsTripIdCache simpleCache
	routeIdCache    skeyCache
	tzCache         tzCache
}

func newLookupCache(db sqlx.Ext) *lookupCache {
	return &lookupCache{
		db: db,
	}
}

func (f *lookupCache) GetRouteID(fvid int, tid string) (int, bool) {
	sk := skey{fvid, tid}
	if a, ok := f.routeIdCache.Get(sk); ok {
		return a, ok
	}
	eid := 0
	err := sqlx.Get(f.db, &eid, "select id from gtfs_routes where feed_version_id = $1 and route_id = $2", fvid, tid)
	f.routeIdCache.Set(sk, eid)
	return eid, err == nil
}

func (f *lookupCache) GetGtfsTripID(id int) (string, bool) {
	if a, ok := f.gtfsTripIdCache.Get(id); ok {
		return a, ok
	}
	q := `select trip_id from gtfs_trips where id = $1 limit 1`
	eid := ""
	err := sqlx.Get(f.db, &eid, q, id)
	f.gtfsTripIdCache.Set(id, eid)
	return eid, err == nil
}

func (f *lookupCache) GetFeedVersionOnestopID(id int) (string, bool) {
	if a, ok := f.fvidFeedCache.Get(id); ok {
		return a, ok
	}
	q := `
	select current_feeds.onestop_id 
	from feed_versions 
	join current_feeds on current_feeds.id = feed_versions.feed_id 
	where feed_versions.id = $1
	limit 1`
	eid := ""
	err := sqlx.Get(
		f.db,
		&eid,
		q,
		id,
	)
	f.fvidFeedCache.Set(id, eid)
	if err != nil {
		return "", false
	}
	return eid, err == nil
}

// StopTimezone looks up the timezone for a stop
func (f *lookupCache) StopTimezone(id int, known string) (*time.Location, bool) {
	// If a timezone is provided, save it and return immediately
	if known != "" {
		log.Trace().Int("stop_id", id).Str("known", known).Msg("tz: using known timezone")
		return f.tzCache.Add(id, known)
	}
	// Check the cache
	if loc, ok := f.tzCache.Get(id); ok {
		log.Trace().Int("stop_id", id).Str("known", known).Str("loc", loc.String()).Msg("tz: using cached timezone")
		return loc, ok
	}
	if id == 0 {
		log.Trace().Int("stop_id", id).Msg("tz: lookup failed, cant find timezone for stops with id=0 unless speciifed explicitly")
		return nil, false
	}
	// Otherwise lookup the timezone
	q := `
		select COALESCE(nullif(s.stop_timezone, ''), nullif(p.stop_timezone, ''), a.agency_timezone)
		from gtfs_stops s
		left join gtfs_stops p on p.id = s.parent_station
		left join lateral (
			select gtfs_agencies.agency_timezone
			from gtfs_agencies
			where gtfs_agencies.feed_version_id = s.feed_version_id
			limit 1
		) a on true
		where s.id = $1
		limit 1`
	tz := ""
	if err := sqlx.Get(f.db, &tz, q, id); err != nil {
		log.Error().Err(err).Int("stop_id", id).Str("known", known).Msg("tz: lookup failed")
		return nil, false
	}
	loc, ok := f.tzCache.Add(id, tz)
	log.Trace().Int("stop_id", id).Str("known", known).Str("loc", loc.String()).Msg("tz: lookup successful")
	return loc, ok
}

// Lookup time.Location by name
func (f *lookupCache) Location(tz string) (*time.Location, bool) {
	return f.tzCache.Location(tz)
}

/////

type skey struct {
	fvid int
	eid  string
}

type skeyCache struct {
	lock   sync.Mutex
	values map[skey]int
}

func (c *skeyCache) Get(key skey) (int, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	a, ok := c.values[key]
	return a, ok
}

func (c *skeyCache) Set(key skey, value int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.values == nil {
		c.values = map[skey]int{}
	}
	c.values[key] = value
}

///

type simpleCache struct {
	lock   sync.Mutex
	values map[int]string
}

func (c *simpleCache) Get(key int) (string, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	a, ok := c.values[key]
	return a, ok
}

func (c *simpleCache) Set(key int, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.values == nil {
		c.values = map[int]string{}
	}
	c.values[key] = value
}

func newSimpleCache() *simpleCache {
	return &simpleCache{
		values: map[int]string{},
	}
}

////

// tzCache saves and manages the timezone location cache
type tzCache struct {
	lock   sync.Mutex
	tzs    map[string]*time.Location
	values map[int]string
}

func newTzCache() *tzCache {
	return &tzCache{
		tzs:    map[string]*time.Location{},
		values: map[int]string{},
	}
}

func (c *tzCache) Get(key int) (*time.Location, bool) {
	var loc *time.Location
	defer c.lock.Unlock()
	c.lock.Lock()
	tz, ok := c.values[key]
	if ok {
		loc, ok = c.tzs[tz] // will be nil if invalid timezone
	}
	return loc, ok
}

func (c *tzCache) Location(tz string) (*time.Location, bool) {
	defer c.lock.Unlock()
	c.lock.Lock()
	loc := c.tzs[tz]
	return loc, loc == nil
}

func (c *tzCache) Add(key int, tz string) (*time.Location, bool) {
	var err error
	var loc *time.Location
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.values == nil {
		c.values = map[int]string{}
	}
	if c.tzs == nil {
		c.tzs = map[string]*time.Location{}
	}
	c.values[key] = tz
	loc, ok := c.tzs[tz]
	if !ok {
		ok = true
		loc, err = time.LoadLocation(tz)
		if err != nil {
			ok = false
			loc = nil
		}
		c.tzs[tz] = loc
	}
	return loc, ok
}
