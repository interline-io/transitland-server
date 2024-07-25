package model

import (
	"time"

	"github.com/interline-io/transitland-lib/tlxy"
)

// This file contains parameters that can be passed to methods for finding/selecting/grouping entities
// These are distinct from WHERE graphql input filters, which are available to users.

type ServiceWindow struct {
	NowLocal     time.Time
	StartDate    time.Time
	EndDate      time.Time
	FallbackWeek time.Time
	Location     *time.Location
}

type StopPlaceParam struct {
	ID    int
	Point tlxy.Point
}

type FrequencyParam struct {
	TripID int
	Limit  *int
}

type FeedVersionFileInfoParam struct {
	FeedVersionID int
	Limit         *int
}

type FeedVersionParam struct {
	FeedID int
	Limit  *int
	Where  *FeedVersionFilter
}

type FeedVersionServiceLevelParam struct {
	FeedVersionID int
	Limit         *int
	Where         *FeedVersionServiceLevelFilter
}

type FeedInfoParam struct {
	FeedVersionID int
	Limit         *int
}

type PathwayParam struct {
	FeedVersionID int
	FromStopID    int
	ToStopID      int
	Limit         *int
	Where         *PathwayFilter
}

type StopTimeParam struct {
	TripID        int
	StopID        int
	FeedVersionID int
	Limit         *int
	Where         *StopTimeFilter
}

type TripStopTimeParam struct {
	TripID        int
	FeedVersionID int
	Limit         *int
	StartTime     *int
	EndTime       *int
	Where         *TripStopTimeFilter
}

type AgencyParam struct {
	FeedVersionID int
	Limit         *int
	OnestopID     *string
	Where         *AgencyFilter
}

type RouteParam struct {
	AgencyID      int
	FeedVersionID int
	Limit         *int
	Where         *RouteFilter
}

type RouteStopParam struct {
	RouteID int
	StopID  int
	Limit   *int
}

type RouteHeadwayParam struct {
	RouteID int
	Limit   *int
}

type RouteGeometryParam struct {
	RouteID int
	Limit   *int
}

type TripParam struct {
	FeedVersionID int
	RouteID       int
	Limit         *int
	ServiceWindow *ServiceWindow
	Where         *TripFilter
}

type StopParam struct {
	FeedVersionID int
	ParentStopID  int
	AgencyID      int
	LevelID       int
	Limit         *int
	Where         *StopFilter
	RouteID       int
}

type LevelParam struct {
	ParentStationID int
	Limit           *int
}

type FeedParam struct {
	OperatorOnestopID string
	Limit             *int
	Where             *FeedFilter
}

type FeedFetchParam struct {
	FeedID int
	Limit  *int
	Where  *FeedFetchFilter
}

type AgencyPlaceParam struct {
	AgencyID int
	Limit    *int
	Where    *AgencyPlaceFilter
}

type OperatorParam struct {
	FeedID int
	Limit  *int
	Where  *OperatorFilter
}

type StopObservationParam struct {
	StopID int
	Limit  *int
	Where  *StopObservationFilter
}

type CalendarDateParam struct {
	ServiceID int
	Limit     *int
	Where     *CalendarDateFilter
}

type CensusGeographyParam struct {
	Radius     *float64
	LayerName  string
	EntityType string
	EntityID   int
	Limit      *int
}

type CensusValueParam struct {
	GeographyID int
	TableNames  string // these have to be comma joined for now, []string cant be used as map key
	Limit       *int
}

type CensusTableParam struct {
	Limit *int
}

type RouteStopBufferParam struct {
	EntityID int
	Radius   *float64
	Limit    *int
}

type RouteStopPatternParam struct {
	RouteID int
}

type SegmentPatternParam struct {
	SegmentID int
	RouteID   int
	Limit     *int
	Where     *SegmentPatternFilter
}

type SegmentParam struct {
	FeedVersionID int
	RouteID       int
	Layer         string
	Limit         *int
	Where         *SegmentFilter
}

type ValidationReportParam struct {
	FeedVersionID int
	Limit         *int
	Where         *ValidationReportFilter
}

type ValidationReportErrorExemplarParam struct {
	ValidationReportGroupID int
	Limit                   *int
}

type ValidationReportErrorGroupParam struct {
	ValidationReportID int
	Limit              *int
}
