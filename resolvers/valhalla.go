package resolvers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/interline-io/transitland-server/model"
)

// PROOF OF CONCEPT

type directionsRequest struct {
	Origin      model.Waypoint
	Destination model.Waypoint
	Mode        model.StepMode
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
}

type valhallaManeuver struct {
	Length      float64 `json:"length"`
	Time        int     `json:"time"`
	TravelMode  string  `json:"travel_mode"`
	Instruction string  `json:"instruction"`
}

func makeValhallaRequest(req valhallaRequest) (*valhallaResponse, error) {
	reqJson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqUrl := fmt.Sprintf(
		"%s/route?json=%s&apikey=%s",
		os.Getenv("VALHALLA_ENDPOINT"),
		url.QueryEscape(string(reqJson)),
		os.Getenv("VALHALLA_API_KEY"),
	)
	res := valhallaResponse{}
	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	fmt.Println("got response:", res)
	return &res, nil
}

func demoValhallaRequest(p directionsRequest) (*model.Directions, error) {
	valReq := valhallaRequest{}
	valReq.Costing = "pedestrian"
	valReq.Locations = append(valReq.Locations, valhallaLocation{Lon: p.Origin.Lon, Lat: p.Origin.Lat})
	valReq.Locations = append(valReq.Locations, valhallaLocation{Lon: p.Destination.Lon, Lat: p.Destination.Lat})
	valRes, err := makeValhallaRequest(valReq)
	if err != nil {
		return nil, err
	}
	dirRes, err := valhallaResponseToDirections(p, valRes)
	if err != nil {
		return nil, err
	}
	return dirRes, nil
}

func valhallaResponseToDirections(p directionsRequest, res *valhallaResponse) (*model.Directions, error) {
	t := time.Now()
	duration := model.Duration{Duration: float64(res.Trip.Summary.Time), Units: model.DurationUnitSeconds}
	distance := model.Distance{Distance: res.Trip.Summary.Length, Units: model.DistanceUnitKilometers}
	endTime := t.Add(time.Duration(duration.Duration) * time.Second)
	dirRes := model.Directions{
		Success:     true,
		Origin:      &p.Origin,
		Destination: &p.Destination,
		StartTime:   &t,
		EndTime:     &endTime,
		Duration:    &duration,
		Distance:    &distance,
	}
	itin := model.Itinerary{}
	for _, leg := range res.Trip.Legs {
		ll := model.Leg{}
		for _, m := range leg.Maneuvers {
			_ = m
			step := model.Step{}
			if m.TravelMode == "pedestrian" {
				step.Mode = model.StepModeWalk
			}
			step.To = &model.Waypoint{}
			step.Distance = &model.Distance{Distance: m.Length, Units: model.DistanceUnitKilometers}
			step.Duration = &model.Duration{Duration: float64(m.Time), Units: model.DurationUnitSeconds}
			step.Instruction = m.Instruction
			ll.Steps = append(ll.Steps, &step)
		}
		if len(ll.Steps) == 0 {
			continue
		}
		ll.Start = &p.Origin
		ll.End = &p.Destination
		ll.StartTime = t
		ll.EndTime = endTime
		itin.Legs = append(itin.Legs, &ll)
	}
	dirRes.Itineraries = append(dirRes.Itineraries, &itin)
	return &dirRes, nil
}
