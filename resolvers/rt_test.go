package resolvers

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Additional tests for RT data on StopResolver
var baseStopQuery = `query($stop_id:String!, $stf:StopTimeFilter!) {
	stops(where: { stop_id: $stop_id }) {
	  id
	  stop_id
	  stop_name
	  stop_times(where:$stf) {
		stop_sequence
		trip {
		  alerts {
			cause
			effect
			severity_level
			url {
				language
				text
			}
			header_text {
				language
				text
			}
			description_text {
				language
				text
			}
			tts_header_text {
				language
				text
			}
			tts_description_text {
				language
				text
			}
			active_period {
				start
				end
			}
		  }
		  trip_id
		  schedule_relationship
		  timestamp
		  route {
			  route_id
			  route_short_name
			  route_long_name
			  agency {
				  agency_id
				  agency_name
			  }
		  }
		}
		arrival {
			scheduled
			estimated
			estimated_utc
			stop_timezone
			delay
			uncertainty
		}
		departure {
			scheduled
			estimated
			estimated_utc
			stop_timezone
			delay
			uncertainty
		}
	  }
	}
  }
`

var baseStopVars = hw{
	"service_date": "2018-05-30",
	"start_time":   57600,
	"end_time":     57900,
}

var baseRTFiles = []rtFile{
	{"BA", "trip_updates", "BA.json"},
	{"CT", "trip_updates", "CT.json"},
}

// rtFetchJson fetches test protobuf in JSON format
// URL is relative to project root
func rtFetchJson(feed string, ftype string, url string, rtfinder model.RTFinder) error {
	var msg pb.FeedMessage
	jdata, err := ioutil.ReadFile(url)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(jdata, &msg); err != nil {
		return err
	}
	rtdata, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("rtdata:%s:%s", feed, ftype)
	return rtfinder.AddData(key, rtdata)
}

type rtTestCase struct {
	name    string
	query   string
	vars    map[string]interface{}
	rtfiles []rtFile
	cb      func(t *testing.T, jj string)
}

type rtFile struct {
	feed  string
	ftype string
	fname string
}

func testRt(t *testing.T, tc rtTestCase) {
	cfg := config.Config{}
	srv, _ := NewServer(cfg, TestDBFinder, TestRTFinder)
	c := client.New(srv)
	for _, rtf := range tc.rtfiles {
		if err := rtFetchJson(rtf.feed, rtf.ftype, testutil.RelPath("test", "data", "rt", rtf.fname), TestRTFinder); err != nil {
			t.Fatal(err)
		}
	}
	var resp map[string]interface{}
	opts := []client.Option{}
	for k, v := range tc.vars {
		opts = append(opts, client.Var(k, v))
	}
	c.MustPost(tc.query, &resp, opts...)
	jj := toJson(resp)
	if tc.cb != nil {
		tc.cb(t, jj)
	}
}

