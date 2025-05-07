package gql

import "testing"

func TestCensusResolver(t *testing.T) {
	vars := hw{}
	testcases := []testcase{
		{
			name:   "basic fields",
			query:  `query { census_datasets {id dataset_name url} }`,
			vars:   vars,
			expect: `{"feeds":[{"file":"server-test.dmfr.json","languages":["en-US"],"name":"Caltrain","onestop_id":"CT","spec":"GTFS"}]}`,
		},
	}
	c, _ := newTestClient(t)
	queryTestcases(t, c, testcases)
}
