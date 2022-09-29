package rest

import (
	"context"
	"testing"
)

func TestRouteRequest(t *testing.T) {
	cfg := testRestConfig()
	routeIds := []string{"1", "12", "14", "15", "16", "17", "19", "20", "24", "25", "275", "30", "31", "32", "33", "34", "35", "36", "360", "37", "38", "39", "400", "42", "45", "46", "48", "5", "51", "6", "60", "7", "75", "8", "9", "96", "97", "570", "571", "572", "573", "574", "800", "PWT", "SKY", "01", "03", "05", "07", "11", "19", "Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"}
	fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	allEnts, err := TestDBFinder.FindRoutes(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	allIds := []string{}
	for _, ent := range allEnts {
		allIds = append(allIds, ent.RouteID)
	}
	testcases := []testRest{
		{"none", RouteRequest{Limit: 1000}, "", "routes.#.route_id", routeIds, 0},
		{"limit:1", RouteRequest{Limit: 1}, "", "routes.#.route_id", nil, 1},
		{"limit:100", RouteRequest{Limit: 100}, "", "routes.#.route_id", nil, len(routeIds)},
		{"search", RouteRequest{Search: "bullet"}, "", "routes.#.route_id", []string{"Bu-130"}, 0},
		{"feed_onestop_id", RouteRequest{FeedOnestopID: "CT"}, "", "routes.#.route_id", []string{"Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"}, 0},
		{"route_type:2", RouteRequest{RouteType: "2"}, "", "routes.#.route_id", []string{"Bu-130", "Li-130", "Lo-130", "Gi-130", "Sp-130"}, 0},
		{"route_type:1", RouteRequest{RouteType: "1"}, "", "routes.#.route_id", []string{"01", "03", "05", "07", "11", "19"}, 0},
		{"feed_onestop_id,route_id", RouteRequest{FeedOnestopID: "BA", RouteID: "19"}, "", "routes.#.route_id", []string{"19"}, 0},
		{"feed_version_sha1", RouteRequest{FeedVersionSHA1: fv}, "", "routes.#.feed_version.sha1", []string{fv, fv, fv, fv, fv, fv}, 0},
		{"operator_onestop_id", RouteRequest{OperatorOnestopID: "o-9q9-bayarearapidtransit"}, "", "routes.#.route_id", []string{"01", "03", "05", "07", "11", "19"}, 0},
		{"lat,lon,radius 100m", RouteRequest{Lon: -122.407974, Lat: 37.784471, Radius: 100}, "", "routes.#.route_id", []string{"01", "05", "07", "11"}, 0},
		{"lat,lon,radius 2000m", RouteRequest{Lon: -122.407974, Lat: 37.784471, Radius: 2000}, "", "routes.#.route_id", []string{"Bu-130", "Li-130", "Lo-130", "Gi-130", "Sp-130", "01", "05", "07", "11"}, 0},
		{"pagination exists", RouteRequest{}, "", "meta.after", nil, 1}, // just check presence
		{"pagination limit 10", RouteRequest{Limit: 10}, "", "routes.#.route_id", allIds[:10], 0},
		{"pagination after 10", RouteRequest{Limit: 10, After: allEnts[10].ID}, "", "routes.#.route_id", allIds[11:21], 0},
		{"feed:route_id", RouteRequest{RouteKey: "BA:01"}, "", "routes.#.route_id", []string{"01"}, 0},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
