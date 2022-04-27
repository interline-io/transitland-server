package xcache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	gcache "github.com/Code-Hex/go-generics-cache"
	"github.com/go-redis/redis/v8"
)

// Item is an item
type Item[K comparable, T any] struct {
	Key        K
	Value      T
	Expiration time.Duration
}

//////////

type LC[K comparable, T any] interface {
	Get(K) (T, bool)
	Set(K, T)
	Delete(K)
}

////////// wrap go-generics-cache

type gcWrap[K comparable, T any] struct {
	Ttl time.Duration
	*gcache.Cache[K, T]
}

func (c *gcWrap[K, T]) Set(key K, value T) {
	c.Cache.Set(key, value, gcache.WithExpiration(c.Ttl))
}

//////////

type Cache[K comparable, T any] struct {
	Ttl       time.Duration
	KeyBytes  func(K) ([]byte, error)
	FromBytes func([]byte) (T, error)
	ToBytes   func(T) ([]byte, error)
	client    *redis.Client
	lock      sync.Mutex
	cache     LC[K, T]
}

func JsonValue[K any](key K) ([]byte, error) {
	return json.Marshal(key)
}

func FromJson[T any](b []byte) (T, error) {
	var a T
	err := json.Unmarshal(b, &a)
	return a, err
}

func New[K comparable, T any](client *redis.Client, ttl time.Duration) *Cache[K, T] {
	c := Cache[K, T]{}
	c.cache = &gcWrap[K, T]{Ttl: ttl, Cache: gcache.New(gcache.AsLRU[K, T]())}
	c.KeyBytes = JsonValue[K]
	c.ToBytes = JsonValue[T]
	c.FromBytes = FromJson[T]
	c.Ttl = ttl
	c.client = client
	return &c
}

func (c *Cache[K, T]) Set(key K, value T) error {
	c.cache.Set(key, value)
	bkey, err := c.KeyBytes(key)
	if err != nil {
		return err
	}
	vb, err := c.ToBytes(value)
	if err != nil {
		return err
	}
	if err := c.client.Set(context.TODO(), string(bkey), vb, c.Ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (c *Cache[K, T]) Get(key K) (T, error) {
	var v T
	v, ok := c.cache.Get(key)
	if ok {
		return v, nil
	}
	bkey, err := c.KeyBytes(key)
	if err != nil {
		return v, err
	}
	lastData := c.client.Get(context.TODO(), string(bkey))
	if err := lastData.Err(); err == redis.Nil {
		return v, nil
	} else if err != nil {
		return v, err
	}
	b, _ := lastData.Bytes()
	v, err = c.FromBytes(b)
	c.Set(key, v)
	return v, err
}
