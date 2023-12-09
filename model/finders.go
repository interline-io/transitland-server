package model

import (
	"context"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/internal/gbfs"

	"github.com/jmoiron/sqlx"
)

type Finders struct {
	Config     config.Config
	Finder     Finder
	RTFinder   RTFinder
	GbfsFinder GbfsFinder
	Checker    Checker
}

// Finder provides all necessary database methods
type Finder interface {
	PermFinder
	EntityFinder
	EntityLoader
}

type PermFinder interface {
	PermFilter(context.Context) *PermFilter
}

// Finder handles basic queries
type EntityFinder interface {
	FindAgencies(context.Context, *int, *Cursor, []int, *AgencyFilter) ([]*Agency, error)
	FindRoutes(context.Context, *int, *Cursor, []int, *RouteFilter) ([]*Route, error)
	FindStops(context.Context, *int, *Cursor, []int, *StopFilter) ([]*Stop, error)
	FindTrips(context.Context, *int, *Cursor, []int, *TripFilter) ([]*Trip, error)
	FindFeedVersions(context.Context, *int, *Cursor, []int, *FeedVersionFilter) ([]*FeedVersion, error)
	FindFeeds(context.Context, *int, *Cursor, []int, *FeedFilter) ([]*Feed, error)
	FindOperators(context.Context, *int, *Cursor, []int, *OperatorFilter) ([]*Operator, error)
	FindPlaces(context.Context, *int, *Cursor, []int, *PlaceAggregationLevel, *PlaceFilter) ([]*Place, error)
	RouteStopBuffer(context.Context, *RouteStopBufferParam) ([]*RouteStopBuffer, error)
	FindFeedVersionServiceWindow(context.Context, int) (time.Time, time.Time, time.Time, error)
	DBX() sqlx.Ext // escape hatch, for now
}

