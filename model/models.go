package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
)

type Feed struct {
	SearchRank *string
	tl.Feed
}

// OnestopID is called FeedID in transitland-lib.
func (f *Feed) OnestopID() (string, error) {
	return f.FeedID, nil
}

type FeedLicense struct {
	tl.FeedLicense
}

type FeedUrls struct {
	tl.FeedUrls
}

type FeedAuthorization struct {
	tl.FeedAuthorization
}

type Agency struct {
	OnestopID       string      `json:"onestop_id"`
	FeedOnestopID   string      `json:"feed_onestop_id"`
	FeedVersionSHA1 string      `json:"feed_version_sha1"`
	Geometry        *tl.Polygon `json:"geometry"`
	SearchRank      *string
	CoifID          *int
	tl.Agency
}

type Calendar struct {
	tl.Calendar
}

type FeedState struct {
	dmfr.FeedState
}

type FeedFetch struct {
	ResponseSha1 tl.String // confusing but easier than alternative fixes
	dmfr.FeedFetch
}

type FeedVersion struct {
	SHA1Dir tl.String `json:"sha1_dir"`
	tl.FeedVersion
}

type Operator struct {
	ID            int
	Generated     bool
	FeedID        int
	FeedOnestopID *string
	SearchRank    *string // internal
	AgencyID      int     // internal
	tl.Operator
}

type Route struct {
	FeedOnestopID                string
	FeedVersionSHA1              string
	OnestopID                    *string
	HeadwaySecondsWeekdayMorning *int
	SearchRank                   *string
	tl.Route
}

type Trip struct {
	RTTripID string // internal: for ADDED trips
	tl.Trip
}

type StopTime struct {
	ServiceDate      tl.Date
	RTTripID         string                        // internal: for ADDED trips
	RTStopTimeUpdate *pb.TripUpdate_StopTimeUpdate // internal
	tl.StopTime
}

type StopTimeEvent struct {
	StopTimezone string      `json:"stop_timezone"`
	Scheduled    tl.WideTime `json:"scheduled"`
	Estimated    tl.WideTime `json:"estimated"`
	EstimatedUtc tl.Time     `json:"estimated_utc"`
	Delay        *int        `json:"delay"`
	Uncertainty  *int        `json:"uncertainty"`
}

type Stop struct {
	FeedOnestopID   string
	FeedVersionSHA1 string
	OnestopID       *string
	SearchRank      *string
	tl.Stop
}

type Frequency struct {
	tl.Frequency
}

type CalendarDate struct {
	tl.CalendarDate
}

type Shape struct {
	tl.Shape
}

type Level struct {
	Geometry tl.Polygon
	tl.Level
}

type FeedInfo struct {
	tl.FeedInfo
}

type Pathway struct {
	tl.Pathway
}

type FeedVersionFileInfo struct {
	dmfr.FeedVersionFileInfo
}

type FeedVersionGtfsImport struct {
	WarningCount             *json.RawMessage `json:"warning_count"`
	EntityCount              *json.RawMessage `json:"entity_count"`
	SkipEntityErrorCount     *json.RawMessage `json:"skip_entity_error_count"`
	SkipEntityReferenceCount *json.RawMessage `json:"skip_entity_reference_count"`
	SkipEntityFilterCount    *json.RawMessage `json:"skip_entity_filter_count"`
	SkipEntityMarkedCount    *json.RawMessage `json:"skip_entity_marked_count"`
	dmfr.FeedVersionImport
}

type FeedVersionServiceLevel struct {
	dmfr.FeedVersionServiceLevel
}

// Support models that don't exist in transitland-lib

// Census models

type CensusValue struct {
	GeographyID int
	TableID     int
	TableValues ValueMap
}

// ValueMap is just a JSONB map[string]interface{}
type ValueMap map[string]interface{}

// Value dump
func (a ValueMap) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan load
func (a *ValueMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
