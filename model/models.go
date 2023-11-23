package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/interline-io/transitland-lib/dmfr"
	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
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
	tl.Trip
	RTTripID string // internal: for ADDED trips
}

type StopTime struct {
	tl.StopTime
	ServiceDate      tl.Date
	RTTripID         string                        // internal: for ADDED trips
	RTStopTimeUpdate *pb.TripUpdate_StopTimeUpdate // internal
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

type RouteStop struct {
	ID       int `json:"id"`
	RouteID  int
	StopID   int
	AgencyID int
}

type RouteHeadway struct {
	ID             int     `json:"id"`
	RouteID        int     `json:"route_id"`
	SelectedStopID int     `json:"selected_stop_id"`
	DirectionID    int     `json:"direction_id"`
	HeadwaySecs    *int    `json:"headway_secs"`
	DowCategory    *int    `json:"dow_category"`
	ServiceDate    tl.Date `json:"service_date"`
	ServiceSeconds *int    `json:"service_seconds"`
	StopTripCount  *int    `json:"stop_trip_count"`
	Departures     tl.Ints
}

type RouteStopPattern struct {
	RouteID       int `json:"id"`
	StopPatternID int `json:"stop_pattern_id"`
	DirectionID   int `json:"direction_id"`
	Count         int `json:"count"`
}

type RouteStopBuffer struct {
	StopPoints     *tl.Geometry `json:"stop_points"`
	StopBuffer     *tl.Geometry `json:"stop_buffer"`
	StopConvexhull *tl.Polygon  `json:"stop_convexhull"`
}

type RouteGeometry struct {
	RouteID               int           `json:"route_id"`
	Generated             bool          `json:"generated"`
	Geometry              tl.LineString `json:"geometry"`
	CombinedGeometry      tl.Geometry   `json:"combined_geometry"`
	Length                tl.Float      `json:"length"`
	MaxSegmentLength      tl.Float      `json:"max_segment_length"`
	FirstPointMaxDistance tl.Float      `json:"first_point_max_distance"`
}

// MTC GTFS+ Extension: route_attributes.txt
type RouteAttribute struct {
	RouteID     int
	Category    tt.Int `json:"Category"`
	Subcategory tt.Int `json:"Subcategory"`
	RunningWay  tt.Int `json:"RunningWay"`
}

type AgencyPlace struct {
	AgencyID int      `json:"agency_id"`
	CityName *string  `json:"city_name" db:"name" `
	Adm0Name *string  `json:"adm0_name" db:"adm0name" `
	Adm1Name *string  `json:"adm1_name" db:"adm1name" `
	Adm0Iso  *string  `json:"adm0_iso" db:"adm0iso" `
	Adm1Iso  *string  `json:"adm1_iso" db:"adm1iso" `
	Rank     *float64 `json:"rank"`
}

// Census models

type CensusGeography struct {
	ID            int         `json:"id"`
	LayerName     string      `json:"layer_name"`
	Geoid         *string     `json:"geoid"`
	Name          *string     `json:"name"`
	Aland         *float64    `json:"aland"`
	Awater        *float64    `json:"awater"`
	Geometry      *tl.Polygon `json:"geometry"`
	MatchEntityID int         // for matching to a stop, route, agency in query
}

type CensusTable struct {
	ID         int
	TableName  string
	TableTitle string
	TableGroup string
}

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

///////////////// Validation

// ValidationResult .
type ValidationResult struct {
	Success              bool                         `json:"success"`
	FailureReason        string                       `json:"failure_reason"`
	Errors               []ValidationResultErrorGroup `json:"errors"`
	Warnings             []ValidationResultErrorGroup `json:"warnings"`
	Sha1                 string                       `json:"sha1"`
	EarliestCalendarDate tl.Date                      `json:"earliest_calendar_date"`
	LatestCalendarDate   tl.Date                      `json:"latest_calendar_date"`
	Files                []FeedVersionFileInfo        `json:"files"`
	ServiceLevels        []FeedVersionServiceLevel    `json:"service_levels"`
	Agencies             []Agency                     `json:"agencies"`
	Routes               []Route                      `json:"routes"`
	Stops                []Stop                       `json:"stops"`
	FeedInfos            []FeedInfo                   `json:"feed_infos"`
	Realtime             []ValidationRealtimeResult   `json:"realtime"`
}

type ValidationRealtimeResult struct {
	Url  string         `json:"url"`
	Json map[string]any `json:"json"`
}

type ValidationResultErrorGroup struct {
	Filename  string                   `json:"filename"`
	ErrorType string                   `json:"error_type"`
	Count     int                      `json:"count"`
	Limit     int                      `json:"limit"`
	Code      int                      `json:"code"`
	Errors    []*ValidationResultError `json:"errors"`
}

type ValidationResultError struct {
	Filename   string         `json:"filename"`
	ErrorType  string         `json:"error_type"`
	EntityID   string         `json:"entity_id"`
	Field      string         `json:"field"`
	Value      string         `json:"value"`
	Message    string         `json:"message"`
	Code       *int           `json:"code"`
	Geometries []*tt.Geometry `json:"geometries"`
}

///////////////////// Fetch

type FeedVersionFetchResult struct {
	FeedVersion  *FeedVersion
	FetchError   *string
	FoundSHA1    bool
	FoundDirSHA1 bool
}

///////////////////// Import

type FeedVersionImportResult struct {
	Success bool
}

///////////////////// Analyst

type StopExternalReference struct {
	ID                  int     `json:"id"`
	TargetFeedOnestopID *string `json:"target_feed_onestop_id"`
	TargetStopID        *string `json:"target_stop_id"`
	Inactive            *bool   `json:"inactive"`
}

// Places

type Place struct {
	Adm0Name  *string `json:"adm0_name" db:"adm0name"`
	Adm1Name  *string `json:"adm1_name" db:"adm1name"`
	CityName  *string `json:"city_name" db:"name"`
	AgencyIDs tt.Ints `json:"-" db:"agency_ids"` // internal
}
