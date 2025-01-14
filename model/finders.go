package model

import (
	"context"
	"io"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-server/internal/gbfs"
)

// Finder provides all necessary database methods
type Finder interface {
	PermFinder
	EntityFinder
	EntityLoader
	EntityMutator
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
	FindFeedVersionServiceWindow(context.Context, int) (*ServiceWindow, error)
	DBX() tldb.Ext // escape hatch, for now
}

type EntityMutator interface {
	StopCreate(ctx context.Context, input StopSetInput) (int, error)
	StopUpdate(ctx context.Context, input StopSetInput) (int, error)
	StopDelete(ctx context.Context, id int) error
	PathwayCreate(ctx context.Context, input PathwaySetInput) (int, error)
	PathwayUpdate(ctx context.Context, input PathwaySetInput) (int, error)
	PathwayDelete(ctx context.Context, id int) error
	LevelCreate(ctx context.Context, input LevelSetInput) (int, error)
	LevelUpdate(ctx context.Context, input LevelSetInput) (int, error)
	LevelDelete(ctx context.Context, id int) error
}

// EntityLoader methods must return items in the same order as the input parameters
type EntityLoader interface {
	// Simple ID loaders
	TripsByID(context.Context, []int) ([]*Trip, []error)
	LevelsByID(context.Context, []int) ([]*Level, []error)
	PathwaysByID(context.Context, []int) ([]*Pathway, []error)
	CalendarsByID(context.Context, []int) ([]*Calendar, []error)
	ShapesByID(context.Context, []int) ([]*Shape, []error)
	FeedVersionsByID(context.Context, []int) ([]*FeedVersion, []error)
	FeedsByID(context.Context, []int) ([]*Feed, []error)
	AgenciesByID(context.Context, []int) ([]*Agency, []error)
	StopsByID(context.Context, []int) ([]*Stop, []error)
	RoutesByID(context.Context, []int) ([]*Route, []error)
	LevelsByParentStationID(context.Context, []LevelParam) ([][]*Level, []error)
	StopExternalReferencesByStopID(context.Context, []int) ([]*StopExternalReference, []error)
	StopObservationsByStopID(context.Context, []StopObservationParam) ([][]*StopObservation, []error)
	TargetStopsByStopID(context.Context, []int) ([]*Stop, []error)
	RouteAttributesByRouteID(context.Context, []int) ([]*RouteAttribute, []error)
	FeedVersionGeometryByID(context.Context, []int) ([]*tt.Polygon, []error)
	CensusTableByID(context.Context, []int) ([]*CensusTable, []error)

	// Segments
	SegmentPatternsByRouteID(context.Context, []SegmentPatternParam) ([][]*SegmentPattern, []error)
	SegmentPatternsBySegmentID(context.Context, []SegmentPatternParam) ([][]*SegmentPattern, []error)
	SegmentsByID(context.Context, []int) ([]*Segment, []error)
	SegmentsByRouteID(context.Context, []SegmentParam) ([][]*Segment, []error)
	SegmentsByFeedVersionID(context.Context, []SegmentParam) ([][]*Segment, []error)

	// Other loaders
	FeedVersionGtfsImportByFeedVersionID(context.Context, []int) ([]*FeedVersionGtfsImport, []error)
	FeedVersionServiceWindowByFeedVersionID(context.Context, []int) ([]*FeedVersionServiceWindow, []error)
	FeedStatesByFeedID(context.Context, []int) ([]*FeedState, []error)
	OperatorsByFeedID(context.Context, []OperatorParam) ([][]*Operator, []error)
	OperatorsByCOIF(context.Context, []int) ([]*Operator, []error)
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
	CensusFieldsByTableID(context.Context, []CensusFieldParam) ([][]*CensusField, []error)

	// Validation reports
	ValidationReportsByFeedVersionID(context.Context, []ValidationReportParam) ([][]*ValidationReport, []error)
	ValidationReportErrorGroupsByValidationReportID(context.Context, []ValidationReportErrorGroupParam) ([][]*ValidationReportErrorGroup, []error)
	ValidationReportErrorExemplarsByValidationReportErrorGroupID(context.Context, []ValidationReportErrorExemplarParam) ([][]*ValidationReportError, []error)
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
	FindStopTimeUpdate(*Trip, *StopTime) (*RTStopTimeUpdate, bool)
	// lookup cache methods
	StopTimezone(int, string) (*time.Location, bool)
	GetGtfsTripID(int) (string, bool)
	GetMessage(string, string) (*pb.FeedMessage, bool)
}

// GbfsFinder manages and looks up GBFS data
type GbfsFinder interface {
	AddData(context.Context, string, gbfs.GbfsFeed) error
	FindBikes(context.Context, *int, *GbfsBikeRequest) ([]*GbfsFreeBikeStatus, error)
	FindDocks(context.Context, *int, *GbfsDockRequest) ([]*GbfsStationInformation, error)
}

type Checker interface {
	authz.CheckerServer
}

type Actions interface {
	StaticFetch(context.Context, string, io.Reader, string) (*FeedVersionFetchResult, error)
	RTFetch(context.Context, string, string, string, string) error
	GbfsFetch(context.Context, string, string) error
	ValidateUpload(context.Context, io.Reader, *string, []string) (*ValidationReport, error)
	FeedVersionUnimport(context.Context, int) (*FeedVersionUnimportResult, error)
	FeedVersionImport(context.Context, int) (*FeedVersionImportResult, error)
	FeedVersionUpdate(context.Context, FeedVersionSetInput) (int, error)
	FeedVersionDelete(context.Context, int) (*FeedVersionDeleteResult, error)
}
