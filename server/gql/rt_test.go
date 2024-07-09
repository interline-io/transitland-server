package gql

import (
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// Additional tests for RT data on StopResolver
var baseStopQuery = `
fragment alert on Alert {
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

query($stop_id:String!, $stf:StopTimeFilter!, $active:Boolean) {
	stops(where: { stop_id: $stop_id }) {
	  id
	  stop_id
	  stop_name
	  alerts(active:$active, limit:5) {
		  ...alert
	  }
	  stop_times(where:$stf) {
		stop_sequence
		trip {
		  alerts(active:$active) {
			...alert
		  }
		  trip_id
		  schedule_relationship
		  timestamp
		  route {
			  route_id
			  route_short_name
			  route_long_name
			  alerts(active:$active) {
				  ...alert
			  }
			  agency {
				  agency_id
				  agency_name
				  alerts(active:$active) {
					  ...alert
				  }
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

func newBaseStopVars() hw {
	return hw{
		"stop_id": "FTVL",
		"stf": hw{
			"service_date": "2018-05-30",
			"start_time":   57600,
			"end_time":     57900,
		},
	}
}

type rtTestCase struct {
	name    string
	query   string
	vars    map[string]interface{}
	rtfiles []testconfig.RTJsonFile
	cb      func(t *testing.T, jj string)
}

func testRt(t *testing.T, tc rtTestCase) {
	t.Run(tc.name, func(t *testing.T) {
		// Create a new RT Finder for each test...
		c, _ := newTestClientWithOpts(t, testconfig.Options{
			When:    "2022-09-01T00:00:00",
			RTJsons: tc.rtfiles,
		})
		var resp map[string]interface{}
		opts := []client.Option{}
		for k, v := range tc.vars {
			opts = append(opts, client.Var(k, v))
		}
		if err := c.Post(tc.query, &resp, opts...); err != nil {
			t.Error(err)
			return
		}
		jj := toJson(resp)
		if tc.cb != nil {
			tc.cb(t, jj)
		}
	})
}

func TestStopRTBasic(t *testing.T) {
	tc := rtTestCase{
		"stop times basic",
		baseStopQuery,
		newBaseStopVars(),
		testconfig.DefaultRTJson(),
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
				sched, _ := tt.NewWideTime(st.Get("arrival.scheduled").String())
				est, _ := tt.NewWideTime(st.Get("arrival.estimated").String())
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

func TestStopRTBasic_ArrivalFallback(t *testing.T) {
	tc := rtTestCase{
		"arrival will use departure if arrival is not present",
		baseStopQuery,
		newBaseStopVars(),
		[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_trip_updates", Fname: "BA-arrival-fallback.json"}},
		func(t *testing.T, jj string) {
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			checkTrip := "1031527WKDY"
			found := false
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, "2018-05-30T23:02:30Z", st.Get("arrival.estimated_utc").String())
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

func TestStopRTBasic_DepartureFallback(t *testing.T) {
	tc := rtTestCase{
		"departure will use arrival if departure is not present",
		baseStopQuery,
		newBaseStopVars(),
		[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_trip_updates", Fname: "BA-departure-fallback.json"}},
		func(t *testing.T, jj string) {
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			checkTrip := "1031527WKDY"
			found := false
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, "2018-05-30T23:02:30Z", st.Get("departure.estimated_utc").String())
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

func TestStopRTBasic_StopIDFallback(t *testing.T) {
	tc := rtTestCase{
		"use stop_id as fallback if no matching stop sequence",
		baseStopQuery,
		newBaseStopVars(),
		[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_trip_updates", Fname: "BA-stop-id-fallback.json"}},
		func(t *testing.T, jj string) {
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			checkTrip := "1031527WKDY"
			found := false
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, checkTrip, st.Get("trip.trip_id").String())
				assert.Equal(t, "2018-05-30T23:02:30Z", st.Get("departure.estimated_utc").String())
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

func TestStopRTBasic_StopIDFallback_NoDoubleVisit(t *testing.T) {
	tc := rtTestCase{
		"do not use stop_id as fallback if stop is visited twice",
		baseStopQuery,
		newBaseStopVars(),
		[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_trip_updates", Fname: "BA-stop-double-visit.json"}},
		func(t *testing.T, jj string) {
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			checkTrip := "1031527WKDY"
			found := false
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, "", st.Get("departure.estimated_utc").String())
			}
			if !found {
				t.Errorf("expected to find trip '%s'", checkTrip)
			}
		},
	}
	testRt(t, tc)
}

func TestStopRTBasic_NoRT(t *testing.T) {
	tc := rtTestCase{
		"no rt matches for trip 2211533WKDY",
		baseStopQuery,
		newBaseStopVars(),
		[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_trip_updates", Fname: "BA-departure-fallback.json"}},
		func(t *testing.T, jj string) {
			a := gjson.Get(jj, "stops.0.stop_times").Array()
			checkTrip := "2211533WKDY"
			found := false
			for _, st := range a {
				if st.Get("trip.trip_id").String() != checkTrip {
					continue
				}
				found = true
				assert.Equal(t, checkTrip, st.Get("trip.trip_id").String())
				assert.Equal(t, "STATIC", st.Get("trip.schedule_relationship").String(), "trip.schedule_relationship")
				assert.Equal(t, "", st.Get("trip.timestamp").String())
				assert.Equal(t, "", st.Get("arrival.estimated_utc").String())
				assert.Equal(t, "", st.Get("departure.estimated_utc").String())
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
		newBaseStopVars(),
		[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_trip_updates", Fname: "BA-added.json"}},
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
		newBaseStopVars(),
		[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_trip_updates", Fname: "BA-added.json"}},
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
	activeVars := newBaseStopVars()
	activeVars["active"] = true
	tcs := []rtTestCase{
		{
			"trip alerts",
			baseStopQuery,
			newBaseStopVars(),
			[]testconfig.RTJsonFile{
				{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"},
			},
			func(t *testing.T, jj string) {
				checkTrip := "1031527WKDY"
				found := false
				a := gjson.Get(jj, "stops.0.stop_times").Array()
				for _, st := range a {
					if st.Get("trip.trip_id").String() != checkTrip {
						continue
					}
					found = true
					alerts := st.Get("trip.alerts").Array()
					if len(alerts) != 2 {
						t.Errorf("got %d alerts, expected 2", len(alerts))
					}
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
		{
			"trip alerts active",
			baseStopQuery,
			activeVars,
			[]testconfig.RTJsonFile{
				{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"},
			},
			func(t *testing.T, jj string) {
				checkTrip := "1031527WKDY"
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
						assert.Equal(t, "Test trip header - active", firstAlert.Get("header_text.0.text").String(), "header_text.0.text")
						assert.Contains(t, firstAlert.Get("description_text.0.text").String(), "trip_id:1031527WKDY", "description_text.0.text")
					} else {
						t.Errorf("got %d alerts, expected 1", len(alerts))
					}
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
	}
	for _, tc := range tcs {
		testRt(t, tc)
	}
}

func TestRouteAlerts(t *testing.T) {
	activeVars := newBaseStopVars()
	activeVars["active"] = true
	tcs := []rtTestCase{
		{
			"stop alerts active",
			baseStopQuery,
			newBaseStopVars(),
			[]testconfig.RTJsonFile{
				{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"},
			},
			func(t *testing.T, jj string) {
				checkTrip := "1031527WKDY"
				sts := gjson.Get(jj, "stops.0.stop_times").Array()
				found := false
				for _, st := range sts {
					if st.Get("trip.trip_id").String() != checkTrip {
						continue
					}
					found = true
					assert.Equal(t, "05", st.Get("trip.route.route_id").String(), "trip.route.route_id")
					alerts := st.Get("trip.route.alerts").Array()
					if len(alerts) != 2 {
						t.Errorf("got %d alerts, expected 2", len(alerts))
					}
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
		{
			"stop alerts active",
			baseStopQuery,
			activeVars,
			[]testconfig.RTJsonFile{
				{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"},
			},
			func(t *testing.T, jj string) {
				checkTrip := "1031527WKDY"
				sts := gjson.Get(jj, "stops.0.stop_times").Array()
				found := false
				for _, st := range sts {
					if st.Get("trip.trip_id").String() != checkTrip {
						continue
					}
					found = true
					assert.Equal(t, "05", st.Get("trip.route.route_id").String(), "trip.route.route_id")
					alerts := st.Get("trip.route.alerts").Array()
					if len(alerts) == 1 {
						firstAlert := alerts[0]
						assert.Equal(t, "Test route header - active", firstAlert.Get("header_text.0.text").String(), "header_text.0.text")
						assert.Contains(t, firstAlert.Get("description_text.0.text").String(), "route_id:05", "description_text.0.text")
					} else {
						t.Errorf("got %d alerts, expected 1", len(alerts))
					}
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
	}
	for _, tc := range tcs {
		testRt(t, tc)
	}

}

func TestStopAlerts(t *testing.T) {
	activeVars := newBaseStopVars()
	activeVars["active"] = true
	tcs := []rtTestCase{
		{
			"stop alerts",
			baseStopQuery,
			newBaseStopVars(),
			[]testconfig.RTJsonFile{
				{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"},
			},
			func(t *testing.T, jj string) {
				alerts := gjson.Get(jj, "stops.0.alerts").Array()
				if len(alerts) != 2 {
					t.Errorf("got %d alerts, expected 2", len(alerts))
				}
			},
		},
		{
			"stop alerts active",
			baseStopQuery,
			activeVars,
			[]testconfig.RTJsonFile{
				{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"},
			},
			func(t *testing.T, jj string) {
				alerts := gjson.Get(jj, "stops.0.alerts").Array()
				if len(alerts) == 1 {
					firstAlert := alerts[0]
					assert.Equal(t, "Test stop header - active", firstAlert.Get("header_text.0.text").String(), "header_text.0.text")
					assert.Contains(t, firstAlert.Get("description_text.0.text").String(), "stop_id:FTVL", "description_text.0.text")
				} else {
					t.Errorf("got %d alerts, expected 1", len(alerts))
				}
			},
		},
	}
	for _, tc := range tcs {
		testRt(t, tc)
	}

}

func TestAgencyAlerts(t *testing.T) {
	activeVars := newBaseStopVars()
	activeVars["active"] = true
	tcs := []rtTestCase{
		{
			"stop alerts",
			baseStopQuery,
			newBaseStopVars(),
			[]testconfig.RTJsonFile{
				{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"},
			},
			func(t *testing.T, jj string) {
				checkTrip := "1031527WKDY"
				sts := gjson.Get(jj, "stops.0.stop_times").Array()
				found := false
				for _, st := range sts {
					if st.Get("trip.trip_id").String() != checkTrip {
						continue
					}
					found = true
					assert.Equal(t, "BART", st.Get("trip.route.agency.agency_id").String(), "trip.route.agency.agency_id")
					alerts := st.Get("trip.route.agency.alerts").Array()
					if len(alerts) != 2 {
						t.Errorf("got %d alerts, expected 2", len(alerts))
					}
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
		{
			"stop alerts active",
			baseStopQuery,
			activeVars,
			[]testconfig.RTJsonFile{{Feed: "BA", Ftype: "realtime_alerts", Fname: "BA-alerts.json"}},
			func(t *testing.T, jj string) {
				checkTrip := "1031527WKDY"
				sts := gjson.Get(jj, "stops.0.stop_times").Array()
				found := false
				for _, st := range sts {
					if st.Get("trip.trip_id").String() != checkTrip {
						continue
					}
					found = true
					assert.Equal(t, "BART", st.Get("trip.route.agency.agency_id").String(), "trip.route.agency.agency_id")
					alerts := st.Get("trip.route.agency.alerts").Array()
					if len(alerts) == 1 {
						firstAlert := alerts[0]
						assert.Equal(t, "Test agency header - active", firstAlert.Get("header_text.0.text").String(), "header_text.0.text")
						assert.Contains(t, firstAlert.Get("description_text.0.text").String(), "agency_id:BART", "description_text.0.text")
					} else {
						t.Errorf("got %d alerts, expected 1", len(alerts))
					}
				}
				if !found {
					t.Errorf("expected to find trip '%s'", checkTrip)
				}
			},
		},
	}
	for _, tc := range tcs {
		testRt(t, tc)
	}
}
