package directions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/model"
)

// PROOF OF CONCEPT

type valhallaHandler struct {
	Client *http.Client
}

func newValhallaHandler(client *http.Client) *valhallaHandler {
	if client == nil {
		client = http.DefaultClient
	}
	return &valhallaHandler{
		Client: client,
	}
}

func (h *valhallaHandler) Request(req model.DirectionRequest) (*model.Directions, error) {
	if err := validateDirectionRequest(req); err != nil {
		return nil, err
	}
	input := valhallaRequest{}
	input.Locations = append(input.Locations, valhallaLocation{Lon: req.From.Lon, Lat: req.From.Lat})
	input.Locations = append(input.Locations, valhallaLocation{Lon: req.To.Lon, Lat: req.To.Lat})
	if req.Mode == model.StepModeAuto {
		input.Costing = "auto"
	} else if req.Mode == model.StepModeBicycle {
		input.Costing = "bicycle"
	} else if req.Mode == model.StepModeWalk {
		input.Costing = "pedestrian"
	}

	departAt := time.Now().In(time.UTC)
	if req.DepartAt == nil {
		departAt = time.Now().In(time.UTC)
		req.DepartAt = &departAt
	} else {
		departAt = *req.DepartAt
	}

	// Prepare response
	ret := model.Directions{
		Origin:      wpiWaypoint(req.From),
		Destination: wpiWaypoint(req.To),
		Success:     true,
		Exception:   nil,
	}

	res, err := makeValhallaRequest(h.Client, input)
	if err != nil || len(res.Trip.Legs) == 0 {
		ret.Success = false
		ret.Exception = aws.String("could not calculate route")
		return &ret, nil
	}

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
		leg.Geometry = tl.NewLineStringFromFlatCoords([]float64{})
		itin.Legs = append(itin.Legs, &leg)
	}
	if len(itin.Legs) > 0 {
		ret.Itineraries = append(ret.Itineraries, &itin)
	}
	return &ret, nil
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
	Time   int     `json:"time"`
	Length float64 `json:"length"`
}

type valhallaLeg struct {
	Shape     string             `json:"shape"`
	Maneuvers []valhallaManeuver `json:"maneuvers"`
	Summary   valhallaSummary    `json:"summary"`
}

type valhallaManeuver struct {
	Length          float64 `json:"length"`
	Time            int     `json:"time"`
	TravelMode      string  `json:"travel_mode"`
	Instruction     string  `json:"instruction"`
	BeginShapeIndex int     `json:"begin_shape_index"`
}

func makeValhallaRequest(client *http.Client, req valhallaRequest) (*valhallaResponse, error) {
	reqUrl := fmt.Sprintf("%s/route", os.Getenv("VALHALLA_ENDPOINT"))
	hreq, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	hreq.Body = io.NopCloser(bytes.NewReader(reqJson))
	hreq.Header.Add("api_key", os.Getenv("VALHALLA_API_KEY"))
	resp, err := client.Do(hreq)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println("response:", string(body))
	res := valhallaResponse{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func valDuration(t int) *model.Duration {
	return &model.Duration{Duration: float64(t), Units: model.DurationUnitSeconds}
}

func valDistance(v float64, units string) *model.Distance {
	return &model.Distance{Distance: v, Units: model.DistanceUnitKilometers}
}
