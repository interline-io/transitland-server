package rest

import (
	"testing"
)

func TestStopRequest(t *testing.T) {
	cfg := testRestConfig()
	fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	osid := "s-9q8yyufxmv-sanfranciscocaltrain"
	bartstops := []string{"12TH", "16TH", "19TH", "19TH_N", "24TH", "ANTC", "ASHB", "BALB", "BAYF", "CAST", "CIVC", "COLS", "COLM", "CONC", "DALY", "DBRK", "DUBL", "DELN", "PLZA", "EMBR", "FRMT", "FTVL", "GLEN", "HAYW", "LAFY", "LAKE", "MCAR", "MCAR_S", "MLBR", "MONT", "NBRK", "NCON", "OAKL", "ORIN", "PITT", "PCTR", "PHIL", "POWL", "RICH", "ROCK", "SBRN", "SFIA", "SANL", "SHAY", "SSAN", "UCTY", "WCRK", "WARM", "WDUB", "WOAK"}
	caltrainRailStops := []string{"70011", "70012", "70021", "70022", "70031", "70032", "70041", "70042", "70051", "70052", "70061", "70062", "70071", "70072", "70081", "70082", "70091", "70092", "70101", "70102", "70111", "70112", "70121", "70122", "70131", "70132", "70141", "70142", "70151", "70152", "70161", "70162", "70171", "70172", "70191", "70192", "70201", "70202", "70211", "70212", "70221", "70222", "70231", "70232", "70241", "70242", "70251", "70252", "70261", "70262", "70271", "70272", "70281", "70282", "70291", "70292", "70301", "70302", "70311", "70312", "70321", "70322"}
	caltrainBusStops := []string{"777402", "777403"}
	_ = caltrainRailStops
	_ = caltrainBusStops
	testcases := []testRest{
		{"basic", StopRequest{}, "", "stops.#.stop_id", nil, 20},                                  // default
		{"onestop_id", StopRequest{OnestopID: osid}, "", "stops.#.onestop_id", []string{osid}, 0}, // default
		{"stop_id", StopRequest{StopID: "70011"}, "", "stops.#.stop_id", []string{"70011"}, 0},    // default
		{"limit:1", StopRequest{Limit: 1}, "", "stops.#.stop_id", nil, 1},
		{"limit:100", StopRequest{Limit: 100}, "", "stops.#.stop_id", nil, 100},
		{"limit:1000", StopRequest{Limit: 1000}, "", "stops.#.stop_id", nil, 114},
		{"feed_onestop_id", StopRequest{FeedOnestopID: "BA", Limit: 100}, "", "stops.#.stop_id", bartstops, 0},
		{"feed_onestop_id,stop_id", StopRequest{FeedOnestopID: "BA", StopID: "12TH"}, "", "stops.#.stop_id", []string{"12TH"}, 0},
		{"feed_version_sha1", StopRequest{FeedVersionSHA1: fv}, "", "stops.#.stop_id", nil, 20},
		{"feed_version_sha1,limit:100", StopRequest{FeedVersionSHA1: fv, Limit: 100}, "", "stops.#.stop_id", nil, 50},
		// {"served_by_route_types=1", StopRequest{ServedByRouteTypes: []int{1}, Limit: 100}, "", "stops.#.stop_id", bartstops, 0},
		// {"served_by_route_types=2", StopRequest{ServedByRouteTypes: []int{2}, Limit: 100}, "", "stops.#.stop_id", caltrainRailStops, 0},
		// {"served_by_route_types=3", StopRequest{ServedByRouteTypes: []int{3}, Limit: 100}, "", "stops.#.stop_id", caltrainBusStops, 0},
		{"served_by_onestop_ids=o-9q9-bayarearapidtransit", StopRequest{ServedByOnestopIds: []string{"o-9q9-bayarearapidtransit"}, Limit: 100}, "", "stops.#.stop_id", bartstops, 0},
		{"served_by_onestop_ids=o-9q9-bayarearapidtransit,o-9q9-caltrain", StopRequest{ServedByOnestopIds: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"}, Limit: 1000}, "", "stops.#.stop_id", nil, 114},
		// {"served_by_onestop_ids=o-9q9-caltrain,served_by_route_types=3", StopRequest{ServedByOnestopIds: []string{"o-9q9-caltrain"}, ServedByRouteTypes: []int{3}, Limit: 100}, "", "stops.#.stop_id", caltrainBusStops, 0},
		{"lat,lon,radius 10m", StopRequest{Lat: -122.407974, Lon: 37.784471, Radius: 10}, "", "stops.#.stop_id", []string{"POWL"}, 0},
		{"lat,lon,radius 2000m", StopRequest{Lat: -122.407974, Lon: 37.784471, Radius: 2000}, "", "stops.#.stop_id", []string{"70011", "70012", "CIVC", "EMBR", "MONT", "POWL"}, 0},
		{"search", StopRequest{Search: "macarthur"}, "", "stops.#.stop_id", []string{"MCAR", "MCAR_S"}, 0}, // default
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
