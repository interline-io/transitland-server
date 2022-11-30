package resolvers

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestTripResolver(t *testing.T) {
	vars := hw{"trip_id": "3850526WKDY"}
	testcases := []testcase{
		{
			name:         "basic fields",
			query:        `query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {trip_id trip_headsign trip_short_name direction_id block_id wheelchair_accessible bikes_allowed stop_pattern_id }}`,
			vars:         vars,
			expect:       `{"trips":[{"bikes_allowed":1,"block_id":"","direction_id":1,"stop_pattern_id":21,"trip_headsign":"Antioch","trip_id":"3850526WKDY","trip_short_name":"","wheelchair_accessible":1}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "calendar",
			query:        `query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {calendar {service_id} }}`,
			vars:         vars,
			expect:       `{"trips":[{"calendar":{"service_id":"WKDY"}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "route",
			query:        `query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {route {route_id} }}`,
			vars:         vars,
			expect:       `{"trips":[{"route":{"route_id":"01"}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "shape",
			query:        `query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {shape {shape_id} }}`,
			vars:         vars,
			expect:       `{"trips":[{"shape":{"shape_id":"02_shp"}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "feed_version",
			query:        `query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {feed_version {sha1} }}`,
			vars:         vars,
			expect:       `{"trips":[{"feed_version":{"sha1":"e535eb2b3b9ac3ef15d82c56575e914575e732e0"}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "stop_times",
			query:        `query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {stop_times {stop_sequence} }}`,
			vars:         vars,
			expect:       ``,
			selector:     "trips.0.stop_times.#.stop_sequence",
			expectSelect: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27"},
		},
		{
			name:         "where trip_id",
			query:        `query{  trips(where:{trip_id:"3850526WKDY"}) {trip_id}}`,
			vars:         vars,
			expect:       ``,
			selector:     "trips.#.trip_id",
			expectSelect: []string{"3850526WKDY"},
		},
		{
			name:         "where service_date",
			query:        `query{trips(where:{feed_onestop_id:"CT",service_date:"2018-05-29"}){trip_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "trips.#.trip_id",
			expectSelect: []string{"101", "103", "305", "207", "309", "211", "313", "215", "217", "319", "221", "323", "225", "227", "329", "231", "233", "135", "237", "139", "143", "147", "151", "155", "257", "159", "261", "263", "365", "267", "269", "371", "273", "375", "277", "279", "381", "283", "385", "287", "289", "191", "193", "195", "197", "199", "102", "104", "206", "208", "310", "212", "314", "216", "218", "320", "222", "324", "226", "228", "330", "232", "134", "236", "138", "142", "146", "150", "152", "254", "156", "258", "360", "262", "264", "366", "268", "370", "272", "274", "376", "278", "380", "282", "284", "386", "288", "190", "192", "194", "196", "198"},
		},
		// license
		{
			name:         "license filter: share_alike_optional = yes",
			query:        `query($lic:LicenseFilter) {trips(limit:1,where: {license: $lic}) {trip_id feed_version{feed{license{share_alike_optional}}}}}`,
			vars:         hw{"lic": hw{"share_alike_optional": "YES"}},
			expect:       ``,
			selector:     "trips.0.feed_version.feed.license.share_alike_optional",
			expectSelect: []string{"yes"},
		},
		{
			name:         "license filter: share_alike_optional = no",
			query:        `query($lic:LicenseFilter) {trips(limit:1,where: {license: $lic}) {trip_id feed_version{feed{license{share_alike_optional}}}}}`,
			vars:         hw{"lic": hw{"share_alike_optional": "NO"}},
			expect:       ``,
			selector:     "trips.0.feed_version.feed.license.share_alike_optional",
			expectSelect: []string{"no"},
		},
		{
			name:         "license filter: create_derived_product = yes",
			query:        `query($lic:LicenseFilter) {trips(limit:1,where: {license: $lic}) {trip_id feed_version{feed{license{create_derived_product}}}}}`,
			vars:         hw{"lic": hw{"create_derived_product": "YES"}},
			expect:       ``,
			selector:     "trips.0.feed_version.feed.license.create_derived_product",
			expectSelect: []string{"yes"},
		},
		{
			name:         "license filter: create_derived_product = no",
			query:        `query($lic:LicenseFilter) {trips(limit:1,where: {license: $lic}) {trip_id feed_version{feed{license{create_derived_product}}}}}`,
			vars:         hw{"lic": hw{"create_derived_product": "NO"}},
			expect:       ``,
			selector:     "trips.0.feed_version.feed.license.create_derived_product",
			expectSelect: []string{"no"},
		},
		{
			name:         "license filter: commercial_use_allowed = yes",
			query:        `query($lic:LicenseFilter) {trips(limit:1,where: {license: $lic}) {trip_id feed_version{feed{license{commercial_use_allowed}}}}}`,
			vars:         hw{"lic": hw{"commercial_use_allowed": "YES"}},
			expect:       ``,
			selector:     "trips.0.feed_version.feed.license.commercial_use_allowed",
			expectSelect: []string{"yes"},
		},
		{
			name:         "license filter: commercial_use_allowed = no",
			query:        `query($lic:LicenseFilter) {trips(limit:1,where: {license: $lic}) {trip_id feed_version{feed{license{commercial_use_allowed}}}}}`,
			vars:         hw{"lic": hw{"commercial_use_allowed": "NO"}},
			expect:       ``,
			selector:     "trips.0.feed_version.feed.license.commercial_use_allowed",
			expectSelect: []string{"no"},
		},

		// TODO: check where feed_version_sha1, feed_onestop_id but only check count
		// TODO: frequencies
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestTripResolver_StopPatternID(t *testing.T) {
	query := `query {
		trips(where: {feed_onestop_id: "BA", trip_id:"3230742WKDY"}) {
		  trip_id
		  stop_pattern_id
		}
	}`
	c := newTestClient()
	var resp map[string]interface{}
	c.MustPost(query, &resp)
	jj := toJson(resp)
	patId := gjson.Get(jj, "trips.0.stop_pattern_id").Int()
	tc := testcase{
		name: "where trip_id",
		query: `query($patid:Int!) {
			trips(where: {feed_onestop_id: "BA", stop_pattern_id:$patid}) {
			  trip_id
			  stop_pattern_id
			}
		  }
		`,
		vars:         hw{"patid": patId},
		expect:       ``,
		selector:     "trips.#.trip_id",
		expectSelect: []string{"3230742WKDY", "3250757WKDY", "3270812WKDY", "3310827WKDY", "3210842WKDY"},
	}
	t.Run(tc.name, func(t *testing.T) {
		testquery(t, c, tc)
	})

}
