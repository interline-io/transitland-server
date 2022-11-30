package resolvers

import (
	"testing"
)

func TestOperatorResolver(t *testing.T) {
	testcases := []testcase{
		{
			name:   "basic fields",
			query:  `query{operators(where:{onestop_id:"o-9q9-bayarearapidtransit"}) {onestop_id}}`,
			vars:   hw{},
			expect: `{"operators":[{"onestop_id":"o-9q9-bayarearapidtransit"}]}`,
		},
		{
			name:         "feeds",
			query:        `query{operators(where:{onestop_id:"o-9q9-bayarearapidtransit"}) {feeds{onestop_id}}}`,
			selector:     "operators.0.feeds.#.onestop_id",
			selectExpect: []string{"BA"},
		},
		{
			name:         "feeds incl rt",
			query:        `query{operators(where:{onestop_id:"o-9q9-caltrain"}) {feeds{onestop_id}}}`,
			selector:     "operators.0.feeds.#.onestop_id",
			selectExpect: []string{"CT", "CT~rt"},
		},
		{
			name:         "feeds only gtfs-rt",
			query:        `query{operators(where:{onestop_id:"o-9q9-caltrain"}) {feeds(where:{spec:GTFS_RT}) {onestop_id}}}`,
			selector:     "operators.0.feeds.#.onestop_id",
			selectExpect: []string{"CT~rt"},
		},
		{
			name:         "feeds only gtfs",
			query:        `query{operators(where:{onestop_id:"o-9q9-caltrain"}) {feeds(where:{spec:GTFS}) {onestop_id}}}`,
			selector:     "operators.0.feeds.#.onestop_id",
			selectExpect: []string{"CT"},
		},
		{
			name:   "tags us_ntd_id=90134",
			query:  `query{operators(where:{tags:{us_ntd_id:"90134"}}) {onestop_id}}`,
			vars:   hw{},
			expect: `{"operators":[{"onestop_id":"o-9q9-caltrain"}]}`,
		},
		{
			name:   "tags us_ntd_id=12345",
			query:  `query{operators(where:{tags:{us_ntd_id:"12345"}}) {onestop_id}}`,
			vars:   hw{},
			expect: `{"operators":[]}`,
		},
		{
			name:   "tags us_ntd_id presence",
			query:  `query{operators(where:{tags:{us_ntd_id:""}}) {onestop_id}}`,
			vars:   hw{},
			expect: `{"operators":[{"onestop_id":"o-9q9-caltrain"}]}`,
		},
		{
			name:         "places iso3166 country",
			query:        `query { operators(where:{adm0_iso: "US"}) {onestop_id}}`,
			selector:     "operators.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit"},
		},
		{
			name:         "places iso3166 state",
			query:        `query { operators(where:{adm1_iso: "US-CA"}) {onestop_id}}`,
			selector:     "operators.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			name:         "places iso3166 state not found",
			query:        `query { operators(where:{adm1_iso: "US-NY"}) {onestop_id}}`,
			selector:     "operators.#.onestop_id",
			selectExpect: []string{},
		},
		{
			name:         "places adm0_name",
			query:        `query { operators(where:{adm0_name: "United States of America"}) {onestop_id}}`,
			selector:     "operators.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit"},
		},
		{
			name:         "places adm1_name",
			query:        `query { operators(where:{adm1_name: "California"}) {onestop_id}}`,
			selector:     "operators.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			name:         "places adm1_name not found",
			query:        `query { operators(where:{adm1_name: "Nowhere"}) {onestop_id}}`,
			selector:     "operators.#.onestop_id",
			selectExpect: []string{},
		},
		{
			name:         "places city_name",
			query:        `query { operators(where:{city_name: "Oakland"}) {onestop_id}}`,
			selector:     "operators.#.onestop_id",
			selectExpect: []string{"o-9q9-bayarearapidtransit"},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
