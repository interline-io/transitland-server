package rtfinder

import (
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestRedisCache(t *testing.T) {
	// redis jobs and cache
	redisUrl := os.Getenv("TL_TEST_REDIS_URL")
	if redisUrl == "" {
		t.Skip("no TL_TEST_REDIS_URL")
		return
	}
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	rtCache := NewRedisCache(client)
	testCache(t, rtCache)
}

// func TestRedisConsumers(t *testing.T) {
// 	// redis jobs and cache
// 	redisUrl := "localhost:6379"
// 	client := redis.NewClient(&redis.Options{Addr: redisUrl})
// 	rtCache := NewRedisCache(client)
// 	rtJobs := NewRedisJobs(client, fmt.Sprintf("queue:%d:%d", os.Getpid(), time.Now().UnixNano()))
// 	testConsumers(t, rtCache, rtJobs)
// }
