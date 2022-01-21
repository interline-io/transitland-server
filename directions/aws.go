package directions

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-server/model"
)

type awsHandler struct {
	Region         string
	CalculatorName string
}

func newAWSHandler(region string, calculator string) *awsHandler {
	return &awsHandler{
		Region:         region,
		CalculatorName: calculator,
	}
}

func (h *awsHandler) Request(req model.DirectionRequest) (*model.Directions, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(h.Region)})
	if err != nil {
		return nil, err
	}
	ls := locationservice.New(sess)

	// Prepare input
	input := locationservice.CalculateRouteInput{
		CalculatorName:      aws.String(h.CalculatorName),
		DeparturePosition:   aws.Float64Slice([]float64{req.From.Lon, req.From.Lat}),
		DestinationPosition: aws.Float64Slice([]float64{req.To.Lon, req.To.Lat}),
		DistanceUnit:        aws.String("Kilometers"),
		IncludeLegGeometry:  aws.Bool(true),
	}
	if req.Mode == model.StepModeAuto {
		input.TravelMode = aws.String("Car")
	} else if req.Mode == model.StepModeWalk {
		input.TravelMode = aws.String("Walking")
	}

	departAt := time.Now().In(time.UTC)
	if req.DepartAt == nil {
		departAt = time.Now().In(time.UTC)
		req.DepartAt = &departAt
		input.DepartNow = aws.Bool(true)
	} else {
		t2 := departAt
		input.DepartureTime = &t2
	}
	if err := input.Validate(); err != nil {
		fmt.Println("error:", err)
		return nil, errors.New("could not validate route input")
	}

	// Prepare response
	ret := model.Directions{
		Origin:      wpiWaypoint(req.From),
		Destination: wpiWaypoint(req.To),
		Success:     true,
		Exception:   nil,
	}
	res, err := ls.CalculateRoute(&input)
	if err != nil || res.Summary == nil {
		ret.Success = false
		ret.Exception = aws.String("could not calculate route")
		return &ret, nil
	}

	// Create itinerary summary
	var distUnits *string
	itin := model.Itinerary{}
	distUnits = res.Summary.DistanceUnit
	itin.Duration = awsDuration(res.Summary.DurationSeconds)
	itin.Distance = awsDistance(res.Summary.Distance, distUnits)
	itin.StartTime = departAt
	if res.Summary.DurationSeconds != nil {
		itin.EndTime = departAt.Add(time.Duration(*res.Summary.DurationSeconds) * time.Second)
	}
	// aws responses have single itineraries
	ret.Duration = itin.Duration
	ret.Distance = itin.Distance
	ret.StartTime = &itin.StartTime
	ret.EndTime = &itin.EndTime
	ret.DataSource = res.Summary.DataSource

	// Create legs for itinerary
	prevLegDepartAt := departAt
	for _, awsleg := range res.Legs {
		if awsleg.DurationSeconds == nil {
			ret.Success = false
			ret.Exception = aws.String("invalid route response")
			return &ret, nil
		}
		leg := model.Leg{}
		prevStepDepartAt := prevLegDepartAt
		for _, awsstep := range awsleg.Steps {
			step := model.Step{}
			step.Duration = awsDuration(awsstep.DurationSeconds)
			step.Distance = awsDistance(awsstep.Distance, distUnits)
			step.StartTime = prevStepDepartAt
			step.EndTime = prevStepDepartAt.Add(time.Duration(*awsstep.DurationSeconds) * time.Second)
			step.To = awsWaypoint(awsstep.EndPosition)
			step.GeometryOffset = awsInt(awsstep.GeometryOffset)
			prevStepDepartAt = step.EndTime
			leg.Steps = append(leg.Steps, &step)
		}
		leg.Duration = awsDuration(awsleg.DurationSeconds)
		leg.Distance = awsDistance(awsleg.Distance, distUnits)
		leg.StartTime = prevLegDepartAt
		leg.EndTime = prevLegDepartAt.Add(time.Duration(*awsleg.DurationSeconds) * time.Second)
		leg.From = awsWaypoint(awsleg.StartPosition)
		leg.To = awsWaypoint(awsleg.EndPosition)
		prevLegDepartAt = leg.EndTime
		if awsleg.Geometry != nil {
			leg.Geometry = awsLineString(awsleg.Geometry.LineString)
		}
		itin.Legs = append(itin.Legs, &leg)
	}
	if len(itin.Legs) > 0 {
		ret.Itineraries = append(ret.Itineraries, &itin)
	}
	return &ret, nil
}

func wpiWaypoint(w *model.WaypointInput) *model.Waypoint {
	return &model.Waypoint{
		Lon:  w.Lon,
		Lat:  w.Lat,
		Name: w.Name,
	}
}

func awsInt(v *int64) int {
	if v == nil {
		return 0
	}
	return int(*v)
}

func awsLineString(v [][]*float64) tl.LineString {
	coords := []float64{}
	for _, coord := range v {
		if len(coord) == 2 && coord[0] != nil && coord[1] != nil {
			coords = append(coords, *coord[0], *coord[1], 0)
		}
	}
	return tl.NewLineStringFromFlatCoords(coords)
}

func awsWaypoint(v []*float64) *model.Waypoint {
	if len(v) != 2 || v[0] == nil || v[1] == nil {
		return nil
	}
	return &model.Waypoint{
		Lon: *v[0],
		Lat: *v[1],
	}
}

func awsDuration(v *float64) *model.Duration {
	if v == nil {
		return nil
	}
	r := model.Duration{
		Duration: *v,
		Units:    model.DurationUnitSeconds,
	}
	return &r
}

func awsDistance(v *float64, units *string) *model.Distance {
	if v == nil || units == nil {
		return nil
	}
	r := model.Distance{}
	switch *units {
	case locationservice.DistanceUnitKilometers:
		r.Units = model.DistanceUnitKilometers
	case locationservice.DistanceUnitMiles:
		r.Units = model.DistanceUnitMiles
	default:
		return nil
	}
	r.Distance = *v
	return &r
}