// EntityLoader methods must return items in the same order as the input parameters
type EntityLoader interface {
	// Simple ID loaders
	TripsByID(context.Context, []int) ([]*Trip, []error)
	LevelsByID(context.Context, []int) ([]*Level, []error)
	CalendarsByID(context.Context, []int) ([]*Calendar, []error)
	ShapesByID(context.Context, []int) ([]*Shape, []error)
	FeedVersionsByID(context.Context, []int) ([]*FeedVersion, []error)
	FeedsByID(context.Context, []int) ([]*Feed, []error)
	AgenciesByID(context.Context, []int) ([]*Agency, []error)
	StopsByID(context.Context, []int) ([]*Stop, []error)
	RoutesByID(context.Context, []int) ([]*Route, []error)
	StopExternalReferencesByStopID(context.Context, []int) ([]*StopExternalReference, []error)
	StopObservationsByStopID(context.Context, []StopObservationParam) ([][]*StopObservation, []error)
	TargetStopsByStopID(context.Context, []int) ([]*Stop, []error)
	RouteAttributesByRouteID(context.Context, []int) ([]*RouteAttribute, []error)
	CensusTableByID(context.Context, []int) ([]*CensusTable, []error)
	FeedVersionGeometryByID(context.Context, []int) ([]*tt.Polygon, []error)

	// Other loaders
	FeedVersionGtfsImportsByFeedVersionID(context.Context, []int) ([]*FeedVersionGtfsImport, []error)
	FeedStatesByFeedID(context.Context, []int) ([]*FeedState, []error)
	OperatorsByFeedID(context.Context, []OperatorParam) ([][]*Operator, []error)
	OperatorsByCOIF(context.Context, []int) ([]*Operator, []error)
	OperatorsByOnestopID(context.Context, []string) ([]*Operator, []error)
	OperatorsByAgencyID(context.Context, []int) ([]*Operator, []error)
	StopPlacesByStopID(context.Context, []StopPlaceParam) ([]*StopPlace, []error)

	// Param loaders
	FeedFetchesByFeedID(context.Context, []FeedFetchParam) ([][]*FeedFetch, []error)
	FeedsByOperatorOnestopID(context.Context, []FeedParam) ([][]*Feed, []error)
	FrequenciesByTripID(context.Context, []FrequencyParam) ([][]*Frequency, []error)
	StopTimesByTripID(context.Context, []TripStopTimeParam) ([][]*StopTime, []error)
	StopTimesByStopID(context.Context, []StopTimeParam) ([][]*StopTime, []error)
	RouteStopsByStopID(context.Context, []RouteStopParam) ([][]*RouteStop, []error)
	StopsByRouteID(context.Context, []StopParam) ([][]*Stop, []error)
	RouteStopsByRouteID(context.Context, []RouteStopParam) ([][]*RouteStop, []error)
	RouteHeadwaysByRouteID(context.Context, []RouteHeadwayParam) ([][]*RouteHeadway, []error)
	RouteStopPatternsByRouteID(context.Context, []RouteStopPatternParam) ([][]*RouteStopPattern, []error)
	FeedVersionFileInfosByFeedVersionID(context.Context, []FeedVersionFileInfoParam) ([][]*FeedVersionFileInfo, []error)
	StopsByParentStopID(context.Context, []StopParam) ([][]*Stop, []error)
	FeedVersionsByFeedID(context.Context, []FeedVersionParam) ([][]*FeedVersion, []error)
	AgencyPlacesByAgencyID(context.Context, []AgencyPlaceParam) ([][]*AgencyPlace, []error)
	RouteGeometriesByRouteID(context.Context, []RouteGeometryParam) ([][]*RouteGeometry, []error)
	TripsByRouteID(context.Context, []TripParam) ([][]*Trip, []error)
	RoutesByAgencyID(context.Context, []RouteParam) ([][]*Route, []error)
	AgenciesByFeedVersionID(context.Context, []AgencyParam) ([][]*Agency, []error)
	AgenciesByOnestopID(context.Context, []AgencyParam) ([][]*Agency, []error)
	StopsByFeedVersionID(context.Context, []StopParam) ([][]*Stop, []error)
	StopsByLevelID(context.Context, []StopParam) ([][]*Stop, []error)
	TripsByFeedVersionID(context.Context, []TripParam) ([][]*Trip, []error)
	FeedInfosByFeedVersionID(context.Context, []FeedInfoParam) ([][]*FeedInfo, []error)
	RoutesByFeedVersionID(context.Context, []RouteParam) ([][]*Route, []error)
	FeedVersionServiceLevelsByFeedVersionID(context.Context, []FeedVersionServiceLevelParam) ([][]*FeedVersionServiceLevel, []error)
	PathwaysByFromStopID(context.Context, []PathwayParam) ([][]*Pathway, []error)
	PathwaysByToStopID(context.Context, []PathwayParam) ([][]*Pathway, []error)
	CalendarDatesByServiceID(context.Context, []CalendarDateParam) ([][]*CalendarDate, []error)
	CensusGeographiesByEntityID(context.Context, []CensusGeographyParam) ([][]*CensusGeography, []error)
	CensusValuesByGeographyID(context.Context, []CensusValueParam) ([][]*CensusValue, []error)
}

// RTFinder manages and looks up RT data
type RTFinder interface {
	AddData(string, []byte) error
	FindTrip(t *Trip) *pb.TripUpdate
	MakeTrip(t *Trip) (*Trip, error)
	FindAlertsForTrip(*Trip, *int, *bool) []*Alert
	FindAlertsForStop(*Stop, *int, *bool) []*Alert
	FindAlertsForRoute(*Route, *int, *bool) []*Alert
	FindAlertsForAgency(*Agency, *int, *bool) []*Alert
	GetAddedTripsForStop(*Stop) []*pb.TripUpdate
	FindStopTimeUpdate(*Trip, *StopTime) (*pb.TripUpdate_StopTimeUpdate, bool)
	// lookup cache methods
	StopTimezone(int, string) (*time.Location, bool)
	GetGtfsTripID(int) (string, bool)
}

// GBFSFinder manages and looks up GBFS data
type GbfsFinder interface {
	AddData(context.Context, string, gbfs.GbfsFeed) error
	FindBikes(context.Context, *int, *GbfsBikeRequest) ([]*GbfsFreeBikeStatus, error)
	FindDocks(context.Context, *int, *GbfsDockRequest) ([]*GbfsStationInformation, error)
}

type Checker interface {
	authz.CheckerServer
}
