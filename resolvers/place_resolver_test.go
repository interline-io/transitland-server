package resolvers

import (
	"testing"
)

func TestPlaceResolver(t *testing.T) {
	q := `query($level: PlaceAggregationLevel) {
		places(level: $level) {
			adm0_name
			adm1_name
			city_name
			count
			operators {
				onestop_id
			}
		}
	}`
	testcases := []testcase{
		{
			name:         "ADM0",
			query:        q,
			vars:         hw{"level": "ADM0"},
			selector:     "places.#.adm0_name",
			selectExpect: []string{"United States of America"},
		},
		{
			name:         "ADM0 count",
			query:        q,
			vars:         hw{"level": "ADM0"},
			selector:     "places.#.count",
			selectExpect: []string{"3"},
		},
		{
			name:         "ADM0_ADM1",
			query:        q,
			vars:         hw{"level": "ADM0_ADM1"},
			selector:     "places.#.adm1_name",
			selectExpect: []string{"California", "Florida"},
		},
		{
			name:         "ADM0_ADM1 count",
			query:        q,
			vars:         hw{"level": "ADM0_ADM1"},
			selector:     "places.#.count",
			selectExpect: []string{"1", "2"},
		},
		{
			name:         "ADM0_ADM1_CITY",
			query:        q,
			vars:         hw{"level": "ADM0_ADM1_CITY"},
			selector:     "places.#.city_name",
			selectExpect: []string{"Berkeley", "Oakland", "San Francisco", "San Jose", "San Mateo", "Tampa", "", ""},
		},
	}
	c, _, _, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			queryTestcase(t, c, tc)
		})
	}
}
