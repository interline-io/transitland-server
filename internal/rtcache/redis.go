package rtcache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/rt/pb"
)

type listener struct {
	source *Source
	ctx    context.Context
	cancel context.CancelFunc
}

func newListener(s *Source, parent context.Context) *listener {
	cc, cf := context.WithCancel(parent)
	return &listener{
		source: s,
		ctx:    cc,
		cancel: cf,
	}
}

type RedisCache struct {
	ctx       context.Context
	lock      sync.Mutex
	client    *redis.Client
	listeners map[string]*listener
}

func NewRedisCache(client *redis.Client) *RedisCache {
	f := RedisCache{
		client:    client,
		listeners: map[string]*listener{},
		ctx:       context.Background(),
	}
	return &f
}

func (f *RedisCache) GetSource(topic string) (*Source, bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if a, ok := f.listeners[topic]; ok {
		return a.source, true
	}
	a, err := f.startListener(topic)
	if err != nil {
		return nil, false
	}
	f.listeners[topic] = a
	return a.source, false
}

func (f *RedisCache) AddFeedMessage(topic string, rtmsg *pb.FeedMessage) error {
	return nil
}

func (f *RedisCache) AddData(topic string, data []byte) error {
	rctx, cc := context.WithTimeout(f.ctx, 5*time.Second)
	defer cc()
	// Set last seen value with 5 min ttl
	if err := f.client.Set(rctx, lastKey(topic), data, 5*time.Minute).Err(); err != nil {
		return err
	}
	// Publish to subscribers
	if err := f.client.Publish(rctx, subKey(topic), data).Err(); err != nil {
		return err
	}
	log.Trace().Str("topic", topic).Int("bytes", len(data)).Msg("cache: added data")
	return nil
}

func (f *RedisCache) Close() error {
	f.lock.Lock()
	defer f.lock.Unlock()
	for k, ls := range f.listeners {
		ls.cancel()
		delete(f.listeners, k)
	}
	return nil
}

func lastKey(topic string) string {
	return fmt.Sprintf("rtfetch:last:%s", topic)
}

func subKey(topic string) string {
	return fmt.Sprintf("rtfetch:sub:%s", topic)
}

func (f *RedisCache) startListener(topic string) (*listener, error) {
	// Create new source
	s, err := NewSource(topic)
	if err != nil {
		return nil, err
	}
	// Add subscription for future data
	ls := newListener(s, f.ctx)
	go func(client *redis.Client, topic string, lst *listener) {
		sub := client.Subscribe(lst.ctx, subKey(topic))
		defer sub.Close()
		subch := sub.Channel()
		for rmsg := range subch {
			if err := s.process([]byte(rmsg.Payload)); err != nil {
				log.Error().Err(err).Str("topic", topic).Int("bytes", len(rmsg.Payload)).Msg("cache: error processing update")
			} else {
				log.Trace().Str("topic", topic).Int("bytes", len(rmsg.Payload)).Msg("cache: processed update")
			}
		}
	}(f.client, topic, ls)
	log.Trace().Str("topic", topic).Msgf("cache: listener created")
	// get the first message
	rctx, cc := context.WithTimeout(f.ctx, 1*time.Second)
	defer cc()
	lastData := f.client.Get(rctx, lastKey(topic))
	if err := lastData.Err(); err == redis.Nil {
		// ok
	} else if err != nil {
		// also ok, hope we get data on future updates
		log.Error().Err(err).Str("topic", topic).Msg("cache: error getting last data for topic")
	} else {
		lb, _ := lastData.Bytes()
		s.process(lb)
	}
	return ls, nil
}
