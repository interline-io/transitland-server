package rest

import (
	"testing"

	"github.com/interline-io/transitland-server/model"
)

func TestFeedRequest(t *testing.T) {
	// fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	testcases := []testRest{
		{
			name:         "basic",
			h:            &FeedRequest{},
			format:       "",
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "test-gbfs", "BA", "HA", "BA~rt", "CT~rt", "test", "EX"},
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
			expectSelect: []string{"BA", "CT", "HA", "EX"},
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
		// spatial
		{
			name:         "lat,lon,radius 100m",
			h:            FeedRequest{Lon: -122.407974, Lat: 37.784471, Radius: 100},
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
			expectLength: 0,
		},
		{
			name:         "lat,lon,radius 2000m",
			h:            FeedRequest{Lon: -122.407974, Lat: 37.784471, Radius: 2000},
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"CT", "BA"},
			expectLength: 0,
		},
		{
			name:         "bbox",
			h:            FeedRequest{Bbox: &restBbox{model.BoundingBox{MinLon: -122.2698781543005, MinLat: 37.80700393130445, MaxLon: -122.2677640139239, MaxLat: 37.8088734037938}}},
			selector:     "feeds.#.onestop_id",
			expectSelect: []string{"BA"},
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
			expectSelect: []string{"CT", "test-gbfs", "HA", "BA~rt", "CT~rt", "test", "EX"},
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
			expectSelect: []string{"CT", "test-gbfs", "HA", "BA~rt", "CT~rt", "test", "EX"},
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
			expectSelect: []string{"CT", "test-gbfs", "HA", "BA~rt", "CT~rt", "test", "EX"},
		},
	}
	srv, te := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, srv, te, tc)
		})
	}

}
