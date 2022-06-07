package resolvers

import (
	"context"
	"testing"
)

func TestFeedResolver(t *testing.T) {
	testcases := []testcase{
		{
			"basic",
			`query { feeds {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA", "CT", "HA", "BA~rt", "CT~rt", "test"},
		},
		{
			"basic fields",
			`query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {name onestop_id spec languages file}}`,
			hw{"onestop_id": "CT"},
			`{"feeds":[{"file":"server-test.dmfr.json","languages":["en-US"],"name":"Caltrain","onestop_id":"CT","spec":"GTFS"}]}`,
			"",
			nil,
		},
		// TODO: authorization,
		// TODO: associated_operators
		{
			"urls",
			`query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {urls { static_current static_historic }}}`,
			hw{"onestop_id": "CT"},
			`{"feeds":[{"urls":{"static_current":"file://test/data/external/caltrain.zip","static_historic":["https://caltrain.com/old_feed.zip"]}}]}`,
			"",
			nil,
		},
		{
			"search by url case insensitive",
			`query($url:String!) { feeds(where:{source_url:{url:$url}}) { onestop_id }}`,
			hw{"url": "file://test/data/external/Caltrain.zip"},
			`{"feeds":[{"onestop_id":"CT"}]}`,
			"",
			nil,
		},
		{
			"search by url case sensitive",
			`query($url:String!) { feeds(where:{source_url:{url:$url, case_sensitive: true}}) { onestop_id }}`,
			hw{"url": "file://test/data/external/Caltrain.zip"},
			`{"feeds":[]}`,
			"",
			nil,
		},
		{
			"search by url with type specified",
			`query($url:String!) { feeds(where:{source_url:{url:$url, type: static_current}}) { onestop_id }}`,
			hw{"url": "file://test/data/external/caltrain.zip"},
			`{"feeds":[{"onestop_id":"CT"}]}`,
			"",
			nil,
		},
		{
			"search by url with type realtime_trip_updates",
			`query($url:String!) { feeds(where:{source_url:{url:$url, type: realtime_trip_updates}}) { onestop_id }}`,
			hw{"url": "file://test/data/rt/BA.json"},
			`{"feeds":[{"onestop_id":"BA~rt"}]}`,
			"",
			nil,
		},
		{
			"search by url with type",
			`query($url:String) { feeds(where:{source_url:{url: $url, type: realtime_trip_updates}}) { onestop_id }}`,
			hw{"url": nil},
			`{"feeds":[{"onestop_id":"BA~rt"},{"onestop_id":"CT~rt"}]}`,
			"",
			nil,
		},
		{
			"license",
			`query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {license {spdx_identifier url use_without_attribution create_derived_product redistribution_allowed commercial_use_allowed share_alike_optional attribution_text attribution_instructions}}}`,
			hw{"onestop_id": "CT"},
			` {"feeds":[{"license":{"attribution_instructions":"test attribution instructions","attribution_text":"data provided by 511.org","commercial_use_allowed":"yes","create_derived_product":"yes","redistribution_allowed":"no","share_alike_optional":"yes","spdx_identifier":"test","url":"http://assets.511.org/pdf/nextgen/developers/511_Data_Agreement_Final.pdf","use_without_attribution":"no"}}]}`,
			"",
			nil,
		},
		{
			"feed_versions",
			`query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {feed_versions { sha1 }}}`,
			hw{"onestop_id": "CT"},
			``,
			"feeds.0.feed_versions.#.sha1",
			[]string{"d2813c293bcfd7a97dde599527ae6c62c98e66c6"},
		},
		{
			"feed_state",
			`query($onestop_id:String!) { feeds(where:{onestop_id:$onestop_id}) {feed_state { feed_version { sha1 }}}}`,
			hw{"onestop_id": "CT"},
			`{"feeds":[{"feed_state":{"feed_version":{"sha1":"d2813c293bcfd7a97dde599527ae6c62c98e66c6"}}}]}`,
			"",
			nil,
		},
		// filters
		{
			"where onestop_id",
			`query { feeds(where:{onestop_id:"test"}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"test"},
		},
		{
			"where spec=gtfs",
			`query { feeds(where:{spec:[GTFS]}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"CT", "BA", "test", "HA"},
		},
		{
			"where spec=gtfs-rt",
			`query { feeds(where:{spec:[GTFS_RT]}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA~rt", "CT~rt"},
		},
		{
			"where fetch_error=true",
			`query { feeds(where:{fetch_error:true}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"test"},
		},
		{
			"where fetch_error=false",
			`query { feeds(where:{fetch_error:false}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA", "CT", "HA"},
		},
		{
			"where import_status=success",
			`query { feeds(where:{import_status:SUCCESS}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA", "CT", "HA"},
		},
		{
			"where import_status=in_progress", // TODO: mock an in-progress import
			`query { feeds(where:{import_status:IN_PROGRESS}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{},
		},
		{
			"where import_status=error", // TODO: mock an in-progress import
			`query { feeds(where:{import_status:ERROR}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{},
		},
		{
			"where search", // TODO: mock an in-progress import
			`query { feeds(where:{search:"cal"}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"CT"},
		},
		{
			"where search ba", // TODO: mock an in-progress import
			`query { feeds(where:{search:"BA"}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA", "BA~rt"},
		},
		{
			"where tags test=ok",
			`query { feeds(where:{tags:{test:"ok"}}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA"},
		},
		{
			"where tags test=ok foo=fail",
			`query { feeds(where:{tags:{test:"ok", foo:"fail"}}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{},
		},
		{
			"where tags test=ok foo=bar",
			`query { feeds(where:{tags:{test:"ok", foo:"bar"}}) {onestop_id}}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA"},
		},
		{
			"where tags test is present",
			`query { feeds(where:{tags:{test:""}}) {onestop_id }}`,
			hw{},
			``,
			"feeds.#.onestop_id",
			[]string{"BA"},
		},
		// feed fetches
		{
			"feed fetches",
			`query { feeds(where:{onestop_id:"BA"}) { onestop_id feed_fetches(limit:1) { success }}}`,
			hw{},
			``,
			"feeds.0.feed_fetches.#.success",
			[]string{"true"},
		},
		{
			"feed fetches failed",
			`query { feeds(where:{onestop_id:"test"}) { onestop_id feed_fetches(limit:1, where:{success:false}) { success }}}`,
			hw{},
			``,
			"feeds.0.feed_fetches.#.success",
			[]string{"false"},
		},
		// multiple queries
		{
			"feed fetches multiple queries 1/2",
			`query { feeds(where:{onestop_id:"BA"}) { onestop_id ok:feed_fetches(limit:1, where:{success:true}) { success } fail:feed_fetches(limit:1, where:{success:false}) { success }}}`,
			hw{},
			``,
			"feeds.0.ok.#.success",
			[]string{"true"},
		},
		{
			"feed fetches multiple queries 2/2",
			`query { feeds(where:{onestop_id:"BA"}) { onestop_id ok:feed_fetches(limit:1, where:{success:true}) { success } fail:feed_fetches(limit:1, where:{success:false}) { success }}}`,
			hw{},
			``,
			"feeds.0.fail.#.success",
			[]string{},
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
			"no cursor",
			"query{feeds(limit:10){id onestop_id}}",
			nil,
			``,
			"feeds.#.onestop_id",
			allIds,
		},
		{
			"after 0",
			"query{feeds(after: 0, limit:10){id onestop_id}}",
			nil,
			``,
			"feeds.#.onestop_id",
			allIds,
		},
		{
			"after 1st",
			"query($after: Int!){feeds(after: $after, limit:10){id onestop_id}}",
			hw{"after": allEnts[1].ID},
			``,
			"feeds.#.onestop_id",
			allIds[2:],
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
