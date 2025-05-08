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
		// {
		// 	name:   "source layers",
		// 	query:  `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name sources { source_name layers }} }`,
		// 	vars:   vars,
		// 	expect: `{"census_datasets":[{"dataset_name":"tiger2024","sources":[{"source_name":"acsdt5y2022-b01001.dat", "layers": ["a", "b"]}]}]}`,
		// },
		// Geographies
		{
			name:              "geographies",
			query:             `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5) { geoid }} }`,
			vars:              vars,
			selector:          "census_datasets.0.geographies.#.geoid",
			selectExpectCount: 5,
		},
		{
			name:         "geographies with layer",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer:"county"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.name",
			selectExpect: []string{"King", "Alameda"},
		},
		{
			name:         "geographies with search",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{search:"king"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"0500000US53033"},
		},
		{
			name:         "geographies with search and layer",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "tract", search:"288.02"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US53033028802"},
		},
	}
	c, _ := newTestClient(t)
	queryTestcases(t, c, testcases)
}
