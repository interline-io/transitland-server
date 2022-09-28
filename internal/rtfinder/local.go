package rtfinder

import (
	"sync"

	"github.com/interline-io/transitland-lib/rt/pb"
)

type LocalCache struct {
	lock    sync.Mutex
	sources map[string]*Source
}

func NewLocalCache() *LocalCache {
	return &LocalCache{
		sources: map[string]*Source{},
	}
}

func (f *LocalCache) GetSource(topic string) (*Source, bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	a, ok := f.sources[topic]
	if ok {
		return a, true
	}
	return nil, false
}

func (f *LocalCache) AddFeedMessage(topic string, rtmsg *pb.FeedMessage) error {
	return nil
}

func (f *LocalCache) AddData(topic string, data []byte) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	s, ok := f.sources[topic]
	if !ok {
		s, _ = NewSource(topic)
		f.sources[topic] = s
	}
	return s.process(data)
}

func (f *LocalCache) Close() error {
	return nil
}
