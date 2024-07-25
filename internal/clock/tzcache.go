package clock

import (
	"sync"
	"time"
)

// TzCache saves and manages the timezone location cache
type TzCache[K comparable] struct {
	lock   sync.Mutex
	tzs    map[string]*time.Location
	values map[K]string
}

func NewTzCache[K comparable]() *TzCache[K] {
	return &TzCache[K]{
		tzs:    map[string]*time.Location{},
		values: map[K]string{},
	}
}

func (c *TzCache[K]) Get(key K) (*time.Location, bool) {
	defer c.lock.Unlock()
	c.lock.Lock()
	tz, ok := c.values[key]
	if ok {
		return c.load(tz)
	}
	return nil, false
}

func (c *TzCache[K]) Location(tz string) (*time.Location, bool) {
	defer c.lock.Unlock()
	c.lock.Lock()
	return c.load(tz)
}
func (c *TzCache[K]) Add(key K, tz string) (*time.Location, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.values[key] = tz
	return c.load(tz)
}

func (c *TzCache[K]) load(tz string) (*time.Location, bool) {
	loc, ok := c.tzs[tz]
	if !ok {
		var err error
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
