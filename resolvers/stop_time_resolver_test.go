package resolvers

import (
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/clock"
)

func TestStopResolver_StopTimes(t *testing.T) {
	vars := hw{"trip_id": "3850526WKDY"}
	testcases := []testcase{
		{
			"basic",
			`query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {trip_id stop_times { arrival_time }} }`,
			vars,
			``,
			"trips.0.stop_times.#.arrival_time",
			[]string{"05:26:00", "05:29:00", "05:33:00", "05:36:00", "05:40:00", "05:43:00", "05:46:00", "05:48:00", "05:50:00", "05:53:00", "05:54:00", "05:56:00", "05:58:00", "06:05:00", "06:08:00", "06:11:00", "06:15:00", "06:17:00", "06:23:00", "06:27:00", "06:32:00", "06:35:00", "06:40:00", "06:43:00", "06:50:00", "07:05:00", "07:13:00"},
		},
		{
			// these are supposed to always be ordered by stop_sequence, so we can directly check the first one.
			"basic fields",
			`query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {trip_id stop_times(limit:1) { arrival_time departure_time stop_sequence stop_headsign pickup_type drop_off_type timepoint interpolated}} }`,
			vars,
			`{"trips":[{"stop_times":[{"arrival_time":"05:26:00","departure_time":"05:26:00","drop_off_type":null,"interpolated":null,"pickup_type":null,"stop_headsign":"Antioch","stop_sequence":1,"timepoint":1}],"trip_id":"3850526WKDY"}]}`,
			"",
			nil,
		},
		{
			// check stops for a trip
			"stop",
			`query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {trip_id stop_times { stop { stop_id } }} }`,
			vars,
			``,
			"trips.0.stop_times.#.stop.stop_id",
			[]string{"SFIA", "SBRN", "SSAN", "COLM", "DALY", "BALB", "GLEN", "24TH", "16TH", "CIVC", "POWL", "MONT", "EMBR", "WOAK", "12TH", "19TH_N", "MCAR", "ROCK", "ORIN", "LAFY", "WCRK", "PHIL", "CONC", "NCON", "PITT", "PCTR", "ANTC"},
		},
		{
			// go through a stop to get trip_ids
			"trip",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {stop_times { trip { trip_id} }} }`,
			hw{"stop_id": "70302"}, // Morgan hill
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"268", "274", "156"},
		},
		// check StopTimeFilter through a stop
		{
			"where service_date start_time end_time",
			`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{service_date:"2018-05-30", start_time: 26000, end_time: 30000}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"07:18:00", "07:24:00", "07:28:00", "07:33:00", "07:39:00", "07:43:00", "07:48:00", "07:54:00", "07:58:00", "08:03:00", "08:09:00", "08:18:00", "07:24:00", "07:39:00", "07:54:00", "08:09:00", "07:16:00", "07:31:00", "07:46:00", "08:01:00", "08:16:00"},
		},
		{
			"where service_date end_time",
			`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{service_date:"2018-05-30", end_time: 20000}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"04:39:00", "04:54:00", "05:09:00", "05:24:00", "04:39:00", "04:54:00", "05:09:00", "05:24:00", "04:31:00", "04:46:00", "05:01:00", "05:16:00", "05:31:00"},
		},
		{
			"where service_date start_time",
			`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{service_date:"2018-05-30", start_time: 76000}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"21:14:00", "21:34:00", "21:54:00", "22:14:00", "22:34:00", "22:54:00", "23:14:00", "23:34:00", "23:54:00", "24:14:00", "24:47:00", "21:14:00", "21:34:00", "21:54:00", "22:14:00", "22:34:00", "22:54:00", "23:14:00", "23:34:00", "23:54:00", "24:14:00", "24:47:00"},
		},
		// accept strings for Start / End
		{
			"start time string",
			`query{ stops(where:{stop_id:"RICH"}) { stop_times(where:{service_date:"2018-05-30", start: "10:00:00", end: "10:10:00"}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.departure_time",
			[]string{"10:02:00", "10:05:00", "10:10:00"},
		},

		// check arrival and departure resolvers
		{
			"arrival departure base case",
			`query{ stops(where:{stop_id:"RICH"}) { stop_times(where:{service_date:"2018-05-30", start_time: 76000, end_time: 76900}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.departure_time",
			[]string{"21:09:00", "21:14:00", "21:15:00"},
		},
		{
			"departures",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{service_date:"2018-05-30", start_time: 76000, end_time: 76900}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"21:15:00"},
		},
		{
			"arrivals",
			`query{ stops(where:{stop_id:"RICH"}) { arrivals(where:{service_date:"2018-05-30", start_time: 76000, end_time: 76900}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.arrivals.#.arrival_time",
			[]string{"21:09:00", "21:14:00"},
		},
		// route_onestop_ids
		{
			"departure route_onestop_ids",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:12:00", "10:20:00", "10:27:00", "10:35:00", "10:42:00", "10:50:00", "10:57:00"},
		},
		{
			"departure route_onestop_ids 1",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{route_onestop_ids: ["r-9q8y-richmond~dalycity~millbrae"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:12:00", "10:27:00", "10:42:00", "10:57:00"},
		},
		{
			"departure route_onestop_ids 2",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		// Allow previous route onestop ids
		// OLD: r-9q9n-fremont~richmond
		// NEW: r-9q9n-warmsprings~southfremont~richmond
		{
			"departure route_onestop_ids use previous id current ok",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: false, route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		{
			"departure route_onestop_ids, use previous id, both at once ok",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: false, route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond","r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		{
			"departure route_onestop_ids, use previous id, both at once, no duplicates",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: true, route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond","r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		{
			"departure route_onestop_ids, use previous id, old, fail",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: false, route_onestop_ids: ["r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{},
		},
		{
			"departure route_onestop_ids, use previous id, old, ok",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: true, route_onestop_ids: ["r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_StopTimes_ServiceDate(t *testing.T) {
	q := `query($stop_id:String!,$sd:Date!,$ed:Boolean){ stops(where:{stop_id:$stop_id}) { stop_times(where:{service_date:$sd, start_time:54000, end_time:57600, use_service_window:$ed}) {service_date arrival_time}}}`
	testcases := []testcase{
		{
			"service date in range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2018-05-29", "ed": true},
			``,
			"stops.0.stop_times.0.service_date",
			[]string{"2018-05-29"}, // expect input date
		},
		{
			"service date after range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": true},
			``,
			"stops.0.stop_times.0.service_date",
			[]string{"2018-06-05"}, // expect adjusted date in window
		},
		{
			"service date before range, friday",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2010-05-28", "ed": true},
			``,
			"stops.0.stop_times.0.service_date",
			[]string{"2018-06-08"}, // expect adjusted date in window
		},
		{
			"service date after range, exact dates",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": false},
			``,
			"stops.0.stop_times.#.service_date",
			[]string{}, // exect no results
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient()
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_StopTimes_WindowDates(t *testing.T) {
	bartWeekdayTimes := []string{"15:01:00", "15:09:00", "15:09:00", "15:16:00", "15:24:00", "15:24:00", "15:31:00", "15:39:00", "15:39:00", "15:46:00", "15:54:00", "15:54:00"}
	bartWeekendTimes := []string{"15:15:00", "15:15:00", "15:35:00", "15:35:00", "15:55:00", "15:55:00"}
	q := `query($stop_id:String!,$sd:Date!,$ed:Boolean){ stops(where:{stop_id:$stop_id}) { stop_times(where:{service_date:$sd, start_time:54000, end_time:57600, use_service_window:$ed}) {arrival_time}}}`
	testcases := []testcase{
		{
			"service date in range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2018-05-29", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekdayTimes,
		},
		{
			"service date after range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekdayTimes,
		},
		{
			"service date after range, exact dates",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": false},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{},
		},
		{
			"service date after range, sunday",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-26", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekendTimes,
		},
		{
			"service date before range, tuesday",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2010-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekdayTimes,
		},
		{
			"fv without feed_info, in window, monday",
			q,
			hw{"stop_id": "70011", "sd": "2019-02-11", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"15:48:00", "15:50:00"},
		},
		{
			"fv without feed_info, before window, friday",
			q,
			hw{"stop_id": "70011", "sd": "2010-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"15:48:00", "15:50:00"},
		},
		{
			"fv without feed_info, after window, tuesday",
			q,
			hw{"stop_id": "70011", "sd": "2030-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"15:48:00", "15:50:00"},
		},
		{
			"fv without feed_info, after window, tuesday, exact date only",
			q,
			hw{"stop_id": "70011", "sd": "2030-05-28", "ed": false},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient()
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_StopTimes_Next(t *testing.T) {
	type tcWithClock struct {
		testcase
		when string
	}
	testcases := []tcWithClock{
		// Relative times
		{
			testcase{
				"where next 3600",
				`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{next:3600}) {arrival_time}}}`,
				hw{},
				``,
				"stops.0.stop_times.#.arrival_time",
				// these should start at 15:00 - 16:00
				[]string{"15:01:00", "15:09:00", "15:09:00", "15:16:00", "15:24:00", "15:24:00", "15:31:00", "15:39:00", "15:39:00", "15:46:00", "15:54:00", "15:54:00"},
			},
			"2018-05-30T22:00:00",
		},
		{
			testcase{
				"where next 1800",
				`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{next:1800}) {arrival_time}}}`,
				hw{},
				``,
				"stops.0.stop_times.#.arrival_time",
				// these should start at 15:00 - 15:30
				[]string{"15:01:00", "15:09:00", "15:09:00", "15:16:00", "15:24:00", "15:24:00"},
			},
			"2018-05-30T22:00:00",
		},
		{
			testcase{
				"where next 900, east coast",
				`query{ stops(where:{stop_id:"6497"}) { stop_times(where:{next:900}) {arrival_time}}}`,
				hw{},
				``,
				"stops.0.stop_times.#.arrival_time",
				// these should start at 18:00 - 18:15
				[]string{"18:00:00", "18:00:00", "18:00:00", "18:00:00", "18:00:00", "18:03:00", "18:10:00", "18:10:00", "18:13:00", "18:14:00", "18:15:00", "18:15:00"},
			},
			"2018-05-30T22:00:00",
		},
		{
			testcase{
				"where next 600, multiple timezones",
				`query{ stops(where:{onestop_ids:["s-dhvrsm227t-universityareatransitcenter", "s-9q9p1wxf72-macarthur"]}) { onestop_id stop_id stop_times(where:{next:600}) {arrival_time}}}`,
				hw{},
				// this test checks the json response because it is too complex for the simple element selector approach
				// we should expect east coast times 18:00-18:10, and west coast times 15:00-15:10
				`{
					"stops": [
					{
						"onestop_id": "s-9q9p1wxf72-macarthur",
						"stop_id": "MCAR",
						"stop_times": [{
							"arrival_time": "15:00:00"
						}, {
							"arrival_time": "15:07:00"
						}]
					}, {
						"onestop_id": "s-9q9p1wxf72-macarthur",
						"stop_id": "MCAR_S",
						"stop_times": [{
							"arrival_time": "15:01:00"
						}, {
							"arrival_time": "15:09:00"
						}, {
							"arrival_time": "15:09:00"
						}]
					},
					{
						"onestop_id": "s-dhvrsm227t-universityareatransitcenter",
						"stop_id": "6497",
						"stop_times": [{
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:03:00"
						}, {
							"arrival_time": "18:10:00"
						}, {
							"arrival_time": "18:10:00"
						}]
					}]
				}`,
				"",
				nil,
			},
			"2018-05-30T22:00:00",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 2018-05-28 22:00:00 +0000 UTC
			// 2018-05-28 15:00:00 -0700 PDT
			when, err := time.Parse("2006-01-02T15:04:05", tc.when)
			if err != nil {
				t.Fatal(err)
			}
			_, _, c := newTestClientWithClock(&clock.Mock{T: when})
			testquery(t, c, tc.testcase)
		})
	}
}
