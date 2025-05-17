package gql

import (
	"github.com/interline-io/transitland-server/model"
)

// This file contains parameters that can be passed to methods for finding/selecting/grouping entities
// These are distinct from WHERE graphql input filters, which are available to users.

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
	Where  *model.FeedVersionFilter
}

type FeedVersionServiceLevelParam struct {
	FeedVersionID int
	Limit         *int
	Where         *model.FeedVersionServiceLevelFilter
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
	Where         *model.PathwayFilter
}

type StopTimeParam struct {
	TripID        int
	StopID        int
	FeedVersionID int
	Limit         *int
	Where         *model.StopTimeFilter
}

type TripStopTimeParam struct {
	TripID        int
	FeedVersionID int
	Limit         *int
	StartTime     *int
	EndTime       *int
	Where         *model.TripStopTimeFilter
}

type AgencyParam struct {
	FeedVersionID int
	Limit         *int
	OnestopID     *string
	Where         *model.AgencyFilter
}

type RouteParam struct {
	AgencyID      int
	FeedVersionID int
	Limit         *int
	Where         *model.RouteFilter
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
	ServiceWindow *model.ServiceWindow
	Where         *model.TripFilter
}

type StopParam struct {
	FeedVersionID int
	ParentStopID  int
	AgencyID      int
	LevelID       int
	Limit         *int
	Where         *model.StopFilter
	RouteID       int
}

type LevelParam struct {
	ParentStationID int
	Limit           *int
}

type FeedParam struct {
	OperatorOnestopID string
	Limit             *int
	Where             *model.FeedFilter
}

type FeedFetchParam struct {
	FeedID int
	Limit  *int
	Where  *model.FeedFetchFilter
}

type AgencyPlaceParam struct {
	AgencyID int
	Limit    *int
	Where    *model.AgencyPlaceFilter
}

type OperatorParam struct {
	FeedID int
	Limit  *int
	Where  *model.OperatorFilter
}

type StopObservationParam struct {
	StopID int
	Limit  *int
	Where  *model.StopObservationFilter
}

type CalendarDateParam struct {
	ServiceID int
	Limit     *int
	Where     *model.CalendarDateFilter
}

type CensusGeographyParam struct {
	EntityType string
	EntityID   int
	DatasetID  int
	Limit      *int
	Where      *model.CensusGeographyFilter
}

type CensusDatasetGeographyParam struct {
	DatasetID int
	Limit     *int
	Where     *model.CensusDatasetGeographyFilter
}

type CensusValueParam struct {
	Dataset    *string
	Geoid      string
	TableNames string // these have to be comma joined for now, []string cant be used as map key
	Limit      *int
}

type CensusTableParam struct {
	Limit *int
}

type CensusFieldParam struct {
	Limit   *int
	TableID int
}

type CensusSourceParam struct {
	DatasetID int
	Limit     *int
	Where     *model.CensusSourceFilter
}

type RouteStopPatternParam struct {
	RouteID int
}

type SegmentPatternParam struct {
	SegmentID int
	RouteID   int
	Limit     *int
	Where     *model.SegmentPatternFilter
}

type SegmentParam struct {
	FeedVersionID int
	RouteID       int
	Layer         string
	Limit         *int
	Where         *model.SegmentFilter
}

type ValidationReportParam struct {
	FeedVersionID int
	Limit         *int
	Where         *model.ValidationReportFilter
}

type ValidationReportErrorExemplarParam struct {
	ValidationReportGroupID int
	Limit                   *int
}

type ValidationReportErrorGroupParam struct {
	ValidationReportID int
	Limit              *int
}
