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
	c, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			queryTestcase(t, c, tc)
		})
	}
}

func TestOperatorResolver_License(t *testing.T) {
	q := `
	query ($lic: LicenseFilter) {
		operators(limit: 10000, where: {license: $lic}) {
		  onestop_id
		}
	  }	  
	`
	selector := `operators.#.onestop_id`
	testcases := []testcase{
		// license: share_alike_optional
		{
			name:               "license filter: share_alike_optional = yes",
			query:              q,
			vars:               hw{"lic": hw{"share_alike_optional": "YES"}},
			selector:           selector,
			selectExpectUnique: []string{"o-dhv-hillsborougharearegionaltransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: share_alike_optional = no",
			query:              q,
			vars:               hw{"lic": hw{"share_alike_optional": "NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-bayarearapidtransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: share_alike_optional = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"share_alike_optional": "EXCLUDE_NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
			selectExpectCount:  3,
		},
		// license: create_derived_product
		{
			name:               "license filter: create_derived_product = yes",
			query:              q,
			vars:               hw{"lic": hw{"create_derived_product": "YES"}},
			selector:           selector,
			selectExpectUnique: []string{"o-dhv-hillsborougharearegionaltransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: create_derived_product = no",
			query:              q,
			vars:               hw{"lic": hw{"create_derived_product": "NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-bayarearapidtransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: create_derived_product = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"create_derived_product": "EXCLUDE_NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
			selectExpectCount:  3,
		},
		// license: commercial_use_allowed
		{
			name:               "license filter: commercial_use_allowed = yes",
			query:              q,
			vars:               hw{"lic": hw{"commercial_use_allowed": "YES"}},
			selector:           selector,
			selectExpectUnique: []string{"o-dhv-hillsborougharearegionaltransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: commercial_use_allowed = no",
			query:              q,
			vars:               hw{"lic": hw{"commercial_use_allowed": "NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-bayarearapidtransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: commercial_use_allowed = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"commercial_use_allowed": "EXCLUDE_NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
			selectExpectCount:  3,
		},
		// license: redistribution_allowed
		{
			name:               "license filter: redistribution_allowed = yes",
			query:              q,
			vars:               hw{"lic": hw{"redistribution_allowed": "YES"}},
			selector:           selector,
			selectExpectUnique: []string{"o-dhv-hillsborougharearegionaltransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: redistribution_allowed = no",
			query:              q,
			vars:               hw{"lic": hw{"redistribution_allowed": "NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-bayarearapidtransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: redistribution_allowed = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"redistribution_allowed": "EXCLUDE_NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
			selectExpectCount:  3,
		},
		// license: use_without_attribution
		{
			name:               "license filter: use_without_attribution = yes",
			query:              q,
			vars:               hw{"lic": hw{"use_without_attribution": "YES"}},
			selector:           selector,
			selectExpectUnique: []string{"o-dhv-hillsborougharearegionaltransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: use_without_attribution = no",
			query:              q,
			vars:               hw{"lic": hw{"use_without_attribution": "NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-bayarearapidtransit"},
			selectExpectCount:  1,
		},
		{
			name:               "license filter: use_without_attribution = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"use_without_attribution": "EXCLUDE_NO"}},
			selector:           selector,
			selectExpectUnique: []string{"o-9q9-caltrain", "o-dhv-hillsborougharearegionaltransit", "o-9qs-demotransitauthority"},
			selectExpectCount:  3,
		},
	}
	c, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			queryTestcase(t, c, tc)
		})
	}
}
