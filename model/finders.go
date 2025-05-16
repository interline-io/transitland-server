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
	FindCensusDatasets(context.Context, *int, *Cursor, []int, *CensusDatasetFilter) ([]*CensusDataset, error)
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

type EntityLoader interface {
	// Simple ID loaders
	TripsByIDs(context.Context, []int) ([]*Trip, []error)
	LevelsByIDs(context.Context, []int) ([]*Level, []error)
	PathwaysByIDs(context.Context, []int) ([]*Pathway, []error)
	CalendarsByIDs(context.Context, []int) ([]*Calendar, []error)
	ShapesByIDs(context.Context, []int) ([]*Shape, []error)
	FeedVersionsByIDs(context.Context, []int) ([]*FeedVersion, []error)
	FeedsByIDs(context.Context, []int) ([]*Feed, []error)
	AgenciesByIDs(context.Context, []int) ([]*Agency, []error)
	StopsByIDs(context.Context, []int) ([]*Stop, []error)
	RoutesByIDs(context.Context, []int) ([]*Route, []error)
	StopExternalReferencesByStopIDs(context.Context, []int) ([]*StopExternalReference, []error)
	TargetStopsByStopIDs(context.Context, []int) ([]*Stop, []error)
	RouteAttributesByRouteIDs(context.Context, []int) ([]*RouteAttribute, []error)
	FeedVersionGeometryByIDs(context.Context, []int) ([]*tt.Polygon, []error)
	CensusTableByIDs(context.Context, []int) ([]*CensusTable, []error)

	// Other loaders
	FeedVersionGtfsImportByFeedVersionIDs(context.Context, []int) ([]*FeedVersionGtfsImport, []error)
	FeedVersionServiceWindowByFeedVersionIDs(context.Context, []int) ([]*FeedVersionServiceWindow, []error)
	FeedStatesByFeedIDs(context.Context, []int) ([]*FeedState, []error)
	OperatorsByCOIFs(context.Context, []int) ([]*Operator, []error)
	OperatorsByAgencyIDs(context.Context, []int) ([]*Operator, []error)

	// Param loaders
	AgenciesByFeedVersionIDs(ctx context.Context, limit *int, where *AgencyFilter, feedVersionIds []int) ([]*Agency, error)
	AgenciesByOnestopIDs(context.Context, *int, *AgencyFilter, []string) ([]*Agency, error)
	AgencyPlacesByAgencyIDs(context.Context, *int, *AgencyPlaceFilter, []int) ([]*AgencyPlace, error)
	CalendarDatesByServiceIDs(context.Context, *int, *CalendarDateFilter, []int) ([]*CalendarDate, error)

	CensusDatasetLayersByDatasetIDs(context.Context, []int) ([][]string, []error)
	CensusFieldsByTableIDs(context.Context, *int, []int) ([]*CensusField, error)
	CensusGeographiesByDatasetIDs(context.Context, *int, *CensusDatasetGeographyFilter, []int) ([]*CensusGeography, error)
	CensusGeographiesByEntityIDs(context.Context, *int, *CensusGeographyFilter, string, []int) ([]*CensusGeography, error)
	CensusSourceLayersBySourceIDs(context.Context, []int) ([][]string, []error)
	CensusSourcesByDatasetIDs(context.Context, *int, *CensusSourceFilter, []int) ([]*CensusSource, error)
	CensusValuesByGeographyIDs(context.Context, *int, []string, []string) ([]*CensusValue, error)

	FeedFetchesByFeedIDs(context.Context, *int, *FeedFetchFilter, []int) ([]*FeedFetch, error)
	FeedInfosByFeedVersionIDs(context.Context, *int, []int) ([]*FeedInfo, error)
	FeedsByOperatorOnestopIDs(context.Context, *int, *FeedFilter, []string) ([]*Feed, error)
	FeedVersionFileInfosByFeedVersionIDs(context.Context, *int, []int) ([]*FeedVersionFileInfo, error)
	FeedVersionsByFeedIDs(context.Context, *int, *FeedVersionFilter, []int) ([]*FeedVersion, error)
	FeedVersionServiceLevelsByFeedVersionIDs(context.Context, *int, *FeedVersionServiceLevelFilter, []int) ([]*FeedVersionServiceLevel, error)

	FrequenciesByTripIDs(context.Context, *int, []int) ([]*Frequency, error)

	PathwaysByFromStopIDs(context.Context, *int, *PathwayFilter, []int) ([]*Pathway, error)
	PathwaysByToStopIDs(context.Context, *int, *PathwayFilter, []int) ([]*Pathway, error)
	LevelsByParentStationIDs(context.Context, *int, []int) ([]*Level, error)

	RouteGeometriesByRouteIDs(context.Context, *int, []int) ([]*RouteGeometry, error)
	RouteHeadwaysByRouteID(context.Context, []RouteHeadwayParam) ([][]*RouteHeadway, []error)
	RoutesByAgencyID(context.Context, []RouteParam) ([][]*Route, []error)
	RoutesByFeedVersionID(context.Context, []RouteParam) ([][]*Route, []error)
	RouteStopPatternsByRouteID(context.Context, []RouteStopPatternParam) ([][]*RouteStopPattern, []error)
	RouteStopsByRouteID(context.Context, []RouteStopParam) ([][]*RouteStop, []error)
	RouteStopsByStopID(context.Context, []RouteStopParam) ([][]*RouteStop, []error)

	StopObservationsByStopID(context.Context, []StopObservationParam) ([][]*StopObservation, []error)
	StopsByFeedVersionID(context.Context, []StopParam) ([][]*Stop, []error)
	StopsByLevelID(context.Context, []StopParam) ([][]*Stop, []error)
	StopsByParentStopID(context.Context, []StopParam) ([][]*Stop, []error)
	StopsByRouteID(context.Context, []StopParam) ([][]*Stop, []error)
	StopTimesByStopID(context.Context, []StopTimeParam) ([][]*StopTime, []error)
	StopTimesByTripID(context.Context, []TripStopTimeParam) ([][]*StopTime, []error)
	StopPlacesByStopID(context.Context, []StopPlaceParam) ([]*StopPlace, []error)

	TripsByFeedVersionID(context.Context, []TripParam) ([][]*Trip, []error)
	TripsByRouteID(context.Context, []TripParam) ([][]*Trip, []error)

	OperatorsByFeedID(context.Context, []OperatorParam) ([][]*Operator, []error)

	// Validation reports
	ValidationReportsByFeedVersionID(context.Context, []ValidationReportParam) ([][]*ValidationReport, []error)
	ValidationReportErrorGroupsByValidationReportID(context.Context, []ValidationReportErrorGroupParam) ([][]*ValidationReportErrorGroup, []error)
	ValidationReportErrorExemplarsByValidationReportErrorGroupID(context.Context, []ValidationReportErrorExemplarParam) ([][]*ValidationReportError, []error)

	// Segments
	SegmentPatternsByRouteID(context.Context, []SegmentPatternParam) ([][]*SegmentPattern, []error)
	SegmentPatternsBySegmentID(context.Context, []SegmentPatternParam) ([][]*SegmentPattern, []error)
	SegmentsByIDs(context.Context, []int) ([]*Segment, []error)
	SegmentsByRouteID(context.Context, []SegmentParam) ([][]*Segment, []error)
	SegmentsByFeedVersionID(context.Context, []SegmentParam) ([][]*Segment, []error)
}

// RTFinder manages and looks up RT data
type RTFinder interface {
	AddData(context.Context, string, []byte) error
	FindTrip(context.Context, *Trip) *pb.TripUpdate
	MakeTrip(context.Context, *Trip) (*Trip, error)
	FindAlertsForTrip(context.Context, *Trip, *int, *bool) []*Alert
	FindAlertsForStop(context.Context, *Stop, *int, *bool) []*Alert
	FindAlertsForRoute(context.Context, *Route, *int, *bool) []*Alert
	FindAlertsForAgency(context.Context, *Agency, *int, *bool) []*Alert
	GetAddedTripsForStop(context.Context, *Stop) []*pb.TripUpdate
	FindStopTimeUpdate(context.Context, *Trip, *StopTime) (*RTStopTimeUpdate, bool)
	// lookup cache methods
	StopTimezone(context.Context, int, string) (*time.Location, bool)
	GetGtfsTripID(context.Context, int) (string, bool)
	GetMessage(context.Context, string, string) (*pb.FeedMessage, bool)
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
