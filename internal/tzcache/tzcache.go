package tzcache

import (
	"sync"
	"time"

	"github.com/interline-io/log"
	"github.com/jmoiron/sqlx"
)

type Cache struct {
	lock                sync.Mutex
	stopTimezoneCache   *tzCache
	agencyTimezoneCache *tzCache
	db                  sqlx.Ext
}

func NewCache(db sqlx.Ext) *Cache {
	return &Cache{
		db:                  db,
		stopTimezoneCache:   newtzCache(),
		agencyTimezoneCache: newtzCache(),
	}
}

func (tzc *Cache) Location(tz string) (*time.Location, bool) {
	return tzc.agencyTimezoneCache.Location(tz)
}

// StopTimezone looks up the timezone for a stop
func (tzc *Cache) StopTimezone(id int, known string) (*time.Location, bool) {
	tzc.lock.Lock()
	defer tzc.lock.Unlock()
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
	return queryTimezone(tzc.stopTimezoneCache, id, known, tzc.db, q, "stop_id")
}

// AgencyTimezone looks up the timezone for an agency
func (tzc *Cache) AgencyTimezone(id int, known string) (*time.Location, bool) {
	tzc.lock.Lock()
	defer tzc.lock.Unlock()
	q := `select agency_timezone from gtfs_agencies where gtfs_agencies.id = $1 limit 1`
	return queryTimezone(tzc.agencyTimezoneCache, id, known, tzc.db, q, "agency_id")
}

func queryTimezone(c *tzCache, id int, known string, db sqlx.Ext, query string, logKey string) (*time.Location, bool) {
	// If a timezone is provided, save it and return immediately
	if known != "" {
		log.Trace().Int(logKey, id).Str("known", known).Msg("tz: using known timezone")
		return c.Add(id, known)
	}

	// Check the cache
	if loc, ok := c.Get(id); ok {
		log.Trace().Int(logKey, id).Str("loc", loc.String()).Msg("tz: using cached timezone")
		return loc, ok
	} else {
		log.Trace().Int(logKey, id).Str("loc", loc.String()).Msg("tz: timezone not in cache")
	}
	if id == 0 {
		log.Trace().Int(logKey, id).Msg("tz: lookup failed, cant find timezone for entity with id=0 unless speciifed explicitly")
		return nil, false
	}
	// Otherwise lookup the timezone
	tz := ""
	if err := sqlx.Get(db, &tz, query, id); err != nil {
		log.Error().Err(err).Int(logKey, id).Msg("tz: lookup failed")
		return nil, false
	}
	loc, ok := c.Add(id, tz)
	log.Trace().Int(logKey, id).Str("loc", loc.String()).Msg("tz: lookup successful")
	return loc, ok
}

///////////////////

type tzCache struct {
	lock   sync.Mutex
	tzs    map[string]*time.Location
	values map[int]string
}

func newtzCache() *tzCache {
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
