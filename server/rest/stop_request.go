package rest

import (
	_ "embed"
	"strconv"
	"strings"

	oa "github.com/getkin/kin-openapi/openapi3"
)

//go:embed stop_request.gql
var stopQuery string

// StopRequest holds options for a /stops request
type StopRequest struct {
	ID                 int       `json:"id,string"`
	StopKey            string    `json:"stop_key"`
	StopID             string    `json:"stop_id"`
	OnestopID          string    `json:"onestop_id"`
	FeedVersionSHA1    string    `json:"feed_version_sha1"`
	FeedOnestopID      string    `json:"feed_onestop_id"`
	Search             string    `json:"search"`
	Bbox               *restBbox `json:"bbox"`
	Lon                float64   `json:"lon,string"`
	Lat                float64   `json:"lat,string"`
	Radius             float64   `json:"radius,string"`
	Format             string    `json:"format"`
	ServedByOnestopIds string    `json:"served_by_onestop_ids"`
	ServedByRouteType  *int      `json:"served_by_route_type,string"`
	IncludeAlerts      bool      `json:"include_alerts,string"`
	IncludeRoutes      bool      `json:"include_routes,string"`
	LicenseFilter
	WithCursor
}

func (r StopRequest) RequestInfo() RequestInfo {
	return RequestInfo{
		Path: "/stops",
		PathItem: &oa.PathItem{
			Extensions: map[string]any{
				"x-alternates": []any{map[string]any{"description": "Request stops in specified format", "method": "get", "path": "/stops.{format}"}, map[string]any{"description": "Request a stop", "method": "get", "path": "/stops/{stop_key}"}, map[string]any{"description": "Request a stop in a specified format", "method": "get", "path": "/stops/{stop_key}.{format}"}},
			},
			Get: &oa.Operation{
				Summary:     "Stops",
				Description: `Search for stops`,
				Responses:   queryToResponses(stopQuery),
				Parameters: oa.Parameters{
					&pref{
						Ref: "#/components/parameters/includeAlertsParam",
					},
					&pref{
						Ref: "#/components/parameters/idParam",
					},
					&pref{
						Value: &param{
							Name:        "stop_key",
							In:          "query",
							Description: `Stop lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs stop_id>' key, or a Onestop ID`,
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
							"x-example-requests": []any{map[string]any{"description": "limit=1", "url": "/stops?limit=1"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/formatParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "format=geojson", "url": "/stops?format=geojson"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/searchParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "search=embarcadero", "url": "/stops?search=embarcadero"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/onestopParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "onestop_id=...", "url": "/stops?onestop_id=s-9q8yyzcny3-embarcadero"}},
						},
					},
					&pref{
						Value: &param{
							Name:        "stop_id",
							In:          "query",
							Description: `Search for records with this GTFS stop_id`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "stop_id=EMBR", "url": "/stops?feed_onestop_id=f-c20-trimet&stop_id=1108"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "served_by_onestop_ids",
							In:          "query",
							Description: `Search stops visited by a route or agency OnestopID. Accepts comma separated values.`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "served_by_onestop_ids=o-9q9-bart,o-9q9-caltrain", "url": "/stops?served_by_onestop_ids=o-9q9-bart,o-9q9-caltrain"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "served_by_route_type",
							In:          "query",
							Description: `Search for stops served by a particular route (vehicle) type`,
							Schema: &sref{
								Value: newSchema("integer", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "served_by_route_type=1", "url": "/stops?served_by_route_type=1"}},
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/sha1Param",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "feed_version_sha1=1c4721d4...", "url": "/stops?feed_version_sha1=1c4721d4e0c9fae1e81f7c79660696e4280ed05b"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/feedParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "feed_onestop_id=f-c20-trimet", "url": "/stops?feed_onestop_id=f-c20-trimet"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/radiusParam",
						Extensions: map[string]any{
							"x-description":      "Search for stops geographically; radius is in meters, requires lon and lat",
							"x-example-requests": []any{map[string]any{"description": "lon=-122&lat=37&radius=1000", "url": "/stops?lon=-122.3&lat=37.8&radius=1000"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/lonParam",
					},
					&pref{
						Ref: "#/components/parameters/latParam",
					},
					&pref{
						Ref: "#/components/parameters/bboxParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/stops?bbox=-122.269,37.807,-122.267,37.808"}},
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
func (r StopRequest) ResponseKey() string { return "stops" }

// Query returns a GraphQL query string and variables.
func (r StopRequest) Query() (string, map[string]any) {
	if r.StopKey == "" {
		// pass
	} else if fsid, eid, ok := strings.Cut(r.StopKey, ":"); ok {
		r.FeedOnestopID = fsid
		r.StopID = eid
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
	if r.Bbox != nil {
		where["bbox"] = r.Bbox.AsJson()
	}
	if r.Search != "" {
		where["search"] = r.Search
	}
	if r.ServedByOnestopIds != "" {
		where["served_by_onestop_ids"] = commaSplit(r.ServedByOnestopIds)
	}
	if r.ServedByRouteType != nil {
		where["served_by_route_type"] = *r.ServedByRouteType
	}
	where["license"] = checkLicenseFilter(r.LicenseFilter)
	return stopQuery, hw{
		"limit":          r.CheckLimit(),
		"after":          r.CheckAfter(),
		"ids":            checkIds(r.ID),
		"include_alerts": r.IncludeAlerts,
		"include_routes": r.IncludeRoutes,
		"where":          where,
	}
}
