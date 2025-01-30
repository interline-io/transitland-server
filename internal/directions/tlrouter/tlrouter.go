package tlrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/directions"
	"github.com/interline-io/transitland-server/model"
)

func init() {
	apikey := os.Getenv("TL_TLROUTER_APIKEY")
	endpoint := os.Getenv("TL_TLROUTER_ENDPOINT")
	if endpoint == "" {
		return
	}
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	if err := directions.RegisterRouter("tlrouter", func() directions.Handler {
		return NewRouter(client, endpoint, apikey)
	}); err != nil {
		panic(err)
	}
}

type Router struct {
	Clock    clock.Clock
	client   *http.Client
	endpoint string
	apikey   string
}

func NewRouter(client *http.Client, endpoint string, apikey string) *Router {
	if client == nil {
		client = http.DefaultClient
	}
	return &Router{
		client:   client,
		endpoint: endpoint,
		apikey:   apikey,
	}
}

func (h *Router) Request(ctx context.Context, req model.DirectionRequest) (*model.Directions, error) {
	if err := directions.ValidateDirectionRequest(req); err != nil {
		return &model.Directions{Success: false, Exception: aws.String("invalid input")}, nil
	}

	// Prepare request
	input := Request{}
	input.FromPlace = RequestLocation{Lat: req.From.Lat, Lon: req.From.Lon}
	input.ToPlace = RequestLocation{Lat: req.To.Lat, Lon: req.To.Lon}
	if req.Mode == model.StepModeTransit {
		input.Mode = "TRANSIT,WALK"
	} else if req.Mode == model.StepModeBicycle {
		input.Mode = "BICYCLE"
	} else if req.Mode == model.StepModeWalk {
		input.Mode = "WALK"
	} else {
		return &model.Directions{Success: false, Exception: aws.String("unsupported travel mode")}, nil
	}
	departAt := time.Now().In(time.UTC)
	if h.Clock != nil {
		departAt = h.Clock.Now()
	}
	if req.DepartAt == nil {
		req.DepartAt = &departAt
	} else {
		departAt = *req.DepartAt
	}
	// Ensure we are in UTC
	departAt = departAt.In(time.UTC)
	input.Time = departAt.Format("15:04:05")
	input.Date = departAt.Format("2006-01-02")

	// Make request
	res, err := makeRequest(ctx, input, h.client, h.endpoint, h.apikey)
	if err != nil || len(res.Plan.Itineraries) == 0 {
		log.For(ctx).Error().Err(err).Msg("tlrouter: failed to calculate route")
		return &model.Directions{Success: false, Exception: aws.String("could not calculate route")}, nil
	}
	// Prepare response
	ret := makeDirections(res, departAt)
	ret.Origin = wpiWaypoint(req.From)
	ret.Destination = wpiWaypoint(req.To)
	ret.Success = true
	ret.Exception = nil
	return ret, nil
}

