package rtcache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type listenChan struct {
	listener chan []byte
	done     chan bool
}

func newListenChan() *listenChan {
	return &listenChan{
		listener: make(chan []byte, 100),
		done:     make(chan bool),
	}
}

type RedisCache struct {
	client    *redis.Client
	listeners []*listenChan
}

func NewRedisCache(client *redis.Client) *RedisCache {

	f := RedisCache{
		client: client,
	}
	return &f
}

func (f *RedisCache) Listen(topic string) (chan []byte, error) {
	lch := newListenChan()
	f.listeners = append(f.listeners, lch)
	// Add subscription for future data
	go func(client *redis.Client, ch *listenChan) {
		defer close(ch.listener)
		// get the first message and send
		lastData := client.Get(context.TODO(), lastKey(topic))
		if err := lastData.Err(); err == redis.Nil {
			ch.listener <- nil
		} else if err != nil {
			panic(err)
		} else {
			lb, _ := lastData.Bytes()
			ch.listener <- lb
		}
		// subscribe for future updates
		sub := client.Subscribe(context.TODO(), subKey(topic))
		defer sub.Close()
		subch := sub.Channel()
		for {
			select {
			case <-lch.done:
				fmt.Printf("cache '%s': done\n", topic)
				return
			case rmsg := <-subch:
				fmt.Printf("cache '%s': sending %d bytes\n", topic, len(rmsg.Payload))
				b := []byte(rmsg.Payload)
				ch.listener <- b
			}
		}
	}(f.client, lch)
	fmt.Printf("cache: '%s' listener created\n", topic)
	return lch.listener, nil
}

func (f *RedisCache) AddData(topic string, data []byte) error {
	// Set last seen value with 5 min ttl
	if err := f.client.Set(context.TODO(), lastKey(topic), data, 5*time.Minute).Err(); err != nil {
		return err
	}
	// Publish to subscribers
	if err := f.client.Publish(context.TODO(), subKey(topic), data).Err(); err != nil {
		return err
	}
	fmt.Printf("cache '%s': added %d bytes\n", topic, len(data))
	return nil
}

func (f *RedisCache) Close() error {
	for _, ch := range f.listeners {
		fmt.Println("closing ch:", ch)
		ch.done <- true
	}
	return nil
}

func lastKey(topic string) string {
	return fmt.Sprintf("rtfetch:last:%s", topic)
}

func subKey(topic string) string {
	return fmt.Sprintf("rtfetch:sub:%s", topic)
}
