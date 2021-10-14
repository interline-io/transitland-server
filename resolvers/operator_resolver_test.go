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
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
