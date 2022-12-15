package rest

import (
	"context"
	"strconv"
	"testing"

	"github.com/tidwall/gjson"
)

func TestTripRequest(t *testing.T) {
	cfg := testRestConfig()
	d, err := makeGraphQLRequest(context.Background(), cfg.srv, `query{routes(where:{feed_onestop_id:"BA",route_id:"11"}) {id onestop_id}}`, nil)
	if err != nil {
		t.Error("failed to get route id for tests")
	}
	routeId := int(gjson.Get(toJson(d), "routes.0.id").Int())
	routeOnestopId := gjson.Get(toJson(d), "routes.0.onestop_id").String()
	d2, err := makeGraphQLRequest(context.Background(), cfg.srv, `query{trips(where:{trip_id:"5132248WKDY"}){id}}`, nil)
	if err != nil {
		t.Error("failed to get route id for tests")
	}
	tripId := int(gjson.Get(toJson(d2), "trips.0.id").Int())

	fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	ctfv := "d2813c293bcfd7a97dde599527ae6c62c98e66c6"
	testcases := []testRest{
		{
			name:         "none",
			h:            TripRequest{},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 20,
		},
		{
			name:         "feed_onestop_id",
			h:            TripRequest{FeedOnestopID: "BA"},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 20,
		},
		{
			name:         "feed_onestop_id ct",
			h:            TripRequest{FeedOnestopID: "CT", Limit: 1000},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 185,
		},
		{
			name:         "feed_version_sha1",
			h:            TripRequest{FeedVersionSHA1: ctfv, Limit: 1000},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 185,
		},
		{
			name:         "feed_version_sha1 ba",
			h:            TripRequest{FeedVersionSHA1: fv, Limit: 1000},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 1000,
		}, // over 1000
		{
			name:         "trip_id",
			h:            TripRequest{TripID: "5132248WKDY"},
			selector:     "trips.#.trip_id",
			expectSelect: []string{"5132248WKDY"},
			expectLength: 0,
		},
		{
			name:         "trip_id,feed_version_id",
			h:            TripRequest{TripID: "5132248WKDY", FeedVersionSHA1: fv},
			selector:     "trips.#.trip_id",
			expectSelect: []string{"5132248WKDY"},
			expectLength: 0,
		},
		{
			name:         "route_id",
			h:            TripRequest{Limit: 1000, RouteID: routeId},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 364,
		},
		{
			name:         "route_id,service_date 1",
			h:            TripRequest{Limit: 1000, RouteID: routeId, ServiceDate: "2018-01-01"},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 0,
		},
		{
			name:         "route_id,service_date 2",
			h:            TripRequest{Limit: 1000, RouteID: routeId, ServiceDate: "2019-01-01"},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 100,
		},
		{
			name:         "route_id,service_date 3",
			h:            TripRequest{Limit: 1000, RouteID: routeId, ServiceDate: "2019-01-02"},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 152,
		},
		{
			name:         "route_id,service_date 4",
			h:            TripRequest{Limit: 1000, RouteID: routeId, ServiceDate: "2020-05-18"},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 0,
		},
		{
			name:         "route_id,trip_id",
			h:            TripRequest{Limit: 1000, RouteID: routeId, TripID: "5132248WKDY"},
			selector:     "trips.#.trip_id",
			expectSelect: []string{"5132248WKDY"},
			expectLength: 0,
		},
		{
			name:         "include_geometry=true",
			h:            TripRequest{TripID: "5132248WKDY", IncludeGeometry: "true"},
			selector:     "trips.0.shape.geometry.type",
			expectSelect: []string{"LineString"},
			expectLength: 0,
		},
		{
			name:         "include_geometry=false",
			h:            TripRequest{TripID: "5132248WKDY", IncludeGeometry: "false"},
			selector:     "trips.0.shape.geometry.type",
			expectSelect: []string{},
			expectLength: 0,
		},
		{
			name:         "does not include stop_times without id",
			h:            TripRequest{TripID: "5132248WKDY"},
			selector:     "trips.0.stop_times.#.stop_sequence",
			expectSelect: nil,
			expectLength: 0,
		},
		{
			name:         "id includes stop_times",
			h:            TripRequest{ID: tripId},
			selector:     "trips.0.stop_times.#.stop_sequence",
			expectSelect: nil,
			expectLength: 18,
		},
		{
			name:         "route_key onestop_id",
			h:            TripRequest{Limit: 1000, RouteKey: routeOnestopId},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 364,
		},
		{
			name:         "route_key int",
			h:            TripRequest{Limit: 1000, RouteKey: strconv.Itoa(routeId)},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 364,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}

func TestTripRequest_Pagination(t *testing.T) {
	testcases := []testRest{
		{
			name:         "limit:1",
			h:            TripRequest{Limit: 1},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 1,
		},
		{
			name:         "limit:100",
			h:            TripRequest{Limit: 100},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 100,
		},
		{
			name:         "limit:1000",
			h:            TripRequest{Limit: 1000},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 1000,
		},
		{
			name:         "limit:10000",
			h:            TripRequest{Limit: 10_000},
			selector:     "trips.#.trip_id",
			expectSelect: nil,
			expectLength: 10_000,
		},
	}
	cfg := testRestConfig()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}

func TestTripRequest_License(t *testing.T) {
	testcases := []testRest{
		{
			name: "license:share_alike_optional yes",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "yes"}}, selector: "trips.#.trip_id",
			expectLength: 14718,
		},
		{
			name: "license:share_alike_optional no",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "no"}}, selector: "trips.#.trip_id",
			expectLength: 2525,
		},
		{
			name: "license:share_alike_optional exclude_no",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "exclude_no"}}, selector: "trips.#.trip_id",
			expectLength: 14903,
		},
		{
			name: "license:commercial_use_allowed yes",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "yes"}}, selector: "trips.#.trip_id",
			expectLength: 14718,
		},
		{
			name: "license:commercial_use_allowed no",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "no"}}, selector: "trips.#.trip_id",
			expectLength: 2525,
		},
		{
			name: "license:commercial_use_allowed exclude_no",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "exclude_no"}}, selector: "trips.#.trip_id",
			expectLength: 14903,
		},
		{
			name: "license:create_derived_product yes",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "yes"}}, selector: "trips.#.trip_id",
			expectLength: 14718,
		},
		{
			name: "license:create_derived_product no",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "no"}}, selector: "trips.#.trip_id",
			expectLength: 2525,
		},
		{
			name: "license:create_derived_product exclude_no",
			h:    TripRequest{Limit: 100_000, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "exclude_no"}}, selector: "trips.#.trip_id",
			expectLength: 14903,
		},
	}
	cfg := testRestConfig()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
