package rest

import (
	"context"
	"testing"
)

func TestRouteRequest(t *testing.T) {
	routeIds := []string{"1", "12", "14", "15", "16", "17", "19", "20", "24", "25", "275", "30", "31", "32", "33", "34", "35", "36", "360", "37", "38", "39", "400", "42", "45", "46", "48", "5", "51", "6", "60", "7", "75", "8", "9", "96", "97", "570", "571", "572", "573", "574", "800", "PWT", "SKY", "01", "03", "05", "07", "11", "19", "Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"}
	fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	testcases := []testRest{
		{
			name:         "none",
			h:            RouteRequest{Limit: 1000},
			selector:     "routes.#.route_id",
			expectSelect: routeIds,
			expectLength: 0,
		},
		{
			name:         "search",
			h:            RouteRequest{Search: "bullet"},
			selector:     "routes.#.route_id",
			expectSelect: []string{"Bu-130"},
			expectLength: 0,
		},
		{
			name:         "feed_onestop_id",
			h:            RouteRequest{FeedOnestopID: "CT"},
			selector:     "routes.#.route_id",
			expectSelect: []string{"Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"},
			expectLength: 0,
		},
		{

			name:         "route_type:2",
			h:            RouteRequest{RouteType: "2"},
			selector:     "routes.#.route_id",
			expectSelect: []string{"Bu-130", "Li-130", "Lo-130", "Gi-130", "Sp-130"},
			expectLength: 0,
		},
		{
			name:         "route_type:1",
			h:            RouteRequest{RouteType: "1"},
			selector:     "routes.#.route_id",
			expectSelect: []string{"01", "03", "05", "07", "11", "19"},
			expectLength: 0,
		},
		{
			name:         "feed_onestop_id,route_id",
			h:            RouteRequest{FeedOnestopID: "BA", RouteID: "19"},
			selector:     "routes.#.route_id",
			expectSelect: []string{"19"},
			expectLength: 0,
		},
		{
			name:         "feed_version_sha1",
			h:            RouteRequest{FeedVersionSHA1: fv},
			selector:     "routes.#.feed_version.sha1",
			expectSelect: []string{fv, fv, fv, fv, fv, fv},
			expectLength: 0,
		},
		{
			name:         "operator_onestop_id",
			h:            RouteRequest{OperatorOnestopID: "o-9q9-bayarearapidtransit"},
			selector:     "routes.#.route_id",
			expectSelect: []string{"01", "03", "05", "07", "11", "19"},
			expectLength: 0,
		},
		{
			name:         "lat,lon,radius 100m",
			h:            RouteRequest{Lon: -122.407974, Lat: 37.784471, Radius: 100},
			selector:     "routes.#.route_id",
			expectSelect: []string{"01", "05", "07", "11"},
			expectLength: 0,
		},
		{
			name:         "lat,lon,radius 2000m",
			h:            RouteRequest{Lon: -122.407974, Lat: 37.784471, Radius: 2000},
			selector:     "routes.#.route_id",
			expectSelect: []string{"Bu-130", "Li-130", "Lo-130", "Gi-130", "Sp-130", "01", "05", "07", "11"},
			expectLength: 0,
		},
		{
			name:         "feed:route_id",
			h:            RouteRequest{RouteKey: "BA:01"},
			selector:     "routes.#.route_id",
			expectSelect: []string{"01"},
			expectLength: 0,
		},
	}
	cfg, _, _, _ := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}

func TestRouteRequest_Pagination(t *testing.T) {
	cfg, dbf, _, _ := testRestConfig(t)
	allEnts, err := dbf.FindRoutes(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	allIds := []string{}
	for _, ent := range allEnts {
		allIds = append(allIds, ent.RouteID)
	}
	testcases := []testRest{
		{
			name:         "limit:1",
			h:            RouteRequest{Limit: 1},
			selector:     "routes.#.route_id",
			expectSelect: nil,
			expectLength: 1,
		},
		{
			name:         "limit:100",
			h:            RouteRequest{Limit: 100},
			selector:     "routes.#.route_id",
			expectSelect: nil,
			expectLength: 57,
		},
		{
			name:         "pagination exists",
			h:            RouteRequest{},
			selector:     "meta.after",
			expectSelect: nil,
			expectLength: 1,
		}, // just check presence
		{
			name:         "pagination limit 10",
			h:            RouteRequest{Limit: 10},
			selector:     "routes.#.route_id",
			expectSelect: allIds[:10],
			expectLength: 0,
		},
		{
			name:         "pagination after 10",
			h:            RouteRequest{Limit: 10, After: allEnts[10].ID},
			selector:     "routes.#.route_id",
			expectSelect: allIds[11:21],
			expectLength: 0,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}

func TestRouteRequest_License(t *testing.T) {
	testcases := []testRest{
		{
			name: "license:share_alike_optional yes",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "yes"}}, selector: "routes.#.route_id",
			expectLength: 45,
		},
		{
			name: "license:share_alike_optional no",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "no"}}, selector: "routes.#.route_id",
			expectLength: 6,
		},
		{
			name: "license:share_alike_optional exclude_no",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseShareAlikeOptional: "exclude_no"}}, selector: "routes.#.route_id",
			expectLength: 51,
		},
		{
			name: "license:commercial_use_allowed yes",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "yes"}}, selector: "routes.#.route_id",
			expectLength: 45,
		},
		{
			name: "license:commercial_use_allowed no",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "no"}}, selector: "routes.#.route_id",
			expectLength: 6,
		},
		{
			name: "license:commercial_use_allowed exclude_no",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseCommercialUseAllowed: "exclude_no"}}, selector: "routes.#.route_id",
			expectLength: 51,
		},
		{
			name: "license:create_derived_product yes",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "yes"}}, selector: "routes.#.route_id",
			expectLength: 45,
		},
		{
			name: "license:create_derived_product no",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "no"}}, selector: "routes.#.route_id",
			expectLength: 6,
		},
		{
			name: "license:create_derived_product exclude_no",
			h:    RouteRequest{Limit: 10_000, LicenseFilter: LicenseFilter{LicenseCreateDerivedProduct: "exclude_no"}}, selector: "routes.#.route_id",
			expectLength: 51,
		},
	}
	cfg, _, _, _ := testRestConfig(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
