package directions

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/xy"
	"github.com/interline-io/transitland-server/model"
)

func init() {
	if err := RegisterRouter("line", func() Handler {
		return &lineRouter{}
	}); err != nil {
		panic(err)
	}
}

// lineRouter is a simple point-to-point handler for testing purposes
type lineRouter struct {
	Clock clock.Clock
}

func (h *lineRouter) Request(req model.DirectionRequest) (*model.Directions, error) {
	// Prepare response
	ret := model.Directions{
		Origin:      wpiWaypoint(req.From),
		Destination: wpiWaypoint(req.To),
		Success:     true,
		Exception:   nil,
	}
	if err := validateDirectionRequest(req); err != nil {
		ret.Success = false
		ret.Exception = aws.String("invalid input")
		return &ret, nil
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

	distance := xy.DistanceHaversine(req.From.Lon, req.From.Lat, req.To.Lon, req.To.Lat) / 1000.0
	speed := 1.0 // m/s
	switch req.Mode {
	case model.StepModeAuto:
		speed = 10
	case model.StepModeBicycle:
		speed = 4
	case model.StepModeWalk:
		speed = 1
	case model.StepModeTransit:
		speed = 5
	}
	duration := float64(distance * 1000 / speed)

	// Create itinerary summary
	itin := model.Itinerary{}
	itin.Duration = valDuration(duration)
	itin.Distance = valDistance(distance, "")
	itin.StartTime = departAt
	itin.EndTime = departAt.Add(time.Duration(duration) * time.Second)

	ret.Duration = itin.Duration
	ret.Distance = itin.Distance
	ret.StartTime = &itin.StartTime
	ret.EndTime = &itin.EndTime
	ret.DataSource = aws.String("LINE")

	// Create legs and steps for itinerary
	step := model.Step{}
	step.Duration = valDuration(duration)
	step.Distance = valDistance(distance, "")
	step.StartTime = departAt
	step.EndTime = departAt.Add(time.Duration(duration) * time.Second)
	step.GeometryOffset = 0

	leg := model.Leg{}
	leg.Steps = append(leg.Steps, &step)
	leg.Duration = valDuration(duration)
	leg.Distance = valDistance(distance, "")
	leg.StartTime = departAt
	leg.EndTime = departAt.Add(time.Duration(duration) * time.Second)
	leg.Geometry = tt.NewLineStringFromFlatCoords([]float64{
		req.From.Lon, req.From.Lat, 0.0,
		req.To.Lon, req.To.Lat, 0.0,
	})

	itin.Legs = append(itin.Legs, &leg)
	if len(itin.Legs) > 0 {
		ret.Itineraries = append(ret.Itineraries, &itin)
	}
	return &ret, nil
}
