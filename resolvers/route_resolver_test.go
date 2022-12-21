package resolvers

import (
	"context"
	"testing"
)

func TestRouteResolver(t *testing.T) {
	vars := hw{"route_id": "03"}
	testcases := []testcase{
		{
			name:         "basic",
			query:        `query {  routes { route_id } }`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"1", "12", "14", "15", "16", "17", "19", "20", "24", "25", "275", "30", "31", "32", "33", "34", "35", "36", "360", "37", "38", "39", "400", "42", "45", "46", "48", "5", "51", "6", "60", "7", "75", "8", "9", "96", "97", "570", "571", "572", "573", "574", "800", "PWT", "SKY", "01", "03", "05", "07", "11", "19", "Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"},
		},
		{
			name:   "basic fields",
			query:  `query($route_id: String!) {  routes(where:{route_id:$route_id}) {onestop_id route_id route_short_name route_long_name route_type route_color route_text_color route_sort_order route_url route_desc feed_version_sha1 feed_onestop_id} }`,
			vars:   vars,
			expect: `{"routes":[{"feed_onestop_id":"BA","feed_version_sha1":"e535eb2b3b9ac3ef15d82c56575e914575e732e0","onestop_id":"r-9q9n-warmsprings~southfremont~richmond","route_color":"ff9933","route_desc":"","route_id":"03","route_long_name":"Warm Springs/South Fremont - Richmond","route_short_name":"","route_sort_order":0,"route_text_color":"","route_type":1,"route_url":"http://www.bart.gov/schedules/bylineresults?route=3"}]}`,
		},
		{
			name:         "geometry",
			query:        `query($route_id: String!) {  routes(where:{route_id:$route_id}) {geometry} }`,
			vars:         vars,
			selector:     "routes.0.geometry.type",
			selectExpect: []string{"MultiLineString"},
		},
		{
			name:   "feed_version",
			query:  `query($route_id: String!) {  routes(where:{route_id:$route_id}) {feed_version{sha1}} }`,
			vars:   vars,
			expect: `{"routes":[{"feed_version":{"sha1":"e535eb2b3b9ac3ef15d82c56575e914575e732e0"}}]}`,
		},
		{
			name:         "trips",
			query:        `query($route_id: String!) {  routes(where:{route_id:$route_id}) {trips{trip_id trip_headsign}} }`,
			vars:         hw{"route_id": "Bu-130"}, // use baby bullet
			selector:     "routes.0.trips.#.trip_id",
			selectExpect: []string{"305", "309", "313", "319", "323", "329", "365", "371", "375", "381", "385", "310", "314", "320", "324", "330", "360", "366", "370", "376", "380", "386", "801", "803", "802", "804"},
		},
		{
			name:         "route_stops",
			query:        `query($route_id: String!) {  routes(where:{route_id:$route_id}) {route_stops{stop{stop_id stop_name}}} }`,
			vars:         vars,
			selector:     "routes.0.route_stops.#.stop.stop_id",
			selectExpect: []string{"12TH", "19TH", "19TH_N", "ASHB", "BAYF", "COLS", "DBRK", "DELN", "PLZA", "FRMT", "FTVL", "HAYW", "LAKE", "MCAR", "MCAR_S", "NBRK", "RICH", "SANL", "SHAY", "UCTY", "WARM"},
		},
		{
			// computations are not stable so just check success
			name:         "geometries",
			query:        `query($route_id: String!) {  routes(where:{route_id:$route_id}) {geometries {generated}} }`,
			vars:         vars,
			selector:     "routes.0.geometries.#.generated",
			selectExpect: []string{"false"},
		},
		{
			name: "route_stop_buffer stop_points 10m",
			query: `query($route_id: String!) { routes(where:{route_id:$route_id}) {route_stop_buffer(radius: 100.0) {stop_points	stop_buffer	stop_convexhull}}}`,
			vars:         vars,
			selector:     "routes.0.route_stop_buffer.stop_points.type",
			selectExpect: []string{"MultiPoint"},
		},
		{
			name: "route_stop_buffer stop_buffer 10m",
			query: `query($route_id: String!) { routes(where:{route_id:$route_id}) {route_stop_buffer(radius: 100.0) {stop_points	stop_buffer	stop_convexhull}}}`,
			vars:         vars,
			selector:     "routes.0.route_stop_buffer.stop_buffer.type",
			selectExpect: []string{"MultiPolygon"},
		},
		{
			name: "route_stop_buffer stop_convexhull 10m",
			query: `query($route_id: String!) { routes(where:{route_id:$route_id}) {route_stop_buffer(radius: 100.0) {stop_points	stop_buffer	stop_convexhull}}}`,
			vars:         vars,
			selector:     "routes.0.route_stop_buffer.stop_convexhull.type",
			selectExpect: []string{"Polygon"},
		},
		{
			// only check dow_category explicitly it's not a stable computation
			name:         "headways",
			query:        `query($route_id: String!) {  routes(where:{route_id:$route_id}) {headways{dow_category departures service_date stop_trip_count stop{stop_id}}} }`,
			vars:         vars,
			selector:     "routes.0.headways.#.dow_category",
			selectExpect: []string{"1", "6", "7", "1", "6", "7"}, // now includes one for each direction and dow category
		},
		{
			name:         "where onestop_id",
			query:        `query {routes(where:{onestop_id:"r-9q9j-bullet"}) {route_id} }`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"Bu-130"},
		},
		{
			name:  "where feed_version_sha1",
			query: `query {routes(where:{feed_version_sha1:"d2813c293bcfd7a97dde599527ae6c62c98e66c6"}) {route_id} }`,

			selector:     "routes.#.route_id",
			selectExpect: []string{"Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"},
		},
		{
			name:         "where feed_onestop_id",
			query:        `query {routes(where:{feed_onestop_id:"CT"}) {route_id} }`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"Bu-130", "Li-130", "Lo-130", "TaSj-130", "Gi-130", "Sp-130"},
		},
		{
			name:         "where route_id",
			query:        `query {routes(where:{route_id:"Lo-130"}) {route_id} }`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"Lo-130"},
		},
		{
			name:         "where route_type=2",
			query:        `query {routes(where:{route_type:2}) {route_id} }`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"Bu-130", "Li-130", "Lo-130", "Gi-130", "Sp-130"},
		},
		{
			name:         "where search",
			query:        `query {routes(where:{search:"warm"}) {route_id} }`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"03", "05"},
		},
		{
			name:         "where search 2",
			query:        `query {routes(where:{search:"bullet"}) {route_id} }`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"Bu-130"},
		},
		// just ensure geometry queries complete successfully; checking coordinates is a pain and flaky.
		{
			name:         "where near 100m",
			query:        `query {routes(where:{near:{lon:-122.407974,lat:37.784471,radius:100.0}}) {route_id route_long_name}}`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"01", "05", "07", "11"},
		},
		{
			name:         "where near 10000m",
			query:        `query {routes(where:{near:{lon:-122.407974,lat:37.784471,radius:10000.0}}) {route_id route_long_name}}`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"Bu-130", "Li-130", "Lo-130", "Gi-130", "Sp-130", "01", "05", "07", "11"},
		},
		{
			name:         "where within polygon",
			query:        `query{routes(where:{within:{type:"Polygon",coordinates:[[[-122.396,37.8],[-122.408,37.79],[-122.393,37.778],[-122.38,37.787],[-122.396,37.8]]]}}){id route_id}}`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"01", "05", "07", "11"},
		},
		{
			name:         "where within polygon big",
			query:        `query{routes(where:{within:{type:"Polygon",coordinates:[[[-122.39481925964355,37.80151060070086],[-122.41653442382812,37.78652126637423],[-122.39662170410156,37.76847577247014],[-122.37301826477051,37.784757615348575],[-122.39481925964355,37.80151060070086]]]}}){id route_id}}`,
			selector:     "routes.#.route_id",
			selectExpect: []string{"Bu-130", "Li-130", "Lo-130", "Gi-130", "Sp-130", "01", "05", "07", "11"},
		},
		// route patterns
		{
			name: "route patterns",
			query: `{
				routes(where: {feed_onestop_id: "BA", route_id: "03"}) {
				  route_id
				  patterns {
					count
					direction_id
					stop_pattern_id
					trips(limit: 1) {
					  trip_id
					}
				  }
				}
			  }`,

			selector:     "routes.0.patterns.#.count",
			selectExpect: []string{"132", "124", "56", "50", "2"},
		},
		// TODO: census_geographies
	}
	c, _, _, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestRouteResolver_PreviousOnestopID(t *testing.T) {
	testcases := []testcase{
		{
			name:         "default",
			query:        `query($osid:String!, $previous:Boolean!) { routes(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous}) { route_id onestop_id }}`,
			vars:         hw{"osid": "r-9q9-antioch~sfia~millbrae", "previous": false},
			selector:     "routes.#.onestop_id",
			selectExpect: []string{"r-9q9-antioch~sfia~millbrae"},
		},
		{
			name:         "old id no result",
			query:        `query($osid:String!, $previous:Boolean!) { routes(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous}) { route_id onestop_id }}`,
			vars:         hw{"osid": "r-9q9-pittsburg~baypoint~sfia~millbrae", "previous": false},
			selector:     "routes.#.onestop_id",
			selectExpect: []string{},
		},
		{
			name:         "old id specify fv",
			query:        `query($osid:String!, $previous:Boolean!) { routes(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous, feed_version_sha1:"dd7aca4a8e4c90908fd3603c097fabee75fea907"}) { route_id onestop_id }}`,
			vars:         hw{"osid": "r-9q9-pittsburg~baypoint~sfia~millbrae", "previous": false},
			selector:     "routes.#.onestop_id",
			selectExpect: []string{"r-9q9-pittsburg~baypoint~sfia~millbrae"},
		},
		{
			name:         "use previous",
			query:        `query($osid:String!, $previous:Boolean!) { routes(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous}) { route_id onestop_id }}`,
			vars:         hw{"osid": "r-9q9-pittsburg~baypoint~sfia~millbrae", "previous": true},
			selector:     "routes.#.onestop_id",
			selectExpect: []string{"r-9q9-pittsburg~baypoint~sfia~millbrae"},
		},
	}
	c, _, _, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestRouteResolver_Cursor(t *testing.T) {
	c, dbf, _, _ := newTestClient(t)
	allEnts, err := dbf.FindRoutes(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	allIds := []string{}
	for _, ent := range allEnts {
		allIds = append(allIds, ent.RouteID)
	}
	testcases := []testcase{
		{
			name:         "no cursor",
			query:        "query{routes(limit:10){feed_version{id} id route_id}}",
			selector:     "routes.#.route_id",
			selectExpect: allIds[:10],
		},
		{
			name:         "after 0",
			query:        "query{routes(after: 0, limit:10){feed_version{id} id route_id}}",
			selector:     "routes.#.route_id",
			selectExpect: allIds[:10],
		},
		{
			name:         "after 10th",
			query:        "query($after: Int!){routes(after: $after, limit:10){feed_version{id} id route_id}}",
			vars:         hw{"after": allEnts[10].ID},
			selector:     "routes.#.route_id",
			selectExpect: allIds[11:21],
		},
		{
			name:         "after last",
			query:        "query($after: Int!){routes(after: $after, limit:10){feed_version{id} id route_id}}",
			vars:         hw{"after": allEnts[len(allEnts)-1].ID},
			selector:     "routes.#.route_id",
			selectExpect: []string{},
		},
		{
			name:         "after invalid id returns no results",
			query:        "query($after: Int!){routes(after: $after, limit:10){feed_version{id} id route_id}}",
			vars:         hw{"after": 10_000_000},
			selector:     "routes.#.route_id",
			selectExpect: []string{},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestRouteResolver_License(t *testing.T) {
	q := `
	query ($lic: LicenseFilter) {
		routes(limit: 10000, where: {license: $lic}) {
		  route_id
		  feed_version {
			feed {
			  onestop_id
			  license {
				share_alike_optional
				create_derived_product
				commercial_use_allowed
				redistribution_allowed
			  }
			}
		  }
		}
	  }	  
	`
	testcases := []testcase{
		// license: share_alike_optional
		{
			name:               "license filter: share_alike_optional = yes",
			query:              q,
			vars:               hw{"lic": hw{"share_alike_optional": "YES"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"HA"},
			selectExpectCount:  45,
		},
		{
			name:               "license filter: share_alike_optional = no",
			query:              q,
			vars:               hw{"lic": hw{"share_alike_optional": "NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"BA"},
			selectExpectCount:  6,
		},
		{
			name:               "license filter: share_alike_optional = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"share_alike_optional": "EXCLUDE_NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"CT", "HA"},
			selectExpectCount:  51,
		},
		// license: create_derived_product
		{
			name:               "license filter: create_derived_product = yes",
			query:              q,
			vars:               hw{"lic": hw{"create_derived_product": "YES"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"HA"},
			selectExpectCount:  45,
		},
		{
			name:               "license filter: create_derived_product = no",
			query:              q,
			vars:               hw{"lic": hw{"create_derived_product": "NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"BA"},
			selectExpectCount:  6,
		},
		{
			name:               "license filter: create_derived_product = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"create_derived_product": "EXCLUDE_NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"CT", "HA"},
			selectExpectCount:  51,
		},
		// license: commercial_use_allowed
		{
			name:               "license filter: commercial_use_allowed = yes",
			query:              q,
			vars:               hw{"lic": hw{"commercial_use_allowed": "YES"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"HA"},
			selectExpectCount:  45,
		},
		{
			name:               "license filter: commercial_use_allowed = no",
			query:              q,
			vars:               hw{"lic": hw{"commercial_use_allowed": "NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"BA"},
			selectExpectCount:  6,
		},
		{
			name:               "license filter: commercial_use_allowed = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"commercial_use_allowed": "EXCLUDE_NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"CT", "HA"},
			selectExpectCount:  51,
		},
		// license: redistribution_allowed
		{
			name:               "license filter: redistribution_allowed = yes",
			query:              q,
			vars:               hw{"lic": hw{"redistribution_allowed": "YES"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"HA"},
			selectExpectCount:  45,
		},
		{
			name:               "license filter: redistribution_allowed = no",
			query:              q,
			vars:               hw{"lic": hw{"redistribution_allowed": "NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"BA"},
			selectExpectCount:  6,
		},
		{
			name:               "license filter: redistribution_allowed = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"redistribution_allowed": "EXCLUDE_NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"CT", "HA"},
			selectExpectCount:  51,
		},
		// license: use_without_attribution
		{
			name:               "license filter: use_without_attribution = yes",
			query:              q,
			vars:               hw{"lic": hw{"use_without_attribution": "YES"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"HA"},
			selectExpectCount:  45,
		},
		{
			name:               "license filter: use_without_attribution = no",
			query:              q,
			vars:               hw{"lic": hw{"use_without_attribution": "NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"BA"},
			selectExpectCount:  6,
		},
		{
			name:               "license filter: use_without_attribution = exclude_no",
			query:              q,
			vars:               hw{"lic": hw{"use_without_attribution": "EXCLUDE_NO"}},
			selector:           "routes.#.feed_version.feed.onestop_id",
			selectExpectUnique: []string{"CT", "HA"},
			selectExpectCount:  51,
		},
	}
	c, _, _, _ := newTestClient(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}
