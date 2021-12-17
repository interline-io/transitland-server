package rest

import (
	_ "embed"
)

//go:embed stop_departure_request.gql
var stopDepartureQuery string

// StopDepartureRequest holds options for a /stops request
type StopDepartureRequest struct {
	StopKey            string  `json:"stop_key"`
	ID                 int     `json:"id,string"`
	Limit              int     `json:"limit,string"`
	After              int     `json:"after,string"`
	StopID             string  `json:"stop_id"`
	OnestopID          string  `json:"onestop_id"`
	FeedVersionSHA1    string  `json:"feed_version_sha1"`
	FeedOnestopID      string  `json:"feed_onestop_id"`
	Search             string  `json:"search"`
	Lon                float64 `json:"lon,string"`
	Lat                float64 `json:"lat,string"`
	Radius             float64 `json:"radius,string"`
	ServedByOnestopIds string  `json:"served_by_onestop_ids"`
}

// ResponseKey returns the GraphQL response entity key.
func (r StopDepartureRequest) ResponseKey() string { return "stops" }

// Query returns a GraphQL query string and variables.
func (r StopDepartureRequest) Query() (string, map[string]interface{}) {
	return stopDepartureQuery, hw{}
}
