package rest

import (
	_ "embed"
	"strconv"
	"strings"
)

//go:embed agency_request.gql
var agencyQuery string

// AgencyRequest holds options for a Route request
type AgencyRequest struct {
	ID              int       `json:"id,string"`
	AgencyKey       string    `json:"agency_key"`
	AgencyID        string    `json:"agency_id"`
	AgencyName      string    `json:"agency_name"`
	OnestopID       string    `json:"onestop_id"`
	FeedVersionSHA1 string    `json:"feed_version_sha1"`
	FeedOnestopID   string    `json:"feed_onestop_id"`
	Search          string    `json:"search"`
	Lon             float64   `json:"lon,string"`
	Lat             float64   `json:"lat,string"`
	Bbox            *restBbox `json:"bbox"`
	Radius          float64   `json:"radius,string"`
	Adm0Name        string    `json:"adm0_name"`
	Adm0Iso         string    `json:"adm0_iso"`
	Adm1Name        string    `json:"adm1_name"`
	Adm1Iso         string    `json:"adm1_iso"`
	CityName        string    `json:"city_name"`
	IncludeAlerts   bool      `json:"include_alerts,string"`
	IncludeRoutes   bool      `json:"include_routes,string"`
	LicenseFilter
	WithCursor
}

// ResponseKey returns the GraphQL response entity key.
func (r AgencyRequest) ResponseKey() string { return "agencies" }

// Query returns a GraphQL query string and variables.
func (r AgencyRequest) Query() (string, map[string]interface{}) {
	if r.AgencyKey == "" {
		// pass
	} else if fsid, eid, ok := strings.Cut(r.AgencyKey, ":"); ok {
		r.FeedOnestopID = fsid
		r.AgencyID = eid
		r.IncludeRoutes = true
	} else if v, err := strconv.Atoi(r.AgencyKey); err == nil {
		r.ID = v
		r.IncludeRoutes = true
	} else {
		r.OnestopID = r.AgencyKey
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
	if r.AgencyID != "" {
		where["agency_id"] = r.AgencyID
	}
	if r.AgencyName != "" {
		where["agency_name"] = r.AgencyName
	}
	if r.Search != "" {
		where["search"] = r.Search
	}
	if r.Lat != 0.0 && r.Lon != 0.0 {
		where["near"] = hw{"lat": r.Lat, "lon": r.Lon, "radius": r.Radius}
	}
	if r.Bbox != nil {
		where["bbox"] = r.Bbox.AsJson()
	}
	if r.Adm0Name != "" {
		where["adm0_name"] = r.Adm0Name
	}
	if r.Adm1Name != "" {
		where["adm1_name"] = r.Adm1Name
	}
	if r.Adm0Iso != "" {
		where["adm0_iso"] = r.Adm0Iso
	}
	if r.Adm1Iso != "" {
		where["adm1_iso"] = r.Adm1Iso
	}
	if r.CityName != "" {
		where["city_name"] = r.CityName
	}
	where["license"] = checkLicenseFilter(r.LicenseFilter)
	return agencyQuery, hw{
		"limit":          r.CheckLimit(),
		"after":          r.CheckAfter(),
		"ids":            checkIds(r.ID),
		"include_alerts": r.IncludeAlerts,
		"include_routes": r.IncludeRoutes,
		"where":          where,
	}
}
