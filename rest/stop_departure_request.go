package rest

import (
	_ "embed"
	"strconv"
)

//go:embed stop_departure_request.gql
var stopDepartureQuery string

// StopDepartureRequest holds options for a /stops/_/departures request
type StopDepartureRequest struct {
	StopKey          string `json:"stop_key"`
	ID               int    `json:"id,string"`
	Limit            int    `json:"limit,string"`
	OnestopID        string `json:"onestop_id"`
	Next             int    `json:"next"`
	ServiceDate      string `json:"service_date"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	IncludeGeometry  bool   `json:"include_geometry,string"`
	UseServiceWindow *bool  `json:"use_service_window,string"`
}

// ResponseKey returns the GraphQL response entity key.
func (r StopDepartureRequest) ResponseKey() string { return "stops" }

// IncludeNext
func (r StopDepartureRequest) IncludeNext() bool { return false }

// Query returns a GraphQL query string and variables.
func (r StopDepartureRequest) Query() (string, map[string]interface{}) {
	if r.StopKey == "" {
		// TODO: add a way to reject request as invalid
	} else if v, err := strconv.Atoi(r.StopKey); err == nil && v > 0 {
		// require an actual ID, not just 0
		r.ID = v
	} else {
		r.OnestopID = r.StopKey
	}
	where := hw{}
	if r.OnestopID != "" {
		where["onestop_id"] = r.OnestopID
	}
	//
	stwhere := hw{}
	if r.UseServiceWindow == nil || *r.UseServiceWindow {
		stwhere["use_service_window"] = true
	}
	if r.ServiceDate != "" {
		stwhere["service_date"] = r.ServiceDate
		stwhere["start"] = r.StartTime
		if r.EndTime != "" {
			stwhere["end"] = r.EndTime
		}
	} else {
		if r.Next == 0 {
			r.Next = 3600
		}
		stwhere["next"] = r.Next
	}
	return stopDepartureQuery, hw{
		"include_geometry": r.IncludeGeometry,
		"limit":            checkLimit(r.Limit),
		"ids":              checkIds(r.ID),
		"where":            where,
		"stop_time_where":  stwhere,
	}
}
