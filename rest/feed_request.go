package rest

import (
	_ "embed"
	"strconv"
)

//go:embed feed_request.gql
var feedQuery string

// FeedRequest holds options for a Route request
type FeedRequest struct {
	FeedKey          string `json:"feed_key"`
	ID               int    `json:"id,string"`
	Limit            int    `json:"limit,string"`
	After            int    `json:"after,string"`
	OnestopID        string `json:"onestop_id"`
	Spec             string `json:"spec"`
	Search           string `json:"search"`
	FetchError       string `json:"fetch_error"`
	TagKey           string `json:"tag_key"`
	TagValue         string `json:"tag_value"`
	URL              string `json:"url"`
	URLType          string `json:"url_type"`
	URLCaseSensitive bool   `json:"url_case_sensitive"`
	LicenseFilter
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
		where["spec"] = []string{r.Spec}
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
	return feedQuery, hw{"limit": checkLimit(r.Limit), "after": checkAfter(r.After), "ids": checkIds(r.ID), "where": where}
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
