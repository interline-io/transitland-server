package resolvers

import (
	"context"
	"testing"
)

func TestFeedResolver(t *testing.T) {
	testcases := []testcase{
		{
			name:         "basic",
			query:        `query { feeds {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "test-gbfs", "BA", "HA", "BA~rt", "CT~rt", "test"},
		},
		{
			name:         "basic fields",
			query:        `query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {name onestop_id spec languages file}}`,
			vars:         hw{"onestop_id": "CT"},
			expect:       `{"feeds":[{"file":"server-test.dmfr.json","languages":["en-US"],"name":"Caltrain","onestop_id":"CT","spec":"GTFS"}]}`,
			selector:     "",
			expectSelect: nil,
		},
		// TODO: authorization,
		// TODO: associated_operators
		{
			name:         "urls",
			query:        `query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {urls { static_current static_historic }}}`,
			vars:         hw{"onestop_id": "CT"},
			expect:       `{"feeds":[{"urls":{"static_current":"file://test/data/external/caltrain.zip","static_historic":["https://caltrain.com/old_feed.zip"]}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "search by url case insensitive",
			query:        `query($url:String!) { feeds(where:{source_url:{url:$url}}) { onestop_id }}`,
			vars:         hw{"url": "file://test/data/external/Caltrain.zip"},
			expect:       `{"feeds":[{"onestop_id":"CT"}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "search by url case sensitive",
			query:        `query($url:String!) { feeds(where:{source_url:{url:$url, case_sensitive: true}}) { onestop_id }}`,
			vars:         hw{"url": "file://test/data/external/Caltrain.zip"},
			expect:       `{"feeds":[]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "search by url with type specified",
			query:        `query($url:String!) { feeds(where:{source_url:{url:$url, type: static_current}}) { onestop_id }}`,
			vars:         hw{"url": "file://test/data/external/caltrain.zip"},
			expect:       `{"feeds":[{"onestop_id":"CT"}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "search by url with type realtime_trip_updates",
			query:        `query($url:String!) { feeds(where:{source_url:{url:$url, type: realtime_trip_updates}}) { onestop_id }}`,
			vars:         hw{"url": "file://test/data/rt/BA.json"},
			expect:       `{"feeds":[{"onestop_id":"BA~rt"}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "search by url with type",
			query:        `query($url:String) { feeds(where:{source_url:{url: $url, type: realtime_trip_updates}}) { onestop_id }}`,
			vars:         hw{"url": nil},
			expect:       `{"feeds":[{"onestop_id":"BA~rt"},{"onestop_id":"CT~rt"}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "license",
			query:        `query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {license {spdx_identifier url use_without_attribution create_derived_product redistribution_allowed commercial_use_allowed share_alike_optional attribution_text attribution_instructions}}}`,
			vars:         hw{"onestop_id": "CT"},
			expect:       ` {"feeds":[{"license":{"attribution_instructions":"test attribution instructions","attribution_text":"data provided by 511.org","commercial_use_allowed":"yes","create_derived_product":"yes","redistribution_allowed":"no","share_alike_optional":"yes","spdx_identifier":"test","url":"http://assets.511.org/pdf/nextgen/developers/511_Data_Agreement_Final.pdf","use_without_attribution":"no"}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		{
			name:         "feed_versions",
			query:        `query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {feed_versions { sha1 }}}`,
			vars:         hw{"onestop_id": "CT"},
			expect:       ``,
			selector:     "feeds.0.feed_versions.#.sha1",
			expectSelect: []string{"d2813c293bcfd7a97dde599527ae6c62c98e66c6"},
		},
		{
			name:         "feed_state",
			query:        `query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {feed_state { feed_version { sha1 }}}}`,
			vars:         hw{"onestop_id": "CT"},
			expect:       `{"feeds":[{"feed_state":{"feed_version":{"sha1":"d2813c293bcfd7a97dde599527ae6c62c98e66c6"}}}]}`,
			selector:     "",
			expectSelect: nil,
		},
		// filters
		{
			name:         "where onestop_id",
			query:        `query { feeds(where:{onestop_id:"test"}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"test"},
		},
		{
			name:         "where spec=gtfs",
			query:        `query { feeds(where:{spec:[GTFS]}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "BA", "test", "HA"},
		},
		{
			name:         "where spec=gtfs-rt",
			query:        `query { feeds(where:{spec:[GTFS_RT]}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA~rt", "CT~rt"},
		},
		{
			name:         "where fetch_error=true",
			query:        `query { feeds(where:{fetch_error:true}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"test"},
		},
		{
			name:         "where fetch_error=false",
			query:        `query { feeds(where:{fetch_error:false}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA", "CT", "HA"},
		},
		{
			name:         "where import_status=success",
			query:        `query { feeds(where:{import_status:SUCCESS}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA", "CT", "HA"},
		},
		{
			name:         "where import_status=in_progress", // TODO: mock an in-progress import
			query:        `query { feeds(where:{import_status:IN_PROGRESS}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{},
		},
		{
			name:         "where import_status=error", // TODO: mock an in-progress import
			query:        `query { feeds(where:{import_status:ERROR}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{},
		},
		{
			name:         "where search", // TODO: mock an in-progress import
			query:        `query { feeds(where:{search:"cal"}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT"},
		},
		{
			name:         "where search ba", // TODO: mock an in-progress import
			query:        `query { feeds(where:{search:"BA"}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA", "BA~rt"},
		},
		{
			name:         "where tags test=ok",
			query:        `query { feeds(where:{tags:{test:"ok"}}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name:         "where tags test=ok foo=fail",
			query:        `query { feeds(where:{tags:{test:"ok", foo:"fail"}}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{},
		},
		{
			name:         "where tags test=ok foo=bar",
			query:        `query { feeds(where:{tags:{test:"ok", foo:"bar"}}) {onestop_id}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name:         "where tags test is present",
			query:        `query { feeds(where:{tags:{test:""}}) {onestop_id }}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		// feed fetches
		{
			name:         "feed fetches",
			query:        `query { feeds(where:{onestop_id:"BA"}) { onestop_id feed_fetches(limit:1) { success }}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.0.feed_fetches.#.success",
			expectSelect: []string{"true"},
		},
		{
			name:         "feed fetches failed",
			query:        `query { feeds(where:{onestop_id:"test"}) { onestop_id feed_fetches(limit:1, where:{success:false}) { success }}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.0.feed_fetches.#.success",
			expectSelect: []string{"false"},
		},
		// multiple queries
		{
			name:         "feed fetches multiple queries 1/2",
			query:        `query { feeds(where:{onestop_id:"BA"}) { onestop_id ok:feed_fetches(limit:1, where:{success:true}) { success } fail:feed_fetches(limit:1, where:{success:false}) { success }}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.0.ok.#.success",
			expectSelect: []string{"true"},
		},
		{
			name:         "feed fetches multiple queries 2/2",
			query:        `query { feeds(where:{onestop_id:"BA"}) { onestop_id ok:feed_fetches(limit:1, where:{success:true}) { success } fail:feed_fetches(limit:1, where:{success:false}) { success }}}`,
			vars:         hw{},
			expect:       ``,
			selector:     "feeds.0.fail.#.success",
			expectSelect: []string{},
		},
		// license
		{
			name:         "license filter: share_alike_optional = yes",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"share_alike_optional": "YES"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "HA"},
		},
		{
			name:         "license filter: share_alike_optional = no",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"share_alike_optional": "NO"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name:         "license filter: create_derived_product = yes",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"create_derived_product": "YES"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "HA"},
		},
		{
			name:         "license filter: create_derived_product = no",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"create_derived_product": "NO"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name:         "license filter: commercial_use_allowed = yes",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"commercial_use_allowed": "YES"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "HA"},
		},
		{
			name:         "license filter: commercial_use_allowed = no",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"commercial_use_allowed": "NO"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name:         "license filter: redistribution_allowed = yes",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"redistribution_allowed": "YES"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"HA"},
		},
		{
			name:         "license filter: redistribution_allowed = no",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"redistribution_allowed": "NO"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "BA"},
		},
		{
			name:         "license filter: use_without_attribution = yes",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"use_without_attribution": "YES"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"HA"},
		},
		{
			name:         "license filter: use_without_attribution = no",
			query:        `query($lic:LicenseFilter) {feeds(where: {license: $lic}) {onestop_id}}`,
			vars:         hw{"lic": hw{"use_without_attribution": "NO"}},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "BA"},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestFeedResolver_Cursor(t *testing.T) {
	allEnts, err := TestDBFinder.FindFeeds(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	allIds := []string{}
	for _, ent := range allEnts {
		allIds = append(allIds, ent.FeedID)
	}
	testcases := []testcase{
		{
			name:         "no cursor",
			query:        "query{feeds(limit:10){id onestop_id}}",
			vars:         nil,
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: allIds,
		},
		{
			name:         "after 0",
			query:        "query{feeds(after: 0, limit:10){id onestop_id}}",
			vars:         nil,
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: allIds,
		},
		{
			name:         "after 1st",
			query:        "query($after: Int!){feeds(after: $after, limit:10){id onestop_id}}",
			vars:         hw{"after": allEnts[1].ID},
			expect:       ``,
			selector:     "feeds.#.onestop_id",
			expectSelect: allIds[2:],
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
