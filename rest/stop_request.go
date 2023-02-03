package rest

import (
	_ "embed"
	"strconv"
	"strings"
)

//go:embed stop_request.gql
var stopQuery string

// StopRequest holds options for a /stops request
type StopRequest struct {
	ID                 int     `json:"id,string"`
	Limit              int     `json:"limit,string"`
	StopKey            string  `json:"stop_key"`
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
	IncludeAlerts      bool    `json:"include_alerts,string"`
	IncludeRoutes      bool    `json:"include_routes,string"`
	LicenseFilter
}

// ResponseKey returns the GraphQL response entity key.
func (r StopRequest) ResponseKey() string { return "stops" }

// Query returns a GraphQL query string and variables.
func (r StopRequest) Query() (string, map[string]interface{}) {
	if r.StopKey == "" {
		// pass
	} else if key := strings.SplitN(r.StopKey, ":", 2); len(key) == 2 {
		r.FeedOnestopID = key[0]
		r.StopID = key[1]
		r.IncludeRoutes = true
	} else if v, err := strconv.Atoi(r.StopKey); err == nil {
		r.ID = v
		r.IncludeRoutes = true
	} else {
		r.OnestopID = r.StopKey
		r.IncludeRoutes = true
	}

	where := hw{}
	if r.FeedVersionSHA1 != "" {
		where["feed_version_sha1"] = r.FeedVersionSHA1
	}
	if r.FeedOnestopID != "" {
		where["feed_onestop_id"] = r.FeedOnestopID
	}
	if r.OnestopID != "" {
		where["onestop_id"] = r.OnestopID
	}
	if r.StopID != "" {
		where["stop_id"] = r.StopID
	}
	if r.Lat != 0.0 && r.Lon != 0.0 {
		where["near"] = hw{"lat": r.Lat, "lon": r.Lon, "radius": r.Radius}
	}
	if r.Search != "" {
		where["search"] = r.Search
	}
	if r.ServedByOnestopIds != "" {
		where["served_by_onestop_ids"] = commaSplit(r.ServedByOnestopIds)
	}
	where["license"] = checkLicenseFilter(r.LicenseFilter)
	return stopQuery, hw{
		"limit":          checkLimit(r.Limit),
		"after":          checkAfter(r.After),
		"ids":            checkIds(r.ID),
		"include_alerts": r.IncludeAlerts,
		"include_routes": r.IncludeRoutes,
		"where":          where,
	}
}
