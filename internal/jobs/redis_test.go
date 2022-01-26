package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRedisJobs(t *testing.T) {
	redisUrl := os.Getenv("TL_REDIS_URL")
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	rtJobs := NewRedisJobs(client, fmt.Sprintf("queue:%d:%d", os.Getpid(), time.Now().UnixNano()))
	testJobs(t, rtJobs)
}
