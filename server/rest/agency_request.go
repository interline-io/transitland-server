package rest

import (
	_ "embed"
	"strconv"
	"strings"

	oa "github.com/getkin/kin-openapi/openapi3"
)

//go:embed agency_request.gql
var agencyQuery string

// AgencyRequest holds options for an Agency request
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

func (r AgencyRequest) RequestInfo() RequestInfo {
	return RequestInfo{
		Path: "/agencies",
		PathItem: &oa.PathItem{
			Extensions: map[string]any{
				"x-alternates": []any{map[string]any{"description": "Request agencies in specified format", "method": "get", "path": "/agencies.{format}"}, map[string]any{"description": "Request an agency", "method": "get", "path": "/agencies/{agency_key}"}, map[string]any{"description": "Request an agency in specified format", "method": "get", "path": "/agencies/{agency_key}.{format}"}},
			},
			Get: &oa.Operation{
				Summary:     "Agencies",
				Description: ``,
				Responses:   queryToResponses(agencyQuery),
				Parameters: oa.Parameters{
					&pref{
						Ref: "#/components/parameters/idParam",
					},
					&pref{
						Value: &param{
							Name:        "agency_key",
							In:          "query",
							Description: `Agency lookup key; can be an integer ID, a '<feed onestop_id>:<gtfs agency_id>' key, or a Onestop ID`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/includeAlertsParam",
					},
					&pref{
						Ref: "#/components/parameters/afterParam",
					},
					&pref{
						Ref: "#/components/parameters/limitParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "limit=1", "url": "/agencies?limit=1"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/formatParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "format=geojson", "url": "/agencies?format=geojson"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/searchParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "search=bart", "url": "/agencies?search=bart"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/onestopParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "onestop_id=o-9q9-caltrain", "url": "/agencies?onestop_id=o-9q9-caltrain"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/sha1Param",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "feed_version_sha1=1c4721d4...", "url": "/agencies?feed_version_sha1=1c4721d4e0c9fae1e81f7c79660696e4280ed05b"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/feedParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/agencies?feed_onestop_id=f-sf~bay~area~rg"}},
						},
					},
					&pref{
						Value: &param{
							Name:        "agency_id",
							In:          "query",
							Description: `Search for records with this GTFS agency_id (string)`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "agency_id=BART", "url": "/agencies?agency_id=BART"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "agency_name",
							In:          "query",
							Description: `Search for records with this GTFS agency_name`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "agency_name=Caltrain", "url": "/agencies?agency_name=Caltrain"}},
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/radiusParam",
						Extensions: map[string]any{
							"x-description":      "Search for agencies geographically, based on stops at this location; radius is in meters, requires lon and lat",
							"x-example-requests": []any{map[string]any{"description": "lon=-122&lat=37&radius=1000", "url": "/agencies?lon=-122.3&lat=37.8&radius=1000"}},
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
							"x-example-requests": []any{map[string]any{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/agencies?bbox=-122.269,37.807,-122.267,37.808"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm0NameParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm0_name=Mexico", "url": "/agencies?adm0_name=Mexico"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm0IsoParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm0_iso=US", "url": "/agencies?adm0_iso=US"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm1NameParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm1_name=California", "url": "/agencies?adm1_name=California"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm1IsoParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm1_iso=US-CA", "url": "/agencies?adm1_iso=US-CA"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/cityNameParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "city_name=Oakland", "url": "/agencies?city_name=Oakland"}},
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
