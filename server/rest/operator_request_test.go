package rest

import (
	"testing"
)

func TestOperatorRequest(t *testing.T) {
	testcases := []testRest{
		{
			name:         "basic",
			h:            OperatorRequest{},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
			expectLength: 0,
		},
		{
			name:         "feed_onestop_id",
			h:            OperatorRequest{FeedOnestopID: "BA"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-bayarearapidtransit"},
			expectLength: 0,
		},
		{
			name:         "onestop_id",
			h:            OperatorRequest{OnestopID: "o-9q9-bayarearapidtransit"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-bayarearapidtransit"},
			expectLength: 0,
		},
		{
			name:         "search",
			h:            OperatorRequest{Search: "bay area"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-bayarearapidtransit"},
			expectLength: 0,
		},
		{
			name:         "tags us_ntd_id=90134",
			h:            OperatorRequest{TagKey: "us_ntd_id", TagValue: "90134"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain"},
			expectLength: 0,
		},
		{
			name:         "tags us_ntd_id present",
			h:            OperatorRequest{TagKey: "us_ntd_id", TagValue: ""},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain"},
			expectLength: 0,
		},
		// {"lat,lon,radius 10m", OperatorRequest{Lon: -122.407974, Lat: 37.784471, Radius: 10}, "", "operators.#.onestop_id", []string{"BART"}, 0},
		// {"lat,lon,radius 2000m", OperatorRequest{Lon: -122.407974, Lat: 37.784471, Radius: 2000}, "", "operators.#.onestop_id", []string{"caltrain-ca-us", "BART"}, 0},
		{
			name:         "adm0name",
			h:            OperatorRequest{Adm0Name: "united states of america"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit", "o-dhv-hillsborougharearegionaltransit"},
			expectLength: 0,
		},
		{
			name:         "adm1name",
			h:            OperatorRequest{Adm1Name: "california"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit"},
			expectLength: 0,
		},
		{
			name:         "adm0iso",
			h:            OperatorRequest{Adm0Iso: "us"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit", "o-dhv-hillsborougharearegionaltransit"},
			expectLength: 0,
		},
		{
			name:         "adm1iso:us-ca",
			h:            OperatorRequest{Adm1Iso: "us-ca"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit"},
			expectLength: 0,
		},
		{
			name:         "adm1iso:us-ny",
			h:            OperatorRequest{Adm1Iso: "us-ny"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{},
			expectLength: 0,
		},
		{
			name:         "city_name:san jose",
			h:            OperatorRequest{CityName: "san jose"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain"},
			expectLength: 0,
		},
		{
			name:         "city_name:oakland",
			h:            OperatorRequest{CityName: "berkeley"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-bayarearapidtransit"},
			expectLength: 0,
		},
		{
			name:         "city_name:new york city",
			h:            OperatorRequest{CityName: "new york city"},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{},
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

func TestOperatorRequest_Pagination(t *testing.T) {
	testcases := []testRest{
		{
			name:         "limit:1",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 1}},
			selector:     "operators.#.onestop_id",
			expectLength: 1,
		},
		{
			name:         "limit:1000",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 1000}},
			selector:     "operators.#.onestop_id",
			expectLength: 4,
		},
	}
	srv, te := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, srv, te, tc)
		})
	}
}

func TestOperatorRequest_License(t *testing.T) {
	testcases := []testRest{
		{
			name:         "license:share_alike_optional yes",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "yes"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-dhv-hillsborougharearegionaltransit"},
		},
		{
			name:         "license:share_alike_optional no",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "no"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-bayarearapidtransit"},
		},
		{
			name:         "license:share_alike_optional exclude_no",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "exclude_no"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
		},
		{
			name:         "license:commercial_use_allowed yes",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "yes"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-dhv-hillsborougharearegionaltransit"},
		},
		{
			name:         "license:commercial_use_allowed no",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "no"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-bayarearapidtransit"},
		},
		{
			name:         "license:commercial_use_allowed exclude_no",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "exclude_no"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
		},
		{
			name:         "license:create_derived_product yes",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "yes"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-dhv-hillsborougharearegionaltransit"},
		},
		{
			name:         "license:create_derived_product no",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "no"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-bayarearapidtransit"},
		},
		{
			name:         "license:create_derived_product exclude_no",
			h:            OperatorRequest{WithCursor: WithCursor{Limit: 10_000}, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "exclude_no"}},
			selector:     "operators.#.onestop_id",
			expectSelect: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
		},
	}
	srv, te := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, srv, te, tc)
		})
	}
}
