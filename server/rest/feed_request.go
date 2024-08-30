package rest

import (
	_ "embed"
	"strconv"
	"strings"

	oa "github.com/getkin/kin-openapi/openapi3"
)

//go:embed feed_request.gql
var feedQuery string

// FeedRequest holds options for a Feed request
type FeedRequest struct {
	FeedKey          string    `json:"feed_key"`
	ID               int       `json:"id,string"`
	OnestopID        string    `json:"onestop_id"`
	Spec             string    `json:"spec"`
	Search           string    `json:"search"`
	FetchError       string    `json:"fetch_error"`
	TagKey           string    `json:"tag_key"`
	TagValue         string    `json:"tag_value"`
	URL              string    `json:"url"`
	URLType          string    `json:"url_type"`
	URLCaseSensitive bool      `json:"url_case_sensitive"`
	Lon              float64   `json:"lon,string"`
	Lat              float64   `json:"lat,string"`
	Radius           float64   `json:"radius,string"`
	Bbox             *restBbox `json:"bbox"`
	LicenseFilter
	WithCursor
}

func (r FeedRequest) RequestInfo() RequestInfo {
	return RequestInfo{
		Path: "/feeds",
		PathItem: &oa.PathItem{
			Extensions: map[string]any{
				"x-alternates": []any{map[string]any{"description": "Request feeds in specified format", "method": "get", "path": "/feeds.{format}"}, map[string]any{"description": "Request a feed by ID or Onestop ID", "method": "get", "path": "/feeds/{feed_key}"}, map[string]any{"description": "Request a feed by ID or Onestop ID in specified format", "method": "get", "path": "/feeds/{feed_key}.{format}"}},
			},
			Get: &oa.Operation{
				Summary:     "Feeds",
				Description: `Search for feeds`,
				Responses:   queryToResponses(feedQuery),
				Parameters: oa.Parameters{
					&pref{
						Ref: "#/components/parameters/idParam",
					},
					&pref{
						Value: &param{
							Name:        "feed_key",
							In:          "query",
							Description: `Feed lookup key; can be an integer ID or a Onestop ID`,
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
							"x-example-requests": []any{map[string]any{"description": "limit=1", "url": "/feeds?limit=1"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/formatParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "format=geojson", "url": "/feeds?format=geojson"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/searchParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "search=caltrain", "url": "/feeds?search=caltrain"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/onestopParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "onestop_id=f-sf~bay~area~rg", "url": "/feeds?onestop_id=f-sf~bay~area~rg"}},
						},
					},
					&pref{
						Value: &param{
							Name:        "spec",
							In:          "query",
							Description: `Type of data contained in this feed`,
							Schema: &sref{
								Value: newSchema("string", "", []any{"gtfs", "gtfs-rt", "gbfs", "mds"}),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "spec=gtfs", "url": "/feeds?spec=gtfs"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "fetch_error",
							In:          "query",
							Description: `Search for feeds with or without a fetch error`,
							Schema: &sref{
								Value: newSchema("string", "", []any{"true", "false"}),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "fetch_error=true", "url": "/feeds?fetch_error=true"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "tag_key",
							In:          "query",
							Description: `Search for feeds with a tag. Combine with tag_value also query for the value of the tag.`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "tag_key=gtfs_data_exchange", "url": "/feeds?tag_key=gtfs_data_exchange"}},
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
								"x-example-requests": []any{map[string]any{"description": "tag_key=unstable_url&tag_value=true", "url": "/feeds?tag_key=unstable_url&tag_value=true"}},
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/radiusParam",
						Extensions: map[string]any{
							"x-description":      "Search for feeds geographically; radius is in meters, requires lon and lat",
							"x-example-requests": []any{map[string]any{"description": "lon=-122&lat=37&radius=1000", "url": "/feeds?lon=-122.3?lat=37.8&radius=1000"}},
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
							"x-example-requests": []any{map[string]any{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/feeds?bbox=-122.269,37.807,-122.267,37.808"}},
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

// ResponseKey .
func (r FeedRequest) ResponseKey() string {
	return "feeds"
}

// Query returns a GraphQL query string and variables.
func (r FeedRequest) Query() (string, map[string]interface{}) {
	if r.FeedKey == "" {
		// pass
	} else if v, err := strconv.Atoi(r.FeedKey); err == nil {
		r.ID = v
	} else {
		r.OnestopID = r.FeedKey
	}
	where := hw{}
	if r.OnestopID != "" {
		where["onestop_id"] = r.OnestopID
	}
	if r.Spec != "" {
		where["spec"] = []string{checkFeedSpecFilterValue(r.Spec)}
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
	if r.TagKey != "" {
		where["tags"] = hw{r.TagKey: r.TagValue}
	}
	if r.FetchError == "true" {
		where["fetch_error"] = true
	} else if r.FetchError == "false" {
		where["fetch_error"] = false
	}
	if r.URL != "" || r.URLType != "" {
		sourceUrl := hw{"case_sensitive": r.URLCaseSensitive}
		if r.URL != "" {
			sourceUrl["url"] = r.URL
		}
		if r.URLType != "" {
			sourceUrl["type"] = r.URLType
		}
		where["source_url"] = sourceUrl
	}
	where["license"] = checkLicenseFilter(r.LicenseFilter)
	return feedQuery, hw{"limit": r.CheckLimit(), "after": r.CheckAfter(), "ids": checkIds(r.ID), "where": where}
}

// ProcessGeoJSON .
func (r FeedRequest) ProcessGeoJSON(response map[string]interface{}) error {
	// This is not ideal. Use gjson?
	entities, ok := response[r.ResponseKey()].([]interface{})
	if ok {
		for _, feature := range entities {
			if f2, ok := feature.(map[string]interface{}); ok {
				if f3, ok := f2["feed_state"].(map[string]interface{}); ok {
					if f4, ok := f3["feed_version"].(hw); ok {
						f2["geometry"] = f4["geometry"]
						delete(f4, "geometry")
					}
				}
			}
		}
	}
	return processGeoJSON(r, response)
}

func checkFeedSpecFilterValue(v string) string {
	v = strings.ToUpper(v)
	switch v {
	case "GTFS-RT":
		return "GTFS_RT"
	}
	return v
}
