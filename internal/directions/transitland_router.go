package directions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/model"
)

func init() {
	endpoint := os.Getenv("TL_TRANSITLAND_ENDPOINT")
	apikey := os.Getenv("TL_TRANSITLAND_API_KEY")
	if endpoint == "" {
		return
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	if err := RegisterRouter("transitland", func() Handler {
		return newTransitlandRouter(client, endpoint, apikey)
	}); err != nil {
		panic(err)
	}
}

type transitlandRouter struct {
	Clock    clock.Clock
	client   *http.Client
	endpoint string
	apikey   string
}

func newTransitlandRouter(client *http.Client, endpoint string, apikey string) *transitlandRouter {
	if client == nil {
		client = http.DefaultClient
	}
	return &transitlandRouter{
		client:   client,
		endpoint: endpoint,
		apikey:   apikey,
	}
}

func (h *transitlandRouter) Request(req model.DirectionRequest) (*model.Directions, error) {
	if err := validateDirectionRequest(req); err != nil {
		return &model.Directions{Success: false, Exception: aws.String("invalid input")}, nil
	}

	// Prepare request
	params := url.Values{}
	params.Set("fromPlace", fmt.Sprintf("%f,%f", req.From.Lat, req.From.Lon))
	params.Set("toPlace", fmt.Sprintf("%f,%f", req.To.Lat, req.To.Lon))
	
	departAt := time.Now()
	if req.DepartAt != nil {
		departAt = *req.DepartAt
	}
	params.Set("date", departAt.Format("2006-01-02"))
	params.Set("time", departAt.Format("15:04:05"))

	// Set mode
	switch req.Mode {
	case model.StepModeAuto:
		params.Set("mode", "CAR")
	case model.StepModeBicycle:
		params.Set("mode", "BICYCLE")
	case model.StepModeWalk:
		params.Set("mode", "WALK")
	default:
		params.Set("mode", "TRANSIT,WALK")
	}

	// Make request
	url := fmt.Sprintf("%s/api/v2/routing/otp/plan?%s", h.endpoint, params.Encode())
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apikey))

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var tlResp transitlandResponse
	if err := json.Unmarshal(body, &tlResp); err != nil {
		return nil, err
	}

	// Convert to model.Directions
	return convertTransitlandResponse(&tlResp), nil
}

type transitlandResponse struct {
	Plan struct {
		Itineraries []struct {
			Duration     float64 `json:"duration"`
			StartTime    int64   `json:"startTime"`
			EndTime      int64   `json:"endTime"`
			WalkTime     float64 `json:"walkTime"`
			TransitTime  float64 `json:"transitTime"`
			WaitingTime  float64 `json:"waitingTime"`
			WalkDistance float64 `json:"walkDistance"`
			Transfers    int     `json:"transfers"`
			Legs         []struct {
				StartTime   int64   `json:"startTime"`
				EndTime     int64   `json:"endTime"`
				Mode        string  `json:"mode"`
				Duration    float64 `json:"duration"`
				Distance    float64 `json:"distance"`
				From        transitlandStop `json:"from"`
				To          transitlandStop `json:"to"`
				LegGeometry struct {
					Points string `json:"points"`
				} `json:"legGeometry"`
			} `json:"legs"`
		} `json:"itineraries"`
	} `json:"plan"`
}

type transitlandStop struct {
	Name      string  `json:"name"`
	StopId    string  `json:"stopId"`
	StopCode  string  `json:"stopCode"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Departure int64   `json:"departure"`
}

func convertTransitlandResponse(tlResp *transitlandResponse) *model.Directions {
	if len(tlResp.Plan.Itineraries) == 0 {
		return &model.Directions{Success: false}
	}

	bestItinerary := tlResp.Plan.Itineraries[0]
	legs := make([]*model.DirectionsLeg, len(bestItinerary.Legs))

	for i, leg := range bestItinerary.Legs {
		legs[i] = &model.DirectionsLeg{
			Distance: &model.Distance{
				Distance: leg.Distance,
				Units:    model.DistanceUnitMeters,
			},
			Duration: &model.Duration{
				Duration: leg.Duration,
				Units:    model.DurationUnitSeconds,
			},
			StartTime: time.Unix(leg.StartTime/1000, 0),
			EndTime:   time.Unix(leg.EndTime/1000, 0),
			Mode:      convertMode(leg.Mode),
			From: &model.Place{
				Name:      leg.From.Name,
				Longitude: leg.From.Lon,
				Latitude:  leg.From.Lat,
			},
			To: &model.Place{
				Name:      leg.To.Name,
				Longitude: leg.To.Lon,
				Latitude:  leg.To.Lat,
			},
			Geometry: leg.LegGeometry.Points,
		}
	}

	return &model.Directions{
		Success: true,
		Duration: &model.Duration{
			Duration: bestItinerary.Duration,
			Units:    model.DurationUnitSeconds,
		},
		Distance: &model.Distance{
			Distance: bestItinerary.WalkDistance,
			Units:    model.DistanceUnitMeters,
		},
		StartTime: time.Unix(bestItinerary.StartTime/1000, 0),
		EndTime:   time.Unix(bestItinerary.EndTime/1000, 0),
		Legs:      legs,
	}
}

func convertMode(mode string) model.StepMode {
	switch mode {
	case "WALK":
		return model.StepModeWalk
	case "BICYCLE":
		return model.StepModeBicycle
	case "CAR":
		return model.StepModeAuto
	case "BUS", "RAIL", "SUBWAY", "TRAM":
		return model.StepModeTransit
	default:
		return model.StepModeUnknown
	}
} 