package gql

import "testing"

func TestCensusResolver(t *testing.T) {
	c, cfg := newTestClient(t)
	geographyId := 0
	if err := cfg.Finder.DBX().QueryRowx(`select id from tl_census_geographies where layer_name = 'tract' and geoid = '1400000US06001403000'`).Scan(&geographyId); err != nil {
		t.Errorf("could not get geography id for test: %s", err.Error())
	}

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
			expect: `{"census_datasets":[{"dataset_name":"tiger2024","layers":["county","place","tract"]}]}`,
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
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer:"county"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.name",
			selectExpect: []string{"King", "Alameda"},
		},
		{
			name:   "dataset geographies with layer and adm names",
			query:  `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer:"county"}) { name geoid adm0_name adm1_name adm0_iso adm1_iso }} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"tiger2024","geographies":[{"adm0_iso":"US","adm0_name":"United States","adm1_iso":"US-WA","adm1_name":"Washington","geoid":"0500000US53033","name":"King"},{"adm0_iso":"US","adm0_name":"United States","adm1_iso":"US-CA","adm1_name":"California","geoid":"0500000US06001","name":"Alameda"}]}]}`,
		},
		{
			name:         "dataset geographies are multipolygon",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{search:"king"}) { name geoid geometry }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geometry.type",
			selectExpect: []string{"MultiPolygon"},
		},
		{
			name:         "dataset geographies with search",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{search:"king"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"0500000US53033"},
		},
		{
			name:         "dataset geographies with search and layer",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer: "tract", search:"288.02"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US53033028802"},
		},
		{
			name:         "dataset geographies near point 1",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer: "tract", location:{near: {lon:-122.270, lat:37.805, radius:1000}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001403501", "1400000US06001403402", "1400000US06001403302", "1400000US06001402802", "1400000US06001403301", "1400000US06001403401", "1400000US06001402801", "1400000US06001401400", "1400000US06001403000", "1400000US06001402600", "1400000US06001403100", "1400000US06001401300", "1400000US06001402900", "1400000US06001401600", "1400000US06001402700", "1400000US06001983200", "1400000US06001403701"},
		},
		{
			name:         "dataset geographies near point 2",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer: "tract", location:{near: {lon:-122.270, lat:37.805, radius:100}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402801", "1400000US06001402900"},
		},
		{
			name:         "dataset geographies near point 3",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer: "tract", location:{near: {lon:-122.270, lat:37.805, radius:10}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402900"},
		},
		{
			name:         "dataset geographies near point 4",
			query:        `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer: "county", location:{near: {lon:-122.270, lat:37.805, radius:1000}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"0500000US06001"},
		},
		{
			name:         "dataset geographies in bbox 1",
			query:        `query($bbox:BoundingBox) { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer: "tract", location:{bbox:$bbox}}) { name geoid }} }`,
			vars:         hw{"bbox": hw{"min_lon": -122.27187746297761, "min_lat": 37.86760085920619, "max_lon": -122.26331772424285, "max_lat": 37.874244507564896}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001982100", "1400000US06001422902", "1400000US06001422901", "1400000US06001422400", "1400000US06001422800"},
		},
		{
			name:         "dataset geographies in bbox 2",
			query:        `query($bbox:BoundingBox) { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer: "tract", location:{bbox:$bbox}}) { name geoid }} }`,
			vars:         hw{"bbox": hw{"min_lon": -122.2698781543005, "min_lat": 37.80700393130445, "max_lon": -122.2677640139239, "max_lat": 37.8088734037938}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402801", "1400000US06001402900"},
		},
		{
			name:         "dataset geographies by id",
			query:        `query($ids:[Int!]) { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{ids:$ids}) { name geoid }} }`,
			vars:         hw{"ids": []int{geographyId}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001403000"},
		},
		{
			name:   "dataset geographies with focus",
			query:  `query($focus: FocusPoint) { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer:"county", location:{focus:$focus}}) { name geoid }} }`,
			vars:   hw{"focus": hw{"lon": -122.270, "lat": 37.805}},
			expect: `{"census_datasets":[{"dataset_name":"tiger2024","geographies":[{"geoid":"0500000US06001","name":"Alameda"},{"geoid":"0500000US53033","name":"King"}]}]}`,
		},
		{
			name:   "dataset geographies with focus 2",
			query:  `query($focus: FocusPoint) { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name geographies(where:{layer:"county", location:{focus:$focus}}) { name geoid }} }`,
			vars:   hw{"focus": hw{"lon": -122.180, "lat": 48.390}},
			expect: `{"census_datasets":[{"dataset_name":"tiger2024","geographies":[{"geoid":"0500000US53033","name":"King"},{"geoid":"0500000US06001","name":"Alameda"}]}]}`,
		},

		// Sources
		{
			name:   "sources",
			query:  `query { census_datasets(where:{dataset_name:"acsdt5y2022"}) {dataset_name sources { source_name }} }`,
			vars:   vars,
			expect: ` {"census_datasets":[{"dataset_name":"acsdt5y2022","sources":[{"source_name":"acsdt5y2022-b01001.dat"},{"source_name":"acsdt5y2022-b01001a.dat"},{"source_name":"acsdt5y2022-b01001b.dat"},{"source_name":"acsdt5y2022-b01001c.dat"},{"source_name":"acsdt5y2022-b01001d.dat"},{"source_name":"acsdt5y2022-b01001e.dat"},{"source_name":"acsdt5y2022-b01001f.dat"},{"source_name":"acsdt5y2022-b01001g.dat"},{"source_name":"acsdt5y2022-b01001h.dat"},{"source_name":"acsdt5y2022-b01001i.dat"}]}]}`,
		},
		// Source layers
		{
			name:   "source layers",
			query:  `query { census_datasets(where:{dataset_name:"tiger2024"}) {dataset_name sources(where:{source_name:"tl_2024_06_tract.zip"}) { source_name layers} } }`,
			vars:   vars,
			expect: `{"census_datasets":[{"dataset_name":"tiger2024","sources":[{"layers":["tract"],"source_name":"tl_2024_06_tract.zip"}]}]}`,
		},
	}
	queryTestcases(t, c, testcases)
}
