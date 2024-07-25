package clock

import (
	"sync"
	"testing"
	"time"
)

func BenchmarkLookupCache_LoadLocation(b *testing.B) {
	t := "America/Los_Angeles"
	time.LoadLocation(t)
	for n := 0; n < b.N; n++ {
		loc, err := time.LoadLocation(t)
		_ = loc
		_ = err
	}
}

func BenchmarkLookupCache_LoadLocationCache(b *testing.B) {
	lock := sync.Mutex{}
	c := map[string]*time.Location{}
	t := "America/Los_Angeles"
	time.LoadLocation(t)
	for n := 0; n < b.N; n++ {
		lock.Lock()
		if _, ok := c[t]; !ok {
			loc, err := time.LoadLocation(t)
			_ = loc
			_ = err
			c[t] = loc
		}
		lock.Unlock()
	}
}

func Benchmark_TzCache(b *testing.B) {
	c := NewTzCache[int]()
	for n := 0; n < b.N; n++ {
		loc, ok := c.Add(n, "America/Los_Angeles")
		_ = loc
		_ = ok
	}
}
