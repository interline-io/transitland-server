package resolvers

import (
	"testing"
)

func TestOperatorResolver(t *testing.T) {
	testcases := []testcase{
		{
			"basic fields",
			`query{operators(where:{onestop_id:"o-9q9-bayarearapidtransit"}) {onestop_id}}`,
			hw{},
			`{"operators":[{"onestop_id":"o-9q9-bayarearapidtransit"}]}`,
			"",
			nil,
		},
		{
			"feeds",
			`query{operators(where:{onestop_id:"o-9q9-bayarearapidtransit"}) {feeds{onestop_id}}}`,
			hw{},
			``,
			"operators.0.feeds.#.onestop_id",
			[]string{"BA"},
		},
		{
			"feeds incl rt",
			`query{operators(where:{onestop_id:"o-9q9-caltrain"}) {feeds{onestop_id}}}`,
			hw{},
			``,
			"operators.0.feeds.#.onestop_id",
			[]string{"CT", "CT~rt"},
		},
		{
			"feeds only gtfs-rt",
			`query{operators(where:{onestop_id:"o-9q9-caltrain"}) {feeds(where:{spec:GTFS_RT}) {onestop_id}}}`,
			hw{},
			``,
			"operators.0.feeds.#.onestop_id",
			[]string{"CT~rt"},
		},
		{
			"feeds only gtfs",
			`query{operators(where:{onestop_id:"o-9q9-caltrain"}) {feeds(where:{spec:GTFS}) {onestop_id}}}`,
			hw{},
			``,
			"operators.0.feeds.#.onestop_id",
			[]string{"CT"},
		},

		{
			"tags us_ntd_id=90134",
			`query{operators(where:{tags:{us_ntd_id:"90134"}}) {onestop_id}}`,
			hw{},
			`{"operators":[{"onestop_id":"o-9q9-caltrain"}]}`,
			"",
			nil,
		},
		{
			"tags us_ntd_id=12345",
			`query{operators(where:{tags:{us_ntd_id:"12345"}}) {onestop_id}}`,
			hw{},
			`{"operators":[]}`,
			"",
			nil,
		},
		{
			"tags us_ntd_id presence",
			`query{operators(where:{tags:{us_ntd_id:""}}) {onestop_id}}`,
			hw{},
			`{"operators":[{"onestop_id":"o-9q9-caltrain"}]}`,
			"",
			nil,
		},
		{
			"places iso3166 country",
			`query { operators(where:{adm0_iso: "US"}) {onestop_id}}`,
			hw{},
			``,
			"operators.#.onestop_id",
			[]string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit"},
		},
		{
			"places iso3166 state",
			`query { operators(where:{adm1_iso: "US-CA"}) {onestop_id}}`,
			hw{},
			``,
			"operators.#.onestop_id",
			[]string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			"places iso3166 state not found",
			`query { operators(where:{adm1_iso: "US-NY"}) {onestop_id}}`,
			hw{},
			``,
			"operators.#.onestop_id",
			[]string{},
		},
		{
			"places adm0_name",
			`query { operators(where:{adm0_name: "United States of America"}) {onestop_id}}`,
			hw{},
			``,
			"operators.#.onestop_id",
			[]string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit"},
		},
		{
			"places adm1_name",
			`query { operators(where:{adm1_name: "California"}) {onestop_id}}`,
			hw{},
			``,
			"operators.#.onestop_id",
			[]string{"o-9q9-bayarearapidtransit", "o-9q9-caltrain"},
		},
		{
			"places adm1_name not found",
			`query { operators(where:{adm1_name: "Nowhere"}) {onestop_id}}`,
			hw{},
			``,
			"operators.#.onestop_id",
			[]string{},
		},
		{
			"places city_name",
			`query { operators(where:{city_name: "Oakland"}) {onestop_id}}`,
			hw{},
			``,
			"operators.#.onestop_id",
			[]string{"o-9q9-bayarearapidtransit"},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
