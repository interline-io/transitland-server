package ecache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
)

type Item[T any] struct {
	Value     T
	ExpiresAt time.Time
	RecheckAt time.Time
}

type Cache[T any] struct {
	topic string
	m     map[string]Item[T]
	lock  sync.Mutex
	redis *redis.Client
}

func NewCache[T any](client *redis.Client, topic string) *Cache[T] {
	return &Cache[T]{
		topic: topic,
		redis: client,
		m:     map[string]Item[T]{},
	}
}

func (e *Cache[T]) GetRecheckKeys() []string {
	e.lock.Lock()
	defer e.lock.Unlock()
	t := time.Now()
	var ret []string
	for k, v := range e.m {
		if v.RecheckAt.IsZero() {
			continue
		}
		if v.RecheckAt.Before(t) {
			ret = append(ret, k)
		}
	}
	return ret
}

func (e *Cache[T]) Get(key string) (T, bool) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if a, ok := e.m[key]; ok {
		return a.Value, true
	}
	v, ok := e.getRedis(key)
	e.m[key] = v
	return v.Value, ok
}

func (e *Cache[T]) Set(key string, value T) {
	e.lock.Lock()
	defer e.lock.Unlock()
	item := Item[T]{
		Value: value,
	}
	e.m[key] = item
	e.setRedis(key, item, 0)
}

func (e *Cache[T]) SetTTL(key string, value T, ttl1 time.Duration, ttl2 time.Duration) {
	e.lock.Lock()
	defer e.lock.Unlock()
	n := time.Now()
	item := Item[T]{
		Value:     value,
		RecheckAt: n.Add(ttl1),
		ExpiresAt: n.Add(ttl2),
	}
	e.m[key] = item
	e.setRedis(key, item, ttl2)
}

func (e *Cache[T]) redisKey(key string) string {
	return fmt.Sprintf("ecache:%s:%s", e.topic, key)
}

func (e *Cache[T]) getRedis(key string) (Item[T], bool) {
	rctx, cc := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cc()
	t := time.Now()
	ld := Item[T]{
		ExpiresAt: t,
		RecheckAt: t,
	}
	ekey := e.redisKey(key)
	lastData := e.redis.Get(rctx, ekey)
	if err := lastData.Err(); err != nil {
		log.Trace().Err(err).Str("key", ekey).Msg("redis read failed")
		return ld, false
	}
	a, err := lastData.Bytes()
	if err != nil {
		log.Trace().Err(err).Str("key", ekey).Msg("redis read failed")
		return ld, false
	}
	if err := json.Unmarshal(a, &ld); err != nil {
		log.Trace().Err(err).Str("key", ekey).Msg("redis read failed during unmarshal")
	}
	log.Trace().Str("key", ekey).Msg("redis read ok")
	return ld, true
}

func (e *Cache[T]) setRedis(key string, item Item[T], ttl time.Duration) error {
	rctx, cc := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cc()
	ekey := e.redisKey(key)
	data, err := json.Marshal(item)
	if err != nil {
		log.Trace().Err(err).Str("key", ekey).Msg("redis write failed during marshal")
		return err
	}
	if err := e.redis.Set(rctx, ekey, data, ttl).Err(); err != nil {
		log.Trace().Err(err).Str("key", ekey).Msg("redis write failed")
	}
	log.Trace().Str("key", ekey).Msg("redis write ok")
	return nil
}
