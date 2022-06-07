package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRedisJobs(t *testing.T) {
	// redis jobs and cache
	redisUrl := os.Getenv("TL_TEST_REDIS_URL")
	if redisUrl == "" {
		t.Skip("no TL_TEST_REDIS_URL")
		return
	}
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	rtJobs := NewRedisJobs(client, fmt.Sprintf("queue:%d:%d", os.Getpid(), time.Now().UnixNano()))
	testJobs(t, rtJobs)
}
