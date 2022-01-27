package rtcache

import (
	"errors"
	"fmt"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"google.golang.org/protobuf/proto"
)

type rtConsumer struct {
	feed         string
	done         chan bool
	entityByTrip map[string]*pb.TripUpdate
	alerts       []*pb.Alert
}

func newRTConsumer() (*rtConsumer, error) {
	f := rtConsumer{
		done:         make(chan bool),
		entityByTrip: map[string]*pb.TripUpdate{},
	}
	return &f, nil
}

func (f *rtConsumer) GetTrip(tid string) (*pb.TripUpdate, bool) {
	// fmt.Printf("consumer '%s': get trip '%s'\n", f.feed, tid)
	a, ok := f.entityByTrip[tid]
	if ok {
		return a, true
	}
	return nil, false
}

func (f *rtConsumer) Start(ch chan []byte) error {
	// fmt.Printf("consumer '%s': start\n", f.feed)
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
				// fmt.Printf("consumer '%s': done\n", f.feed)
				return
			case rtdata := <-ch:
				// fmt.Printf("consumer '%s': received %d bytes\n", f.feed, len(rtdata))
				if err := f.process(rtdata); err != nil {
					fmt.Println("error processing rt data")
				}
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
		// fmt.Printf("consumer '%s': timeout waiting for first entity\n", f.feed)
		return errors.New("timeout waiting for first entity")
	case <-ready:
		// fmt.Printf("consumer '%s': ready!\n", f.feed)
		return nil
	}
}

func (f *rtConsumer) process(rtdata []byte) error {
	if len(rtdata) == 0 {
		// fmt.Printf("consumer '%s': received no data\n", f.feed)
		return nil
	}
	rtmsg := pb.FeedMessage{}
	if err := proto.Unmarshal(rtdata, &rtmsg); err != nil {
		return err
	}
	defaultTimestamp := rtmsg.GetHeader().GetTimestamp()
	a := map[string]*pb.TripUpdate{}
	for _, ent := range rtmsg.Entity {
		if v := ent.TripUpdate; v != nil {
			// Set default timestamp
			if v.Timestamp == nil {
				v.Timestamp = &defaultTimestamp
			}
			tid := v.GetTrip().GetTripId()
			a[tid] = v
		}
		if v := ent.Alert; v != nil {
			f.alerts = append(f.alerts, v)
		}
		// todo: vehicle positions...
	}
	// fmt.Printf("consumer '%s': processed trips: %s\n", f.feed, strings.Join(tids, ","))
	f.entityByTrip = a
	return nil
}
