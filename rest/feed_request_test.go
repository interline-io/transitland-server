package rest

import (
	"testing"
)

func TestFeedRequest(t *testing.T) {
	cfg := testRestConfig()
	// fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	testcases := []testRest{
		{"basic", &FeedRequest{}, "", "feeds.#.onestop_id", []string{"CT", "BA", "BA~rt", "test"}, 0},
		{"onestop_id", &FeedRequest{OnestopID: "BA"}, "", "feeds.#.onestop_id", []string{"BA"}, 0},
		{"spec", &FeedRequest{Spec: "gtfs-rt"}, "", "feeds.#.onestop_id", []string{"BA~rt"}, 0},
		{"search", &FeedRequest{Search: "ba"}, "", "feeds.#.onestop_id", []string{"BA~rt", "BA"}, 0},
		{"fetch_error true", &FeedRequest{FetchError: "true"}, "", "feeds.#.onestop_id", []string{"test"}, 0},
		{"fetch_error false", &FeedRequest{FetchError: "false"}, "", "feeds.#.onestop_id", []string{"BA", "CT"}, 0},
		{"tags test=ok", &FeedRequest{TagKey: "test", TagValue: "ok"}, "", "feeds.#.onestop_id", []string{"BA"}, 0},
		{"tags foo present", &FeedRequest{TagKey: "foo", TagValue: ""}, "", "feeds.#.onestop_id", []string{"BA"}, 0},
	}
	for _, tc := range testcases {
		testquery(t, cfg, tc)
	}
}
