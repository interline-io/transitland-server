package gql

// import graph gophers with your other imports
import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	dataloader "github.com/graph-gophers/dataloader/v7"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-server/model"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
	waitTime   = 2 * time.Millisecond
	maxBatch   = 100
)

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	AgenciesByFeedVersionIDs                                      *dataloader.Loader[agencyLoaderParam, []*model.Agency]
	AgenciesByIDs                                                 *dataloader.Loader[int, *model.Agency]
	AgenciesByOnestopIDs                                          *dataloader.Loader[agencyLoaderParam, []*model.Agency]
	AgencyPlacesByAgencyIDs                                       *dataloader.Loader[agencyPlaceLoaderParam, []*model.AgencyPlace]
	CalendarDatesByServiceIDs                                     *dataloader.Loader[calendarDateLoaderParam, []*model.CalendarDate]
	CalendarsByIDs                                                *dataloader.Loader[int, *model.Calendar]
	CensusDatasetLayersByDatasetIDs                               *dataloader.Loader[int, []string]
	CensusSourceLayersBySourceIDs                                 *dataloader.Loader[int, []string]
	CensusFieldsByTableIDs                                        *dataloader.Loader[censusFieldLoaderParam, []*model.CensusField]
	CensusGeographiesByDatasetIDs                                 *dataloader.Loader[censusDatasetGeographyLoaderParam, []*model.CensusGeography]
	CensusGeographiesByEntityIDs                                  *dataloader.Loader[censusGeographyLoaderParam, []*model.CensusGeography]
	CensusSourcesByDatasetIDs                                     *dataloader.Loader[censusSourceLoaderParam, []*model.CensusSource]
	CensusTableByIDs                                              *dataloader.Loader[int, *model.CensusTable]
	CensusValuesByGeographyIDs                                    *dataloader.Loader[censusValueLoaderParam, []*model.CensusValue]
	FeedFetchesByFeedIDs                                          *dataloader.Loader[feedFetchLoaderParam, []*model.FeedFetch]
	FeedInfosByFeedVersionIDs                                     *dataloader.Loader[feedInfoLoaderParam, []*model.FeedInfo]
	FeedsByIDs                                                    *dataloader.Loader[int, *model.Feed]
	FeedsByOperatorOnestopIDs                                     *dataloader.Loader[feedLoaderParam, []*model.Feed]
	FeedStatesByFeedIDs                                           *dataloader.Loader[int, *model.FeedState]
	FeedVersionFileInfosByFeedVersionIDs                          *dataloader.Loader[feedVersionFileInfoLoaderParam, []*model.FeedVersionFileInfo]
	FeedVersionGeometryByIDs                                      *dataloader.Loader[int, *tt.Polygon]
	FeedVersionGtfsImportByFeedVersionIDs                         *dataloader.Loader[int, *model.FeedVersionGtfsImport]
	FeedVersionsByFeedIDs                                         *dataloader.Loader[feedVersionLoaderParam, []*model.FeedVersion]
	FeedVersionsByIDs                                             *dataloader.Loader[int, *model.FeedVersion]
	FeedVersionServiceLevelsByFeedVersionIDs                      *dataloader.Loader[feedVersionServiceLevelLoaderParam, []*model.FeedVersionServiceLevel]
	FeedVersionServiceWindowByFeedVersionIDs                      *dataloader.Loader[int, *model.FeedVersionServiceWindow]
	FrequenciesByTripIDs                                          *dataloader.Loader[frequencyLoaderParam, []*model.Frequency]
	LevelsByIDs                                                   *dataloader.Loader[int, *model.Level]
	LevelsByParentStationIDs                                      *dataloader.Loader[levelLoaderParam, []*model.Level]
	OperatorsByAgencyIDs                                          *dataloader.Loader[int, *model.Operator]
	OperatorsByCOIFs                                              *dataloader.Loader[int, *model.Operator]
	OperatorsByFeedIDs                                            *dataloader.Loader[operatorLoaderParam, []*model.Operator]
	PathwaysByFromStopIDs                                         *dataloader.Loader[pathwayLoaderParam, []*model.Pathway]
	PathwaysByIDs                                                 *dataloader.Loader[int, *model.Pathway]
	PathwaysByToStopID                                            *dataloader.Loader[pathwayLoaderParam, []*model.Pathway]
	RouteAttributesByRouteIDs                                     *dataloader.Loader[int, *model.RouteAttribute]
	RouteGeometriesByRouteIDs                                     *dataloader.Loader[routeGeometryLoaderParam, []*model.RouteGeometry]
	RouteHeadwaysByRouteIDs                                       *dataloader.Loader[routeHeadwayLoaderParam, []*model.RouteHeadway]
	RoutesByAgencyIDs                                             *dataloader.Loader[routeLoaderParam, []*model.Route]
	RoutesByFeedVersionIDs                                        *dataloader.Loader[routeLoaderParam, []*model.Route]
	RoutesByIDs                                                   *dataloader.Loader[int, *model.Route]
	RouteStopPatternsByRouteIDs                                   *dataloader.Loader[routeStopPatternLoaderParam, []*model.RouteStopPattern]
	RouteStopsByRouteIDs                                          *dataloader.Loader[routeStopLoaderParam, []*model.RouteStop]
	RouteStopsByStopIDs                                           *dataloader.Loader[routeStopLoaderParam, []*model.RouteStop]
	SegmentPatternsByRouteIDs                                     *dataloader.Loader[segmentPatternLoaderParam, []*model.SegmentPattern]
	SegmentPatternsBySegmentIDs                                   *dataloader.Loader[segmentPatternLoaderParam, []*model.SegmentPattern]
	SegmentsByFeedVersionIDs                                      *dataloader.Loader[segmentLoaderParam, []*model.Segment]
	SegmentsByIDs                                                 *dataloader.Loader[int, *model.Segment]
	SegmentsByRouteIDs                                            *dataloader.Loader[segmentLoaderParam, []*model.Segment]
	ShapesByIDs                                                   *dataloader.Loader[int, *model.Shape]
	StopExternalReferencesByStopIDs                               *dataloader.Loader[int, *model.StopExternalReference]
	StopObservationsByStopIDs                                     *dataloader.Loader[stopObservationLoaderParam, []*model.StopObservation]
	StopPlacesByStopID                                            *dataloader.Loader[model.StopPlaceParam, *model.StopPlace]
	StopsByFeedVersionIDs                                         *dataloader.Loader[stopLoaderParam, []*model.Stop]
	StopsByIDs                                                    *dataloader.Loader[int, *model.Stop]
	StopsByLevelIDs                                               *dataloader.Loader[stopLoaderParam, []*model.Stop]
	StopsByParentStopIDs                                          *dataloader.Loader[stopLoaderParam, []*model.Stop]
	StopsByRouteIDs                                               *dataloader.Loader[stopLoaderParam, []*model.Stop]
	StopTimesByStopIDs                                            *dataloader.Loader[stopTimeLoaderParam, []*model.StopTime]
	StopTimesByTripIDs                                            *dataloader.Loader[tripStopTimeLoaderParam, []*model.StopTime]
	TargetStopsByStopIDs                                          *dataloader.Loader[int, *model.Stop]
	TripsByFeedVersionIDs                                         *dataloader.Loader[tripLoaderParam, []*model.Trip]
	TripsByIDs                                                    *dataloader.Loader[int, *model.Trip]
	TripsByRouteIDs                                               *dataloader.Loader[tripLoaderParam, []*model.Trip]
	ValidationReportErrorExemplarsByValidationReportErrorGroupIDs *dataloader.Loader[validationReportErrorExemplarLoaderParam, []*model.ValidationReportError]
	ValidationReportErrorGroupsByValidationReportIDs              *dataloader.Loader[validationReportErrorGroupLoaderParam, []*model.ValidationReportErrorGroup]
	ValidationReportsByFeedVersionIDs                             *dataloader.Loader[validationReportLoaderParam, []*model.ValidationReport]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(dbf model.Finder, batchSize int, stopTimeBatchSize int) *Loaders {
	if batchSize == 0 {
		batchSize = maxBatch
	}
	if stopTimeBatchSize == 0 {
		stopTimeBatchSize = maxBatch
	}
	loaders := &Loaders{
		AgenciesByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []agencyLoaderParam) ([][]*model.Agency, []error) {
				return paramGroupQuery(
					params,
					func(p agencyLoaderParam) (int, *model.AgencyFilter, *int) {
						return p.FeedVersionID, p.Where, p.Limit
					},
					func(keys []int, where *model.AgencyFilter, limit *int) (ents []*model.Agency, err error) {
						return dbf.AgenciesByFeedVersionIDs(ctx, limit, where, keys)
					},
					func(ent *model.Agency) int {
						return ent.FeedVersionID
					},
				)
			},
		),
		AgenciesByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.AgenciesByIDs),
		AgenciesByOnestopIDs: withWaitAndCapacity(waitTime, batchSize,
			func(ctx context.Context, params []agencyLoaderParam) ([][]*model.Agency, []error) {
				return paramGroupQuery(
					params,
					func(p agencyLoaderParam) (string, *model.AgencyFilter, *int) {
						a := ""
						if p.OnestopID != nil {
							a = *p.OnestopID
						}
						return a, p.Where, p.Limit
					},
					func(keys []string, where *model.AgencyFilter, limit *int) (ents []*model.Agency, err error) {
						return dbf.AgenciesByOnestopIDs(ctx, limit, where, keys)
					},
					func(ent *model.Agency) string {
						return ent.OnestopID
					},
				)
			}),
		AgencyPlacesByAgencyIDs: withWaitAndCapacity(waitTime, batchSize,
			func(ctx context.Context, params []agencyPlaceLoaderParam) ([][]*model.AgencyPlace, []error) {
				return paramGroupQuery(
					params,
					func(p agencyPlaceLoaderParam) (int, *model.AgencyPlaceFilter, *int) {
						return p.AgencyID, p.Where, p.Limit
					},
					func(keys []int, where *model.AgencyPlaceFilter, limit *int) (ents []*model.AgencyPlace, err error) {
						return dbf.AgencyPlacesByAgencyIDs(ctx, limit, where, keys)
					},
					func(ent *model.AgencyPlace) int {
						return ent.AgencyID
					},
				)

			}),
		CalendarDatesByServiceIDs: withWaitAndCapacity(waitTime, batchSize,
			func(ctx context.Context, params []calendarDateLoaderParam) ([][]*model.CalendarDate, []error) {
				return paramGroupQuery(
					params,
					func(p calendarDateLoaderParam) (int, *model.CalendarDateFilter, *int) {
						return p.ServiceID, p.Where, p.Limit
					},
					func(keys []int, where *model.CalendarDateFilter, limit *int) (ents []*model.CalendarDate, err error) {
						return dbf.CalendarDatesByServiceIDs(ctx, limit, where, keys)
					},
					func(ent *model.CalendarDate) int {
						return ent.ServiceID.Int()
					},
				)
			}),
		CalendarsByIDs:                  withWaitAndCapacity(waitTime, batchSize, dbf.CalendarsByIDs),
		CensusDatasetLayersByDatasetIDs: withWaitAndCapacity(waitTime, batchSize, dbf.CensusDatasetLayersByDatasetIDs),
		CensusSourceLayersBySourceIDs:   withWaitAndCapacity(waitTime, batchSize, dbf.CensusSourceLayersBySourceIDs),
		CensusFieldsByTableIDs: withWaitAndCapacity(waitTime, batchSize,
			func(ctx context.Context, params []censusFieldLoaderParam) ([][]*model.CensusField, []error) {
				return paramGroupQuery(
					params,
					func(p censusFieldLoaderParam) (int, *censusFieldLoaderParam, *int) {
						return p.TableID, nil, p.Limit
					},
					func(keys []int, where *censusFieldLoaderParam, limit *int) (ents []*model.CensusField, err error) {
						return dbf.CensusFieldsByTableIDs(ctx, limit, keys)
					},
					func(ent *model.CensusField) int {
						return ent.TableID
					},
				)
			}),
		CensusGeographiesByDatasetIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []censusDatasetGeographyLoaderParam) ([][]*model.CensusGeography, []error) {
				return paramGroupQuery(
					params,
					func(p censusDatasetGeographyLoaderParam) (int, *model.CensusDatasetGeographyFilter, *int) {
						return p.DatasetID, p.Where, p.Limit
					},
					func(keys []int, p *model.CensusDatasetGeographyFilter, limit *int) (ents []*model.CensusGeography, err error) {
						return dbf.CensusGeographiesByDatasetIDs(ctx, limit, p, keys)
					},
					func(ent *model.CensusGeography) int {
						return ent.DatasetID
					},
				)
			}),
		CensusGeographiesByEntityIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []censusGeographyLoaderParam) ([][]*model.CensusGeography, []error) {
				return paramGroupQuery(
					params,
					func(p censusGeographyLoaderParam) (int, *censusGeographyLoaderParam, *int) {
						rp := censusGeographyLoaderParam{
							EntityType: p.EntityType,
							Where:      p.Where,
						}
						return p.EntityID, &rp, p.Limit
					},
					func(keys []int, param *censusGeographyLoaderParam, limit *int) (ents []*model.CensusGeography, err error) {
						return dbf.CensusGeographiesByEntityIDs(ctx, limit, param.Where, param.EntityType, keys)
					},
					func(ent *model.CensusGeography) int {
						return ent.MatchEntityID
					},
				)
			}),
		CensusSourcesByDatasetIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []censusSourceLoaderParam) ([][]*model.CensusSource, []error) {
				return paramGroupQuery(
					params,
					func(p censusSourceLoaderParam) (int, *model.CensusSourceFilter, *int) {
						return p.DatasetID, p.Where, p.Limit
					},
					func(keys []int, where *model.CensusSourceFilter, limit *int) (ents []*model.CensusSource, err error) {
						return dbf.CensusSourcesByDatasetIDs(ctx, limit, where, keys)
					},
					func(ent *model.CensusSource) int {
						return ent.DatasetID
					},
				)
			}),
		CensusTableByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.CensusTableByIDs),
		CensusValuesByGeographyIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []censusValueLoaderParam) ([][]*model.CensusValue, []error) {
				return paramGroupQuery(
					params,
					func(p censusValueLoaderParam) (string, string, *int) {
						return p.Geoid, p.TableNames, p.Limit
					},
					func(keys []string, tableNames string, limit *int) (ents []*model.CensusValue, err error) {
						return nil, nil
						// return dbf.CensusValuesByGeographyIDs(ctx, limit, where, keys)
					},
					func(ent *model.CensusValue) string {
						return ent.Geoid
					},
				)
			}),
		FeedFetchesByFeedIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []feedFetchLoaderParam) ([][]*model.FeedFetch, []error) {
				return paramGroupQuery(
					params,
					func(p feedFetchLoaderParam) (int, *model.FeedFetchFilter, *int) {
						return p.FeedID, p.Where, p.Limit
					},
					func(keys []int, where *model.FeedFetchFilter, limit *int) (ents []*model.FeedFetch, err error) {
						return dbf.FeedFetchesByFeedIDs(ctx, limit, where, keys)
					},
					func(ent *model.FeedFetch) int {
						return ent.FeedID
					},
				)
			}),
		FeedInfosByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []feedInfoLoaderParam) ([][]*model.FeedInfo, []error) {
				return paramGroupQuery(
					params,
					func(p feedInfoLoaderParam) (int, bool, *int) {
						return p.FeedVersionID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.FeedInfo, err error) {
						return dbf.FeedInfosByFeedVersionIDs(ctx, limit, keys)
					},
					func(ent *model.FeedInfo) int {
						return ent.FeedVersionID
					},
				)
			}),
		FeedsByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.FeedsByIDs),
		FeedsByOperatorOnestopIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []feedLoaderParam) ([][]*model.Feed, []error) {
				return paramGroupQuery(
					params,
					func(p feedLoaderParam) (string, *model.FeedFilter, *int) {
						return p.OperatorOnestopID, p.Where, p.Limit
					},
					func(keys []string, where *model.FeedFilter, limit *int) (ents []*model.Feed, err error) {
						return dbf.FeedsByOperatorOnestopIDs(ctx, limit, where, keys)
					},
					func(ent *model.Feed) string {
						return ent.WithOperatorOnestopID.String()
					},
				)
			}),
		FeedStatesByFeedIDs: withWaitAndCapacity(waitTime, batchSize, dbf.FeedStatesByFeedIDs),
		FeedVersionFileInfosByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []feedVersionFileInfoLoaderParam) ([][]*model.FeedVersionFileInfo, []error) {
				return paramGroupQuery(
					params,
					func(p feedVersionFileInfoLoaderParam) (int, bool, *int) {
						return p.FeedVersionID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.FeedVersionFileInfo, err error) {
						return dbf.FeedVersionFileInfosByFeedVersionIDs(ctx, limit, keys)
					},
					func(ent *model.FeedVersionFileInfo) int {
						return ent.FeedVersionID
					},
				)
			}),
		FeedVersionGeometryByIDs:              withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionGeometryByIDs),
		FeedVersionGtfsImportByFeedVersionIDs: withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionGtfsImportByFeedVersionIDs),
		FeedVersionsByFeedIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []feedVersionLoaderParam) ([][]*model.FeedVersion, []error) {
				return paramGroupQuery(
					params,
					func(p feedVersionLoaderParam) (int, *model.FeedVersionFilter, *int) {
						return p.FeedID, p.Where, p.Limit
					},
					func(keys []int, where *model.FeedVersionFilter, limit *int) ([]*model.FeedVersion, error) {
						return dbf.FeedVersionsByFeedIDs(ctx, limit, where, keys)
					},
					func(ent *model.FeedVersion) int {
						return ent.FeedID
					},
				)
			}),
		FeedVersionsByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionsByIDs),
		FeedVersionServiceLevelsByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []feedVersionServiceLevelLoaderParam) ([][]*model.FeedVersionServiceLevel, []error) {
				return paramGroupQuery(
					params,
					func(p feedVersionServiceLevelLoaderParam) (int, *model.FeedVersionServiceLevelFilter, *int) {
						return p.FeedVersionID, p.Where, p.Limit
					},
					func(keys []int, where *model.FeedVersionServiceLevelFilter, limit *int) (ents []*model.FeedVersionServiceLevel, err error) {
						return dbf.FeedVersionServiceLevelsByFeedVersionIDs(ctx, limit, where, keys)
					},
					func(ent *model.FeedVersionServiceLevel) int {
						return ent.FeedVersionID
					},
				)
			}),
		FeedVersionServiceWindowByFeedVersionIDs: withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionServiceWindowByFeedVersionIDs),
		FrequenciesByTripIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []frequencyLoaderParam) ([][]*model.Frequency, []error) {
				return paramGroupQuery(
					params,
					func(p frequencyLoaderParam) (int, bool, *int) {
						return p.TripID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.Frequency, err error) {
						return dbf.FrequenciesByTripIDs(ctx, limit, keys)
					},
					func(ent *model.Frequency) int {
						return ent.TripID.Int()
					},
				)
			}),
		LevelsByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.LevelsByIDs),
		LevelsByParentStationIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []levelLoaderParam) ([][]*model.Level, []error) {
				return paramGroupQuery(
					params,
					func(p levelLoaderParam) (int, bool, *int) {
						return p.ParentStationID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.Level, err error) {
						return dbf.LevelsByParentStationIDs(ctx, limit, keys)
					},
					func(ent *model.Level) int {
						return ent.ParentStation.Int()
					},
				)
			}),
		OperatorsByAgencyIDs: withWaitAndCapacity(waitTime, batchSize, dbf.OperatorsByAgencyIDs),
		OperatorsByCOIFs:     withWaitAndCapacity(waitTime, batchSize, dbf.OperatorsByCOIFs),
		OperatorsByFeedIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []operatorLoaderParam) ([][]*model.Operator, []error) {
				return paramGroupQuery(
					params,
					func(p operatorLoaderParam) (int, *model.OperatorFilter, *int) {
						return p.FeedID, p.Where, p.Limit
					},
					func(keys []int, where *model.OperatorFilter, limit *int) (ents []*model.Operator, err error) {
						return dbf.OperatorsByFeedIDs(ctx, limit, where, keys)
					},
					func(ent *model.Operator) int {
						return ent.FeedID
					},
				)
			},
		),
		PathwaysByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.PathwaysByIDs),
		PathwaysByFromStopIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []pathwayLoaderParam) ([][]*model.Pathway, []error) {
				return paramGroupQuery(
					params,
					func(p pathwayLoaderParam) (int, *model.PathwayFilter, *int) {
						return p.FromStopID, p.Where, p.Limit
					},
					func(keys []int, where *model.PathwayFilter, limit *int) (ents []*model.Pathway, err error) {
						return dbf.PathwaysByFromStopIDs(ctx, limit, where, keys)
					},
					func(ent *model.Pathway) int {
						return ent.FromStopID.Int()
					},
				)
			}),
		PathwaysByToStopID: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []pathwayLoaderParam) ([][]*model.Pathway, []error) {
				return paramGroupQuery(
					params,
					func(p pathwayLoaderParam) (int, *model.PathwayFilter, *int) {
						return p.ToStopID, p.Where, p.Limit
					},
					func(keys []int, where *model.PathwayFilter, limit *int) (ents []*model.Pathway, err error) {
						return dbf.PathwaysByToStopIDs(ctx, limit, where, keys)
					},
					func(ent *model.Pathway) int {
						return ent.FromStopID.Int()
					},
				)
			}),
		RouteAttributesByRouteIDs: withWaitAndCapacity(waitTime, batchSize, dbf.RouteAttributesByRouteIDs),
		RouteGeometriesByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []routeGeometryLoaderParam) ([][]*model.RouteGeometry, []error) {
				return paramGroupQuery(
					params,
					func(p routeGeometryLoaderParam) (int, bool, *int) {
						return p.RouteID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.RouteGeometry, err error) {
						return dbf.RouteGeometriesByRouteIDs(ctx, limit, keys)
					},
					func(ent *model.RouteGeometry) int {
						return ent.RouteID
					},
				)
			}),
		RouteHeadwaysByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []routeHeadwayLoaderParam) ([][]*model.RouteHeadway, []error) {
				return paramGroupQuery(
					params,
					func(p routeHeadwayLoaderParam) (int, bool, *int) {
						return p.RouteID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.RouteHeadway, err error) {
						return dbf.RouteHeadwaysByRouteIDs(ctx, limit, keys)
					},
					func(ent *model.RouteHeadway) int {
						return ent.RouteID
					},
				)
			}),
		RoutesByAgencyIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []routeLoaderParam) ([][]*model.Route, []error) {
				return paramGroupQuery(
					params,
					func(p routeLoaderParam) (int, *model.RouteFilter, *int) {
						return p.AgencyID, p.Where, p.Limit
					},
					func(keys []int, where *model.RouteFilter, limit *int) (ents []*model.Route, err error) {
						return dbf.RoutesByAgencyIDs(ctx, limit, where, keys)
					},
					func(ent *model.Route) int {
						return ent.AgencyID.Int()
					},
				)
			}),
		RoutesByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []routeLoaderParam) ([][]*model.Route, []error) {
				return paramGroupQuery(
					params,
					func(p routeLoaderParam) (int, *model.RouteFilter, *int) {
						return p.FeedVersionID, p.Where, p.Limit
					},
					func(keys []int, where *model.RouteFilter, limit *int) (ents []*model.Route, err error) {
						return dbf.RoutesByFeedVersionIDs(ctx, limit, where, keys)
					},
					func(ent *model.Route) int {
						return ent.FeedVersionID
					},
				)
			}),
		RoutesByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.RoutesByIDs),
		RouteStopPatternsByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []routeStopPatternLoaderParam) ([][]*model.RouteStopPattern, []error) {
				return paramGroupQuery(
					params,
					func(p routeStopPatternLoaderParam) (int, bool, *int) {
						return p.RouteID, false, nil
					},
					func(keys []int, where bool, limit *int) (ents []*model.RouteStopPattern, err error) {
						return dbf.RouteStopPatternsByRouteIDs(ctx, limit, keys)
					},
					func(ent *model.RouteStopPattern) int {
						return ent.RouteID
					},
				)
			}),
		RouteStopsByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []routeStopLoaderParam) ([][]*model.RouteStop, []error) {
				return paramGroupQuery(
					params,
					func(p routeStopLoaderParam) (int, bool, *int) {
						return p.RouteID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.RouteStop, err error) {
						return dbf.RouteStopsByRouteIDs(ctx, limit, keys)
					},
					func(ent *model.RouteStop) int {
						return ent.RouteID
					},
				)
			}),
		RouteStopsByStopIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []routeStopLoaderParam) ([][]*model.RouteStop, []error) {
				return paramGroupQuery(
					params,
					func(p routeStopLoaderParam) (int, bool, *int) {
						return p.StopID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) (ents []*model.RouteStop, err error) {
						return dbf.RouteStopsByStopIDs(ctx, limit, keys)
					},
					func(ent *model.RouteStop) int {
						return ent.StopID
					},
				)
			}),
		SegmentPatternsByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []segmentPatternLoaderParam) ([][]*model.SegmentPattern, []error) {
				return paramGroupQuery(
					params,
					func(p segmentPatternLoaderParam) (int, *model.SegmentPatternFilter, *int) {
						return p.RouteID, p.Where, p.Limit
					},
					func(keys []int, where *model.SegmentPatternFilter, limit *int) (ents []*model.SegmentPattern, err error) {
						return dbf.SegmentPatternsByRouteIDs(ctx, limit, where, keys)
					},
					func(ent *model.SegmentPattern) int {
						return ent.RouteID
					},
				)
			},
		),
		SegmentPatternsBySegmentIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []segmentPatternLoaderParam) ([][]*model.SegmentPattern, []error) {
				return paramGroupQuery(
					params,
					func(p segmentPatternLoaderParam) (int, *model.SegmentPatternFilter, *int) {
						return p.SegmentID, p.Where, p.Limit
					},
					func(keys []int, where *model.SegmentPatternFilter, limit *int) (ents []*model.SegmentPattern, err error) {
						return dbf.SegmentPatternsBySegmentIDs(ctx, limit, where, keys)
					},
					func(ent *model.SegmentPattern) int {
						return ent.SegmentID
					},
				)
			},
		),
		SegmentsByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []segmentLoaderParam) ([][]*model.Segment, []error) {
				return paramGroupQuery(
					params,
					func(p segmentLoaderParam) (int, *model.SegmentFilter, *int) {
						return p.FeedVersionID, p.Where, p.Limit
					},
					func(keys []int, where *model.SegmentFilter, limit *int) (ents []*model.Segment, err error) {
						return dbf.SegmentsByFeedVersionIDs(ctx, limit, where, keys)
					},
					func(ent *model.Segment) int {
						return ent.FeedVersionID
					},
				)
			},
		),
		SegmentsByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.SegmentsByIDs),
		SegmentsByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []segmentLoaderParam) ([][]*model.Segment, []error) {
				return paramGroupQuery(
					params,
					func(p segmentLoaderParam) (int, *model.SegmentFilter, *int) {
						return p.RouteID, p.Where, p.Limit
					},
					func(keys []int, where *model.SegmentFilter, limit *int) (ents []*model.Segment, err error) {
						return dbf.SegmentsByRouteIDs(ctx, limit, where, keys)
					},
					func(ent *model.Segment) int {
						return ent.WithRouteID
					},
				)
			},
		),
		ShapesByIDs:                     withWaitAndCapacity(waitTime, batchSize, dbf.ShapesByIDs),
		StopExternalReferencesByStopIDs: withWaitAndCapacity(waitTime, batchSize, dbf.StopExternalReferencesByStopIDs),
		StopObservationsByStopIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []stopObservationLoaderParam) ([][]*model.StopObservation, []error) {
				return paramGroupQuery(
					params,
					func(p stopObservationLoaderParam) (int, *model.StopObservationFilter, *int) {
						return p.StopID, p.Where, p.Limit
					},
					func(keys []int, where *model.StopObservationFilter, limit *int) (ents []*model.StopObservation, err error) {
						return dbf.StopObservationsByStopIDs(ctx, limit, where, keys)
					},
					func(ent *model.StopObservation) int {
						return ent.StopID
					},
				)
			}),
		StopPlacesByStopID: withWaitAndCapacity(waitTime, batchSize, dbf.StopPlacesByStopID),
		StopsByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []stopLoaderParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p stopLoaderParam) (int, *model.StopFilter, *int) {
						return p.FeedVersionID, p.Where, p.Limit
					},
					func(keys []int, where *model.StopFilter, limit *int) (ents []*model.Stop, err error) {
						return dbf.StopsByFeedVersionIDs(ctx, limit, where, keys)
					},
					func(ent *model.Stop) int {
						return ent.FeedVersionID
					},
				)
			}),
		StopsByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.StopsByIDs),
		StopsByLevelIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []stopLoaderParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p stopLoaderParam) (int, *model.StopFilter, *int) {
						return p.LevelID, p.Where, p.Limit
					},
					func(keys []int, where *model.StopFilter, limit *int) (ents []*model.Stop, err error) {
						return dbf.StopsByLevelIDs(ctx, limit, where, keys)
					},
					func(ent *model.Stop) int {
						return ent.LevelID.Int()
					},
				)
			}),
		StopsByParentStopIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []stopLoaderParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p stopLoaderParam) (int, *model.StopFilter, *int) {
						return p.ParentStopID, p.Where, p.Limit
					},
					func(keys []int, where *model.StopFilter, limit *int) (ents []*model.Stop, err error) {
						return dbf.StopsByParentStopIDs(ctx, limit, where, keys)
					},
					func(ent *model.Stop) int {
						return ent.ParentStation.Int()
					},
				)
			}),
		StopsByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []stopLoaderParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p stopLoaderParam) (int, *model.StopFilter, *int) {
						return p.RouteID, p.Where, p.Limit
					},
					func(keys []int, where *model.StopFilter, limit *int) (ents []*model.Stop, err error) {
						return dbf.StopsByRouteIDs(ctx, limit, where, keys)
					},
					func(ent *model.Stop) int {
						return ent.WithRouteID.Int()
					},
				)
			}),
		StopTimesByStopIDs: withWaitAndCapacity(
			waitTime,
			stopTimeBatchSize,
			func(ctx context.Context, params []stopTimeLoaderParam) ([][]*model.StopTime, []error) {
				return paramGroupQuery(
					params,
					func(p stopTimeLoaderParam) (model.FVPair, *model.StopTimeFilter, *int) {
						return model.FVPair{FeedVersionID: p.FeedVersionID, EntityID: p.StopID}, p.Where, p.Limit
					},
					func(keys []model.FVPair, where *model.StopTimeFilter, limit *int) (ents []*model.StopTime, err error) {
						return dbf.StopTimesByStopIDs(ctx, limit, where, keys)
					},
					func(ent *model.StopTime) model.FVPair {
						return model.FVPair{FeedVersionID: ent.FeedVersionID, EntityID: ent.StopID.Int()}
					},
				)
			},
		),
		StopTimesByTripIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []tripStopTimeLoaderParam) ([][]*model.StopTime, []error) {
				return paramGroupQuery(
					params,
					func(p tripStopTimeLoaderParam) (model.FVPair, *model.TripStopTimeFilter, *int) {
						return model.FVPair{FeedVersionID: p.FeedVersionID, EntityID: p.TripID}, p.Where, p.Limit
					},
					func(keys []model.FVPair, where *model.TripStopTimeFilter, limit *int) (ents []*model.StopTime, err error) {
						return dbf.StopTimesByTripIDs(ctx, limit, where, keys)
					},
					func(ent *model.StopTime) model.FVPair {
						return model.FVPair{FeedVersionID: ent.FeedVersionID, EntityID: ent.TripID.Int()}
					},
				)
			},
		),
		TargetStopsByStopIDs: withWaitAndCapacity(waitTime, batchSize, dbf.TargetStopsByStopIDs),
		TripsByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []tripLoaderParam) ([][]*model.Trip, []error) {
				return paramGroupQuery(
					params,
					func(p tripLoaderParam) (int, *model.TripFilter, *int) {
						return p.FeedVersionID, p.Where, p.Limit
					},
					func(keys []int, where *model.TripFilter, limit *int) (ents []*model.Trip, err error) {
						return dbf.TripsByFeedVersionIDs(ctx, limit, where, keys)
					},
					func(ent *model.Trip) int {
						return ent.FeedVersionID
					},
				)
			},
		),
		TripsByIDs: withWaitAndCapacity(waitTime, batchSize, dbf.TripsByIDs),
		TripsByRouteIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []tripLoaderParam) ([][]*model.Trip, []error) {
				return paramGroupQuery(
					params,
					func(p tripLoaderParam) (model.FVPair, *model.TripFilter, *int) {
						return model.FVPair{EntityID: p.RouteID, FeedVersionID: p.FeedVersionID}, p.Where, p.Limit
					},
					func(keys []model.FVPair, where *model.TripFilter, limit *int) (ents []*model.Trip, err error) {
						return dbf.TripsByRouteIDs(ctx, limit, where, keys)
					},
					func(ent *model.Trip) model.FVPair {
						return model.FVPair{EntityID: ent.RouteID.Int(), FeedVersionID: ent.FeedVersionID}
					},
				)
			},
		),
		ValidationReportErrorExemplarsByValidationReportErrorGroupIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []validationReportErrorExemplarLoaderParam) ([][]*model.ValidationReportError, []error) {
				return paramGroupQuery(
					params,
					func(p validationReportErrorExemplarLoaderParam) (int, bool, *int) {
						return p.ValidationReportGroupID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) ([]*model.ValidationReportError, error) {
						return dbf.ValidationReportErrorExemplarsByValidationReportErrorGroupIDs(ctx, limit, keys)
					},
					func(ent *model.ValidationReportError) int { return ent.ValidationReportErrorGroupID },
				)
			},
		),
		ValidationReportErrorGroupsByValidationReportIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []validationReportErrorGroupLoaderParam) ([][]*model.ValidationReportErrorGroup, []error) {
				return paramGroupQuery(
					params,
					func(p validationReportErrorGroupLoaderParam) (int, bool, *int) {
						return p.ValidationReportID, false, p.Limit
					},
					func(keys []int, where bool, limit *int) ([]*model.ValidationReportErrorGroup, error) {
						return dbf.ValidationReportErrorGroupsByValidationReportIDs(ctx, limit, keys)
					},
					func(ent *model.ValidationReportErrorGroup) int { return ent.ValidationReportID },
				)
			},
		),
		ValidationReportsByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []validationReportLoaderParam) ([][]*model.ValidationReport, []error) {
				return paramGroupQuery(
					params,
					func(p validationReportLoaderParam) (int, *model.ValidationReportFilter, *int) {
						return p.FeedVersionID, p.Where, p.Limit
					},
					func(keys []int, where *model.ValidationReportFilter, limit *int) ([]*model.ValidationReport, error) {
						return dbf.ValidationReportsByFeedVersionIDs(ctx, limit, where, keys)
					},
					func(ent *model.ValidationReport) int { return ent.FeedVersionID },
				)

			},
		),
	}
	return loaders
}

func loaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is per request scoped loaders/cache
		// Is this OK to use as a long term cache?
		ctx := r.Context()
		cfg := model.ForContext(ctx)
		loaders := NewLoaders(cfg.Finder, cfg.LoaderBatchSize, cfg.LoaderStopTimeBatchSize)
		nextCtx := context.WithValue(ctx, loadersKey, loaders)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
}

// LoaderFor returns the dataloader for a given context
func LoaderFor(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// withWait is a helper that sets a default with time, with less manually specifying type params
func withWaitAndCapacity[
	T any,
	ParamT comparable,
](d time.Duration, size int, cb func(context.Context, []ParamT) ([]T, []error)) *dataloader.Loader[ParamT, T] {
	return dataloader.NewBatchedLoader(unwrapResult(cb), dataloader.WithWait[ParamT, T](d), dataloader.WithBatchCapacity[ParamT, T](size))
}

// unwrap function adapts existing Finder methods to dataloader Result type
func unwrapResult[
	T any,
	ParamT comparable,
](
	cb func(context.Context, []ParamT) ([]T, []error),
) func(context.Context, []ParamT) []*dataloader.Result[T] {
	x := func(ctx context.Context, ps []ParamT) []*dataloader.Result[T] {
		a, errs := cb(ctx, ps)
		if len(a) != len(ps) {
			log.For(ctx).Trace().Msgf("error in dataloader, result len %d did not match param length %d", len(a), len(ps))
			return nil
		}
		ret := make([]*dataloader.Result[T], len(ps))
		for idx := range ps {
			var err error
			if idx < len(errs) {
				err = errs[idx]
			}
			var data T
			if idx < len(a) {
				data = a[idx]
			}
			ret[idx] = &dataloader.Result[T]{Data: data, Error: err}
		}
		return ret
	}
	return x
}

////////////

// Multiple param sets

func paramGroupQuery[
	K comparable,
	P any,
	W any,
	R any,
](
	params []P,
	paramFunc func(P) (K, W, *int),
	queryFunc func([]K, W, *int) ([]*R, error),
	keyFunc func(*R) K,
) ([][]*R, []error) {
	// Create return value
	ret := make([][]*R, len(params))
	errs := make([]error, len(params))

	// Group params by JSON representation
	type paramGroupItem[K comparable, M any] struct {
		Limit *int
		Where M
	}
	type paramGroup[K comparable, M any] struct {
		Index []int
		Keys  []K
		Limit *int
		Where M
	}
	paramGroups := map[string]paramGroup[K, W]{}
	for i, param := range params {
		// Get values from supplied func
		key, where, limit := paramFunc(param)

		// Convert to paramGroupItem
		item := paramGroupItem[K, W]{
			Limit: limit,
			Where: where,
		}

		// Use the JSON representation of Where and Limit as the key
		jj, err := json.Marshal(paramGroupItem[K, W]{Where: item.Where, Limit: item.Limit})
		if err != nil {
			// TODO: log and expand error
			errs[i] = err
			continue
		}
		paramGroupKey := string(jj)

		// Add index and key
		a, ok := paramGroups[paramGroupKey]
		if !ok {
			a = paramGroup[K, W]{Where: item.Where, Limit: item.Limit}
		}
		a.Index = append(a.Index, i)
		a.Keys = append(a.Keys, key)
		paramGroups[paramGroupKey] = a
	}

	// Process each param group
	for _, pgroup := range paramGroups {
		// Run query function
		ents, err := queryFunc(pgroup.Keys, pgroup.Where, pgroup.Limit)

		// Group using keyFunc and merge into output
		limit := uint64(1000)
		if a := checkLimit(pgroup.Limit); a != nil {
			limit = uint64(*a)
		}
		bykey := map[K][]*R{}
		for _, ent := range ents {
			key := keyFunc(ent)
			bykey[key] = append(bykey[key], ent)
		}
		for keyidx, key := range pgroup.Keys {
			idx := pgroup.Index[keyidx]
			gi := bykey[key]
			if err != nil {
				errs[idx] = err
			}
			if uint64(len(gi)) <= limit {
				ret[idx] = gi
			} else {
				ret[idx] = gi[0:limit]
			}
		}
	}
	return ret, errs
}
