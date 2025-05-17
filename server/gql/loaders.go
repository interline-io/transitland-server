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
	AgenciesByFeedVersionIDs                                     *dataloader.Loader[model.AgencyParam, []*model.Agency]
	AgenciesByIDs                                                *dataloader.Loader[int, *model.Agency]
	AgenciesByOnestopIDs                                         *dataloader.Loader[model.AgencyParam, []*model.Agency]
	AgencyPlacesByAgencyIDs                                      *dataloader.Loader[model.AgencyPlaceParam, []*model.AgencyPlace]
	CalendarDatesByServiceIDs                                    *dataloader.Loader[model.CalendarDateParam, []*model.CalendarDate]
	CalendarsByIDs                                               *dataloader.Loader[int, *model.Calendar]
	CensusDatasetLayersByDatasetIDs                              *dataloader.Loader[int, []string]
	CensusSourceLayersBySourceIDs                                *dataloader.Loader[int, []string]
	CensusFieldsByTableIDs                                       *dataloader.Loader[model.CensusFieldParam, []*model.CensusField]
	CensusGeographiesByDatasetIDs                                *dataloader.Loader[model.CensusDatasetGeographyParam, []*model.CensusGeography]
	CensusGeographiesByEntityIDs                                 *dataloader.Loader[model.CensusGeographyParam, []*model.CensusGeography]
	CensusSourcesByDatasetIDs                                    *dataloader.Loader[model.CensusSourceParam, []*model.CensusSource]
	CensusTableByIDs                                             *dataloader.Loader[int, *model.CensusTable]
	CensusValuesByGeographyIDs                                   *dataloader.Loader[model.CensusValueParam, []*model.CensusValue]
	FeedFetchesByFeedIDs                                         *dataloader.Loader[model.FeedFetchParam, []*model.FeedFetch]
	FeedInfosByFeedVersionIDs                                    *dataloader.Loader[model.FeedInfoParam, []*model.FeedInfo]
	FeedsByIDs                                                   *dataloader.Loader[int, *model.Feed]
	FeedsByOperatorOnestopIDs                                    *dataloader.Loader[model.FeedParam, []*model.Feed]
	FeedStatesByFeedIDs                                          *dataloader.Loader[int, *model.FeedState]
	FeedVersionFileInfosByFeedVersionIDs                         *dataloader.Loader[model.FeedVersionFileInfoParam, []*model.FeedVersionFileInfo]
	FeedVersionGeometryByIDs                                     *dataloader.Loader[int, *tt.Polygon]
	FeedVersionGtfsImportByFeedVersionIDs                        *dataloader.Loader[int, *model.FeedVersionGtfsImport]
	FeedVersionsByFeedIDs                                        *dataloader.Loader[model.FeedVersionParam, []*model.FeedVersion]
	FeedVersionsByIDs                                            *dataloader.Loader[int, *model.FeedVersion]
	FeedVersionServiceLevelsByFeedVersionIDs                     *dataloader.Loader[model.FeedVersionServiceLevelParam, []*model.FeedVersionServiceLevel]
	FeedVersionServiceWindowByFeedVersionIDs                     *dataloader.Loader[int, *model.FeedVersionServiceWindow]
	FrequenciesByTripIDs                                         *dataloader.Loader[model.FrequencyParam, []*model.Frequency]
	LevelsByIDs                                                  *dataloader.Loader[int, *model.Level]
	LevelsByParentStationIDs                                     *dataloader.Loader[model.LevelParam, []*model.Level]
	OperatorsByAgencyIDs                                         *dataloader.Loader[int, *model.Operator]
	OperatorsByCOIFs                                             *dataloader.Loader[int, *model.Operator]
	OperatorsByFeedID                                            *dataloader.Loader[model.OperatorParam, []*model.Operator]
	PathwaysByFromStopIDs                                        *dataloader.Loader[model.PathwayParam, []*model.Pathway]
	PathwaysByIDs                                                *dataloader.Loader[int, *model.Pathway]
	PathwaysByToStopID                                           *dataloader.Loader[model.PathwayParam, []*model.Pathway]
	RouteAttributesByRouteIDs                                    *dataloader.Loader[int, *model.RouteAttribute]
	RouteGeometriesByRouteIDs                                    *dataloader.Loader[model.RouteGeometryParam, []*model.RouteGeometry]
	RouteHeadwaysByRouteIDs                                      *dataloader.Loader[model.RouteHeadwayParam, []*model.RouteHeadway]
	RoutesByAgencyIDs                                            *dataloader.Loader[model.RouteParam, []*model.Route]
	RoutesByFeedVersionIDs                                       *dataloader.Loader[model.RouteParam, []*model.Route]
	RoutesByIDs                                                  *dataloader.Loader[int, *model.Route]
	RouteStopPatternsByRouteIDs                                  *dataloader.Loader[model.RouteStopPatternParam, []*model.RouteStopPattern]
	RouteStopsByRouteIDs                                         *dataloader.Loader[model.RouteStopParam, []*model.RouteStop]
	RouteStopsByStopIDs                                          *dataloader.Loader[model.RouteStopParam, []*model.RouteStop]
	SegmentPatternsByRouteID                                     *dataloader.Loader[model.SegmentPatternParam, []*model.SegmentPattern]
	SegmentPatternsBySegmentID                                   *dataloader.Loader[model.SegmentPatternParam, []*model.SegmentPattern]
	SegmentsByFeedVersionID                                      *dataloader.Loader[model.SegmentParam, []*model.Segment]
	SegmentsByIDs                                                *dataloader.Loader[int, *model.Segment]
	SegmentsByRouteID                                            *dataloader.Loader[model.SegmentParam, []*model.Segment]
	ShapesByIDs                                                  *dataloader.Loader[int, *model.Shape]
	StopExternalReferencesByStopIDs                              *dataloader.Loader[int, *model.StopExternalReference]
	StopObservationsByStopIDs                                    *dataloader.Loader[model.StopObservationParam, []*model.StopObservation]
	StopPlacesByStopID                                           *dataloader.Loader[model.StopPlaceParam, *model.StopPlace]
	StopsByFeedVersionIDs                                        *dataloader.Loader[model.StopParam, []*model.Stop]
	StopsByIDs                                                   *dataloader.Loader[int, *model.Stop]
	StopsByLevelIDs                                              *dataloader.Loader[model.StopParam, []*model.Stop]
	StopsByParentStopIDs                                         *dataloader.Loader[model.StopParam, []*model.Stop]
	StopsByRouteIDs                                              *dataloader.Loader[model.StopParam, []*model.Stop]
	StopTimesByStopID                                            *dataloader.Loader[model.StopTimeParam, []*model.StopTime]
	StopTimesByTripID                                            *dataloader.Loader[model.TripStopTimeParam, []*model.StopTime]
	TargetStopsByStopIDs                                         *dataloader.Loader[int, *model.Stop]
	TripsByFeedVersionIDs                                        *dataloader.Loader[model.TripParam, []*model.Trip]
	TripsByIDs                                                   *dataloader.Loader[int, *model.Trip]
	TripsByRouteIDs                                              *dataloader.Loader[model.TripParam, []*model.Trip]
	ValidationReportErrorExemplarsByValidationReportErrorGroupID *dataloader.Loader[model.ValidationReportErrorExemplarParam, []*model.ValidationReportError]
	ValidationReportErrorGroupsByValidationReportID              *dataloader.Loader[model.ValidationReportErrorGroupParam, []*model.ValidationReportErrorGroup]
	ValidationReportsByFeedVersionID                             *dataloader.Loader[model.ValidationReportParam, []*model.ValidationReport]
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
			func(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
				return paramGroupQuery(
					params,
					func(p model.AgencyParam) (int, *model.AgencyFilter, *int) {
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
			func(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
				return paramGroupQuery(
					params,
					func(p model.AgencyParam) (string, *model.AgencyFilter, *int) {
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
			func(ctx context.Context, params []model.AgencyPlaceParam) ([][]*model.AgencyPlace, []error) {
				return paramGroupQuery(
					params,
					func(p model.AgencyPlaceParam) (int, *model.AgencyPlaceFilter, *int) {
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
			func(ctx context.Context, params []model.CalendarDateParam) ([][]*model.CalendarDate, []error) {
				return paramGroupQuery(
					params,
					func(p model.CalendarDateParam) (int, *model.CalendarDateFilter, *int) {
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
			func(ctx context.Context, params []model.CensusFieldParam) ([][]*model.CensusField, []error) {
				return paramGroupQuery(
					params,
					func(p model.CensusFieldParam) (int, *model.CensusFieldParam, *int) {
						return p.TableID, nil, p.Limit
					},
					func(keys []int, where *model.CensusFieldParam, limit *int) (ents []*model.CensusField, err error) {
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
			func(ctx context.Context, params []model.CensusDatasetGeographyParam) ([][]*model.CensusGeography, []error) {
				return paramGroupQuery(
					params,
					func(p model.CensusDatasetGeographyParam) (int, *model.CensusDatasetGeographyFilter, *int) {
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
			func(ctx context.Context, params []model.CensusGeographyParam) ([][]*model.CensusGeography, []error) {
				return paramGroupQuery(
					params,
					func(p model.CensusGeographyParam) (int, *model.CensusGeographyParam, *int) {
						rp := model.CensusGeographyParam{
							EntityType: p.EntityType,
							Where:      p.Where,
						}
						return p.EntityID, &rp, p.Limit
					},
					func(keys []int, param *model.CensusGeographyParam, limit *int) (ents []*model.CensusGeography, err error) {
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
			func(ctx context.Context, params []model.CensusSourceParam) ([][]*model.CensusSource, []error) {
				return paramGroupQuery(
					params,
					func(p model.CensusSourceParam) (int, *model.CensusSourceFilter, *int) {
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
			func(ctx context.Context, params []model.CensusValueParam) ([][]*model.CensusValue, []error) {
				return paramGroupQuery(
					params,
					func(p model.CensusValueParam) (string, string, *int) {
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
			func(ctx context.Context, params []model.FeedFetchParam) ([][]*model.FeedFetch, []error) {
				return paramGroupQuery(
					params,
					func(p model.FeedFetchParam) (int, *model.FeedFetchFilter, *int) {
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
			func(ctx context.Context, params []model.FeedInfoParam) ([][]*model.FeedInfo, []error) {
				return paramGroupQuery(
					params,
					func(p model.FeedInfoParam) (int, bool, *int) {
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
			func(ctx context.Context, params []model.FeedParam) ([][]*model.Feed, []error) {
				return paramGroupQuery(
					params,
					func(p model.FeedParam) (string, *model.FeedFilter, *int) {
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
			func(ctx context.Context, params []model.FeedVersionFileInfoParam) ([][]*model.FeedVersionFileInfo, []error) {
				return paramGroupQuery(
					params,
					func(p model.FeedVersionFileInfoParam) (int, bool, *int) {
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
			func(ctx context.Context, params []model.FeedVersionParam) ([][]*model.FeedVersion, []error) {
				return paramGroupQuery(
					params,
					func(p model.FeedVersionParam) (int, *model.FeedVersionFilter, *int) {
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
			func(ctx context.Context, params []model.FeedVersionServiceLevelParam) ([][]*model.FeedVersionServiceLevel, []error) {
				return paramGroupQuery(
					params,
					func(p model.FeedVersionServiceLevelParam) (int, *model.FeedVersionServiceLevelFilter, *int) {
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
			func(ctx context.Context, params []model.FrequencyParam) ([][]*model.Frequency, []error) {
				return paramGroupQuery(
					params,
					func(p model.FrequencyParam) (int, bool, *int) {
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
			func(ctx context.Context, params []model.LevelParam) ([][]*model.Level, []error) {
				return paramGroupQuery(
					params,
					func(p model.LevelParam) (int, bool, *int) {
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
		OperatorsByFeedID:    withWaitAndCapacity(waitTime, batchSize, dbf.OperatorsByFeedID),
		PathwaysByIDs:        withWaitAndCapacity(waitTime, batchSize, dbf.PathwaysByIDs),
		PathwaysByFromStopIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
				return paramGroupQuery(
					params,
					func(p model.PathwayParam) (int, *model.PathwayFilter, *int) {
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
			func(ctx context.Context, params []model.PathwayParam) ([][]*model.Pathway, []error) {
				return paramGroupQuery(
					params,
					func(p model.PathwayParam) (int, *model.PathwayFilter, *int) {
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
			func(ctx context.Context, params []model.RouteGeometryParam) ([][]*model.RouteGeometry, []error) {
				return paramGroupQuery(
					params,
					func(p model.RouteGeometryParam) (int, bool, *int) {
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
			func(ctx context.Context, params []model.RouteHeadwayParam) ([][]*model.RouteHeadway, []error) {
				return paramGroupQuery(
					params,
					func(p model.RouteHeadwayParam) (int, bool, *int) {
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
			func(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
				return paramGroupQuery(
					params,
					func(p model.RouteParam) (int, *model.RouteFilter, *int) {
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
			func(ctx context.Context, params []model.RouteParam) ([][]*model.Route, []error) {
				return paramGroupQuery(
					params,
					func(p model.RouteParam) (int, *model.RouteFilter, *int) {
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
			func(ctx context.Context, params []model.RouteStopPatternParam) ([][]*model.RouteStopPattern, []error) {
				return paramGroupQuery(
					params,
					func(p model.RouteStopPatternParam) (int, bool, *int) {
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
			func(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
				return paramGroupQuery(
					params,
					func(p model.RouteStopParam) (int, bool, *int) {
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
			func(ctx context.Context, params []model.RouteStopParam) ([][]*model.RouteStop, []error) {
				return paramGroupQuery(
					params,
					func(p model.RouteStopParam) (int, bool, *int) {
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
		SegmentPatternsByRouteID:        withWaitAndCapacity(waitTime, batchSize, dbf.SegmentPatternsByRouteID),
		SegmentPatternsBySegmentID:      withWaitAndCapacity(waitTime, batchSize, dbf.SegmentPatternsBySegmentID),
		SegmentsByFeedVersionID:         withWaitAndCapacity(waitTime, batchSize, dbf.SegmentsByFeedVersionID),
		SegmentsByIDs:                   withWaitAndCapacity(waitTime, batchSize, dbf.SegmentsByIDs),
		SegmentsByRouteID:               withWaitAndCapacity(waitTime, batchSize, dbf.SegmentsByRouteID),
		ShapesByIDs:                     withWaitAndCapacity(waitTime, batchSize, dbf.ShapesByIDs),
		StopExternalReferencesByStopIDs: withWaitAndCapacity(waitTime, batchSize, dbf.StopExternalReferencesByStopIDs),
		StopObservationsByStopIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []model.StopObservationParam) ([][]*model.StopObservation, []error) {
				return paramGroupQuery(
					params,
					func(p model.StopObservationParam) (int, *model.StopObservationFilter, *int) {
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
			func(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p model.StopParam) (int, *model.StopFilter, *int) {
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
			func(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p model.StopParam) (int, *model.StopFilter, *int) {
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
			func(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p model.StopParam) (int, *model.StopFilter, *int) {
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
			func(ctx context.Context, params []model.StopParam) ([][]*model.Stop, []error) {
				return paramGroupQuery(
					params,
					func(p model.StopParam) (int, *model.StopFilter, *int) {
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
		StopTimesByStopID:    withWaitAndCapacity(waitTime, stopTimeBatchSize, dbf.StopTimesByStopID),
		StopTimesByTripID:    withWaitAndCapacity(waitTime, batchSize, dbf.StopTimesByTripID),
		TargetStopsByStopIDs: withWaitAndCapacity(waitTime, batchSize, dbf.TargetStopsByStopIDs),
		TripsByFeedVersionIDs: withWaitAndCapacity(
			waitTime,
			batchSize,
			func(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
				return paramGroupQuery(
					params,
					func(p model.TripParam) (int, *model.TripFilter, *int) {
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
			func(ctx context.Context, params []model.TripParam) ([][]*model.Trip, []error) {
				return paramGroupQuery(
					params,
					func(p model.TripParam) (model.FVPair, *model.TripFilter, *int) {
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
		ValidationReportErrorExemplarsByValidationReportErrorGroupID: withWaitAndCapacity(waitTime, batchSize, dbf.ValidationReportErrorExemplarsByValidationReportErrorGroupID),
		ValidationReportErrorGroupsByValidationReportID:              withWaitAndCapacity(waitTime, batchSize, dbf.ValidationReportErrorGroupsByValidationReportID),
		ValidationReportsByFeedVersionID:                             withWaitAndCapacity(waitTime, batchSize, dbf.ValidationReportsByFeedVersionID),
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
		// limit := checkLimit(pgroup.Limit)
		limit := uint64(1000)
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
