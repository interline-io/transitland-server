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
	return operatorQuery, hw{"limit": checkLimit(r.Limit), "after": checkAfter(r.After), "ids": checkIds(r.ID), "where": where}
}
