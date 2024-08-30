package rest

import (
	_ "embed"
	"strconv"

	oa "github.com/getkin/kin-openapi/openapi3"
)

//go:embed operator_request.gql
var operatorQuery string

// OperatorRequest holds options for an Operator request
type OperatorRequest struct {
	OperatorKey   string    `json:"operator_key"`
	ID            int       `json:"id,string"`
	OnestopID     string    `json:"onestop_id"`
	FeedOnestopID string    `json:"feed_onestop_id"`
	Search        string    `json:"search"`
	TagKey        string    `json:"tag_key"`
	TagValue      string    `json:"tag_value"`
	Lon           float64   `json:"lon,string"`
	Lat           float64   `json:"lat,string"`
	Bbox          *restBbox `json:"bbox"`
	Radius        float64   `json:"radius,string"`
	Adm0Name      string    `json:"adm0_name"`
	Adm0Iso       string    `json:"adm0_iso"`
	Adm1Name      string    `json:"adm1_name"`
	Adm1Iso       string    `json:"adm1_iso"`
	CityName      string    `json:"city_name"`
	IncludeAlerts bool      `json:"include_alerts,string"`
	LicenseFilter
	WithCursor
}

func (r OperatorRequest) RequestInfo() RequestInfo {
	return RequestInfo{
		Path: "/operators",
		PathItem: &oa.PathItem{
			Extensions: map[string]any{
				"x-alternates": []any{map[string]any{"description": "Request operators in specified format", "method": "get", "path": "/operators.{format}"}, map[string]any{"description": "Request an operator by Onestop ID", "method": "get", "path": "/operators/{onestop_id}"}},
			},
			Get: &oa.Operation{
				Summary:     "Operators",
				Description: `Search for operators`,
				Responses:   queryToResponses(operatorQuery),
				Parameters: oa.Parameters{
					&pref{
						Ref: "#/components/parameters/idParam",
					},
					&pref{
						Ref: "#/components/parameters/afterParam",
					},
					&pref{
						Ref: "#/components/parameters/limitParam",
					},
					&pref{
						Ref: "#/components/parameters/searchParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "search=bart", "url": "/operators?search=caltrain"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/includeAlertsParam",
					},
					&pref{
						Ref: "#/components/parameters/onestopParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "onestop_id=o-9q9-caltrain", "url": "/operators?onestop_id=o-9q9-caltrain"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/feedParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/operators?feed_onestop_id=f-sf~bay~area~rg"}},
						},
					},
					&pref{
						Value: &param{
							Name:        "tag_key",
							In:          "query",
							Description: `Search for operators with a tag. Combine with tag_value also query for the value of the tag.`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "tag_key=us_ntd_id", "url": "/operators?tag_key=us_ntd_id"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "tag_value",
							In:          "query",
							Description: `Search for feeds tagged with a given value. Must be combined with tag_key.`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "tag_key=us_ntd_id&tag_value=40029", "url": "/operators?tag_key=us_ntd_id&tag_value=40029"}},
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm0NameParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm0_name=Mexico", "url": "/operators?adm0_name=Mexico"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm0IsoParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm0_iso=US", "url": "/operators?adm0_iso=US"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm1NameParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm1_name=California", "url": "/operators?adm1_name=California"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/adm1IsoParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "adm1_iso=US-CA", "url": "/operators?adm1_iso=US-CA"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/cityNameParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "city_name=Oakland", "url": "/operators?city_name=Oakland"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/radiusParam",
						Extensions: map[string]any{
							"x-description":      "Search for operators geographically, based on stops at this location; radius is in meters, requires lon and lat",
							"x-example-requests": []any{map[string]any{"description": "lon=-122&lat=37&radius=1000", "url": "/operators?lon=-122.3&lat=37.8&radius=1000"}},
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
							"x-example-requests": []any{map[string]any{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/operators?bbox=-122.269,37.807,-122.267,37.808"}},
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
func (r OperatorRequest) ResponseKey() string { return "operators" }

// Query returns a GraphQL query string and variables.
func (r OperatorRequest) Query() (string, map[string]interface{}) {
	if r.OperatorKey == "" {
		// pass
	} else if v, err := strconv.Atoi(r.OperatorKey); err == nil {
		r.ID = v
	} else {
		r.OnestopID = r.OperatorKey
	}
	where := hw{}
	where["merged"] = true
	if r.FeedOnestopID != "" {
		where["feed_onestop_id"] = r.FeedOnestopID
	}
	if r.OnestopID != "" {
		where["onestop_id"] = r.OnestopID
	}
	if r.Search != "" {
		where["search"] = r.Search
	}
	if r.TagKey != "" {
		where["tags"] = hw{r.TagKey: r.TagValue}
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
	return operatorQuery, hw{
		"limit":          r.CheckLimit(),
		"after":          r.CheckAfter(),
		"ids":            checkIds(r.ID),
		"include_alerts": r.IncludeAlerts,
		"where":          where,
	}
}
