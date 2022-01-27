package model

import (
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"

	"github.com/jmoiron/sqlx"
)

// Finder provides all necessary database methods
type Finder interface {
	EntityFinder
	EntityLoader
}

// Finder handles basic queries
type EntityFinder interface {
	FindAgencies(limit *int, after *int, ids []int, where *AgencyFilter) ([]*Agency, error)
	FindRoutes(limit *int, after *int, ids []int, where *RouteFilter) ([]*Route, error)
	FindStops(limit *int, after *int, ids []int, where *StopFilter) ([]*Stop, error)
	FindTrips(limit *int, after *int, ids []int, where *TripFilter) ([]*Trip, error)
	FindFeedVersions(limit *int, after *int, ids []int, where *FeedVersionFilter) ([]*FeedVersion, error)
	FindFeeds(limit *int, after *int, ids []int, where *FeedFilter) ([]*Feed, error)
	FindOperators(limit *int, after *int, ids []int, where *OperatorFilter) ([]*Operator, error)
	RouteStopBuffer(*RouteStopBufferParam) ([]*RouteStopBuffer, error)
	DBX() sqlx.Ext // escape hatch, for now
}

// EntityLoader methods must return items in the same order as the input parameters
type EntityLoader interface {
	// Simple ID loaders
	TripsByID([]int) ([]*Trip, []error)
	LevelsByID([]int) ([]*Level, []error)
	CalendarsByID([]int) ([]*Calendar, []error)
	ShapesByID([]int) ([]*Shape, []error)
	FeedVersionsByID([]int) ([]*FeedVersion, []error)
	FeedsByID([]int) ([]*Feed, []error)
	AgenciesByID([]int) ([]*Agency, []error)
	StopsByID([]int) ([]*Stop, []error)
	RoutesByID([]int) ([]*Route, []error)
	CensusTableByID([]int) ([]*CensusTable, []error)
	// Other loaders
	FeedVersionGtfsImportsByFeedVersionID([]int) ([]*FeedVersionGtfsImport, []error)
	FeedStatesByFeedID([]int) ([]*FeedState, []error)
	OperatorsByFeedID([]OperatorParam) ([][]*Operator, []error)
	OperatorsByCOIF([]int) ([]*Operator, []error)
	// Param loaders
	FrequenciesByTripID([]FrequencyParam) ([][]*Frequency, []error)
	StopTimesByTripID([]StopTimeParam) ([][]*StopTime, []error)
	StopTimesByStopID([]StopTimeParam) ([][]*StopTime, []error)
	RouteStopsByStopID([]RouteStopParam) ([][]*RouteStop, []error)
	StopsByRouteID([]StopParam) ([][]*Stop, []error)
	RouteStopsByRouteID([]RouteStopParam) ([][]*RouteStop, []error)
	RouteHeadwaysByRouteID([]RouteHeadwayParam) ([][]*RouteHeadway, []error)
	FeedVersionFileInfosByFeedVersionID([]FeedVersionFileInfoParam) ([][]*FeedVersionFileInfo, []error)
	StopsByParentStopID([]StopParam) ([][]*Stop, []error)
	FeedVersionsByFeedID([]FeedVersionParam) ([][]*FeedVersion, []error)
	AgencyPlacesByAgencyID([]AgencyPlaceParam) ([][]*AgencyPlace, []error)
	RouteGeometriesByRouteID([]RouteGeometryParam) ([][]*RouteGeometry, []error)
	TripsByRouteID([]TripParam) ([][]*Trip, []error)
	RoutesByAgencyID([]RouteParam) ([][]*Route, []error)
	AgenciesByFeedVersionID([]AgencyParam) ([][]*Agency, []error)
	AgenciesByOnestopID([]AgencyParam) ([][]*Agency, []error)
	StopsByFeedVersionID([]StopParam) ([][]*Stop, []error)
	TripsByFeedVersionID([]TripParam) ([][]*Trip, []error)
	FeedInfosByFeedVersionID([]FeedInfoParam) ([][]*FeedInfo, []error)
	RoutesByFeedVersionID([]RouteParam) ([][]*Route, []error)
	FeedVersionServiceLevelsByFeedVersionID([]FeedVersionServiceLevelParam) ([][]*FeedVersionServiceLevel, []error)
	PathwaysByFromStopID([]PathwayParam) ([][]*Pathway, []error)
	PathwaysByToStopID([]PathwayParam) ([][]*Pathway, []error)
	CalendarDatesByServiceID([]CalendarDateParam) ([][]*CalendarDate, []error)
	CensusGeographiesByEntityID([]CensusGeographyParam) ([][]*CensusGeography, []error)
	CensusValuesByGeographyID([]CensusValueParam) ([][]*CensusValue, []error)
}

// RTFinder manages and looks up RT data
type RTFinder interface {
	AddData(string, []byte) error
	FindTrip(t *Trip) *pb.TripUpdate
	FindMakeTrip(t *Trip) (*Trip, error)
	FindAlertsForTrip(*Trip) []*Alert
	FindAlertsForStop(*Stop) []*Alert
	FindAlertsForRoute(*Route) []*Alert
	FindAlertsForAgency(*Agency) []*Alert
	GetAddedTripsForStop(*Stop) []*pb.TripUpdate
	FindStopTimeUpdate(*Trip, *StopTime) (*pb.TripUpdate_StopTimeUpdate, bool)
	// lookup cache methods
	StopTimezone(int, string) (*time.Location, bool)
	GetGtfsTripID(int) (string, bool)
}
