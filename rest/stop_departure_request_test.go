package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestStopDepartureRequest(t *testing.T) {
	bp := func(v bool) *bool {
		return &v
	}
	sid := "s-9q9nfsxn67-fruitvale"
	testcases := []testRest{
		{
			name:         "basic",
			h:            StopDepartureRequest{StopKey: sid},
			format:       "",
			selector:     "stops.#.stop_id",
			expectSelect: nil,
			expectLength: 1,
		},
		{
			name:         "departure 10:00:00",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", Limit: 5},
			format:       "",
			selector:     "stops.0.departures.#.departure_time",
			expectSelect: []string{"10:02:00", "10:02:00", "10:05:00", "10:09:00", "10:12:00"},
			expectLength: 0,
		},
		{
			name:         "departure 10:00:00 to 10:10:00",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00"},
			format:       "",
			selector:     "stops.0.departures.#.departure_time",
			expectSelect: []string{"10:02:00", "10:02:00", "10:05:00", "10:09:00"},
			expectLength: 0,
		},
		{
			name:         "include_geometry=true",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00", IncludeGeometry: true},
			format:       "",
			selector:     "stops.0.departures.0.trip.shape.geometry.type",
			expectSelect: []string{"LineString"},
			expectLength: 0,
		},
		{
			name:         "include_geometry=false",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00", IncludeGeometry: false},
			format:       "",
			selector:     "stops.0.departures.0.trip.shape.geometry.type",
			expectSelect: []string{},
			expectLength: 0,
		},
		{
			name:         "use_service_window=true",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00", UseServiceWindow: bp(true)},
			format:       "",
			selector:     "stops.0.departures.#.service_date",
			expectSelect: []string{"2018-06-04", "2018-06-04", "2018-06-04", "2018-06-04"},
			expectLength: 0,
		},
		{
			name:         "use_service_window=false",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00", UseServiceWindow: bp(false)},
			format:       "",
			selector:     "stops.0.departures.#.service_date",
			expectSelect: []string{},
			expectLength: 0,
		},
		{
			name:         "use_service_window=false good date",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2018-06-04", StartTime: "10:00:00", EndTime: "10:10:00", UseServiceWindow: bp(false)},
			format:       "",
			selector:     "stops.0.departures.#.service_date",
			expectSelect: []string{"2018-06-04", "2018-06-04", "2018-06-04", "2018-06-04"},
			expectLength: 0,
		},
		{
			name:         "selects best service window date",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2022-05-30", StartTime: "10:00:00", EndTime: "10:10:00"},
			format:       "",
			selector:     "stops.0.departures.#.service_date",
			expectSelect: []string{"2018-06-04", "2018-06-04", "2018-06-04", "2018-06-04"},
			expectLength: 0,
		},
		{
			name:         "no pagination",
			h:            StopDepartureRequest{StopKey: sid, ServiceDate: "2018-06-04", Limit: 1},
			format:       "",
			selector:     "meta.next",
			expectSelect: []string{},
			expectLength: 0,
		},
		{
			name:         "requires valid stop key",
			h:            StopDepartureRequest{StopKey: "0"},
			format:       "",
			selector:     "stops.0.onestop_id",
			expectSelect: []string{},
			expectLength: 0,
		},
		{
			name:         "requires valid stop key 2",
			h:            StopDepartureRequest{StopKey: "-1"},
			format:       "",
			selector:     "stops.0.onestop_id",
			expectSelect: []string{},
			expectLength: 0,
		},
		{
			name:         "feed_key",
			h:            StopDepartureRequest{StopKey: "BA:FTVL"},
			format:       "",
			selector:     "stops.0.stop_id",
			expectSelect: []string{"FTVL"},
			expectLength: 0,
		},
		//
		{
			name: "include_alerts:true",
			h:    StopDepartureRequest{StopKey: "BA:FTVL", ServiceDate: "2018-05-30", IncludeAlerts: true},
			f: func(t *testing.T, jj string) {
				a := gjson.Get(jj, "stops.0.alerts").Array()
				assert.Equal(t, 2, len(a), "alert count")
			},
		},
		{
			name: "include_alerts:false",
			h:    StopDepartureRequest{StopKey: "BA:FTVL", ServiceDate: "2018-05-30", IncludeAlerts: false},
			f: func(t *testing.T, jj string) {
				a := gjson.Get(jj, "stops.0.alerts").Array()
				assert.Equal(t, 0, len(a), "alert count")
			},
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
	cfg, _, _, _ := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
