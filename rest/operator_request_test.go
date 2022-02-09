package rest

import (
	"testing"
)

func TestOperatorRequest(t *testing.T) {
	cfg := testRestConfig()
	testcases := []testRest{
		{"basic", OperatorRequest{}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit"}, 0},
		{"feed_onestop_id", OperatorRequest{FeedOnestopID: "BA"}, "", "operators.#.onestop_id", []string{"o-9q9-bayarearapidtransit"}, 0},
		{"onestop_id", OperatorRequest{OnestopID: "o-9q9-bayarearapidtransit"}, "", "operators.#.onestop_id", []string{"o-9q9-bayarearapidtransit"}, 0},
		{"search", OperatorRequest{Search: "bay area"}, "", "operators.#.onestop_id", []string{"o-9q9-bayarearapidtransit"}, 0},
		{"tags us_ntd_id=90134", OperatorRequest{TagKey: "us_ntd_id", TagValue: "90134"}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain"}, 0},
		{"tags us_ntd_id present", OperatorRequest{TagKey: "us_ntd_id", TagValue: ""}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain"}, 0},
		// {"lat,lon,radius 10m", OperatorRequest{Lon: -122.407974, Lat: 37.784471, Radius: 10}, "", "operators.#.onestop_id", []string{"BART"}, 0},
		// {"lat,lon,radius 2000m", OperatorRequest{Lon: -122.407974, Lat: 37.784471, Radius: 2000}, "", "operators.#.onestop_id", []string{"caltrain-ca-us", "BART"}, 0},
		{"adm0name", OperatorRequest{Adm0Name: "united states of america"}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit"}, 0},
		{"adm1name", OperatorRequest{Adm1Name: "california"}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit"}, 0},
		{"adm0iso", OperatorRequest{Adm0Iso: "us"}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit"}, 0},
		{"adm1iso:us-ca", OperatorRequest{Adm1Iso: "us-ca"}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain", "o-9q9-bayarearapidtransit"}, 0},
		{"adm1iso:us-ny", OperatorRequest{Adm1Iso: "us-ny"}, "", "operators.#.onestop_id", []string{}, 0},
		{"city_name:san jose", OperatorRequest{CityName: "san jose"}, "", "operators.#.onestop_id", []string{"o-9q9-caltrain"}, 0},
		{"city_name:oakland", OperatorRequest{CityName: "berkeley"}, "", "operators.#.onestop_id", []string{"o-9q9-bayarearapidtransit"}, 0},
		{"city_name:new york city", OperatorRequest{CityName: "new york city"}, "", "operators.#.onestop_id", []string{}, 0},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
