package rtfinder

import (
	"testing"

	"github.com/interline-io/transitland-dbutil/testutil"
)

func TestRedisCache(t *testing.T) {
	// redis jobs and cache
	if a, ok := testutil.CheckTestRedisClient(); !ok {
		t.Skip(a)
		return
	}
	client := testutil.MustOpenTestRedisClient(t)
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
