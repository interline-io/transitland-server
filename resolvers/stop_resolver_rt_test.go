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

func TestStopRTResolver(t *testing.T) {
	baseQuery := `query($stop_id:String!, $stf:StopTimeFilter!) {
		stops(where: { stop_id: $stop_id }) {
		  id
		  stop_id
		  stop_name
		  stop_times(where:$stf) {
			stop_sequence
			trip {
			  trip_id
			  schedule_relationship
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

	type rtFile struct {
		feed  string
		ftype string
		fname string
	}

	baseSTVars := hw{"service_date": "2018-05-30", "start_time": 57600, "end_time": 57900}
	baseRTFiles := []rtFile{
		{"BA", "trip_updates", "BA.json"},
		{"CT", "trip_updates", "CT.json"},
	}
	testcases := []struct {
		name    string
		query   string
		vars    map[string]interface{}
		rtfiles []rtFile
		cb      func(t *testing.T, jj string)
	}{
		{
			"basic",
			baseQuery,
			hw{"stop_id": "FTVL", "stf": baseSTVars},
			baseRTFiles,
			func(t *testing.T, jj string) {
				// A little more explicit version of the string check test
				a := gjson.Get(jj, "stops.0.stop_times").Array()
				delay := 30
				assert.Equal(t, 3, len(a))
				for _, st := range a {
					assert.Equal(t, "America/Los_Angeles", st.Get("arrival.stop_timezone").String())
					assert.Equal(t, delay, int(st.Get("arrival.delay").Int()))
					assert.Equal(t, "America/Los_Angeles", st.Get("departure.stop_timezone").String())
					assert.Equal(t, delay, int(st.Get("departure.delay").Int()))
					sched, _ := tl.NewWideTime(st.Get("arrival.scheduled").String())
					est, _ := tl.NewWideTime(st.Get("arrival.estimated").String())
					assert.Equal(t, sched.Seconds+int(delay), est.Seconds)
				}
				checkTrip := "1011630WKDY"
				found := false
				for _, st := range a {
					if st.Get("trip.trip_id").String() != checkTrip {
						found = true
						continue
					}
					assert.Equal(t, checkTrip, st.Get("trip.trip_id").String())
					assert.Equal(t, "2018-05-31T00:05:30Z", st.Get("arrival.estimated_utc").String())
					assert.Equal(t, "2018-05-31T00:05:30Z", st.Get("departure.estimated_utc").String())
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
		{
			"added",
			baseQuery,
			hw{"stop_id": "FTVL", "stf": baseSTVars},
			[]rtFile{{"BA", "trip_updates", "BA-added.json"}},
			func(t *testing.T, jj string) {
				checkTrip := "-123"
				found := false
				a := gjson.Get(jj, "stops.0.stop_times").Array()
				assert.Equal(t, 4, len(a))
				for _, st := range a {
					if st.Get("trip.trip_id").String() != checkTrip {
						found = true
						continue
					}
					assert.Equal(t, checkTrip, st.Get("trip.trip_id").String())
					assert.Equal(t, "05", st.Get("trip.route.route_id").String())
					assert.Equal(t, "ADDED", st.Get("trip.schedule_relationship").String())
					assert.Equal(t, "", st.Get("arrival.scheduled").String())
					assert.Equal(t, "", st.Get("departure.scheduled").String())
					assert.Equal(t, "2018-05-30T23:02:32Z", st.Get("arrival.estimated_utc").String())
					assert.Equal(t, "2018-05-30T23:02:32Z", st.Get("departure.estimated_utc").String())
					assert.Equal(t, 12, int(st.Get("stop_sequence").Int()))
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
		{
			"canceled",
			baseQuery,
			hw{"stop_id": "FTVL", "stf": baseSTVars},
			[]rtFile{{"BA", "trip_updates", "BA-added.json"}},
			func(t *testing.T, jj string) {
				checkTrip := "2211533WKDY"
				found := false
				a := gjson.Get(jj, "stops.0.stop_times").Array()
				assert.Equal(t, 4, len(a))
				for _, st := range a {
					if st.Get("trip.trip_id").String() != checkTrip {
						found = true
						continue
					}
					assert.Equal(t, checkTrip, st.Get("trip.trip_id").String())
					assert.Equal(t, "03", st.Get("trip.route.route_id").String())
					assert.Equal(t, "CANCELED", st.Get("trip.schedule_relationship").String())
					assert.Equal(t, "16:02:00", st.Get("arrival.scheduled").String())
					assert.Equal(t, "16:02:00", st.Get("departure.scheduled").String())
					assert.Equal(t, "", st.Get("arrival.estimated_utc").String())
					assert.Equal(t, "", st.Get("departure.estimated_utc").String())
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}
