package gql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestCensusResolver(t *testing.T) {
	c, cfg := newTestClient(t)
	geographyId := 0
	if err := cfg.Finder.DBX().QueryRowx(`select id from tl_census_geographies where geoid = '1400000US06001403000'`).Scan(&geographyId); err != nil {
		t.Errorf("could not get geography id for test: %s", err.Error())
	}

	vars := hw{}
	testcases := []testcase{
		// Datasets
		{
			name:   "dataset basic fields",
			query:  `query { census_datasets {name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"name":"acsdt5y2022"},{"name":"tiger2024"}]}`,
		},
		{
			name:   "dataset filter by name",
			query:  `query { census_datasets(where:{name:"acsdt5y2022"}) {name} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"name":"acsdt5y2022"}]}`,
		},
		{
			name:   "dataset filter by search",
			query:  `query { census_datasets(where:{search:"tiger"}) {name } }`,
			vars:   vars,
			expect: `{"census_datasets":[{"name":"tiger2024"}]}`,
		},
		// Dataset layers
		{
			name:   "dataset layers",
			query:  `query { census_datasets(where:{name:"tiger2024"}) {name layers { name description }} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"layers":[{"description":"Layer: uac20","name":"uac20"},{"description":"Layer: cbsa","name":"cbsa"},{"description":"Layer: csa","name":"csa"},{"description":"Layer: state","name":"state"},{"description":"Layer: county","name":"county"},{"description":"Layer: place","name":"place"},{"description":"Layer: tract","name":"tract"}],"name":"tiger2024"}]}`,
		},
		{
			name:   "dataset layer geographies",
			query:  `query { census_datasets(where:{name:"tiger2024"}) {name layers { name geographies(where:{search:"ala"}) { name } }} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"layers":[{"geographies":null,"name":"uac20"},{"geographies":null,"name":"cbsa"},{"geographies":null,"name":"csa"},{"geographies":null,"name":"state"},{"geographies":[{"name":"Alameda"}],"name":"county"},{"geographies":[{"name":"Acalanes Ridge"}],"name":"place"},{"geographies":null,"name":"tract"}],"name":"tiger2024"}]}`,
		},
		// Dataset Geographies
		{
			name:              "dataset geographies",
			query:             `query { census_datasets(where:{name:"tiger2024"}) {name geographies(limit:5) { geoid }} }`,
			vars:              vars,
			selector:          "census_datasets.0.geographies.#.geoid",
			selectExpectCount: 5,
		},
		{
			name:         "dataset geographies with layer",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer:"county"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.name",
			selectExpect: []string{"King", "Alameda"},
		},
		{
			name:   "dataset geographies with layer and adm names",
			query:  `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer:"county"}) { name geoid adm0_name adm1_name adm0_iso adm1_iso }} }`,
			vars:   vars,
			expect: `{"census_datasets":[{"name":"tiger2024","geographies":[{"adm0_iso":"US","adm0_name":"United States","adm1_iso":"US-WA","adm1_name":"Washington","geoid":"0500000US53033","name":"King"},{"adm0_iso":"US","adm0_name":"United States","adm1_iso":"US-CA","adm1_name":"California","geoid":"0500000US06001","name":"Alameda"}]}]}`,
		},
		{
			name:         "dataset geographies are multipolygon",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{search:"king"}) { name geoid geometry }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geometry.type",
			selectExpect: []string{"MultiPolygon"},
		},
		{
			name:         "dataset geographies with search",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{search:"king"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"0500000US53033"},
		},
		{
			name:         "dataset geographies with search and layer",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "tract", search:"288.02"}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US53033028802"},
		},
		{
			name:         "dataset geographies near point 1",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "tract", location:{near: {lon:-122.270, lat:37.805, radius:1000}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001403501", "1400000US06001403402", "1400000US06001403302", "1400000US06001402802", "1400000US06001403301", "1400000US06001403401", "1400000US06001402801", "1400000US06001401400", "1400000US06001403000", "1400000US06001402600", "1400000US06001403100", "1400000US06001401300", "1400000US06001402900", "1400000US06001401600", "1400000US06001402700", "1400000US06001983200", "1400000US06001403701"},
		},
		{
			name:         "dataset geographies near point 2",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "tract", location:{near: {lon:-122.270, lat:37.805, radius:100}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402801", "1400000US06001402900"},
		},
		{
			name:         "dataset geographies near point 3",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "tract", location:{near: {lon:-122.270, lat:37.805, radius:10}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402900"},
		},
		{
			name:         "dataset geographies near point 4",
			query:        `query { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "county", location:{near: {lon:-122.270, lat:37.805, radius:1000}}}) { name geoid }} }`,
			vars:         vars,
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"0500000US06001"},
		},
		{
			name:         "dataset geographies in bbox 1",
			query:        `query($bbox:BoundingBox) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "tract", location:{bbox:$bbox}}) { name geoid }} }`,
			vars:         hw{"bbox": hw{"min_lon": -122.27187746297761, "min_lat": 37.86760085920619, "max_lon": -122.26331772424285, "max_lat": 37.874244507564896}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001982100", "1400000US06001422902", "1400000US06001422901", "1400000US06001422400", "1400000US06001422800"},
		},
		{
			name:         "dataset geographies in bbox 2",
			query:        `query($bbox:BoundingBox) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "tract", location:{bbox:$bbox}}) { name geoid }} }`,
			vars:         hw{"bbox": hw{"min_lon": -122.2698781543005, "min_lat": 37.80700393130445, "max_lon": -122.2677640139239, "max_lat": 37.8088734037938}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001402801", "1400000US06001402900"},
		},
		{
			name:  "dataset geographies in polygon - tract",
			query: `query($feature:Polygon) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "tract", location:{within:$feature}}) { name geoid geometry_area intersection_area }} }`,
			vars: hw{"feature": hw{"type": "Polygon", "coordinates": [][][]float64{{
				{-122.27463277683867, 37.805635064682264},
				{-122.28006473340696, 37.80461858815316},
				{-122.27406099193678, 37.801456127261474},
				{-122.2754189810789, 37.79671218203016},
				{-122.27041586318674, 37.799648945955155},
				{-122.26398328303992, 37.79863238703946},
				{-122.26791430424078, 37.80247264731531},
				{-122.26441212171653, 37.80693387544258},
				{-122.269558185834, 37.806199767818995},
				{-122.27313184147101, 37.81066077079923},
				{-122.27463277683867, 37.805635064682264},
			}}}},
			f: func(t *testing.T, jj string) {
				// Sum areas to expected amount
				expectIntersectionArea := 829385.7985148486
				expectGeometryArea := 4755614.60179
				gotIntersectionArea := 0.0
				gotGeometryArea := 0.0
				a := gjson.Get(jj, "census_datasets.0.geographies").Array()
				for _, v := range a {
					gotIntersectionArea += v.Get("intersection_area").Float()
					gotGeometryArea += v.Get("geometry_area").Float()
				}
				assert.InDelta(t, expectIntersectionArea, gotIntersectionArea, 1.0)
				assert.InDelta(t, expectGeometryArea, gotGeometryArea, 1.0)
				assert.Equal(t, 11, len(a), "should have 11 geographies in polygon")
			},
		},
		{
			name:  "dataset geographies in polygon - county",
			query: `query($feature:Polygon) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "county", location:{within:$feature}}) { name geoid geometry_area intersection_area }} }`,
			vars: hw{"feature": hw{"type": "Polygon", "coordinates": [][][]float64{{
				{-122.27463277683867, 37.805635064682264},
				{-122.28006473340696, 37.80461858815316},
				{-122.27406099193678, 37.801456127261474},
				{-122.2754189810789, 37.79671218203016},
				{-122.27041586318674, 37.799648945955155},
				{-122.26398328303992, 37.79863238703946},
				{-122.26791430424078, 37.80247264731531},
				{-122.26441212171653, 37.80693387544258},
				{-122.269558185834, 37.806199767818995},
				{-122.27313184147101, 37.81066077079923},
				{-122.27463277683867, 37.805635064682264},
			}}}},
			f: func(t *testing.T, jj string) {
				// Sum areas to expected amount
				expectIntersectionArea := 829385.7985148486
				expectGeometryArea := 2126920288.43
				gotIntersectionArea := 0.0
				gotGeometryArea := 0.0
				a := gjson.Get(jj, "census_datasets.0.geographies").Array()
				for _, v := range a {
					gotIntersectionArea += v.Get("intersection_area").Float()
					gotGeometryArea += v.Get("geometry_area").Float()
				}
				assert.InDelta(t, expectIntersectionArea, gotIntersectionArea, 1.0)
				assert.InDelta(t, expectGeometryArea, gotGeometryArea, 1.0)
				assert.Equal(t, 1, len(a), "should have 1 geographies in polygon")
			},
		},
		{
			name:  "dataset geographies in big polygon - county",
			query: `query($feature:Polygon) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer: "county", location:{within:$feature}}) { name geoid geometry_area intersection_area }} }`,
			vars: hw{"feature": hw{"type": "Polygon", "coordinates": [][][]float64{{
				{-123.77489413290716, 38.794161309061735},
				{-122.69431950796763, 35.52679604934255},
				{-119.9104881819854, 37.991860068760204},
				{-123.77489413290716, 38.794161309061735},
			}}}},
			f: func(t *testing.T, jj string) {
				// Should be equal to the area of Alameda County
				expectGeometryArea := 2126920288.43
				expectIntersectionArea := 2126920288.43
				intersectionArea := 0.0
				geometryArea := 0.0
				a := gjson.Get(jj, "census_datasets.0.geographies").Array()
				for _, v := range a {
					intersectionArea += v.Get("intersection_area").Float()
					geometryArea += v.Get("geometry_area").Float()
				}
				assert.InDelta(t, expectIntersectionArea, intersectionArea, 1.0)
				assert.InDelta(t, expectGeometryArea, geometryArea, 1.0)
				assert.Equal(t, 1, len(a), "should have 1 geographies in polygon")
			},
		},
		{
			name:         "dataset geographies by id",
			query:        `query($ids:[Int!]) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{ids:$ids}) { name geoid }} }`,
			vars:         hw{"ids": []int{geographyId}},
			selector:     "census_datasets.0.geographies.#.geoid",
			selectExpect: []string{"1400000US06001403000"},
		},
		{
			name:   "dataset geographies with focus",
			query:  `query($focus: FocusPoint) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer:"county", location:{focus:$focus}}) { name geoid }} }`,
			vars:   hw{"focus": hw{"lon": -122.270, "lat": 37.805}},
			expect: `{"census_datasets":[{"name":"tiger2024","geographies":[{"geoid":"0500000US06001","name":"Alameda"},{"geoid":"0500000US53033","name":"King"}]}]}`,
		},
		{
			name:   "dataset geographies with focus 2",
			query:  `query($focus: FocusPoint) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{layer:"county", location:{focus:$focus}}) { name geoid }} }`,
			vars:   hw{"focus": hw{"lon": -122.180, "lat": 48.390}},
			expect: `{"census_datasets":[{"name":"tiger2024","geographies":[{"geoid":"0500000US53033","name":"King"},{"geoid":"0500000US06001","name":"Alameda"}]}]}`,
		},
		{
			name:   "dataset geographies layer",
			query:  `query($ids:[Int!]) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{ids:$ids}) { name geoid layer { name }}} }`,
			vars:   hw{"ids": []int{geographyId}},
			expect: `{"census_datasets":[{"geographies":[{"geoid":"1400000US06001403000","layer":{"name":"tract"},"name":"4030"}],"name":"tiger2024"}]}`,
		},
		{
			name:   "dataset geographies source",
			query:  `query($ids:[Int!]) { census_datasets(where:{name:"tiger2024"}) {name geographies(where:{ids:$ids}) { name geoid source { name }}} }`,
			vars:   hw{"ids": []int{geographyId}},
			expect: `{"census_datasets":[{"geographies":[{"geoid":"1400000US06001403000","name":"4030","source":{"name":"tl_2024_06_tract.zip"}}],"name":"tiger2024"}]}`,
		},
		// Sources
		{
			name:   "sources",
			query:  `query { census_datasets(where:{name:"acsdt5y2022"}) {name sources { name }} }`,
			vars:   vars,
			expect: ` {"census_datasets":[{"name":"acsdt5y2022","sources":[{"name":"acsdt5y2022-b01001.dat"},{"name":"acsdt5y2022-b01001a.dat"},{"name":"acsdt5y2022-b01001b.dat"},{"name":"acsdt5y2022-b01001c.dat"},{"name":"acsdt5y2022-b01001d.dat"},{"name":"acsdt5y2022-b01001e.dat"},{"name":"acsdt5y2022-b01001f.dat"},{"name":"acsdt5y2022-b01001g.dat"},{"name":"acsdt5y2022-b01001h.dat"},{"name":"acsdt5y2022-b01001i.dat"}]}]}`,
		},
		// Source layers
		{
			name:   "source layers",
			query:  `query { census_datasets(where:{name:"tiger2024"}) {name sources(where:{name:"tl_2024_06_tract.zip"}) {name layers { name description }} } }`,
			vars:   vars,
			expect: `{"census_datasets":[{"name":"tiger2024","sources":[{"layers":[{"description":"Layer: tract","name":"tract"}],"name":"tl_2024_06_tract.zip"}]}]}`,
		},
		{
			name:   "source geographies",
			query:  `query { census_datasets(where:{name:"tiger2024"}) {name sources(where:{name:"tl_2024_us_county.zip"}) {name geographies(where:{search:"ala"}) { name } } } }`,
			vars:   vars,
			expect: `{"census_datasets":[{"name":"tiger2024","sources":[{"geographies":[{"name":"Alameda"}],"name":"tl_2024_us_county.zip"}]}]}`,
		},
		// Agency geographies
		{
			name:   "agency geographies",
			query:  `query { agencies { census_geographies { name } } }`,
			vars:   vars,
			expect: `{"census_datasets":[{"name":"tiger2024","sources":[{"geographies":[{"name":"Alameda"}],"name":"tl_2024_us_county.zip"}]}]}`,
		},
	}
	queryTestcases(t, c, testcases)
}
