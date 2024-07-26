package model

import (
	"encoding/json"

	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
)

type Feed struct {
	SearchRank *string
	tl.Feed
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

type RTStopTimeUpdate struct {
	LastDelay      *int32
	StopTimeUpdate *pb.TripUpdate_StopTimeUpdate
	TripUpdate     *pb.TripUpdate
}

type StopTime struct {
	ServiceDate      tl.Date
	Date             tl.Date
	RTTripID         string            // internal: for ADDED trips
	RTStopTimeUpdate *RTStopTimeUpdate // internal
	tl.StopTime
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
	Geometry      tl.Polygon
	ParentStation tt.Key
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

// Some enum helpers

var specTypeMap = map[string]FeedSpecTypes{
	"gtfs":    FeedSpecTypesGtfs,
	"gtfs-rt": FeedSpecTypesGtfsRt,
	"gbfs":    FeedSpecTypesGbfs,
	"mds":     FeedSpecTypesMds,
}

func (f FeedSpecTypes) ToDBString() string {
	for k, v := range specTypeMap {
		if f == v {
			return k
		}
	}
	return ""
}

func (f FeedSpecTypes) FromDBString(s string) *FeedSpecTypes {
	a, ok := specTypeMap[s]
	if !ok {
		return nil
	}
	return &a
}
