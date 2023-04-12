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
			expectSelect: []string{"e535eb2b3b9ac3ef15d82c56575e914575e732e0", "d2813c293bcfd7a97dde599527ae6c62c98e66c6", "c969427f56d3a645195dd8365cde6d7feae7e99b", "dd7aca4a8e4c90908fd3603c097fabee75fea907"},
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
	}
	srv, te := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, srv, te, tc)
		})
	}
}
