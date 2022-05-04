package rest

import (
	"testing"
)

func TestFeedRequest(t *testing.T) {
	cfg := testRestConfig()
	// fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	testcases := []testRest{
		{"basic", &FeedRequest{}, "", "feeds.#.onestop_id", []string{"CT", "BA", "HA", "BA~rt", "CT~rt", "test"}, 0},
		{"onestop_id", &FeedRequest{OnestopID: "BA"}, "", "feeds.#.onestop_id", []string{"BA"}, 0},
		{"spec", &FeedRequest{Spec: "GTFS_RT"}, "", "feeds.#.onestop_id", []string{"BA~rt", "CT~rt"}, 0},
		{"search", &FeedRequest{Search: "ba"}, "", "feeds.#.onestop_id", []string{"BA~rt", "BA"}, 0},
		{"fetch_error true", &FeedRequest{FetchError: "true"}, "", "feeds.#.onestop_id", []string{"test"}, 0},
		{"fetch_error false", &FeedRequest{FetchError: "false"}, "", "feeds.#.onestop_id", []string{"BA", "CT", "HA"}, 0},
		{"tags test=ok", &FeedRequest{TagKey: "test", TagValue: "ok"}, "", "feeds.#.onestop_id", []string{"BA"}, 0},
		{"tags foo present", &FeedRequest{TagKey: "foo", TagValue: ""}, "", "feeds.#.onestop_id", []string{"BA"}, 0},
		{"url type", &FeedRequest{URLType: "realtime_trip_updates"}, "", "feeds.#.onestop_id", []string{"BA~rt", "CT~rt"}, 0},
		{"url source", &FeedRequest{URL: "file://test/data/external/caltrain.zip"}, "", "feeds.#.onestop_id", []string{"CT"}, 0},
		{"url source and type", &FeedRequest{URL: "file://test/data/external/caltrain.zip", URLType: "static_current"}, "", "feeds.#.onestop_id", []string{"CT"}, 0},
		{"url source case insensitive", &FeedRequest{URL: "file://test/data/external/Caltrain.zip", URLCaseSensitive: false}, "", "feeds.#.onestop_id", []string{"CT"}, 0},
		{"url source case sensitive", &FeedRequest{URL: "file://test/data/external/Caltrain.zip", URLCaseSensitive: true}, "", "feeds.#.onestop_id", []string{}, 0},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