func makeRequest(ctx context.Context, req Request, client *http.Client, endpoint string, apikey string) (*PlanResponse, error) {
	reqUrl := fmt.Sprintf("%s/route", endpoint)
	hreq, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	hreq.Body = io.NopCloser(bytes.NewReader(reqJson))
	hreq.Header.Add("api_key", apikey)
	log.TraceCheck(func() {
		log.For(ctx).Trace().Str("url", hreq.URL.String()).Str("body", string(reqJson)).Msg("tlrouter: request")
	})
	resp, err := client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := PlanResponse{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func makeDirections(res *PlanResponse, departAt time.Time) *model.Directions {
	// Map PlanResponse to Directions
	ret := model.Directions{}
	ret.DataSource = aws.String("OSM, Transitland")
	for _, vitin := range res.Plan.Itineraries {
		itin := model.Itinerary{}
		itin.From = &model.Waypoint{Lon: res.Plan.From.Lon, Lat: res.Plan.From.Lat, Name: aws.String(res.Plan.From.Name)}
		itin.To = &model.Waypoint{Lon: res.Plan.To.Lon, Lat: res.Plan.To.Lat, Name: aws.String(res.Plan.To.Name)}
		itin.Duration = makeDuration(float64(vitin.Duration))
		itin.Distance = makeDistance(vitin.Distance, "m")
		itin.StartTime = departAt
		itin.EndTime = departAt.Add(time.Duration(vitin.StartTime) * time.Millisecond)

		// Create legs for itinerary
		prevLegDepartAt := departAt
		for _, vleg := range vitin.Legs {
			leg := model.Leg{}
			prevStepDepartAt := prevLegDepartAt
			for _, vstep := range vleg.Steps {
				_ = vstep
				step := model.Step{}
				prevStepDepartAt = step.EndTime
				leg.Steps = append(leg.Steps, &step)
			}
			leg.Duration = makeDuration(float64(vleg.StartTime))
			leg.Distance = makeDistance(vleg.Distance, "m")
			leg.StartTime = prevLegDepartAt
			leg.EndTime = prevLegDepartAt.Add(time.Duration(vleg.StartTime) * time.Millisecond)
			prevLegDepartAt = leg.EndTime
			_ = prevStepDepartAt
			leg.Geometry = tt.NewLineStringFromFlatCoords([]float64{})
			// TODO: decode points
			itin.Legs = append(itin.Legs, &leg)
		}
		if len(itin.Legs) > 0 {
			ret.Itineraries = append(ret.Itineraries, &itin)
		}
	}
	if len(ret.Itineraries) > 0 {
		r0 := ret.Itineraries[0]
		ret.Duration = r0.Duration
		ret.Distance = r0.Distance
		ret.StartTime = &r0.StartTime
		ret.EndTime = &r0.EndTime
	}
	return &ret
}

type Request struct {
	// Required options
	FromPlace RequestLocation `json:"fromPlace"`
	ToPlace   RequestLocation `json:"toPlace"`
	Time      string          `json:"time"`
	Date      string          `json:"date"`

	// Advanced options
	MaxItineraries      int     `json:"maxItineraries"`
	Mode                string  `json:"mode"`
	ArriveBy            bool    `json:"arriveBy"`
	MaxWalkingDistance  float64 `json:"maxWalkingDistance"`
	WalkingSpeed        float64 `json:"walkingSpeed"`
	MaxK                int     `json:"maxK"`
	MaxTripTime         int     `json:"maxTripTime"`
	TransferTimePenalty int     `json:"transferTimePenalty"`
}

type RequestLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func wpiWaypoint(w *model.WaypointInput) *model.Waypoint {
	if w == nil {
		return nil
	}
	return &model.Waypoint{
		Lon:  w.Lon,
		Lat:  w.Lat,
		Name: w.Name,
	}
}

func makeDuration(t float64) *model.Duration {
	return &model.Duration{Duration: float64(t), Units: model.DurationUnitSeconds}
}

func makeDistance(v float64, units string) *model.Distance {
	_ = units
	return &model.Distance{Distance: v, Units: model.DistanceUnitKilometers}
}

// Generated from example.json

type PlanResponse struct {
	Plan Plan `json:"plan"`
}

type Plan struct {
	Date        int64       `json:"date"`
	From        Location    `json:"from"`
	To          Location    `json:"to"`
	Itineraries []Itinerary `json:"itineraries"`
}

type Location struct {
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Name          string  `json:"name"`
	StopOnestopId string  `json:"stopOnestopId"`
}

type Itinerary struct {
	Duration        int64   `json:"duration"`
	Distance        float64 `json:"distance"`
	StartTime       int64   `json:"startTime"`
	EndTime         int64   `json:"endTime"`
	WalkTime        int64   `json:"walkTime"`
	WalkDistance    float64 `json:"walkDistance"`
	TransitTime     int64   `json:"transitTime"`
	TransitDistance float64 `json:"transitDistance"`
	WaitingTime     int64   `json:"waitingTime"`
	Transfers       int     `json:"transfers"`
	Legs            []Leg   `json:"legs"`
}

type Leg struct {
	StartTime         int64       `json:"startTime"`
	EndTime           int64       `json:"endTime"`
	Distance          float64     `json:"distance"`
	Duration          int64       `json:"duration"`
	Mode              string      `json:"mode"`
	TransitLeg        bool        `json:"transitLeg"`
	From              LegLocation `json:"from"`
	To                LegLocation `json:"to"`
	Steps             []Step      `json:"steps"`
	LegGeometry       Geometry    `json:"legGeometry"`
	AgencyId          string      `json:"agencyId,omitempty"`
	AgencyName        string      `json:"agencyName,omitempty"`
	RouteShortName    string      `json:"routeShortName,omitempty"`
	RouteLongName     string      `json:"routeLongName,omitempty"`
	RouteType         int         `json:"routeType,omitempty"`
	RouteId           string      `json:"routeId,omitempty"`
	RouteColor        string      `json:"routeColor,omitempty"`
	RouteTextColor    string      `json:"routeTextColor,omitempty"`
	RouteOnestopId    string      `json:"routeOnestopId,omitempty"`
	TripId            string      `json:"tripId,omitempty"`
	Headsign          string      `json:"headsign,omitempty"`
	FeedId            string      `json:"feedId,omitempty"`
	FeedVersionSha1   string      `json:"feedVersionSha1,omitempty"`
	IntermediateStops []Stop      `json:"intermediateStops,omitempty"`
}

type LegLocation struct {
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Name          string  `json:"name"`
	Departure     int64   `json:"departure"`
	StopId        string  `json:"stopId,omitempty"`
	StopCode      string  `json:"stopCode,omitempty"`
	StopIndex     int     `json:"stopIndex,omitempty"`
	StopSequence  int     `json:"stopSequence,omitempty"`
	StopOnestopId string  `json:"stopOnestopId"`
}

type Step struct {
	// Define fields for steps if needed
}

type Geometry struct {
	Points string `json:"points"`
	Length int    `json:"length"`
}

type Stop struct {
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Name          string  `json:"name"`
	Departure     int64   `json:"departure"`
	StopId        string  `json:"stopId"`
	StopCode      string  `json:"stopCode"`
	StopIndex     int     `json:"stopIndex"`
	StopSequence  int     `json:"stopSequence"`
	StopOnestopId string  `json:"stopOnestopId"`
}
