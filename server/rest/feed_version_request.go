package rest

import (
	_ "embed"
	"strconv"
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
