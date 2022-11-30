package resolvers

import (
	"context"
	"testing"
)

func TestAgencyResolver(t *testing.T) {
	vars := hw{"agency_id": "caltrain-ca-us"}
	testcases := []testcase{
		{
			name:  "basic",
			query: `query { agencies {agency_id}}`,

			selector:     "agencies.#.agency_id",
			selectExpect: []string{"caltrain-ca-us", "BART", ""},
		},
		{
			name:   "basic fields",
			query:  `query($agency_id:String!) { agencies(where:{agency_id:$agency_id}) {onestop_id agency_id agency_name agency_lang agency_phone agency_timezone agency_url agency_email agency_fare_url feed_version_sha1 feed_onestop_id}}`,
			vars:   vars,
			expect: `{"agencies":[{"agency_email":"","agency_fare_url":"","agency_id":"caltrain-ca-us","agency_lang":"en","agency_name":"Caltrain","agency_phone":"800-660-4287","agency_timezone":"America/Los_Angeles","agency_url":"http://www.caltrain.com","feed_onestop_id":"CT","feed_version_sha1":"d2813c293bcfd7a97dde599527ae6c62c98e66c6","onestop_id":"o-9q9-caltrain"}]}`,
		},
		{
			// just ensure this query completes successfully; checking coordinates is a pain and flaky.
			name:         "geometry",
			query:        `query($agency_id:String!) { agencies(where:{agency_id:$agency_id}) {geometry}}`,
			vars:         vars,
			selector:     "agencies.0.geometry.type",
			selectExpect: []string{"Polygon"},
		},
		{
			name:  "near 100m",
			query: `query {agencies(where:{near:{lon:-122.407974,lat:37.784471,radius:100.0}}) {agency_id}}`,

			selector:     "agencies.#.agency_id",
			selectExpect: []string{"BART"},
		},
		{
			name:  "near 10000m",
			query: `query {agencies(where:{near:{lon:-122.407974,lat:37.784471,radius:10000.0}}) {agency_id}}`,

			selector:     "agencies.#.agency_id",
			selectExpect: []string{"caltrain-ca-us", "BART"},
		},
		{
			name:  "within polygon",
			query: `query{agencies(where:{within:{type:"Polygon",coordinates:[[[-122.39803791046143,37.794626736533836],[-122.40106344223022,37.792303711508595],[-122.3965573310852,37.789641468930114],[-122.3938751220703,37.792354581451946],[-122.39803791046143,37.794626736533836]]]}}){agency_id}}`,

			selector:     "agencies.#.agency_id",
			selectExpect: []string{"BART"},
		},
		{
			name:  "within polygon big",
			query: `query{agencies(where:{within:{type:"Polygon",coordinates:[[[-122.39481925964355,37.80151060070086],[-122.41653442382812,37.78652126637423],[-122.39662170410156,37.76847577247014],[-122.37301826477051,37.784757615348575],[-122.39481925964355,37.80151060070086]]]}}){id agency_id}}`,

			selector:     "agencies.#.agency_id",
			selectExpect: []string{"caltrain-ca-us", "BART"},
		},
		{
			name:   "feed_version",
			query:  `query($agency_id:String!) { agencies(where:{agency_id:$agency_id}) {feed_version { sha1 }}}`,
			vars:   vars,
			expect: `{"agencies":[{"feed_version":{"sha1":"d2813c293bcfd7a97dde599527ae6c62c98e66c6"}}]}`,
		},
		{
			name:         "routes",
			query:        `query($agency_id:String!) { agencies(where:{agency_id:$agency_id}) {routes { route_id }}}`,
			vars:         vars,
			selector:     "agencies.0.routes.#.route_id",
			selectExpect: []string{"Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"},
		},
		// places should test filters because it's not a root resolver
		{
			name:         "places",
			query:        `query($agency_id:String!) { agencies(where:{agency_id:$agency_id}) {places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.0.places.#.city_name",
			selectExpect: []string{"San Mateo", "San Francisco", "San Jose", ""},
		},
		{
			name:         "places rank 0.25",
			query:        `query($agency_id:String!) { agencies(where:{agency_id:$agency_id}) {places(where:{min_rank:0.25}) {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.0.places.#.city_name",
			selectExpect: []string{"San Mateo", "San Jose", ""},
		},
		{
			name:         "places rank 0.75",
			query:        `query($agency_id:String!) { agencies(where:{agency_id:$agency_id}) {places(where:{min_rank:0.75}) {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.0.places.#.adm1_name",
			selectExpect: []string{"California"},
		},
		// place iso codes
		{
			name:         "places iso3166 country",
			query:        `query { agencies(where:{adm0_iso: "US"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit"},
		},
		{
			name:         "places iso3166 state",
			query:        `query { agencies(where:{adm1_iso: "US-CA"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			name:         "places iso3166 state lowercase",
			query:        `query { agencies(where:{adm1_iso: "us-ca"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			name:         "places iso3166 state and country",
			query:        `query { agencies(where:{adm0_iso: "us", adm1_iso: "us-ca"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			name:         "places iso3166 state and city",
			query:        `query { agencies(where:{city_name: "oakland", adm1_iso: "us-ca"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit"},
		},
		{
			name:         "places iso3166 state and city no result",
			query:        `query { agencies(where:{city_name: "test", adm1_iso: "us-ca"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{},
		},
		{
			name:         "places iso3166 state no results",
			query:        `query { agencies(where:{adm1_iso: "US-NY"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{},
		},
		{
			name:         "places state",
			query:        `query { agencies(where:{adm1_name: "California"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			name:         "places state no result",
			query:        `query { agencies(where:{adm1_name: "New York"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{},
		},
		{
			name:         "places city",
			query:        `query { agencies(where:{city_name: "Berkeley"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit"},
		},
		{
			name:         "places city 2",
			query:        `query { agencies(where:{city_name: "San Jose"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-caltrain"},
		},
		{
			name:         "places city 2 lowercase",
			query:        `query { agencies(where:{city_name: "san jose"}) {onestop_id places {adm0_name adm1_name city_name}}}`,
			vars:         vars,
			selector:     "agencies.#.onestop_id",
			selectExpect: []string{"o-9q9-caltrain"},
		},
		// search
		{
			name:         "search",
			query:        `query($search:String!) { agencies(where:{search:$search}) {agency_id}}`,
			vars:         hw{"search": "Bay Area"},
			selector:     "agencies.#.agency_id",
			selectExpect: []string{"BART"},
		},
		{
			name:         "search",
			query:        `query($search:String!) { agencies(where:{search:$search}) {agency_id}}`,
			vars:         hw{"search": "caltrain"},
			selector:     "agencies.#.agency_id",
			selectExpect: []string{"caltrain-ca-us"},
		},
		// TODO
		// {"census_geographies", }
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestAgencyResolver_Cursor(t *testing.T) {
	allEnts, err := TestDBFinder.FindAgencies(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	allIds := []string{}
	for _, ent := range allEnts {
		allIds = append(allIds, ent.AgencyID)
	}
	testcases := []testcase{
		{
			name:         "no cursor",
			query:        "query{agencies(limit:10){feed_version{id} id agency_id}}",
			vars:         nil,
			selector:     "agencies.#.agency_id",
			selectExpect: allIds,
		},
		{
			name:         "after 0",
			query:        "query{agencies(after: 0, limit:10){feed_version{id} id agency_id}}",
			vars:         nil,
			selector:     "agencies.#.agency_id",
			selectExpect: allIds,
		},
		{
			name:         "after 1st",
			query:        "query($after: Int!){agencies(after: $after, limit:10){feed_version{id} id agency_id}}",
			vars:         hw{"after": allEnts[1].ID},
			selector:     "agencies.#.agency_id",
			selectExpect: allIds[2:],
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
