package rest

import (
	_ "embed"
	"strconv"
	"strings"

	oa "github.com/getkin/kin-openapi/openapi3"
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

func (r RouteRequest) RequestInfo() RequestInfo {
	return RequestInfo{
		Path: "/routes",
		PathItem: &oa.PathItem{
			Extensions: map[string]any{
				"x-alternates": []any{map[string]any{"description": "Request routes in specified format", "method": "get", "path": "/routes.{format}"}, map[string]any{"description": "Request a route", "method": "get", "path": "/routes/{route_key}"}, map[string]any{"description": "Request a route in a specified format", "method": "get", "path": "/routes/{route_key}.{format}"}},
			},
			Get: &oa.Operation{
				Summary:     "Routes",
				Description: `Search for routes`,
				Responses:   queryToResponses(routeQuery),
				Parameters: oa.Parameters{
					&pref{
						Ref: "#/components/parameters/idParam",
					},
					&pref{
						Value: &param{
							Name:        "route_key",
							In:          "query",
							Description: `Route lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs route_id>' key, or a Onestop ID`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/afterParam",
					},
					&pref{
						Ref: "#/components/parameters/limitParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "limit=1", "url": "/routes?limit=1"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/formatParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "format=png", "url": "/routes?format=png&feed_onestop_id=f-dr5r7-nycdotsiferry"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/includeAlertsParam",
					},
					&pref{
						Ref: "#/components/parameters/searchParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "search=daly+city", "url": "/routes?search=daly+city"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/onestopParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "onestop_id=r-9q9j-l1", "url": "/routes?onestop_id=r-9q9j-l1"}},
						},
					},
					&pref{
						Value: &param{
							Name:        "route_id",
							In:          "query",
							Description: `Search for records with this GTFS route_id`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "route_id=Bu-130", "url": "/routes?feed_onestop_id=f-sf~bay~area~rg&route_id=AC:10"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "route_type",
							In:          "query",
							Description: `Search for routes with this GTFS route (vehicle) type`,
							Schema: &sref{
								Value: newSchema("integer", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "route_type=1", "url": "/routes?route_type=1"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "operator_onestop_id",
							In:          "query",
							Description: `Search for records by operator OnestopID`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "operator_onestop_id=...", "url": "/routes?operator_onestop_id=o-9q9-caltrain"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "include_geometry",
							In:          "query",
							Description: `Include route geometry`,
							Schema: &sref{
								Value: newSchema("string", "", []any{"true", "false"}),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "include_geometry=true", "url": "/routes?include_geometry=true"}},
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/sha1Param",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "feed_version_sha1=041ffeec...", "url": "/routes?feed_version_sha1=041ffeec98316e560bc2b91960f7150ad329bd5f"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/feedParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/routes?feed_onestop_id=f-sf~bay~area~rg"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/radiusParam",
						Extensions: map[string]any{
							"x-description":      "Search for routes geographically, based on stops at this location; radius is in meters, requires lon and lat",
							"x-example-requests": []any{map[string]any{"description": "lon=-122&lat=37&radius=1000", "url": "/routes?lon=-122.3&lat=37.8&radius=1000"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/latParam",
					},
					&pref{
						Ref: "#/components/parameters/lonParam",
					},
					&pref{
						Ref: "#/components/parameters/bboxParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/routes?bbox=-122.269,37.807,-122.267,37.808"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/licenseCommercialUseAllowedParam",
					},
					&pref{
						Ref: "#/components/parameters/licenseShareAlikeOptionalParam",
					},
					&pref{
						Ref: "#/components/parameters/licenseCreateDerivedProductParam",
					},
					&pref{
						Ref: "#/components/parameters/licenseRedistributionAllowedParam",
					},
					&pref{
						Ref: "#/components/parameters/licenseUseWithoutAttributionParam",
					},
				},
			},
		},
	}
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
