package rtcache

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

type RTFinder struct {
	cache    Cache
	fetchers map[string]*rtConsumer
	lock     sync.Mutex
	lc       *lookupCache
}

func NewRTFinder(cache Cache, db sqlx.Ext) *RTFinder {
	return &RTFinder{
		cache:    cache,
		lc:       newLookupCache(db),
		fetchers: map[string]*rtConsumer{},
	}
}

func (f *RTFinder) AddData(topic string, data []byte) error {
	return f.cache.AddData(topic, data)
}

func (f *RTFinder) GetGtfsTripID(id int) (string, bool) {
	return f.lc.GetGtfsTripID(id)
}

func (f *RTFinder) StopTimezone(id int, known string) (*time.Location, bool) {
	return f.lc.StopTimezone(id, known)
}

func (f *RTFinder) FindTrip(t *model.Trip) *pb.TripUpdate {
	topic, _ := f.lc.GetFeedVersionOnestopID(t.FeedVersionID)
	a, _ := f.getTrip(topic, t.TripID)
	return a
}

func (f *RTFinder) FindAlertsForTrip(t *model.Trip) []*model.Alert {
	topic, _ := f.lc.GetFeedVersionOnestopID(t.FeedVersionID)
	a, err := f.getListener(getTopicKey(topic, "alerts"))
	if err != nil {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, alert := range a.alerts {
		for _, s := range alert.GetInformedEntity() {
			if s.Trip == nil {
				continue
			}
			if s.Trip.GetTripId() == t.TripID {
				foundAlerts = append(foundAlerts, makeAlert(alert))
			}
		}
	}
	return foundAlerts
}

func (f *RTFinder) FindAlertsForRoute(t *model.Route) []*model.Alert {
	topic, _ := f.lc.GetFeedVersionOnestopID(t.FeedVersionID)
	a, err := f.getListener(getTopicKey(topic, "alerts"))
	if err != nil {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, alert := range a.alerts {
		for _, s := range alert.GetInformedEntity() {
			if s.Trip != nil {
				continue
			}
			if s.GetRouteId() == t.RouteID {
				foundAlerts = append(foundAlerts, makeAlert(alert))
			}
		}
	}
	return foundAlerts
}

func (f *RTFinder) FindAlertsForAgency(t *model.Agency) []*model.Alert {
	topic, _ := f.lc.GetFeedVersionOnestopID(t.FeedVersionID)
	a, err := f.getListener(getTopicKey(topic, "alerts"))
	if err != nil {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, alert := range a.alerts {
		for _, s := range alert.GetInformedEntity() {
			if s.Trip != nil {
				continue
			}
			if s.GetAgencyId() == t.AgencyID {
				foundAlerts = append(foundAlerts, makeAlert(alert))
			}
		}
	}
	return foundAlerts
}

func (f *RTFinder) FindAlertsForStop(t *model.Stop) []*model.Alert {
	topic, _ := f.lc.GetFeedVersionOnestopID(t.FeedVersionID)
	a, err := f.getListener(getTopicKey(topic, "alerts"))
	if err != nil {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, alert := range a.alerts {
		for _, s := range alert.GetInformedEntity() {
			if s.StopId == nil {
				continue
			}
			if s.GetStopId() == t.StopID {
				foundAlerts = append(foundAlerts, makeAlert(alert))
			}
		}
	}
	return foundAlerts
}

func (f *RTFinder) FindStopTimeUpdate(t *model.Trip, st *model.StopTime) (*pb.TripUpdate_StopTimeUpdate, bool) {
	topic, _ := f.lc.GetFeedVersionOnestopID(t.FeedVersionID)
	tid := t.TripID
	seq := st.StopSequence
	rtTrip, rtok := f.getTrip(topic, tid)
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
func (f *RTFinder) GetAddedTripsForStop(t *model.Stop) []*pb.TripUpdate {
	topic, _ := f.lc.GetFeedVersionOnestopID(t.FeedVersionID)
	sid := t.StopID
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

func (f *RTFinder) MakeTrip(obj *model.Trip) (*model.Trip, error) {
	t := model.Trip{}
	t.FeedVersionID = obj.FeedVersionID
	t.TripID = obj.TripID
	t.RTTripID = obj.RTTripID
	if rtTrip := f.FindTrip(&t); rtTrip != nil {
		rtt := rtTrip.Trip
		rid, _ := f.lc.GetRouteID(obj.FeedVersionID, rtt.GetRouteId())
		t.RouteID = strconv.Itoa(rid)
		t.DirectionID = int(rtt.GetDirectionId())
		return &t, nil
	}
	return nil, errors.New("not found")
}

func (f *RTFinder) getTrip(topic string, tid string) (*pb.TripUpdate, bool) {
	if tid == "" {
		panic("no tid")
	}
	a, err := f.getListener(getTopicKey(topic, "trip_updates"))
	if err != nil {
		return nil, false
	}
	trip, ok := a.GetTrip(tid)
	return trip, ok
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
