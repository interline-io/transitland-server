package xcache

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

type TestBlob struct {
	data []byte
}

func (t *TestBlob) FromBytes(b []byte) error {
	t.data = b
	return nil
}

func (t TestBlob) Bytes() []byte {
	return t.data
}

func TestCache(t *testing.T) {
	// redis jobs and cache
	redisUrl := os.Getenv("TL_TEST_REDIS_URL")
	if redisUrl == "" {
		t.Skip("no TL_TEST_REDIS_URL")
		return
	}
	client := redis.NewClient(&redis.Options{Addr: redisUrl})
	c := New[string, *TestBlob](client, 1*time.Minute)
	c.KeyBytes = func(s string) ([]byte, error) { return []byte(s), nil }
	c.ToBytes = func(b *TestBlob) ([]byte, error) { return b.data, nil }
	if err := c.Set("test", &TestBlob{data: []byte("ok")}); err != nil {
		t.Fatal(err)
	}
	a, err := c.Get("test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("a:", a)
}
