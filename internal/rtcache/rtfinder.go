package rtcache

import (
	"sync"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

type RTFinder struct {
	cache    Cache
	fetchers map[string]*rtConsumer
	lock     sync.Mutex
	*lookupCache
}

func NewRTFinder(cache Cache, db sqlx.Ext) *RTFinder {
	return &RTFinder{
		cache:       cache,
		lookupCache: newLookupCache(db),
		fetchers:    map[string]*rtConsumer{},
	}
}

func (f *RTFinder) AddData(topic string, data []byte) error {
	return f.cache.AddData(topic, data)
}

func (f *RTFinder) GetTrip(topic string, tid string) (*pb.TripUpdate, bool) {
	a, err := f.getListener(getTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil, false
	}
	trip, ok := a.GetTrip(tid)
	return trip, ok
}

func (f *RTFinder) FindAlerts(topic string, agencyId string, routeId string, routeType int, tripId string, tripDirection int, stopId string) []*model.Alert {
	a, err := f.getListener(getTopicKey(topic, "alerts"))
	if err != nil {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, alert := range a.alerts {
		for _, s := range alert.GetInformedEntity() {
			// fmt.Println("checking informed entity:", s)
			// fmt.Printf("filter: topic '%s' agency '%s' route '%s' route_type '%d' trip '%s' dir '%d' stop '%s'\n", topic, agencyId, routeId, routeType, tripId, tripDirection, stopId)
			found := true
			if (tripId != "" || s.Trip != nil) && s.Trip.GetTripId() != tripId {
				// fmt.Println("exclude trip")
				found = false
			}
			if s.AgencyId != nil && s.GetAgencyId() != agencyId {
				// fmt.Println("exclude agency")
				found = false
			}
			if s.RouteId != nil && s.GetRouteId() != routeId {
				// fmt.Println("exclude route")
				found = false
			}
			if s.StopId != nil && s.GetStopId() != stopId {
				// fmt.Println("exclude stop")
				found = false
			}
			if s.DirectionId != nil && int(s.GetDirectionId()) != tripDirection {
				// fmt.Println("exclude trip direction")
				found = false
			}
			// fmt.Println("found:", found)
			// TODO: route type
			if found {
				foundAlerts = append(foundAlerts, makeAlert(alert))
			}
		}
	}
	return foundAlerts
}

type tripAgencyRoute struct {
	RouteID  string
	AgencyID string
}

func (f *RTFinder) FindAlertsForTrip(t *model.Trip) []*model.Alert {
	topic, _ := f.GetFeedVersionOnestopID(t.FeedVersionID)
	v := tripAgencyRoute{}
	sqlx.Get(f.db, &v, "select r.route_id,a.agency_id from gtfs_routes r join gtfs_agencies a on a.id = r.agency_id where r.id = $1 limit 1", t.RouteID)
	return f.FindAlerts(
		topic,
		v.AgencyID,
		v.RouteID,
		0,
		t.TripID,
		t.DirectionID,
		"",
	)
}

func (f *RTFinder) FindStopTimeUpdate(topic string, tid string, sid string, seq int) (*pb.TripUpdate_StopTimeUpdate, bool) {
	rtTrip, rtok := f.GetTrip(topic, tid)
	if !rtok {
		return nil, false
	}
	for _, ste := range rtTrip.StopTimeUpdate {
		// Must match on StopSequence
		// TODO: allow matching on stop_id if stop_sequence is not provided
		if int(ste.GetStopSequence()) == seq {
			return ste, true
		}
	}
	return nil, false
}

// TODO: put this method on consumer and wrap, as with GetTrip
func (f *RTFinder) GetAddedTripsForStop(topic string, sid string) []*pb.TripUpdate {
	a, err := f.getListener(getTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil
	}
	// TODO: index more efficiently
	var ret []*pb.TripUpdate
	for _, trip := range a.entityByTrip {
		if trip.Trip.GetScheduleRelationship() != pb.TripDescriptor_ADDED {
			continue
		}
		for _, ste := range trip.StopTimeUpdate {
			if ste.GetStopId() == sid {
				ret = append(ret, trip)
				break // continue to next trip
			}
		}
	}
	return ret
}

func (f *RTFinder) getListener(topicKey string) (*rtConsumer, error) {
	f.lock.Lock()
	a, ok := f.fetchers[topicKey]
	if !ok {
		ch, err := f.cache.Listen(topicKey)
		// Failed to create listener
		if err != nil {
			// fmt.Printf("manager: '%s' failed to create listener\n", topicKey)
			return nil, err
		}
		// fmt.Printf("manager: '%s' listener created\n", topicKey)
		a, _ = newRTConsumer()
		a.feed = topicKey
		a.Start(ch)
		// fmt.Printf("manager: '%s' consumer started\n", topicKey)
		f.fetchers[topicKey] = a
	}
	f.lock.Unlock()
	return a, nil
}

func makeAlert(a *pb.Alert) *model.Alert {
	r := model.Alert{}
	r.Cause = pstr(a.Cause.String())
	r.Effect = pstr(a.Effect.String())
	for _, tr := range a.ActivePeriod {
		rttr := model.RTTimeRange{}
		if tr.Start != nil {
			v := int(*tr.Start)
			rttr.Start = &v
		}
		if tr.End != nil {
			v := int(*tr.End)
			rttr.Start = &v
		}
		r.ActivePeriod = append(r.ActivePeriod, &rttr)
	}
	r.HeaderText = newTranslation(a.HeaderText)
	r.DescriptionText = newTranslation(a.DescriptionText)
	r.TtsHeaderText = newTranslation(a.TtsHeaderText)
	r.TtsDescriptionText = newTranslation(a.TtsDescriptionText)
	r.URL = newTranslation(a.Url)
	r.SeverityLevel = pstr(a.SeverityLevel.String())
	return &r
}

func pstr(v string) *string {
	if v == "" {
		return nil
	}
	v2 := v
	return &v2
}

func newTranslation(v *pb.TranslatedString) []*model.RTTranslation {
	if v == nil {
		return nil
	}
	var ret []*model.RTTranslation
	for _, tr := range v.Translation {
		ntr := model.RTTranslation{
			Language: tr.Language,
		}
		if tr.Text != nil {
			ntr.Text = *tr.Text
		}
		ret = append(ret, &ntr)
	}
	return ret
}
