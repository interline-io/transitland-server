package rtcache

import (
	"context"
	"testing"
)

func TestLocalJobs(t *testing.T) {
	ctx := context.Background()
	rtJobs := NewLocalJobs(ctx)
	testJobs(t, rtJobs)
}

func TestLocalCache(t *testing.T) {
	ctx := context.Background()
	rtCache := NewLocalCache(ctx)
	testCache(t, rtCache)
}

func TestLocalConsumers(t *testing.T) {
	ctx := context.Background()
	rtJobs := NewLocalJobs(ctx)
	rtCache := NewLocalCache(ctx)
	testConsumers(t, rtCache, rtJobs)
}
