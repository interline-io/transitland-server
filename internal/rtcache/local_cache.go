package rtcache

import (
	"sync"
)

type LocalCache struct {
	lock      sync.Mutex
	lastData  map[string][]byte
	listeners map[string][]chan []byte
}

func NewLocalCache() *LocalCache {
	return &LocalCache{
		lastData:  map[string][]byte{},
		listeners: make(map[string][]chan []byte),
	}
}

func (f *LocalCache) Listen(topic string) (chan []byte, error) {
	c := make(chan []byte, 1000)
	c <- f.lastData[topic]
	f.lock.Lock()
	f.listeners[topic] = append(f.listeners[topic], c)
	f.lock.Unlock()
	// fmt.Printf("cache: '%s' listener created\n", topic)
	return c, nil
}

func (f *LocalCache) AddData(topic string, data []byte) error {
	f.lock.Lock()
	f.lastData[topic] = data
	f.lock.Unlock()
	for _, c := range f.listeners[topic] {
		c <- data
	}
	// fmt.Printf("cache: '%s' added %d bytes\n", topic, len(data))
	return nil
}

func (f *LocalCache) Close() error {
	return nil
}