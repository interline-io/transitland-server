package gql

import "testing"

func TestCensusResolver(t *testing.T) {
	vars := hw{}
	testcases := []testcase{
		{
			name:   "basic fields",
			query:  `query { census_datasets {dataset_name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"acsdt5y2022"},{"dataset_name":"tiger2024"}]}`,
		},
		{
			name:   "filter by dataset_name",
			query:  `query { census_datasets(where:{dataset_name:"acsdt5y2022"}) {dataset_name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"acsdt5y2022"}]}`,
		},
		{
			name:   "filter by search",
			query:  `query { census_datasets(where:{search:"tiger"}) {dataset_name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"tiger2024"}]}`,
		},
		// Sources
		{
			name:   "sources",
			query:  `query { census_datasets(where:{dataset_name:"acsdt5y2022"}) {dataset_name sources { source_name }} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"acsdt5y2022","sources":[{"source_name":"acsdt5y2022-b01001.dat"}]}]}`,
		},
	}
	c, _ := newTestClient(t)
	queryTestcases(t, c, testcases)
}
