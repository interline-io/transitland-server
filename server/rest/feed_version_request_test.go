package rest

import (
	"testing"
)

func TestFeedVersionRequest(t *testing.T) {
	fv := "d2813c293bcfd7a97dde599527ae6c62c98e66c6"
	testcases := []testRest{
		{
			name:         "basic",
			h:            FeedVersionRequest{},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"e535eb2b3b9ac3ef15d82c56575e914575e732e0", "d2813c293bcfd7a97dde599527ae6c62c98e66c6", "c969427f56d3a645195dd8365cde6d7feae7e99b", "dd7aca4a8e4c90908fd3603c097fabee75fea907", "43e2278aa272879c79460582152b04e7487f0493"},
			expectLength: 0,
		},
		{
			name:         "limit:1",
			h:            FeedVersionRequest{Limit: 1},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{},
			expectLength: 1,
		},
		{
			name:         "sha1",
			h:            FeedVersionRequest{FeedVersionKey: fv},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{fv},
			expectLength: 0,
		},
		{
			name:         "feed_onestop_id,limit:100",
			h:            FeedVersionRequest{Limit: 100, FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"e535eb2b3b9ac3ef15d82c56575e914575e732e0", "dd7aca4a8e4c90908fd3603c097fabee75fea907"},
			expectLength: 0,
		},
		{
			name:         "fetched_after",
			h:            FeedVersionRequest{FetchedAfter: "2009-08-07T06:05:04.3Z", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"dd7aca4a8e4c90908fd3603c097fabee75fea907", "e535eb2b3b9ac3ef15d82c56575e914575e732e0"},
		},
		{
			name:         "fetched_after 2",
			h:            FeedVersionRequest{FetchedAfter: "2123-04-05T06:07:08.9Z", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{},
		},
		{
			name:         "fetched_before",
			h:            FeedVersionRequest{FetchedBefore: "2123-04-05T06:07:08.9Z", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"dd7aca4a8e4c90908fd3603c097fabee75fea907", "e535eb2b3b9ac3ef15d82c56575e914575e732e0"},
		},
		{
			name:         "fetched_before 2",
			h:            FeedVersionRequest{FetchedBefore: "2009-08-07T06:05:04.3Z", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{},
		},
		{
			name:         "covers_start_date",
			h:            FeedVersionRequest{CoversStartDate: "2016-12-31", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"dd7aca4a8e4c90908fd3603c097fabee75fea907"},
		},
		{
			name:         "covers_start_date 2",
			h:            FeedVersionRequest{CoversStartDate: "2010-01-01", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{},
		},
		{
			name:         "covers_end_date",
			h:            FeedVersionRequest{CoversEndDate: "2016-12-31", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"dd7aca4a8e4c90908fd3603c097fabee75fea907"},
		},
		{
			name:         "covers_end_date 2",
			h:            FeedVersionRequest{CoversEndDate: "2040-01-01", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{},
		},
		{
			name:         "covers_start_date and covers_end_date",
			h:            FeedVersionRequest{CoversStartDate: "2016-12-01", CoversEndDate: "2016-12-31", FeedOnestopID: "BA"},
			format:       "",
			selector:     "feed_versions.#.sha1",
			expectSelect: []string{"dd7aca4a8e4c90908fd3603c097fabee75fea907"},
		},
	}
	srv, te := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, srv, te, tc)
		})
	}
}