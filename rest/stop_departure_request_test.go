package rest

import (
	"testing"
)

func TestStopDepartureRequest(t *testing.T) {
	cfg := testRestConfig()
	sid := "s-9q9nfsxn67-fruitvale"
	testcases := []testRest{
		{"basic", StopDepartureRequest{StopKey: sid}, "", "stops.#.stop_id", nil, 1},
		{
			"departure 10:00:00",
			StopDepartureRequest{StopKey: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", Limit: 5},
			"",
			"stops.0.departures.#.departure_time",
			[]string{"10:02:00", "10:02:00", "10:05:00", "10:09:00", "10:12:00"},
			0,
		},
		{
			"departure 10:00:00 to 10:10:00",
			StopDepartureRequest{StopKey: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00"},
			"",
			"stops.0.departures.#.departure_time",
			[]string{"10:02:00", "10:02:00", "10:05:00", "10:09:00"},
			0,
		},
		{
			"selects best service window date",
			StopDepartureRequest{StopKey: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00"},
			"",
			"stops.0.departures.#.service_date",
			[]string{"2018-06-04", "2018-06-04", "2018-06-04", "2018-06-04"},
			0,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
