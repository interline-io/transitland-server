package resolvers

import (
	"testing"
)

func TestFeedVersionResolver(t *testing.T) {
	vars := hw{"feed_version_sha1": "d2813c293bcfd7a97dde599527ae6c62c98e66c6"}
	testcases := []testcase{
		{
			name:         "basic",
			query:        `query {  feed_versions {sha1} }`,
			vars:         hw{},
			expect:       ``,
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"e535eb2b3b9ac3ef15d82c56575e914575e732e0", "d2813c293bcfd7a97dde599527ae6c62c98e66c6", "c969427f56d3a645195dd8365cde6d7feae7e99b", "dd7aca4a8e4c90908fd3603c097fabee75fea907"},
		},
		{
			name:         "basic fields",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {sha1 url earliest_calendar_date latest_calendar_date name description} }`,
			vars:         vars,
			expect:       `{"feed_versions":[{"description":null,"earliest_calendar_date":"2017-10-02","latest_calendar_date":"2019-10-06","name":null,"sha1":"d2813c293bcfd7a97dde599527ae6c62c98e66c6","url":"file://test/data/external/caltrain.zip"}]}`,
			selector:     "",
			expectSelect: nil,
		},
		// children
		{
			name:         "feed",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {feed{onestop_id}} }`,
			vars:         vars,
			expect:       `{"feed_versions":[{"feed":{"onestop_id":"CT"}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "feed_version_gtfs_import",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {feed_version_gtfs_import{success in_progress}} }`,
			vars:         vars,
			expect:       `{"feed_versions":[{"feed_version_gtfs_import":{"in_progress":false,"success":true}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "feed_infos",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {feed_infos {feed_publisher_name feed_publisher_url feed_lang feed_version feed_start_date feed_end_date}} }`,
			vars:         hw{"feed_version_sha1": "e535eb2b3b9ac3ef15d82c56575e914575e732e0"}, // check BART instead
			expect:       `{"feed_versions":[{"feed_infos":[{"feed_end_date":"2019-07-01","feed_lang":"en","feed_publisher_name":"Bay Area Rapid Transit","feed_publisher_url":"http://www.bart.gov","feed_start_date":"2018-05-26","feed_version":"47"}]}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "files",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {files {name rows sha1 header csv_like size}} }`,
			vars:         vars,
			expect:       ``,
			selector:     "feed_versions.0.files.#.name",
			expectSelect: []string{"agency.txt", "calendar.txt", "calendar_attributes.txt", "calendar_dates.txt", "directions.txt", "fare_attributes.txt", "fare_rules.txt", "farezone_attributes.txt", "frequencies.txt", "realtime_routes.txt", "routes.txt", "shapes.txt", "stop_attributes.txt", "stop_times.txt", "stops.txt", "transfers.txt", "trips.txt"},
		},
		{
			name:         "agencies",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {agencies {agency_id}} }`,
			vars:         vars,
			expect:       ``,
			selector:     "feed_versions.0.agencies.#.agency_id",
			expectSelect: []string{"caltrain-ca-us"},
		},
		{
			name:         "routes",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {routes {route_id}} }`,
			vars:         vars,
			expect:       ``,
			selector:     "feed_versions.0.routes.#.route_id",
			expectSelect: []string{"Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"},
		},
		{
			name:         "stops",
			query:        `query($feed_version_sha1: String!) {  feed_versions(where:{sha1:$feed_version_sha1}) {stops {stop_id}} }`,
			vars:         vars,
			expect:       ``,
			selector:     "feed_versions.0.stops.#.stop_id",
			expectSelect: []string{"70011", "70012", "70021", "70022", "70031", "70032", "70041", "70042", "70051", "70052", "70061", "70062", "70071", "70072", "70081", "70082", "70091", "70092", "70101", "70102", "70111", "70112", "70121", "70122", "70131", "70132", "70141", "70142", "70151", "70152", "70161", "70162", "70171", "70172", "70191", "70192", "70201", "70202", "70211", "70212", "70221", "70222", "70231", "70232", "70241", "70242", "70251", "70252", "70261", "70262", "70271", "70272", "70281", "70282", "70291", "70292", "70301", "70302", "70311", "70312", "70321", "70322", "777402", "777403"},
		},
		// where
		{
			name:         "where feed_onestop_id",
			query:        `query{feed_versions(where:{feed_onestop_id:"CT"}) {sha1} }`,
			vars:         hw{},
			expect:       ``,
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"d2813c293bcfd7a97dde599527ae6c62c98e66c6"},
		},
		{
			name:         "where sha1",
			query:        `query{feed_versions(where:{sha1:"d2813c293bcfd7a97dde599527ae6c62c98e66c6"}) {sha1} }`,
			vars:         hw{},
			expect:       ``,
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"d2813c293bcfd7a97dde599527ae6c62c98e66c6"},
		},
		{
			name:         "where import_status success",
			query:        `query{feed_versions(where:{feed_onestop_id:"CT", import_status:SUCCESS}) {sha1} }`,
			vars:         hw{},
			expect:       ``,
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"d2813c293bcfd7a97dde599527ae6c62c98e66c6"},
		},
		// there isnt a fv with this import status in test db
		{
			name:         "where import_status error",
			query:        `query{feed_versions(where:{feed_onestop_id:"CT", import_status:ERROR}) {sha1} }`,
			vars:         hw{},
			expect:       ``,
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{},
		},
		// there isnt a fv with this import status in test db
		{
			name:         "where import_status error",
			query:        `query{feed_versions(where:{feed_onestop_id:"CT", import_status:IN_PROGRESS}) {sha1} }`,
			vars:         hw{},
			expect:       ``,
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
