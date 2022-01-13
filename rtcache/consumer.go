package rtcache

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"google.golang.org/protobuf/proto"
)

type RTConsumerManager struct {
	cache    Cache
	fetchers map[string]*RTConsumer
	lock     sync.Mutex
}

func NewRTConsumerManager(cache Cache) *RTConsumerManager {
	return &RTConsumerManager{
		cache:    cache,
		fetchers: map[string]*RTConsumer{},
	}
}

func (f *RTConsumerManager) getListener(topic string) (*RTConsumer, error) {
	f.lock.Lock()
	a, ok := f.fetchers[topic]
	if !ok {
		ch, err := f.cache.Listen(topic)
		// Failed to create listener
		if err != nil {
			fmt.Printf("manager: '%s' failed to create listener\n", topic)
			return nil, err
		}
		fmt.Printf("manager: '%s' listener created\n", topic)
		a, _ = NewRTConsumer()
		a.feed = topic
		a.Start(ch)
		fmt.Printf("manager: '%s' consumer started\n", topic)
		f.fetchers[topic] = a
	}
	f.lock.Unlock()
	return a, nil
}

func (f *RTConsumerManager) GetTripIDs(topic string) ([]string, error) {
	var ret []string
	a, err := f.getListener(topic)
	if err != nil {
		return nil, err
	}
	for k := range a.entityByTrip {
		ret = append(ret, k)
	}
	return ret, nil
}

func (f *RTConsumerManager) GetTrip(topic string, tid string) (*pb.TripUpdate, bool) {
	a, err := f.getListener(topic)
	if err != nil {
		return nil, false
	}
	trip, ok := a.GetTrip(tid)
	return trip, ok
}

/////////

type RTConsumer struct {
	feed         string
	done         chan bool
	entityByTrip map[string]*pb.TripUpdate
}

func NewRTConsumer() (*RTConsumer, error) {
	f := RTConsumer{
		done:         make(chan bool),
		entityByTrip: map[string]*pb.TripUpdate{},
	}
	return &f, nil
}

func (f *RTConsumer) GetTrip(tid string) (*pb.TripUpdate, bool) {
	fmt.Printf("consumer '%s': get trip '%s'\n", f.feed, tid)
	a, ok := f.entityByTrip[tid]
	if ok {
		return a, true
	}
	return nil, false
}

func (f *RTConsumer) Start(ch chan []byte) error {
	fmt.Printf("consumer '%s': start\n", f.feed)
	f.entityByTrip = map[string]*pb.TripUpdate{}
	timeout := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	ready := make(chan bool)
	go func() {
		for {
			select {
			case <-f.done:
				fmt.Printf("consumer '%s': done\n", f.feed)
				return
			case rtdata := <-ch:
				fmt.Printf("consumer '%s': received %d bytes\n", f.feed, len(rtdata))
				f.process(rtdata)
				if ready != nil {
					ready <- true
					close(ready)
					ready = nil
				}
			}
		}
	}()
	// wait for first entity
	select {
	case <-timeout:
		fmt.Printf("consumer '%s': timeout waiting for first entity\n", f.feed)
		return errors.New("timeout waiting for first entity")
	case <-ready:
		fmt.Printf("consumer '%s': ready!\n", f.feed)
		return nil
	}
}

func (f *RTConsumer) process(rtdata []byte) error {
	rtmsg := pb.FeedMessage{}
	if err := proto.Unmarshal(rtdata, &rtmsg); err != nil {
		return err
	}
	a := map[string]*pb.TripUpdate{}
	tids := []string{}
	for _, ent := range rtmsg.Entity {
		if v := ent.TripUpdate; v != nil {
			tid := v.GetTrip().GetTripId()
			tids = append(tids, tid)
			a[tid] = v
		}
		// todo: handle alerts and vehicle positions...
	}
	fmt.Printf("consumer '%s': processed trips: %s\n", f.feed, strings.Join(tids, ","))
	f.entityByTrip = a
	return nil
}
