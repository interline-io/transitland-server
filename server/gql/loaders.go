package gql

// import graph gophers with your other imports
import (
	"context"
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
	AgenciesByFeedVersionID                                      *dataloader.Loader[model.AgencyParam, []*model.Agency]
	AgenciesByID                                                 *dataloader.Loader[int, *model.Agency]
	AgenciesByOnestopID                                          *dataloader.Loader[model.AgencyParam, []*model.Agency]
	AgencyPlacesByAgencyID                                       *dataloader.Loader[model.AgencyPlaceParam, []*model.AgencyPlace]
	CalendarDatesByServiceID                                     *dataloader.Loader[model.CalendarDateParam, []*model.CalendarDate]
	CalendarsByID                                                *dataloader.Loader[int, *model.Calendar]
	CensusGeographiesByEntityID                                  *dataloader.Loader[model.CensusGeographyParam, []*model.CensusGeography]
	CensusTableByID                                              *dataloader.Loader[int, *model.CensusTable]
	CensusFieldsByTableID                                        *dataloader.Loader[model.CensusFieldParam, []*model.CensusField]
	CensusValuesByGeographyID                                    *dataloader.Loader[model.CensusValueParam, []*model.CensusValue]
	FeedFetchesByFeedID                                          *dataloader.Loader[model.FeedFetchParam, []*model.FeedFetch]
	FeedInfosByFeedVersionID                                     *dataloader.Loader[model.FeedInfoParam, []*model.FeedInfo]
	FeedsByID                                                    *dataloader.Loader[int, *model.Feed]
	FeedsByOperatorOnestopID                                     *dataloader.Loader[model.FeedParam, []*model.Feed]
	FeedStatesByFeedID                                           *dataloader.Loader[int, *model.FeedState]
	FeedVersionFileInfosByFeedVersionID                          *dataloader.Loader[model.FeedVersionFileInfoParam, []*model.FeedVersionFileInfo]
	FeedVersionGeometryByID                                      *dataloader.Loader[int, *tt.Polygon]
	FeedVersionGtfsImportByFeedVersionID                         *dataloader.Loader[int, *model.FeedVersionGtfsImport]
	FeedVersionServiceWindowByFeedVersionID                      *dataloader.Loader[int, *model.FeedVersionServiceWindow]
	FeedVersionsByFeedID                                         *dataloader.Loader[model.FeedVersionParam, []*model.FeedVersion]
	FeedVersionsByID                                             *dataloader.Loader[int, *model.FeedVersion]
	FeedVersionServiceLevelsByFeedVersionID                      *dataloader.Loader[model.FeedVersionServiceLevelParam, []*model.FeedVersionServiceLevel]
	FrequenciesByTripID                                          *dataloader.Loader[model.FrequencyParam, []*model.Frequency]
	LevelsByID                                                   *dataloader.Loader[int, *model.Level]
	LevelsByParentStationID                                      *dataloader.Loader[model.LevelParam, []*model.Level]
	OperatorsByAgencyID                                          *dataloader.Loader[int, *model.Operator]
	OperatorsByCOIF                                              *dataloader.Loader[int, *model.Operator]
	OperatorsByFeedID                                            *dataloader.Loader[model.OperatorParam, []*model.Operator]
	PathwaysByFromStopID                                         *dataloader.Loader[model.PathwayParam, []*model.Pathway]
	PathwaysByID                                                 *dataloader.Loader[int, *model.Pathway]
	PathwaysByToStopID                                           *dataloader.Loader[model.PathwayParam, []*model.Pathway]
	RouteAttributesByRouteID                                     *dataloader.Loader[int, *model.RouteAttribute]
	RouteGeometriesByRouteID                                     *dataloader.Loader[model.RouteGeometryParam, []*model.RouteGeometry]
	RouteHeadwaysByRouteID                                       *dataloader.Loader[model.RouteHeadwayParam, []*model.RouteHeadway]
	RoutesByAgencyID                                             *dataloader.Loader[model.RouteParam, []*model.Route]
	RoutesByFeedVersionID                                        *dataloader.Loader[model.RouteParam, []*model.Route]
	RoutesByID                                                   *dataloader.Loader[int, *model.Route]
	RouteStopPatternsByRouteID                                   *dataloader.Loader[model.RouteStopPatternParam, []*model.RouteStopPattern]
	RouteStopsByRouteID                                          *dataloader.Loader[model.RouteStopParam, []*model.RouteStop]
	RouteStopsByStopID                                           *dataloader.Loader[model.RouteStopParam, []*model.RouteStop]
	SegmentPatternsByRouteID                                     *dataloader.Loader[model.SegmentPatternParam, []*model.SegmentPattern]
	SegmentPatternsBySegmentID                                   *dataloader.Loader[model.SegmentPatternParam, []*model.SegmentPattern]
	SegmentsByID                                                 *dataloader.Loader[int, *model.Segment]
	SegmentsByRouteID                                            *dataloader.Loader[model.SegmentParam, []*model.Segment]
	SegmentsByFeedVersionID                                      *dataloader.Loader[model.SegmentParam, []*model.Segment]
	ShapesByID                                                   *dataloader.Loader[int, *model.Shape]
	StopExternalReferencesByStopID                               *dataloader.Loader[int, *model.StopExternalReference]
	StopObservationsByStopID                                     *dataloader.Loader[model.StopObservationParam, []*model.StopObservation]
	StopPlacesByStopID                                           *dataloader.Loader[model.StopPlaceParam, *model.StopPlace]
	StopsByFeedVersionID                                         *dataloader.Loader[model.StopParam, []*model.Stop]
	StopsByID                                                    *dataloader.Loader[int, *model.Stop]
	StopsByLevelID                                               *dataloader.Loader[model.StopParam, []*model.Stop]
	StopsByParentStopID                                          *dataloader.Loader[model.StopParam, []*model.Stop]
	StopsByRouteID                                               *dataloader.Loader[model.StopParam, []*model.Stop]
	StopTimesByStopID                                            *dataloader.Loader[model.StopTimeParam, []*model.StopTime]
	StopTimesByTripID                                            *dataloader.Loader[model.TripStopTimeParam, []*model.StopTime]
	TargetStopsByStopID                                          *dataloader.Loader[int, *model.Stop]
	TripsByFeedVersionID                                         *dataloader.Loader[model.TripParam, []*model.Trip]
	TripsByID                                                    *dataloader.Loader[int, *model.Trip]
	TripsByRouteID                                               *dataloader.Loader[model.TripParam, []*model.Trip]
	ValidationReportErrorExemplarsByValidationReportErrorGroupID *dataloader.Loader[model.ValidationReportErrorExemplarParam, []*model.ValidationReportError]
	ValidationReportErrorGroupsByValidationReportID              *dataloader.Loader[model.ValidationReportErrorGroupParam, []*model.ValidationReportErrorGroup]
	ValidationReportsByFeedVersionID                             *dataloader.Loader[model.ValidationReportParam, []*model.ValidationReport]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(dbf model.Finder) *Loaders {
	loaders := &Loaders{
		AgenciesByFeedVersionID:              withWaitAndCapacity(waitTime, maxBatch, dbf.AgenciesByFeedVersionID),
		AgenciesByID:                         withWaitAndCapacity(waitTime, maxBatch, dbf.AgenciesByID),
		AgenciesByOnestopID:                  withWaitAndCapacity(waitTime, maxBatch, dbf.AgenciesByOnestopID),
		AgencyPlacesByAgencyID:               withWaitAndCapacity(waitTime, maxBatch, dbf.AgencyPlacesByAgencyID),
		CalendarDatesByServiceID:             withWaitAndCapacity(waitTime, maxBatch, dbf.CalendarDatesByServiceID),
		CalendarsByID:                        withWaitAndCapacity(waitTime, maxBatch, dbf.CalendarsByID),
		CensusGeographiesByEntityID:          withWaitAndCapacity(waitTime, maxBatch, dbf.CensusGeographiesByEntityID),
		CensusTableByID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.CensusTableByID),
		CensusFieldsByTableID:                withWaitAndCapacity(waitTime, maxBatch, dbf.CensusFieldsByTableID),
		CensusValuesByGeographyID:            withWaitAndCapacity(waitTime, maxBatch, dbf.CensusValuesByGeographyID),
		FeedFetchesByFeedID:                  withWaitAndCapacity(waitTime, maxBatch, dbf.FeedFetchesByFeedID),
		FeedInfosByFeedVersionID:             withWaitAndCapacity(waitTime, maxBatch, dbf.FeedInfosByFeedVersionID),
		FeedsByID:                            withWaitAndCapacity(waitTime, maxBatch, dbf.FeedsByID),
		FeedsByOperatorOnestopID:             withWaitAndCapacity(waitTime, maxBatch, dbf.FeedsByOperatorOnestopID),
		FeedStatesByFeedID:                   withWaitAndCapacity(waitTime, maxBatch, dbf.FeedStatesByFeedID),
		FeedVersionFileInfosByFeedVersionID:  withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionFileInfosByFeedVersionID),
		FeedVersionGeometryByID:              withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionGeometryByID),
		FeedVersionGtfsImportByFeedVersionID: withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionGtfsImportByFeedVersionID), FeedVersionServiceWindowByFeedVersionID: withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionServiceWindowByFeedVersionID),
		FeedVersionsByFeedID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionsByFeedID),
		FeedVersionsByID:                        withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionsByID),
		FeedVersionServiceLevelsByFeedVersionID: withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionServiceLevelsByFeedVersionID),
		FrequenciesByTripID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.FrequenciesByTripID),
		LevelsByID:                              withWaitAndCapacity(waitTime, maxBatch, dbf.LevelsByID),
		LevelsByParentStationID:                 withWaitAndCapacity(waitTime, maxBatch, dbf.LevelsByParentStationID),
		OperatorsByAgencyID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.OperatorsByAgencyID),
		OperatorsByCOIF:                         withWaitAndCapacity(waitTime, maxBatch, dbf.OperatorsByCOIF),
		OperatorsByFeedID:                       withWaitAndCapacity(waitTime, maxBatch, dbf.OperatorsByFeedID),
		PathwaysByFromStopID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.PathwaysByFromStopID),
		PathwaysByID:                            withWaitAndCapacity(waitTime, maxBatch, dbf.PathwaysByID),
		PathwaysByToStopID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.PathwaysByToStopID),
		RouteAttributesByRouteID:                withWaitAndCapacity(waitTime, maxBatch, dbf.RouteAttributesByRouteID),
		RouteGeometriesByRouteID:                withWaitAndCapacity(waitTime, maxBatch, dbf.RouteGeometriesByRouteID),
		RouteHeadwaysByRouteID:                  withWaitAndCapacity(waitTime, maxBatch, dbf.RouteHeadwaysByRouteID),
		RoutesByAgencyID:                        withWaitAndCapacity(waitTime, maxBatch, dbf.RoutesByAgencyID),
		RoutesByFeedVersionID:                   withWaitAndCapacity(waitTime, maxBatch, dbf.RoutesByFeedVersionID),
		RoutesByID:                              withWaitAndCapacity(waitTime, maxBatch, dbf.RoutesByID),
		RouteStopPatternsByRouteID:              withWaitAndCapacity(waitTime, maxBatch, dbf.RouteStopPatternsByRouteID),
		RouteStopsByRouteID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.RouteStopsByRouteID),
		RouteStopsByStopID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.RouteStopsByStopID),
		SegmentPatternsByRouteID:                withWaitAndCapacity(waitTime, maxBatch, dbf.SegmentPatternsByRouteID),
		SegmentPatternsBySegmentID:              withWaitAndCapacity(waitTime, maxBatch, dbf.SegmentPatternsBySegmentID),
		SegmentsByID:                            withWaitAndCapacity(waitTime, maxBatch, dbf.SegmentsByID),
		SegmentsByRouteID:                       withWaitAndCapacity(waitTime, maxBatch, dbf.SegmentsByRouteID),
		SegmentsByFeedVersionID:                 withWaitAndCapacity(waitTime, maxBatch, dbf.SegmentsByFeedVersionID),
		ShapesByID:                              withWaitAndCapacity(waitTime, maxBatch, dbf.ShapesByID),
		StopExternalReferencesByStopID:          withWaitAndCapacity(waitTime, maxBatch, dbf.StopExternalReferencesByStopID),
		StopObservationsByStopID:                withWaitAndCapacity(waitTime, maxBatch, dbf.StopObservationsByStopID),
		StopPlacesByStopID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.StopPlacesByStopID),
		StopsByFeedVersionID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByFeedVersionID),
		StopsByID:                               withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByID),
		StopsByLevelID:                          withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByLevelID),
		StopsByParentStopID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByParentStopID),
		StopsByRouteID:                          withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByRouteID),
		StopTimesByStopID:                       withWaitAndCapacity(waitTime, 1, dbf.StopTimesByStopID),
		StopTimesByTripID:                       withWaitAndCapacity(waitTime, maxBatch, dbf.StopTimesByTripID),
		TargetStopsByStopID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.TargetStopsByStopID),
		TripsByFeedVersionID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.TripsByFeedVersionID),
		TripsByID:                               withWaitAndCapacity(waitTime, maxBatch, dbf.TripsByID),
		TripsByRouteID:                          withWaitAndCapacity(waitTime, maxBatch, dbf.TripsByRouteID),
		ValidationReportErrorExemplarsByValidationReportErrorGroupID: withWaitAndCapacity(waitTime, maxBatch, dbf.ValidationReportErrorExemplarsByValidationReportErrorGroupID),
		ValidationReportErrorGroupsByValidationReportID:              withWaitAndCapacity(waitTime, maxBatch, dbf.ValidationReportErrorGroupsByValidationReportID),
		ValidationReportsByFeedVersionID:                             withWaitAndCapacity(waitTime, maxBatch, dbf.ValidationReportsByFeedVersionID),
	}
	return loaders
}

func loaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is per request scoped loaders/cache
		// Is this OK to use as a long term cache?
		ctx := r.Context()
		cfg := model.ForContext(ctx)
		loaders := NewLoaders(cfg.Finder)
		nextCtx := context.WithValue(ctx, loadersKey, loaders)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Loaders {
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
			log.Trace().Msgf("error in dataloader, result len %d did not match param length %d", len(a), len(ps))
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
