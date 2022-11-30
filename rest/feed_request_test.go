package rest

import (
	"testing"
)

func TestFeedRequest(t *testing.T) {
	cfg := testRestConfig()
	// fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	testcases := []testRest{
		{
			name:         "basic",
			h:            &FeedRequest{},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "test-gbfs", "BA", "HA", "BA~rt", "CT~rt", "test"},
			expectLength: 0,
		},
		{
			name:         "onestop_id",
			h:            &FeedRequest{OnestopID: "BA"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
			expectLength: 0,
		},
		{
			name:         "spec",
			h:            &FeedRequest{Spec: "GTFS_RT"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA~rt", "CT~rt"},
			expectLength: 0,
		},
		{
			name:         "search",
			h:            &FeedRequest{Search: "ba"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA~rt", "BA"},
			expectLength: 0,
		},
		{
			name:         "fetch_error true",
			h:            &FeedRequest{FetchError: "true"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"test"},
			expectLength: 0,
		},
		{
			name:         "fetch_error false",
			h:            &FeedRequest{FetchError: "false"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA", "CT", "HA"},
			expectLength: 0,
		},
		{
			name:         "tags test=ok",
			h:            &FeedRequest{TagKey: "test", TagValue: "ok"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
			expectLength: 0,
		},
		{
			name:         "tags foo present",
			h:            &FeedRequest{TagKey: "foo", TagValue: ""},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
			expectLength: 0,
		},
		{
			name:         "url type",
			h:            &FeedRequest{URLType: "realtime_trip_updates"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA~rt", "CT~rt"},
			expectLength: 0,
		},
		{
			name:         "url source",
			h:            &FeedRequest{URL: "file://test/data/external/caltrain.zip"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT"},
			expectLength: 0,
		},
		{
			name:         "url source and type",
			h:            &FeedRequest{URL: "file://test/data/external/caltrain.zip", URLType: "static_current"},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT"},
			expectLength: 0,
		},
		{
			name:         "url source case insensitive",
			h:            &FeedRequest{URL: "file://test/data/external/Caltrain.zip", URLCaseSensitive: false},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT"},
			expectLength: 0,
		},
		{
			name:         "url source case sensitive",
			h:            &FeedRequest{URL: "file://test/data/external/Caltrain.zip", URLCaseSensitive: true},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{},
			expectLength: 0,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}

func TestFeedRequest_License(t *testing.T) {
	testcases := []testRest{
		{
			name: "license:share_alike_optional yes",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "yes"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"HA"},
		},
		{
			name: "license:share_alike_optional no",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "no"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name: "license:share_alike_optional exclude_no",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "exclude_no"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"CT", "test-gbfs", "HA", "BA~rt", "CT~rt", "test"},
		},
		{
			name: "license:commercial_use_allowed yes",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "yes"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"HA"},
		},
		{
			name: "license:commercial_use_allowed no",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "no"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name: "license:commercial_use_allowed exclude_no",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "exclude_no"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"CT", "test-gbfs", "HA", "BA~rt", "CT~rt", "test"},
		},
		{
			name: "license:create_derived_product yes",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "yes"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"HA"},
		},
		{
			name: "license:create_derived_product no",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "no"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
		},
		{
			name: "license:create_derived_product exclude_no",
			h:    FeedRequest{LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "exclude_no"}}, selector: "feeds.#.onestop_id",
			expectSelect: []string{"CT", "test-gbfs", "HA", "BA~rt", "CT~rt", "test"},
		},
	}
	cfg := testRestConfig()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}

}
