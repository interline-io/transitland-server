package rest

import (
	"testing"
)

func TestStopDepartureRequest(t *testing.T) {
	bp := func(v bool) *bool {
		return &v
	}
	cfg := testRestConfig()
	sid := "s-9q9nfsxn67-fruitvale"
	testcases := []testRest{
		{"basic", StopDepartureRequest{Key: sid}, "", "stops.#.stop_id", nil, 1},
		{
			"departure 10:00:00",
			StopDepartureRequest{Key: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", Limit: 5},
			"",
			"stops.0.departures.#.departure_time",
			[]string{"10:02:00", "10:02:00", "10:05:00", "10:09:00", "10:12:00"},
			0,
		},
		{
			"departure 10:00:00 to 10:10:00",
			StopDepartureRequest{Key: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00"},
			"",
			"stops.0.departures.#.departure_time",
			[]string{"10:02:00", "10:02:00", "10:05:00", "10:09:00"},
			0,
		},
		{
			"include_geometry=true",
			StopDepartureRequest{Key: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00", IncludeGeometry: true},
			"",
			"stops.0.departures.0.trip.shape.geometry.type",
			[]string{"LineString"},
			0,
		},
		{
			"include_geometry=false",
			StopDepartureRequest{Key: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00", IncludeGeometry: false},
			"",
			"stops.0.departures.0.trip.shape.geometry.type",
			[]string{},
			0,
		},
		{
			"use_service_window=true",
			StopDepartureRequest{Key: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00", UseServiceWindow: bp(true)},
			"",
			"stops.0.departures.#.service_date",
			[]string{"2018-06-04", "2018-06-04", "2018-06-04", "2018-06-04"},
			0,
		},
		{
			"use_service_window=false",
			StopDepartureRequest{Key: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00", UseServiceWindow: bp(false)},
			"",
			"stops.0.departures.#.service_date",
			[]string{},
			0,
		},
		{
			"use_service_window=false good date",
			StopDepartureRequest{Key: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00", UseServiceWindow: bp(false)},
			"",
			"stops.0.departures.#.service_date",
			[]string{"2018-06-04", "2018-06-04", "2018-06-04", "2018-06-04"},
			0,
		},
		{
			"selects best service window date",
			StopDepartureRequest{Key: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00"},
			"",
			"stops.0.departures.#.service_date",
			[]string{"2018-06-04", "2018-06-04", "2018-06-04", "2018-06-04"},
			0,
		},
		{
			"no pagination",
			StopDepartureRequest{Key: sid, ServiceDate: "2018-06-04", Limit: 1},
			"",
			"meta.next",
			[]string{},
			0,
		},
		{
			"requires valid stop key",
			StopDepartureRequest{Key: "0"},
			"",
			"stops.0.onestop_id",
			[]string{},
			0,
		},
		{
			"requires valid stop key 2",
			StopDepartureRequest{Key: "-1"},
			"",
			"stops.0.onestop_id",
			[]string{},
			0,
		},
		{
			"feed_key",
			StopDepartureRequest{Key: "BA:FTVL"},
			"",
			"stops.0.stop_id",
			[]string{"FTVL"},
			0,
		},
		// TODO
		// {
		// 	"requires valid stop key 3",
		// 	StopDepartureRequest{StopKey: ""},
		// 	"",
		// 	"stops.0.onestop_id",
		// 	[]string{},
		// 	0,
		// },
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
