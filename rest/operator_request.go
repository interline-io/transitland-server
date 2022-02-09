package rest

import (
	_ "embed"
	"strconv"
)

//go:embed operator_request.gql
var operatorQuery string

// OperatorRequest holds options for a Route request
type OperatorRequest struct {
	OperatorKey   string `json:"operator_key"`
	ID            int    `json:"id,string"`
	Limit         int    `json:"limit,string"`
	After         int    `json:"after,string"`
	OnestopID     string `json:"onestop_id"`
	FeedOnestopID string `json:"feed_onestop_id"`
	Search        string `json:"search"`
	TagKey        string `json:"tag_key"`
	TagValue      string `json:"tag_value"`
	Adm0Name      string `json:"adm0_name"`
	Adm0Iso       string `json:"adm0_iso"`
	Adm1Name      string `json:"adm1_name"`
	Adm1Iso       string `json:"adm1_iso"`
	CityName      string `json:"city_name"`
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
	return operatorQuery, hw{"limit": checkLimit(r.Limit), "after": checkAfter(r.After), "ids": checkIds(r.ID), "where": where}
}
