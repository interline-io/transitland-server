package rest

import (
	_ "embed"
	"strconv"

	oa "github.com/getkin/kin-openapi/openapi3"
)

//go:embed feed_version_request.gql
var feedVersionQuery string

// FeedVersionRequest holds options for a Feed Version request
type FeedVersionRequest struct {
	FeedVersionKey  string    `json:"feed_version_key"`
	FeedKey         string    `json:"feed_key"`
	ID              int       `json:"id,string"`
	FeedID          int       `json:"feed_id,string"`
	FeedOnestopID   string    `json:"feed_onestop_id"`
	Sha1            string    `json:"sha1"`
	FetchedBefore   string    `json:"fetched_before"`
	FetchedAfter    string    `json:"fetched_after"`
	CoversStartDate string    `json:"covers_start_date"`
	CoversEndDate   string    `json:"covers_end_date"`
	Lon             float64   `json:"lon,string"`
	Lat             float64   `json:"lat,string"`
	Radius          float64   `json:"radius,string"`
	Bbox            *restBbox `json:"bbox"`
	WithCursor
}

func (r FeedVersionRequest) RequestInfo() RequestInfo {
	return RequestInfo{
		Path: "/feed_versions",
		PathItem: &oa.PathItem{
			Extensions: map[string]any{
				"x-alternates": []any{map[string]any{"description": "Request feed versions in specified format", "method": "get", "path": "/feeds_versions.{format}"}, map[string]any{"description": "Request a feed version by ID or SHA1", "method": "get", "path": "/feeds_versions/{feed_version_key}"}, map[string]any{"description": "Request a feed version by ID or SHA1 in specified format", "method": "get", "path": "/feeds_versions/{feed_version_key}.{format}"}, map[string]any{"description": "Request feed versions by feed ID or OnestopID", "method": "get", "path": "/feeds/{feed_key}/feed_versions"}},
			},
			Get: &oa.Operation{
				Summary:     "Feed Versions",
				Description: `Search for feed versions`,
				Responses:   queryToResponses(feedVersionQuery),
				Parameters: oa.Parameters{
					&pref{
						Ref: "#/components/parameters/idParam",
					},
					&pref{
						Value: &param{
							Name:        "feed_version_key",
							In:          "query",
							Description: `Feed version lookup key; can be an integer ID or a SHA1 value`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "feed_key",
							In:          "query",
							Description: `Feed lookup key; can be an integer ID or Onestop ID`,
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
							"x-example-requests": []any{map[string]any{"description": "limit=1", "url": "/feed_versions?limit=1"}},
						},
					},
					&pref{
						Ref: "#/components/parameters/formatParam",
						Extensions: map[string]any{
							"x-example-requests": []any{map[string]any{"description": "format=geojson", "url": "/feed_versions?format=geojson"}},
						},
					},
					&pref{
						Value: &param{
							Name:        "sha1",
							In:          "query",
							Description: `Feed version SHA1`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "sha1=e535eb2b3...", "url": "/feed_versions?sha1=dd7aca4a8e4c90908fd3603c097fabee75fea907"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "feed_onestop_id",
							In:          "query",
							Description: `Feed OnestopID`,
							Schema: &sref{
								Value: newSchema("string", "", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "feed_onestop_id=f-sf~bay~area~rg", "url": "/feed_versions?feed_onestop_id=f-sf~bay~area~rg"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "fetched_before",
							In:          "query",
							Description: `Filter for feed versions fetched earlier than given date time in UTC`,
							Schema: &sref{
								Value: newSchema("string", "datetime", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "fetched_before=2023-01-01T00:00:00Z", "url": "/feed_versions?fetched_before=2023-01-01T00:00:00Z"}},
							},
						},
					},
					&pref{
						Value: &param{
							Name:        "fetched_after",
							In:          "query",
							Description: `Filter for feed versions fetched since given date time in UTC`,
							Schema: &sref{
								Value: newSchema("string", "datetime", nil),
							},
							Extensions: map[string]any{
								"x-example-requests": []any{map[string]any{"description": "fetched_after=2023-01-01T00:00:00Z", "url": "/feed_versions?fetched_after=2023-01-01T00:00:00Z"}},
							},
						},
					},
					&pref{
						Ref: "#/components/parameters/radiusParam",
						Extensions: map[string]any{
							"x-description":      "Search for feed versions geographically; radius is in meters, requires lon and lat",
							"x-example-requests": []any{map[string]any{"description": "lon=-122&lat=37&radius=1000", "url": "/feed_versions?lon=-122.3&lat=37.8&radius=1000"}},
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
							"x-example-requests": []any{map[string]any{"description": "bbox=-122.269,37.807,-122.267,37.808", "url": "/feed_versions?bbox=-122.269,37.807,-122.267,37.808"}},
						},
					},
				},
			},
		},
	}
}

// Query returns a GraphQL query string and variables.
func (r FeedVersionRequest) Query() (string, map[string]interface{}) {
	// Handle feed key
	if r.FeedKey == "" {
		// pass
	} else if v, err := strconv.Atoi(r.FeedKey); err == nil {
		r.FeedID = v
	} else {
		r.FeedOnestopID = r.FeedKey
	}
	// Handle feed version key
	if r.FeedVersionKey == "" {
		// pass
	} else if v, err := strconv.Atoi(r.FeedVersionKey); err == nil {
		r.ID = v
	} else {
		r.Sha1 = r.FeedVersionKey
	}
	where := hw{}
	if r.FeedID > 0 {
		where["feed_ids"] = []int{r.FeedID}
	}
	if r.FeedOnestopID != "" {
		where["feed_onestop_id"] = r.FeedOnestopID
	}
	if r.Sha1 != "" {
		where["sha1"] = r.Sha1
	}
	if r.Lat != 0.0 && r.Lon != 0.0 {
		where["near"] = hw{"lat": r.Lat, "lon": r.Lon, "radius": r.Radius}
	}
	if r.Bbox != nil {
		where["bbox"] = r.Bbox.AsJson()
	}
	whereCovers := hw{}
	if r.CoversStartDate != "" {
		whereCovers["start_date"] = r.CoversStartDate
	}
	if r.CoversEndDate != "" {
		whereCovers["end_date"] = r.CoversEndDate
	}
	if r.FetchedAfter != "" {
		whereCovers["fetched_after"] = r.FetchedAfter
	}
	if r.FetchedBefore != "" {
		whereCovers["fetched_before"] = r.FetchedBefore
	}
	if len(whereCovers) > 0 {
		where["covers"] = whereCovers
	}
	return feedVersionQuery, hw{"limit": r.CheckLimit(), "after": r.CheckAfter(), "ids": checkIds(r.ID), "where": where}
}

// ResponseKey .
func (r FeedVersionRequest) ResponseKey() string {
	return "feed_versions"
}
