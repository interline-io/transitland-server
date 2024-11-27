package rest

import (
	"context"
	"time"

	"github.com/interline-io/transitland-server/internal/directions"
	"github.com/interline-io/transitland-server/model"
)

// AutoTrafficMode represents the traffic mode for auto routing
type AutoTrafficMode string

const (
	SpeedLimits AutoTrafficMode = "speed_limits"
	LiveTraffic AutoTrafficMode = "live_traffic"
)

// EtaRequest handles requests for estimated time of arrival between two points
type EtaRequest struct {
	FromLat         float64         `json:"fromlat,string"`
	FromLon         float64         `json:"fromlon,string"`
	ToLat           float64         `json:"tolat,string"`
	ToLon           float64         `json:"tolon,string"`
	Mode            string          `json:"mode"`
	DepartAt        string          `json:"departure"`
	AutoTrafficMode AutoTrafficMode `json:"auto_traffic_mode"`
	Date            string          `json:"date"`
	Time            string          `json:"time"`
	WithCursor
}

func (r *EtaRequest) SetDateTime(departAt time.Time) {
	r.Date = departAt.Format("2006-01-02")
	r.Time = departAt.Format("15:04:05")
}

func (r *EtaRequest) Query(ctx context.Context) (string, map[string]interface{}) {
	vars := hw{}
	vars["from"] = hw{"lat": r.FromLat, "lon": r.FromLon}
	vars["to"] = hw{"lat": r.ToLat, "lon": r.ToLon}

	// Parse departure time or use current time
	departAt := time.Now()
	if r.DepartAt != "" {
		if t, err := time.Parse(time.RFC3339, r.DepartAt); err == nil {
			departAt = t
		}
	}
	r.SetDateTime(departAt)

	// Handle mode selection
	modes := []model.StepMode{model.StepModeAuto, model.StepModeBicycle, model.StepModeWalk, model.StepModeTransit}
	if r.Mode != "" {
		switch r.Mode {
		case "auto":
			modes = []model.StepMode{model.StepModeAuto}
		case "bicycle":
			modes = []model.StepMode{model.StepModeBicycle}
		case "walk":
			modes = []model.StepMode{model.StepModeWalk}
		case "transit":
			modes = []model.StepMode{model.StepModeTransit}
		}
	}

	// Set default AutoTrafficMode if not provided
	if r.AutoTrafficMode == "" {
		r.AutoTrafficMode = SpeedLimits
	}

	// Build response
	response := make(map[string]interface{})
	for _, mode := range modes {
		req := model.DirectionRequest{
			From:     &model.WaypointInput{Lat: r.FromLat, Lon: r.FromLon},
			To:       &model.WaypointInput{Lat: r.ToLat, Lon: r.ToLon},
			Mode:     mode,
			DepartAt: &departAt,
		}

		var dir *model.Directions
		var err error
		var usedRouter string

		// Select appropriate router based on mode and traffic settings
		switch mode {
		case model.StepModeAuto:
			if r.AutoTrafficMode == LiveTraffic {
				if router := directions.GetRouter("aws"); router != nil {
					dir, err = router.Request(req)
					usedRouter = "aws"
				}
			} else {
				if router := directions.GetRouter("valhalla"); router != nil {
					dir, err = router.Request(req)
					usedRouter = "valhalla"
				}
			}
		case model.StepModeTransit:
			if router := directions.GetRouter("transitland"); router != nil {
				dir, err = router.Request(req)
				usedRouter = "transitland"
			}
		default:
			if router := directions.GetRouter("valhalla"); router != nil {
				dir, err = router.Request(req)
				usedRouter = "valhalla"
			}
		}

		if err == nil && dir != nil {
			result := map[string]interface{}{
				"duration": dir.Duration,
				"distance": dir.Distance,
				"router":   usedRouter,
			}

			// Add mode-specific details
			if mode == model.StepModeAuto {
				result["auto_traffic_mode"] = r.AutoTrafficMode
			} else if mode == model.StepModeTransit && dir.Legs != nil {
				// Add transit-specific details
				var transitDetails struct {
					NumTransfers int      `json:"num_transfers"`
					Modes       []string  `json:"modes"`
					WalkTime    *float64  `json:"walk_time,omitempty"`
					WaitTime    *float64  `json:"wait_time,omitempty"`
					StartTime   time.Time `json:"start_time"`
					EndTime     time.Time `json:"end_time"`
				}

				transitDetails.StartTime = dir.StartTime
				transitDetails.EndTime = dir.EndTime

				// Calculate number of transfers and collect unique modes
				modesMap := make(map[string]bool)
				var transfers int
				for i, leg := range dir.Legs {
					if i > 0 && leg.Mode == model.StepModeTransit {
						transfers++
					}
					modesMap[string(leg.Mode)] = true
				}
				transitDetails.NumTransfers = transfers

				// Convert modes map to slice
				for mode := range modesMap {
					transitDetails.Modes = append(transitDetails.Modes, mode)
				}

				result["transit_details"] = transitDetails
			}

			response[string(mode)] = result
		}
	}

	return "", response
}

func (r *EtaRequest) ResponseKey() string {
	return "eta"
} 