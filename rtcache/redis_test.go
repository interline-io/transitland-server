package rtcache

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRedisJobs(t *testing.T) {
	redisUrl := "localhost:6379"
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	rtJobs := NewRedisJobs(client, fmt.Sprintf("queue:%d:%d", os.Getpid(), time.Now().UnixNano()))
	testJobs(t, rtJobs)
}

func TestRedisCache(t *testing.T) {
	// redis jobs and cache
	redisUrl := "localhost:6379"
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	rtCache := NewRedisCache(client)
	testCache(t, rtCache)
}

func TestRedisConsumers(t *testing.T) {
	// redis jobs and cache
	redisUrl := "localhost:6379"
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	rtCache := NewRedisCache(client)
	rtJobs := NewRedisJobs(client, fmt.Sprintf("queue:%d:%d", os.Getpid(), time.Now().UnixNano()))
	testConsumers(t, rtCache, rtJobs)
}
