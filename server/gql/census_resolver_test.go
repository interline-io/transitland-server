package gql

import "testing"

func TestCensusResolver(t *testing.T) {
	vars := hw{}
	testcases := []testcase{
		// Datasets
		{
			name:   "dataset basic fields",
			query:  `query { census_datasets {dataset_name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"acsdt5y2022"},{"dataset_name":"tiger2024"}]}`,
		},
		{
			name:   "dataset filter by dataset_name",
			query:  `query { census_datasets(where:{dataset_name:"acsdt5y2022"}) {dataset_name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"acsdt5y2022"}]}`,
		},
		{
			name:   "dataset filter by search",
			query:  `query { census_datasets(where:{search:"tiger"}) {dataset_name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"tiger2024"}]}`,
		},
		// Dataset layers
		{
			name:   "dataset layers",
			query:  `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name layers} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"tiger2024","layers":["county","tract"]}]}`,
		},
		// Dataset Geographies
		{
			name:              "dataset geographies",
			query:             `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5) { geoid }} }`,
			vars:              vars,
			selector:          "census_datasets.0.geographies.#.geoid",
			selectExpectCount: 5,
		},
		{
			name:         "dataset geographies with layer",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer:"county"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.name",
			selectExpect: []string{"King", "Alameda"},
		},
		{
			name:         "dataset geographies with search",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{search:"king"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"0500000US53033"},
		},
		{
			name:         "dataset geographies with search and layer",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "tract", search:"288.02"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US53033028802"},
		},
		// Sources
		{
			name:   "sources",
			query:  `query { census_datasets(where:{dataset_name:"acsdt5y2022"}) {dataset_name sources { source_name }} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"acsdt5y2022","sources":[{"source_name":"acsdt5y2022-b01001.dat"}]}]}`,
		},
		// Source layers
		{
			name:   "source layers",
			query:  `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name sources(where:{source_name:"tl_2024_06_tract.zip"}) { source_name layers} } }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"tiger2024","sources":[{"layers":["tract"],"source_name":"tl_2024_06_tract.zip"}]}]}`,
		},
	}
	c, _ := newTestClient(t)
	queryTestcases(t, c, testcases)
}
