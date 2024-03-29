package rest

import (
	_ "embed"
	"strconv"
	"strings"
)

//go:embed route_request.gql
var routeQuery string

// RouteRequest holds options for a Route request
type RouteRequest struct {
	ID                int       `json:"id,string"`
	RouteKey          string    `json:"route_key"`
	AgencyKey         string    `json:"agency_key"`
	RouteID           string    `json:"route_id"`
	RouteType         string    `json:"route_type"`
	OnestopID         string    `json:"onestop_id"`
	OperatorOnestopID string    `json:"operator_onestop_id"`
	Format            string    `json:"format"`
	Search            string    `json:"search"`
	AgencyID          int       `json:"agency_id,string"`
	FeedVersionSHA1   string    `json:"feed_version_sha1"`
	FeedOnestopID     string    `json:"feed_onestop_id"`
	Lon               float64   `json:"lon,string"`
	Lat               float64   `json:"lat,string"`
	Radius            float64   `json:"radius,string"`
	Bbox              *restBbox `json:"bbox"`
	IncludeGeometry   bool      `json:"include_geometry,string"`
	IncludeAlerts     bool      `json:"include_alerts,string"`
	IncludeStops      bool      `json:"include_stops,string"`
	LicenseFilter
	WithCursor
}

// ResponseKey returns the GraphQL response entity key.
func (r RouteRequest) ResponseKey() string { return "routes" }

// Query returns a GraphQL query string and variables.
func (r RouteRequest) Query() (string, map[string]interface{}) {
	// These formats will need geometries included
	if r.ID > 0 || r.Format == "geojson" || r.Format == "geojsonl" || r.Format == "png" {
		r.IncludeGeometry = true
	}

	// Handle operator key
	if r.AgencyKey == "" {
		// pass
	} else if v, err := strconv.Atoi(r.AgencyKey); err == nil {
		r.AgencyID = v
	} else {
		r.OperatorOnestopID = r.AgencyKey
	}
	// Handle route key
	if r.RouteKey == "" {
		// pass
	} else if fsid, eid, ok := strings.Cut(r.RouteKey, ":"); ok {
		r.FeedOnestopID = fsid
		r.RouteID = eid
		r.IncludeGeometry = true
		r.IncludeStops = true
	} else if v, err := strconv.Atoi(r.RouteKey); err == nil {
		r.ID = v
		r.IncludeGeometry = true
		r.IncludeStops = true
	} else {
		r.OnestopID = r.RouteKey
		r.IncludeGeometry = true
		r.IncludeStops = true
	}

	where := hw{}
	if r.FeedVersionSHA1 != "" {
		where["feed_version_sha1"] = r.FeedVersionSHA1
	}
	if r.FeedOnestopID != "" {
		where["feed_onestop_id"] = r.FeedOnestopID
	}
	if r.RouteID != "" {
		where["route_id"] = r.RouteID
	}
	if r.RouteType != "" {
		where["route_type"] = r.RouteType
	}
	if r.OnestopID != "" {
		where["onestop_id"] = r.OnestopID
	}
	if r.OperatorOnestopID != "" {
		where["operator_onestop_id"] = r.OperatorOnestopID
	}
	if r.AgencyID > 0 {
		where["agency_ids"] = []int{r.AgencyID}
	}
	if r.Lat != 0.0 && r.Lon != 0.0 {
		where["near"] = hw{"lat": r.Lat, "lon": r.Lon, "radius": r.Radius}
	}
	if r.Bbox != nil {
		where["bbox"] = r.Bbox.AsJson()
	}
	if r.Search != "" {
		where["search"] = r.Search
	}
	where["license"] = checkLicenseFilter(r.LicenseFilter)
	return routeQuery, hw{
		"limit":            r.CheckLimit(),
		"after":            r.CheckAfter(),
		"ids":              checkIds(r.ID),
		"where":            where,
		"include_alerts":   r.IncludeAlerts,
		"include_geometry": r.IncludeGeometry,
		"include_stops":    r.IncludeStops,
	}
}
