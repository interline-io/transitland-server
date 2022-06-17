package rtcache

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

// Cache provides a method for looking up and listening for changed RT data
type Cache interface {
	AddFeedMessage(string, *pb.FeedMessage) error
	AddData(string, []byte) error
	GetSource(string) (*Source, bool)
	Close() error
}

////////

type RTFinder struct {
	cache Cache
	lc    *lookupCache
}

func NewRTFinder(cache Cache, db sqlx.Ext) *RTFinder {
	return &RTFinder{
		cache: cache,
		lc:    newLookupCache(db),
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
	topics, ok := f.lc.GetFeedVersionRTFeeds(t.FeedVersionID)
	if !ok {
		return nil
	}
	for _, topic := range topics {
		if a, ok := f.getTrip(topic, t.TripID); ok {
			return a
		}
	}
	return nil
}

func (f *RTFinder) FindAlertsForTrip(t *model.Trip) []*model.Alert {
	topics, ok := f.lc.GetFeedVersionRTFeeds(t.FeedVersionID)
	if !ok {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, topic := range topics {
		a, ok := f.cache.GetSource(getTopicKey(topic, "realtime_alerts"))
		if !ok {
			return nil
		}
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
	}
	return foundAlerts
}

func (f *RTFinder) FindAlertsForRoute(t *model.Route) []*model.Alert {
	topics, ok := f.lc.GetFeedVersionRTFeeds(t.FeedVersionID)
	if !ok {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, topic := range topics {
		a, ok := f.cache.GetSource(getTopicKey(topic, "realtime_alerts"))
		if !ok {
			continue
		}
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
	}
	return foundAlerts
}

func (f *RTFinder) FindAlertsForAgency(t *model.Agency) []*model.Alert {
	topics, ok := f.lc.GetFeedVersionRTFeeds(t.FeedVersionID)
	if !ok {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, topic := range topics {
		a, ok := f.cache.GetSource(getTopicKey(topic, "realtime_alerts"))
		if !ok {
			continue
		}
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
	}
	return foundAlerts
}

func (f *RTFinder) FindAlertsForStop(t *model.Stop) []*model.Alert {
	topics, ok := f.lc.GetFeedVersionRTFeeds(t.FeedVersionID)
	if !ok {
		return nil
	}
	var foundAlerts []*model.Alert
	for _, topic := range topics {
		a, ok := f.cache.GetSource(getTopicKey(topic, "realtime_alerts"))
		if !ok {
			continue
		}
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
	}
	return foundAlerts
}

func (f *RTFinder) FindStopTimeUpdate(t *model.Trip, st *model.StopTime) (*pb.TripUpdate_StopTimeUpdate, bool) {
	topics, ok := f.lc.GetFeedVersionRTFeeds(t.FeedVersionID)
	if !ok {
		return nil, false
	}
	tid := t.TripID
	seq := st.StopSequence
	for _, topic := range topics {
		rtTrip, rtok := f.getTrip(topic, tid)
		if !rtok {
			continue
		}
		// Match on stop sequence
		for _, ste := range rtTrip.StopTimeUpdate {
			if int(ste.GetStopSequence()) == seq {
				log.Trace().Str("trip_id", t.TripID).Int("seq", seq).Msgf("found stop time update on trip_id/stop_sequence")
				return ste, true
			}
		}
		// If no match on stop sequence, match on stop_id if stop is not visited twice
		check := map[string]int{}
		for _, ste := range rtTrip.StopTimeUpdate {
			check[ste.GetStopId()] += 1
		}
		sid, ok := f.lc.GetGtfsStopID(atoi(st.StopID))
		if !ok {
			continue
		}
		for _, ste := range rtTrip.StopTimeUpdate {
			stid := ste.GetStopId()
			if sid == stid && check[stid] == 1 {
				log.Trace().Str("trip_id", t.TripID).Str("stop_id", sid).Msgf("found stop time update on trip_id/stop_id")
				return ste, true
			}
		}
	}
	log.Trace().Str("trip_id", t.TripID).Int("seq", seq).Msgf("no stop time update found")
	return nil, false
}

// TODO: put this method on consumer and wrap, as with GetTrip
func (f *RTFinder) GetAddedTripsForStop(t *model.Stop) []*pb.TripUpdate {
	sid := t.StopID
	topics, ok := f.lc.GetFeedVersionRTFeeds(t.FeedVersionID)
	if !ok {
		return nil
	}
	var ret []*pb.TripUpdate
	for _, topic := range topics {
		a, ok := f.cache.GetSource(getTopicKey(topic, "realtime_trip_updates"))
		if !ok {
			continue
		}
		// TODO: index more efficiently
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
		rid, ok := f.lc.GetRouteID(obj.FeedVersionID, rtt.GetRouteId())
		if !ok {
			return nil, errors.New("not found")
		}
		t.RouteID = strconv.Itoa(rid)
		t.DirectionID = int(rtt.GetDirectionId())
		return &t, nil
	}
	return nil, errors.New("not found")
}

func (f *RTFinder) getTrip(topic string, tid string) (*pb.TripUpdate, bool) {
	if tid == "" {
		return nil, false
	}
	a, ok := f.cache.GetSource(getTopicKey(topic, "realtime_trip_updates"))
	if !ok {
		return nil, false
	}
	trip, ok := a.GetTrip(tid)
	return trip, ok
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

func atoi(v string) int {
	a, _ := strconv.Atoi(v)
	return a
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

func getTopicKey(topic string, t string) string {
	return fmt.Sprintf("rtdata:%s:%s", topic, t)
}
