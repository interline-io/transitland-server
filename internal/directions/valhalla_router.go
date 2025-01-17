package directions

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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-mw/caches/httpcache"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/model"
)

func init() {
	endpoint := os.Getenv("TL_VALHALLA_ENDPOINT")
	apikey := os.Getenv("TL_VALHALLA_API_KEY")
	if endpoint == "" {
		return
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	if os.Getenv("TL_DIRECTIONS_ENABLE_CACHE") != "" {
		client.Transport = httpcache.NewCache(nil, nil, httpcache.NewTTLCache(16*1024, 24*time.Hour))
	}
	if err := RegisterRouter("valhalla", func() Handler {
		return newValhallaRouter(client, endpoint, apikey)
	}); err != nil {
		panic(err)
	}
}

type valhallaRouter struct {
	Clock    clock.Clock
	client   *http.Client
	endpoint string
	apikey   string
}

func newValhallaRouter(client *http.Client, endpoint string, apikey string) *valhallaRouter {
	if client == nil {
		client = http.DefaultClient
	}
	return &valhallaRouter{
		client:   client,
		endpoint: endpoint,
		apikey:   apikey,
	}
}

func (h *valhallaRouter) Request(ctx context.Context, req model.DirectionRequest) (*model.Directions, error) {
	if err := validateDirectionRequest(req); err != nil {
		return &model.Directions{Success: false, Exception: aws.String("invalid input")}, nil
	}

	// Prepare request
	input := valhallaRequest{}
	input.Locations = append(input.Locations, valhallaLocation{Lon: req.From.Lon, Lat: req.From.Lat})
	input.Locations = append(input.Locations, valhallaLocation{Lon: req.To.Lon, Lat: req.To.Lat})
	if req.Mode == model.StepModeAuto {
		input.Costing = "auto"
	} else if req.Mode == model.StepModeBicycle {
		input.Costing = "bicycle"
	} else if req.Mode == model.StepModeWalk {
		input.Costing = "pedestrian"
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

	// Make request
	res, err := makeValRequest(ctx, input, h.client, h.endpoint, h.apikey)
	if err != nil || len(res.Trip.Legs) == 0 {
		log.For(ctx).Error().Err(err).Msg("valhalla router failed to calculate route")
		return &model.Directions{Success: false, Exception: aws.String("could not calculate route")}, nil
	}
	// Prepare response
	ret := makeValDirections(res, departAt)
	ret.Origin = wpiWaypoint(req.From)
	ret.Destination = wpiWaypoint(req.To)
	ret.Success = true
	ret.Exception = nil
	return ret, nil
}

func makeValRequest(ctx context.Context, req valhallaRequest, client *http.Client, endpoint string, apikey string) (*valhallaResponse, error) {
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
	log.For(ctx).Debug().Str("url", hreq.URL.String()).Str("body", string(reqJson)).Msg("valhalla request")
	resp, err := client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := valhallaResponse{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func makeValDirections(res *valhallaResponse, departAt time.Time) *model.Directions {
	ret := model.Directions{}
	// Create itinerary summary
	itin := model.Itinerary{}
	itin.Duration = valDuration(res.Trip.Summary.Time)
	itin.Distance = valDistance(res.Trip.Summary.Length, res.Units)
	itin.StartTime = departAt
	itin.EndTime = departAt.Add(time.Duration(res.Trip.Summary.Time) * time.Second)
	// valhalla responses have single itineraries
	ret.Duration = itin.Duration
	ret.Distance = itin.Distance
	ret.StartTime = &itin.StartTime
	ret.EndTime = &itin.EndTime
	ret.DataSource = aws.String("OSM")

	// Create legs for itinerary
	prevLegDepartAt := departAt
	for _, vleg := range res.Trip.Legs {
		leg := model.Leg{}
		prevStepDepartAt := prevLegDepartAt
		for _, vstep := range vleg.Maneuvers {
			step := model.Step{}
			step.Duration = valDuration(vstep.Time)
			step.Distance = valDistance(vstep.Length, res.Units)
			step.StartTime = prevStepDepartAt
			step.EndTime = prevStepDepartAt.Add(time.Duration(vstep.Time) * time.Second)
			// step.To = vstep.
			step.GeometryOffset = vstep.BeginShapeIndex
			prevStepDepartAt = step.EndTime
			leg.Steps = append(leg.Steps, &step)
		}
		leg.Duration = valDuration(vleg.Summary.Time)
		leg.Distance = valDistance(vleg.Summary.Length, res.Units)
		leg.StartTime = prevLegDepartAt
		leg.EndTime = prevLegDepartAt.Add(time.Duration(vleg.Summary.Time) * time.Second)
		// leg.From = awsWaypoint(awsleg.StartPosition)
		// leg.To = awsWaypoint(awsleg.EndPosition)
		prevLegDepartAt = leg.EndTime
		leg.Geometry = tt.NewLineStringFromFlatCoords([]float64{})
		itin.Legs = append(itin.Legs, &leg)
	}
	if len(itin.Legs) > 0 {
		ret.Itineraries = append(ret.Itineraries, &itin)
	}
	return &ret
}

type valhallaRequest struct {
	Locations []valhallaLocation `json:"locations"`
	Costing   string             `json:"costing"`
}

type valhallaLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type valhallaResponse struct {
	Trip  valhallaTrip `json:"trip"`
	Units string       `json:"units"`
}

type valhallaTrip struct {
	Legs    []valhallaLeg   `json:"legs"`
	Summary valhallaSummary `json:"summary"`
}

type valhallaSummary struct {
	Time   float64 `json:"time"`
	Length float64 `json:"length"`
}

type valhallaLeg struct {
	Shape     string             `json:"shape"`
	Maneuvers []valhallaManeuver `json:"maneuvers"`
	Summary   valhallaSummary    `json:"summary"`
}

type valhallaManeuver struct {
	Length          float64 `json:"length"`
	Time            float64 `json:"time"`
	TravelMode      string  `json:"travel_mode"`
	Instruction     string  `json:"instruction"`
	BeginShapeIndex int     `json:"begin_shape_index"`
}

func valDuration(t float64) *model.Duration {
	return &model.Duration{Duration: float64(t), Units: model.DurationUnitSeconds}
}

func valDistance(v float64, units string) *model.Distance {
	_ = units
	return &model.Distance{Distance: v, Units: model.DistanceUnitKilometers}
}
