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
		{
			name:         "dataset geographies near point 1",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "tract", near: {lon:-122.270, lat:37.805, radius:1000}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001983200", "1400000US06001403402", "1400000US06001403302", "1400000US06001402802", "1400000US06001403301"},
		},
		{
			name:         "dataset geographies near point 2",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "tract", near: {lon:-122.270, lat:37.805, radius:100}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402801", "1400000US06001402900"},
		},
		{
			name:         "dataset geographies near point 3",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "tract", near: {lon:-122.270, lat:37.805, radius:10}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402900"},
		},
		{
			name:         "dataset geographies near point 4",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "county", near: {lon:-122.270, lat:37.805, radius:1000}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"0500000US06001"},
		},
		{
			name:         "dataset geographies in bbox 1",
			query:        `query($bbox:BoundingBox) { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "tract", bbox:$bbox}) { name geoid }} }`,
			vars:         hw{"bbox": hw{"min_lon": -122.27187746297761, "min_lat": 37.86760085920619, "max_lon": -122.26331772424285, "max_lat": 37.874244507564896}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001982100", "1400000US06001422902", "1400000US06001422901", "1400000US06001422400", "1400000US06001422800"},
		},
		{
			name:         "dataset geographies in bbox 2",
			query:        `query($bbox:BoundingBox) { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(limit:5, where:{layer: "tract", bbox:$bbox}) { name geoid }} }`,
			vars:         hw{"bbox": hw{"min_lon": -122.2698781543005, "min_lat": 37.80700393130445, "max_lon": -122.2677640139239, "max_lat": 37.8088734037938}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402801", "1400000US06001402900"},
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