func TestStopRTBasic(t *testing.T) {
	tc := rtTestCase{
		"stop times basic",
		baseStopQuery,
		hw{"stop_id": "FTVL", "stf": baseStopVars},
		baseRTFiles,
		func(t *testing.T, jj string) {
			// A little more explicit version of the string check test
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			delay := 30
			assert.Equal(t, 3, len(a))
			for _, st := range a {
				assert.Equal(t, "America/Los_Angeles", st.Get("arrival.stop_timezone").String(), "arrival.stop_timezone")
				assert.Equal(t, delay, int(st.Get("arrival.delay").Int()), "arrival.delay")
				assert.Equal(t, "America/Los_Angeles", st.Get("departure.stop_timezone").String(), "departure.stop_timezone")
				assert.Equal(t, delay, int(st.Get("departure.delay").Int()), "departure.delay")
				sched, _ := tl.NewWideTime(st.Get("arrival.scheduled").String())
				est, _ := tl.NewWideTime(st.Get("arrival.estimated").String())
				assert.Equal(t, sched.Seconds+int(delay), est.Seconds, "arrival.scheduled + delay = arrival.estimated for this test")
			}
			checkTrip := "1031527WKDY"
			found := false
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, checkTrip, st.Get("trip.trip_id").String())
				assert.Equal(t, "2018-05-30T22:27:30Z", st.Get("trip.timestamp").String())
				assert.Equal(t, "2018-05-30T23:02:30Z", st.Get("arrival.estimated_utc").String())
				assert.Equal(t, "2018-05-30T23:02:30Z", st.Get("departure.estimated_utc").String())
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

func TestStopRTAddedTrip(t *testing.T) {
	tc := rtTestCase{
		"stop times added trip",
		baseStopQuery,
		hw{"stop_id": "FTVL", "stf": baseStopVars},
		[]rtFile{{"BA", "trip_updates", "BA-added.json"}},
		func(t *testing.T, jj string) {
			checkTrip := "-123"
			found := false
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			assert.Equal(t, 4, len(a))
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, checkTrip, st.Get("trip.trip_id").String(), "trip.trip_id")
				assert.Equal(t, "05", st.Get("trip.route.route_id").String(), "trip.route.route_id")
				assert.Equal(t, "ADDED", st.Get("trip.schedule_relationship").String(), "trip.schedule_relationship")
				assert.Equal(t, "", st.Get("arrival.scheduled").String(), "arrival.scheduled")
				assert.Equal(t, "", st.Get("departure.scheduled").String(), "departure.scheduled")
				assert.Equal(t, "2018-05-30T23:02:32Z", st.Get("arrival.estimated_utc").String(), "arrival.estimated_utc")
				assert.Equal(t, "2018-05-30T23:02:32Z", st.Get("departure.estimated_utc").String(), "departure.estimated_utc")
				assert.Equal(t, 12, int(st.Get("stop_sequence").Int()), "stop_sequence")
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

func TestStopRTCanceledTrip(t *testing.T) {
	tc := rtTestCase{
		"stop times canceled trip",
		baseStopQuery,
		hw{"stop_id": "FTVL", "stf": baseStopVars},
		[]rtFile{{"BA", "trip_updates", "BA-added.json"}},
		func(t *testing.T, jj string) {
			checkTrip := "2211533WKDY"
			found := false
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			assert.Equal(t, 4, len(a))
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, checkTrip, st.Get("trip.trip_id").String(), "trip.trip_id")
				assert.Equal(t, "03", st.Get("trip.route.route_id").String(), "trip.route.route_id")
				assert.Equal(t, "CANCELED", st.Get("trip.schedule_relationship").String(), "trip.schedule_relationship")
				assert.Equal(t, "16:02:00", st.Get("arrival.scheduled").String(), "arrival.scheduled")
				assert.Equal(t, "16:02:00", st.Get("departure.scheduled").String(), "departure.scheduled")
				assert.Equal(t, "", st.Get("arrival.estimated_utc").String(), "arrival.estimated_utc")
				assert.Equal(t, "", st.Get("departure.estimated_utc").String(), "departure.estimated_utc")
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

func TestTripAlerts(t *testing.T) {
	tc := rtTestCase{
		"trip alerts",
		baseStopQuery,
		hw{"stop_id": "FTVL", "stf": baseStopVars},
		[]rtFile{{"BA", "trip_updates", "BA-added.json"}, {"BA", "alerts", "BA-alerts.json"}},
		func(t *testing.T, jj string) {
			checkTrip := "2211533WKDY"
			found := false
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				alerts := st.Get("trip.alerts").Array()
				if len(alerts) == 1 {
					firstAlert := alerts[0]
					assert.Equal(t, "UNKNOWN_CAUSE", firstAlert.Get("cause").String(), "cause")
					assert.Equal(t, "UNKNOWN_EFFECT", firstAlert.Get("effect").String(), "effect")
					assert.Equal(t, "UNKNOWN_SEVERITY", firstAlert.Get("severity_level").String(), "severity_level")
					assert.Equal(t, "Test trip header", firstAlert.Get("header_text.0.text").String(), "header_text.0.text")
					assert.Equal(t, "Test trip description", firstAlert.Get("description_text.0.text").String(), "description_text.0.text")
				} else {
					t.Error("expected exactly 1 alert")
				}
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

// TODO
// TestStopAlerts
// TestRouteAlerts
// TestAgencyAlerts
