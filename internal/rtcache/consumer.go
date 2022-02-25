package rtcache

import (
	"errors"
	"time"

	"github.com/interline-io/transitland-lib/log"
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
	log.Debug().Str("feed_id", f.feed).Str("trip", tid).Msg("consumer: get trip")
	a, ok := f.entityByTrip[tid]
	if ok {
		return a, true
	}
	return nil, false
}

func (f *rtConsumer) Start(ch chan []byte) error {
	log.Debug().Str("feed_id", f.feed).Msg("consumer: start")
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
				log.Debug().Str("feed_id", f.feed).Msg("consumer: done")
				return
			case rtdata := <-ch:
				log.Debug().Str("feed_id", f.feed).Int("bytes", len(rtdata)).Msg("consumer: received data")
				if err := f.process(rtdata); err != nil {
					log.Error().Err(err).Str("feed_id", f.feed).Msg("consumer: error processing rt data")
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
		log.Debug().Str("feed_id", f.feed).Msg("consumer: timed out waiting for first entity")
		return errors.New("timeout waiting for first entity")
	case <-ready:
		log.Debug().Str("feed_id", f.feed).Msg("consumer: ready")
		return nil
	}
}

func (f *rtConsumer) process(rtdata []byte) error {
	if len(rtdata) == 0 {
		log.Debug().Str("feed_id", f.feed).Msg("consumer: received no data")
		return nil
	}
	rtmsg := pb.FeedMessage{}
	if err := proto.Unmarshal(rtdata, &rtmsg); err != nil {
		return err
	}
	defaultTimestamp := rtmsg.GetHeader().GetTimestamp()
	a := map[string]*pb.TripUpdate{}
	var alerts []*pb.Alert
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
			alerts = append(alerts, v)
		}
		// todo: vehicle positions...
	}
	log.Debug().Str("feed_id", f.feed).Int("trip_updates", len(a)).Int("alerts", len(alerts)).Msg("consumer: processed trips")
	f.entityByTrip = a
	f.alerts = alerts
	return nil
}
