package rtcache

import (
	"testing"
)

func TestLocalJobs(t *testing.T) {
	rtJobs := NewLocalJobs()
	testJobs(t, rtJobs)
}

func TestLocalCache(t *testing.T) {
	rtCache := NewLocalCache()
	testCache(t, rtCache)
}

func TestLocalConsumers(t *testing.T) {
	rtJobs := NewLocalJobs()
	rtCache := NewLocalCache()
	testConsumers(t, rtCache, rtJobs)
}
