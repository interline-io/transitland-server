package rest

import (
	_ "embed"
	"strconv"
)

//go:embed stop_departure_request.gql
var stopDepartureQuery string

// StopDepartureRequest holds options for a /stops request
type StopDepartureRequest struct {
	StopKey         string `json:"stop_key"`
	ID              int    `json:"id,string"`
	Limit           int    `json:"limit,string"`
	After           int    `json:"after,string"`
	StopID          string `json:"stop_id"`
	OnestopID       string `json:"onestop_id"`
	Next            int    `json:"next"`
	ServiceDate     string `json:"service_date"`
	StartTime       int    `json:"start_time"`
	EndTime         int    `json:"end_time"`
	IncludeGeometry bool   `json:"include_geometry,string"`
}

// ResponseKey returns the GraphQL response entity key.
func (r StopDepartureRequest) ResponseKey() string { return "stops" }

// Query returns a GraphQL query string and variables.
func (r StopDepartureRequest) Query() (string, map[string]interface{}) {
	if r.StopKey == "" {
		// pass
	} else if v, err := strconv.Atoi(r.StopKey); err == nil {
		r.ID = v
	} else {
		r.OnestopID = r.StopKey
	}
	where := hw{}
	if r.OnestopID != "" {
		where["onestop_id"] = r.OnestopID
	}
	stwhere := hw{
		"use_service_window": true,
	}
	if r.ServiceDate != "" {
		stwhere["service_date"] = r.ServiceDate
		stwhere["start_time"] = r.StartTime
		if r.EndTime > 0 {
			stwhere["end_time"] = r.EndTime
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
		"after":            checkAfter(r.After),
		"ids":              checkIds(r.ID),
		"where":            where,
		"stop_time_where":  stwhere,
	}
}
