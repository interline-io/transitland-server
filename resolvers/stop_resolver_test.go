package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/clock"
)

func TestStopResolver(t *testing.T) {
	bartStops := []string{"12TH", "16TH", "19TH", "19TH_N", "24TH", "ANTC", "ASHB", "BALB", "BAYF", "CAST", "CIVC", "COLS", "COLM", "CONC", "DALY", "DBRK", "DUBL", "DELN", "PLZA", "EMBR", "FRMT", "FTVL", "GLEN", "HAYW", "LAFY", "LAKE", "MCAR", "MCAR_S", "MLBR", "MONT", "NBRK", "NCON", "OAKL", "ORIN", "PITT", "PCTR", "PHIL", "POWL", "RICH", "ROCK", "SBRN", "SFIA", "SANL", "SHAY", "SSAN", "UCTY", "WCRK", "WARM", "WDUB", "WOAK"}
	caltrainRailStops := []string{"70011", "70012", "70021", "70022", "70031", "70032", "70041", "70042", "70051", "70052", "70061", "70062", "70071", "70072", "70081", "70082", "70091", "70092", "70101", "70102", "70111", "70112", "70121", "70122", "70131", "70132", "70141", "70142", "70151", "70152", "70161", "70162", "70171", "70172", "70191", "70192", "70201", "70202", "70211", "70212", "70221", "70222", "70231", "70232", "70241", "70242", "70251", "70252", "70261", "70262", "70271", "70272", "70281", "70282", "70291", "70292", "70301", "70302", "70311", "70312", "70321", "70322"}
	caltrainBusStops := []string{"777402", "777403"}
	caltrainStops := []string{}
	caltrainStops = append(caltrainStops, caltrainRailStops...)
	caltrainStops = append(caltrainStops, caltrainBusStops...)
	allStops := []string{}
	allStops = append(allStops, bartStops...)
	allStops = append(allStops, caltrainStops...)
	vars := hw{"stop_id": "MCAR"}
	testcases := []testcase{
		{
			"basic",
			`query($feed_version_sha1:String!) { stops(where:{feed_version_sha1:$feed_version_sha1}) { stop_id } }`, // just check BART
			hw{"feed_version_sha1": "e535eb2b3b9ac3ef15d82c56575e914575e732e0"},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"basic fields",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {onestop_id feed_version_sha1 feed_onestop_id location_type stop_code stop_desc stop_id stop_name stop_timezone stop_url wheelchair_boarding zone_id} }`,
			vars,
			`{"stops":[{"feed_onestop_id":"BA","feed_version_sha1":"e535eb2b3b9ac3ef15d82c56575e914575e732e0","location_type":0,"onestop_id":"s-9q9p1wxf72-macarthur","stop_code":"","stop_desc":"","stop_id":"MCAR","stop_name":"MacArthur","stop_timezone":"","stop_url":"http://www.bart.gov/stations/MCAR/","wheelchair_boarding":1,"zone_id":"MCAR"}]}`,
			"",
			nil,
		},
		{
			// just ensure this query completes successfully; checking coordinates is a pain and flaky.
			"geometry",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {geometry} }`,
			vars,
			``,
			"stops.0.geometry.type",
			[]string{"Point"},
		},
		{
			"feed_version",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {feed_version_sha1} }`,
			vars,
			`{"stops":[{"feed_version_sha1":"e535eb2b3b9ac3ef15d82c56575e914575e732e0"}]}`,
			"",
			nil,
		},
		{
			"route_stops",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {route_stops{route{route_id route_short_name}}} }`,
			vars,
			``,
			"stops.0.route_stops.#.route.route_id",
			[]string{"01", "03", "07"},
		},
		{
			"where near 10m",
			`query {stops(where:{near:{lon:-122.407974,lat:37.784471,radius:10.0}}) {stop_id onestop_id geometry}}`,
			vars,
			``,
			"stops.#.stop_id",
			[]string{"POWL"},
		},
		{
			"where near 2000m",
			`query {stops(where:{near:{lon:-122.407974,lat:37.784471,radius:2000.0}}) {stop_id onestop_id geometry}}`,
			vars,
			``,
			"stops.#.stop_id",
			[]string{"70011", "70012", "CIVC", "EMBR", "MONT", "POWL"},
		},
		{
			"where within polygon",
			`query{stops(where:{within:{type:"Polygon",coordinates:[[[-122.396,37.8],[-122.408,37.79],[-122.393,37.778],[-122.38,37.787],[-122.396,37.8]]]}}){id stop_id}}`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"EMBR", "MONT"},
		},
		{
			"where onestop_id",
			`query{stops(where:{onestop_id:"s-9q9k658fd1-sanjosediridoncaltrain"}) {stop_id} }`,
			vars,
			``,
			"stops.0.stop_id",
			[]string{"70262"},
		},
		{
			"where feed_version_sha1",
			`query($feed_version_sha1:String!) { stops(where:{feed_version_sha1:$feed_version_sha1}) { stop_id } }`, // just check BART
			hw{"feed_version_sha1": "e535eb2b3b9ac3ef15d82c56575e914575e732e0"},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"where feed_onestop_id",
			`query{stops(where:{feed_onestop_id:"BA"}) { stop_id } }`, // just check BART
			hw{},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"where stop_id",
			`query{stops(where:{stop_id:"12TH"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"12TH"},
		},
		{
			"where search",
			`query{stops(where:{search:"macarthur"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"MCAR", "MCAR_S"},
		},
		{
			"where search 2",
			`query{stops(where:{search:"ftvl"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"FTVL"},
		},
		{
			"where search 3",
			`query{stops(where:{search:"warm springs"}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			[]string{"WARM"},
		},
		// served_by_onestop_ids
		{
			"served_by_onestop_ids=o-9q9-bayarearapidtransit",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			bartStops,
		},
		{
			"served_by_onestop_ids=o-9q9-caltrain",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-caltrain"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// caltrain stops minus a couple non-service stops
			caltrainStops,
		},
		{
			"served_by_onestop_ids=r-9q9-antioch~sfia~millbrae",
			`query{stops(where:{served_by_onestop_ids:["r-9q9-antioch~sfia~millbrae"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// yellow line stops
			[]string{"12TH", "16TH", "19TH", "19TH_N", "24TH", "ANTC", "BALB", "CIVC", "COLM", "CONC", "DALY", "EMBR", "GLEN", "LAFY", "MCAR", "MCAR_S", "MLBR", "MONT", "NCON", "ORIN", "PITT", "PCTR", "PHIL", "POWL", "ROCK", "SBRN", "SFIA", "SSAN", "WCRK", "WOAK"},
		},
		{
			"served_by_onestop_ids=r-9q9-antioch~sfia~millbrae,r-9q8y-richmond~dalycity~millbrae",
			`query{stops(where:{served_by_onestop_ids:["r-9q9-antioch~sfia~millbrae","r-9q8y-richmond~dalycity~millbrae"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// combination of yellow and red line stops
			[]string{"12TH", "16TH", "19TH", "19TH_N", "24TH", "ANTC", "ASHB", "BALB", "CIVC", "COLM", "CONC", "DALY", "DBRK", "DELN", "PLZA", "EMBR", "GLEN", "LAFY", "MCAR", "MCAR_S", "MLBR", "MONT", "NBRK", "NCON", "ORIN", "PITT", "PCTR", "PHIL", "POWL", "RICH", "ROCK", "SBRN", "SFIA", "SSAN", "WCRK", "WOAK"},
		},
		{
			"served_by_onestop_ids=o-9q9-bayarearapidtransit,r-9q9-antioch~sfia~millbrae",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit","r-9q9-antioch~sfia~millbrae"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// all bart stops
			bartStops,
		},
		{
			"served_by_onestop_ids=o-9q9-bayarearapidtransit,o-9q9-caltrain",
			`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit","o-9q9-caltrain"]}) { stop_id } }`,
			hw{},
			``,
			"stops.#.stop_id",
			// all stops
			allStops,
		},
		// {
		// 	"served_by_route_types=2,served_by_onestop_ids=o-9q9-bayarearapidtransit,o-9q9-caltrain",
		// 	`query{stops(where:{served_by_onestop_ids:["o-9q9-bayarearapidtransit","o-9q9-caltrain"], served_by_route_types:[2]}) { stop_id } }`,
		// 	hw{},
		// 	``,
		// 	"stops.#.stop_id",
		// 	caltrainRailStops,
		// },
		// TODO: parent, children; test data has no stations.
		// TODO: level, pathways_from_stop, pathways_to_stop: test data has no pathways...
		// TODO: census_geographies
		// stop_times
		{
			"stop_times",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {stop_times { trip { trip_id} }} }`,
			hw{"stop_id": "70302"}, // Morgan hill
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"268", "274", "156"},
		},
		{
			"stop_times where weekday_morning",
			`query($stop_id: String!, $service_date:Date!) {  stops(where:{stop_id:$stop_id}) {stop_times(where:{service_date:$service_date, start_time:21600, end_time:25200}) { trip { trip_id} }} }`,
			hw{"stop_id": "MCAR", "service_date": "2018-05-29"},
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"3830503WKDY", "3850526WKDY", "3610541WKDY", "3630556WKDY", "3650611WKDY", "2210533WKDY", "2230548WKDY", "2250603WKDY", "2270618WKDY", "4410518WKDY", "4430533WKDY", "4450548WKDY", "4470603WKDY"},
		},
		{
			"stop_times where sunday_morning",
			`query($stop_id: String!, $service_date:Date!) {  stops(where:{stop_id:$stop_id}) {stop_times(where:{service_date:$service_date, start_time:21600, end_time:36000}) { trip { trip_id} }} }`,
			hw{"stop_id": "MCAR", "service_date": "2018-05-27"},
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"3730756SUN", "3750757SUN", "3770801SUN", "3790821SUN", "3610841SUN", "3630901SUN", "2230800SUN", "2250748SUN", "2270808SUN", "2290828SUN", "2310848SUN", "2330908SUN"},
		},
		{
			"stop_times where saturday_evening",
			`query($stop_id: String!, $service_date:Date!) {  stops(where:{stop_id:$stop_id}) {stop_times(where:{service_date:$service_date, start_time:57600, end_time:72000}) { trip { trip_id} }} }`,
			hw{"stop_id": "MCAR", "service_date": "2018-05-26"},
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"3611521SAT", "3631541SAT", "3651601SAT", "3671621SAT", "3691641SAT", "3711701SAT", "3731721SAT", "3751741SAT", "3771801SAT", "3791821SAT", "3611841SAT", "3631901SAT", "2231528SAT", "2251548SAT", "2271608SAT", "2291628SAT", "2311648SAT", "2331708SAT", "2351728SAT", "2211748SAT", "2231808SAT", "2251828SAT", "2271848SAT", "2291908SAT", "4471533SAT", "4491553SAT", "4511613SAT", "4531633SAT", "4411653SAT", "4431713SAT", "4451733SAT", "4471753SAT", "4491813SAT", "4511833SAT", "4531853SAT"},
		},
		// TODO: census_geographies
		// TODO: route_stop_buffer
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_Cursor(t *testing.T) {
	// First 1000 stops...
	allEnts, err := TestDBFinder.FindStops(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	allIds := []string{}
	for _, st := range allEnts {
		allIds = append(allIds, st.StopID)
	}
	testcases := []testcase{
		{
			"no cursor",
			"query{stops(limit:100){feed_version{id} id stop_id}}",
			nil,
			``,
			"stops.#.stop_id",
			allIds[:100],
		},
		{
			"after 0",
			"query{stops(after: 0, limit:100){feed_version{id} id stop_id}}",
			nil,
			``,
			"stops.#.stop_id",
			allIds[:100],
		},
		{
			"after 10th",
			"query($after: Int!){stops(after: $after, limit:10){feed_version{id} id stop_id}}",
			hw{"after": allEnts[10].ID},
			``,
			"stops.#.stop_id",
			allIds[11:21],
		},
		{
			"after invalid id behaves like (0,0)",
			"query($after: Int!){stops(after: $after, limit:10){feed_version{id} id stop_id}}",
			hw{"after": 10_000_000},
			``,
			"stops.#.stop_id",
			allIds[:10],
		},

		// TODO: uncomment after schema changes
		// {
		// 	"no cursor",
		// 	"query($cursor: Cursor!){stops(after: $cursor, limit:100){feed_version{id} id stop_id}}",
		// 	hw{"cursor": 0},
		// 	``,
		// 	"stops.#.stop_id",
		// 	stopIds[:100],
		// },
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_PreviousOnestopID(t *testing.T) {
	testcases := []testcase{
		{
			"default",
			`query($osid:String!, $previous:Boolean!) { stops(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous}) { stop_id onestop_id }}`,
			hw{"osid": "s-9q9nfsxn67-fruitvale", "previous": false},
			``,
			"stops.#.onestop_id",
			[]string{"s-9q9nfsxn67-fruitvale"},
		},
		{
			"old id no result",
			`query($osid:String!, $previous:Boolean!) { stops(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous}) { stop_id onestop_id }}`,
			hw{"osid": "s-9q9nfswzpg-fruitvale", "previous": false},
			``,
			"stops.#.onestop_id",
			[]string{},
		},
		{
			"old id no specify fv",
			`query($osid:String!, $previous:Boolean!) { stops(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous, feed_version_sha1:"dd7aca4a8e4c90908fd3603c097fabee75fea907"}) { stop_id onestop_id }}`,
			hw{"osid": "s-9q9nfswzpg-fruitvale", "previous": false},
			``,
			"stops.#.onestop_id",
			[]string{"s-9q9nfswzpg-fruitvale"},
		},
		{
			"use previous",
			`query($osid:String!, $previous:Boolean!) { stops(where:{onestop_id:$osid, allow_previous_onestop_ids:$previous}) { stop_id onestop_id }}`,
			hw{"osid": "s-9q9nfswzpg-fruitvale", "previous": true},
			``,
			"stops.#.onestop_id",
			[]string{"s-9q9nfswzpg-fruitvale"},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_StopTimes(t *testing.T) {
	vars := hw{"trip_id": "3850526WKDY"}
	testcases := []testcase{
		{
			"basic",
			`query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {trip_id stop_times { arrival_time }} }`,
			vars,
			``,
			"trips.0.stop_times.#.arrival_time",
			[]string{"05:26:00", "05:29:00", "05:33:00", "05:36:00", "05:40:00", "05:43:00", "05:46:00", "05:48:00", "05:50:00", "05:53:00", "05:54:00", "05:56:00", "05:58:00", "06:05:00", "06:08:00", "06:11:00", "06:15:00", "06:17:00", "06:23:00", "06:27:00", "06:32:00", "06:35:00", "06:40:00", "06:43:00", "06:50:00", "07:05:00", "07:13:00"},
		},
		{
			// these are supposed to always be ordered by stop_sequence, so we can directly check the first one.
			"basic fields",
			`query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {trip_id stop_times(limit:1) { arrival_time departure_time stop_sequence stop_headsign pickup_type drop_off_type timepoint interpolated}} }`,
			vars,
			`{"trips":[{"stop_times":[{"arrival_time":"05:26:00","departure_time":"05:26:00","drop_off_type":null,"interpolated":null,"pickup_type":null,"stop_headsign":"Antioch","stop_sequence":1,"timepoint":1}],"trip_id":"3850526WKDY"}]}`,
			"",
			nil,
		},
		{
			// check stops for a trip
			"stop",
			`query($trip_id: String!) {  trips(where:{trip_id:$trip_id}) {trip_id stop_times { stop { stop_id } }} }`,
			vars,
			``,
			"trips.0.stop_times.#.stop.stop_id",
			[]string{"SFIA", "SBRN", "SSAN", "COLM", "DALY", "BALB", "GLEN", "24TH", "16TH", "CIVC", "POWL", "MONT", "EMBR", "WOAK", "12TH", "19TH_N", "MCAR", "ROCK", "ORIN", "LAFY", "WCRK", "PHIL", "CONC", "NCON", "PITT", "PCTR", "ANTC"},
		},
		{
			// go through a stop to get trip_ids
			"trip",
			`query($stop_id: String!) {  stops(where:{stop_id:$stop_id}) {stop_times { trip { trip_id} }} }`,
			hw{"stop_id": "70302"}, // Morgan hill
			``,
			"stops.0.stop_times.#.trip.trip_id",
			[]string{"268", "274", "156"},
		},
		// check StopTimeFilter through a stop
		{
			"where service_date start_time end_time",
			`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{service_date:"2018-05-30", start_time: 26000, end_time: 30000}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"07:18:00", "07:24:00", "07:28:00", "07:33:00", "07:39:00", "07:43:00", "07:48:00", "07:54:00", "07:58:00", "08:03:00", "08:09:00", "08:18:00", "07:24:00", "07:39:00", "07:54:00", "08:09:00", "07:16:00", "07:31:00", "07:46:00", "08:01:00", "08:16:00"},
		},
		{
			"where service_date end_time",
			`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{service_date:"2018-05-30", end_time: 20000}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"04:39:00", "04:54:00", "05:09:00", "05:24:00", "04:39:00", "04:54:00", "05:09:00", "05:24:00", "04:31:00", "04:46:00", "05:01:00", "05:16:00", "05:31:00"},
		},
		{
			"where service_date start_time",
			`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{service_date:"2018-05-30", start_time: 76000}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"21:14:00", "21:34:00", "21:54:00", "22:14:00", "22:34:00", "22:54:00", "23:14:00", "23:34:00", "23:54:00", "24:14:00", "24:47:00", "21:14:00", "21:34:00", "21:54:00", "22:14:00", "22:34:00", "22:54:00", "23:14:00", "23:34:00", "23:54:00", "24:14:00", "24:47:00"},
		},
		// check arrival and departure resolvers
		{
			"arrival departure base case",
			`query{ stops(where:{stop_id:"RICH"}) { stop_times(where:{service_date:"2018-05-30", start_time: 76000, end_time: 76900}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.stop_times.#.departure_time",
			[]string{"21:09:00", "21:14:00", "21:15:00"},
		},
		{
			"departures",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{service_date:"2018-05-30", start_time: 76000, end_time: 76900}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"21:15:00"},
		},
		{
			"arrivals",
			`query{ stops(where:{stop_id:"RICH"}) { arrivals(where:{service_date:"2018-05-30", start_time: 76000, end_time: 76900}) {arrival_time}}}`,
			hw{},
			``,
			"stops.0.arrivals.#.arrival_time",
			[]string{"21:09:00", "21:14:00"},
		},
		// route_onestop_ids
		{
			"departure route_onestop_ids",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:12:00", "10:20:00", "10:27:00", "10:35:00", "10:42:00", "10:50:00", "10:57:00"},
		},
		{
			"departure route_onestop_ids 1",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{route_onestop_ids: ["r-9q8y-richmond~dalycity~millbrae"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:12:00", "10:27:00", "10:42:00", "10:57:00"},
		},
		{
			"departure route_onestop_ids 2",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		// Allow previous route onestop ids
		// OLD: r-9q9n-fremont~richmond
		// NEW: r-9q9n-warmsprings~southfremont~richmond
		{
			"departure route_onestop_ids use previous id current ok",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: false, route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		{
			"departure route_onestop_ids, use previous id, both at once ok",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: false, route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond","r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		{
			"departure route_onestop_ids, use previous id, both at once, no duplicates",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: true, route_onestop_ids: ["r-9q9n-warmsprings~southfremont~richmond","r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
		{
			"departure route_onestop_ids, use previous id, old, fail",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: false, route_onestop_ids: ["r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{},
		},
		{
			"departure route_onestop_ids, use previous id, old, ok",
			`query{ stops(where:{stop_id:"RICH"}) { departures(where:{allow_previous_route_onestop_ids: true, route_onestop_ids: ["r-9q9n-fremont~richmond"], service_date:"2018-05-30", start_time: 36000, end_time: 39600}) {departure_time}}}`,
			hw{},
			``,
			"stops.0.departures.#.departure_time",
			[]string{"10:05:00", "10:20:00", "10:35:00", "10:50:00"},
		},
	}
	c := newTestClient()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_StopTimes_ServiceDate(t *testing.T) {
	q := `query($stop_id:String!,$sd:Date!,$ed:Boolean){ stops(where:{stop_id:$stop_id}) { stop_times(where:{service_date:$sd, start_time:54000, end_time:57600, use_service_window:$ed}) {service_date arrival_time}}}`
	testcases := []testcase{
		{
			"service date in range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2018-05-29", "ed": true},
			``,
			"stops.0.stop_times.0.service_date",
			[]string{"2018-05-29"}, // expect input date
		},
		{
			"service date after range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": true},
			``,
			"stops.0.stop_times.0.service_date",
			[]string{"2018-06-05"}, // expect adjusted date in window
		},
		{
			"service date before range, friday",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2010-05-28", "ed": true},
			``,
			"stops.0.stop_times.0.service_date",
			[]string{"2018-06-08"}, // expect adjusted date in window
		},
		{
			"service date after range, exact dates",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": false},
			``,
			"stops.0.stop_times.#.service_date",
			[]string{}, // exect no results
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient()
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_StopTimes_WindowDates(t *testing.T) {
	bartWeekdayTimes := []string{"15:01:00", "15:09:00", "15:09:00", "15:16:00", "15:24:00", "15:24:00", "15:31:00", "15:39:00", "15:39:00", "15:46:00", "15:54:00", "15:54:00"}
	bartWeekendTimes := []string{"15:15:00", "15:15:00", "15:35:00", "15:35:00", "15:55:00", "15:55:00"}
	q := `query($stop_id:String!,$sd:Date!,$ed:Boolean){ stops(where:{stop_id:$stop_id}) { stop_times(where:{service_date:$sd, start_time:54000, end_time:57600, use_service_window:$ed}) {arrival_time}}}`
	testcases := []testcase{
		{
			"service date in range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2018-05-29", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekdayTimes,
		},
		{
			"service date after range",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekdayTimes,
		},
		{
			"service date after range, exact dates",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-28", "ed": false},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{},
		},
		{
			"service date after range, sunday",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2030-05-26", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekendTimes,
		},
		{
			"service date before range, tuesday",
			q,
			hw{"stop_id": "MCAR_S", "sd": "2010-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			bartWeekdayTimes,
		},
		{
			"fv without feed_info, in window, monday",
			q,
			hw{"stop_id": "70011", "sd": "2019-02-11", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"15:48:00", "15:50:00"},
		},
		{
			"fv without feed_info, before window, friday",
			q,
			hw{"stop_id": "70011", "sd": "2010-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"15:48:00", "15:50:00"},
		},
		{
			"fv without feed_info, after window, tuesday",
			q,
			hw{"stop_id": "70011", "sd": "2030-05-28", "ed": true},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{"15:48:00", "15:50:00"},
		},
		{
			"fv without feed_info, after window, tuesday, exact date only",
			q,
			hw{"stop_id": "70011", "sd": "2030-05-28", "ed": false},
			``,
			"stops.0.stop_times.#.arrival_time",
			[]string{},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient()
			testquery(t, c, tc)
		})
	}
}

func TestStopResolver_StopTimes_Next(t *testing.T) {
	type tcWithClock struct {
		testcase
		when string
	}
	testcases := []tcWithClock{
		// Relative times
		{
			testcase{
				"where next 3600",
				`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{next:3600}) {arrival_time}}}`,
				hw{},
				``,
				"stops.0.stop_times.#.arrival_time",
				// these should start at 15:00 - 16:00
				[]string{"15:01:00", "15:09:00", "15:09:00", "15:16:00", "15:24:00", "15:24:00", "15:31:00", "15:39:00", "15:39:00", "15:46:00", "15:54:00", "15:54:00"},
			},
			"2018-05-30T22:00:00",
		},
		{
			testcase{
				"where next 1800",
				`query{ stops(where:{stop_id:"MCAR_S"}) { stop_times(where:{next:1800}) {arrival_time}}}`,
				hw{},
				``,
				"stops.0.stop_times.#.arrival_time",
				// these should start at 15:00 - 15:30
				[]string{"15:01:00", "15:09:00", "15:09:00", "15:16:00", "15:24:00", "15:24:00"},
			},
			"2018-05-30T22:00:00",
		},
		{
			testcase{
				"where next 900, east coast",
				`query{ stops(where:{stop_id:"6497"}) { stop_times(where:{next:900}) {arrival_time}}}`,
				hw{},
				``,
				"stops.0.stop_times.#.arrival_time",
				// these should start at 18:00 - 18:15
				[]string{"18:00:00", "18:00:00", "18:00:00", "18:00:00", "18:00:00", "18:03:00", "18:10:00", "18:10:00", "18:13:00", "18:14:00", "18:15:00", "18:15:00"},
			},
			"2018-05-30T22:00:00",
		},
		{
			testcase{
				"where next 600, multiple timezones",
				`query{ stops(where:{onestop_ids:["s-dhvrsm227t-universityareatransitcenter", "s-9q9p1wxf72-macarthur"]}) { onestop_id stop_id stop_times(where:{next:600}) {arrival_time}}}`,
				hw{},
				// this test checks the json response because it is too complex for the simple element selector approach
				// we should expect east coast times 18:00-18:10, and west coast times 15:00-15:10
				`{
					"stops": [
					{
						"onestop_id": "s-9q9p1wxf72-macarthur",
						"stop_id": "MCAR",
						"stop_times": [{
							"arrival_time": "15:00:00"
						}, {
							"arrival_time": "15:07:00"
						}]
					}, {
						"onestop_id": "s-9q9p1wxf72-macarthur",
						"stop_id": "MCAR_S",
						"stop_times": [{
							"arrival_time": "15:01:00"
						}, {
							"arrival_time": "15:09:00"
						}, {
							"arrival_time": "15:09:00"
						}]
					},
					{
						"onestop_id": "s-dhvrsm227t-universityareatransitcenter",
						"stop_id": "6497",
						"stop_times": [{
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:00:00"
						}, {
							"arrival_time": "18:03:00"
						}, {
							"arrival_time": "18:10:00"
						}, {
							"arrival_time": "18:10:00"
						}]
					}]
				}`,
				"",
				nil,
			},
			"2018-05-30T22:00:00",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// 2018-05-28 22:00:00 +0000 UTC
			// 2018-05-28 15:00:00 -0700 PDT
			when, err := time.Parse("2006-01-02T15:04:05", tc.when)
			if err != nil {
				t.Fatal(err)
			}
			c := newTestClientWithClock(&clock.Mock{T: when})
			testquery(t, c, tc.testcase)
		})
	}
}
