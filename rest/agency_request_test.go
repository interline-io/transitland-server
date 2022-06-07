package rest

import (
	"context"
	"testing"
)

func TestAgencyRequest(t *testing.T) {
	cfg := testRestConfig()
	fv := "e535eb2b3b9ac3ef15d82c56575e914575e732e0"
	allEnts, err := TestDBFinder.FindAgencies(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	allIds := []string{}
	for _, ent := range allEnts {
		allIds = append(allIds, ent.AgencyID)
	}
	testcases := []testRest{
		{"basic", AgencyRequest{}, "", "agencies.#.agency_id", []string{"caltrain-ca-us", "BART", ""}, 0},
		{"limit:1", AgencyRequest{Limit: 1}, "", "agencies.#.agency_id", nil, 1}, // this used to be caltrain but now bart is imported first.
		{"feed_version_sha1", AgencyRequest{FeedVersionSHA1: fv}, "", "agencies.#.agency_id", []string{"BART"}, 0},
		{"feed_onestop_id", AgencyRequest{FeedOnestopID: "BA"}, "", "agencies.#.agency_id", []string{"BART"}, 0},
		{"feed_onestop_id,agency_id", AgencyRequest{FeedOnestopID: "BA", AgencyID: "BART"}, "", "agencies.#.agency_id", []string{"BART"}, 0},
		{"agency_id", AgencyRequest{AgencyID: "BART"}, "", "agencies.#.agency_id", []string{"BART"}, 0},
		{"agency_name", AgencyRequest{AgencyName: "Bay Area Rapid Transit"}, "", "agencies.#.agency_name", []string{"Bay Area Rapid Transit"}, 0},
		{"onestop_id", AgencyRequest{OnestopID: "o-9q9-bayarearapidtransit"}, "", "agencies.#.onestop_id", []string{"o-9q9-bayarearapidtransit"}, 0},
		{"onestop_id,feed_version_sha1", AgencyRequest{OnestopID: "o-9q9-bayarearapidtransit", FeedVersionSHA1: fv}, "", "agencies.#.feed_version.sha1", []string{fv}, 0},
		{"agency_key onestop_id", AgencyRequest{AgencyKey: "o-9q9-bayarearapidtransit"}, "", "agencies.#.onestop_id", []string{"o-9q9-bayarearapidtransit"}, 0},
		{"lat,lon,radius 10m", AgencyRequest{Lon: -122.407974, Lat: 37.784471, Radius: 10}, "", "agencies.#.agency_id", []string{"BART"}, 0},
		{"lat,lon,radius 2000m", AgencyRequest{Lon: -122.407974, Lat: 37.784471, Radius: 2000}, "", "agencies.#.agency_id", []string{"caltrain-ca-us", "BART"}, 0},
		{"search", AgencyRequest{Search: "caltrain"}, "", "agencies.#.agency_id", []string{"caltrain-ca-us"}, 0},
		{"adm0name", AgencyRequest{Adm0Name: "united states of america"}, "", "agencies.#.agency_id", []string{"caltrain-ca-us", "BART", ""}, 0},
		{"adm1name", AgencyRequest{Adm1Name: "california"}, "", "agencies.#.agency_id", []string{"caltrain-ca-us", "BART"}, 0},
		{"adm0iso", AgencyRequest{Adm0Iso: "us"}, "", "agencies.#.agency_id", []string{"caltrain-ca-us", "BART", ""}, 0},
		{"adm1iso:us-ca", AgencyRequest{Adm1Iso: "us-ca"}, "", "agencies.#.agency_id", []string{"caltrain-ca-us", "BART"}, 0},
		{"adm1iso:us-ny", AgencyRequest{Adm1Iso: "us-ny"}, "", "agencies.#.agency_id", []string{}, 0},
		{"city_name:san jose", AgencyRequest{CityName: "san jose"}, "", "agencies.#.agency_id", []string{"caltrain-ca-us"}, 0},
		{"city_name:oakland", AgencyRequest{CityName: "berkeley"}, "", "agencies.#.agency_id", []string{"BART"}, 0},
		{"city_name:new york city", AgencyRequest{CityName: "new york city"}, "", "agencies.#.agency_id", []string{}, 0},
		{"pagination exists", AgencyRequest{}, "", "meta.after", nil, 1}, // just check presence
		{"pagination limit 1", AgencyRequest{Limit: 1}, "", "agencies.#.agency_id", allIds[:1], 0},
		{"pagination after 1", AgencyRequest{Limit: 1, After: allEnts[0].ID}, "", "agencies.#.agency_id", allIds[1:2], 0},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, cfg, tc)
		})
	}
}
