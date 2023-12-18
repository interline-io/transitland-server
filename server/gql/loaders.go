package gql

// import graph gophers with your other imports
import (
	"context"
	"net/http"
	"time"

	dataloader "github.com/graph-gophers/dataloader/v7"
	"github.com/interline-io/log"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/model"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
	waitTime   = 2 * time.Millisecond
	maxBatch   = 1_000
)

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	FeedStatesByFeedID                      *dataloader.Loader[int, *model.FeedState]
	AgenciesByID                            *dataloader.Loader[int, *model.Agency]
	CalendarsByID                           *dataloader.Loader[int, *model.Calendar]
	FeedsByID                               *dataloader.Loader[int, *model.Feed]
	RoutesByID                              *dataloader.Loader[int, *model.Route]
	ShapesByID                              *dataloader.Loader[int, *model.Shape]
	StopsByID                               *dataloader.Loader[int, *model.Stop]
	FeedVersionsByID                        *dataloader.Loader[int, *model.FeedVersion]
	LevelsByID                              *dataloader.Loader[int, *model.Level]
	TripsByID                               *dataloader.Loader[int, *model.Trip]
	OperatorsByCOIF                         *dataloader.Loader[int, *model.Operator]
	OperatorsByOnestopID                    *dataloader.Loader[string, *model.Operator]
	OperatorsByAgencyID                     *dataloader.Loader[int, *model.Operator]
	FeedVersionGtfsImportsByFeedVersionID   *dataloader.Loader[int, *model.FeedVersionGtfsImport]
	CensusTableByID                         *dataloader.Loader[int, *model.CensusTable]
	FeedVersionGeometryByID                 *dataloader.Loader[int, *tt.Polygon]
	StopPlacesByStopID                      *dataloader.Loader[model.StopPlaceParam, *model.StopPlace]
	FeedVersionsByFeedID                    *dataloader.Loader[model.FeedVersionParam, []*model.FeedVersion]
	FeedFetchesByFeedID                     *dataloader.Loader[model.FeedFetchParam, []*model.FeedFetch]
	OperatorsByFeedID                       *dataloader.Loader[model.OperatorParam, []*model.Operator]
	AgenciesByOnestopID                     *dataloader.Loader[model.AgencyParam, []*model.Agency]
	FeedVersionServiceLevelsByFeedVersionID *dataloader.Loader[model.FeedVersionServiceLevelParam, []*model.FeedVersionServiceLevel]
	FeedVersionFileInfosByFeedVersionID     *dataloader.Loader[model.FeedVersionFileInfoParam, []*model.FeedVersionFileInfo]
	AgenciesByFeedVersionID                 *dataloader.Loader[model.AgencyParam, []*model.Agency]
	RoutesByFeedVersionID                   *dataloader.Loader[model.RouteParam, []*model.Route]
	StopsByFeedVersionID                    *dataloader.Loader[model.StopParam, []*model.Stop]
	TripsByFeedVersionID                    *dataloader.Loader[model.TripParam, []*model.Trip]
	FeedInfosByFeedVersionID                *dataloader.Loader[model.FeedInfoParam, []*model.FeedInfo]
	FeedsByOperatorOnestopID                *dataloader.Loader[model.FeedParam, []*model.Feed]
	StopsByRouteID                          *dataloader.Loader[model.StopParam, []*model.Stop]
	StopsByParentStopID                     *dataloader.Loader[model.StopParam, []*model.Stop]
	AgencyPlacesByAgencyID                  *dataloader.Loader[model.AgencyPlaceParam, []*model.AgencyPlace]
	RouteGeometriesByRouteID                *dataloader.Loader[model.RouteGeometryParam, []*model.RouteGeometry]
	TripsByRouteID                          *dataloader.Loader[model.TripParam, []*model.Trip]
	FrequenciesByTripID                     *dataloader.Loader[model.FrequencyParam, []*model.Frequency]
	StopTimesByTripID                       *dataloader.Loader[model.TripStopTimeParam, []*model.StopTime]
	StopTimesByStopID                       *dataloader.Loader[model.StopTimeParam, []*model.StopTime]
	RouteStopsByRouteID                     *dataloader.Loader[model.RouteStopParam, []*model.RouteStop]
	RouteStopPatternsByRouteID              *dataloader.Loader[model.RouteStopPatternParam, []*model.RouteStopPattern]
	RouteStopsByStopID                      *dataloader.Loader[model.RouteStopParam, []*model.RouteStop]
	RouteHeadwaysByRouteID                  *dataloader.Loader[model.RouteHeadwayParam, []*model.RouteHeadway]
	RoutesByAgencyID                        *dataloader.Loader[model.RouteParam, []*model.Route]
	PathwaysByFromStopID                    *dataloader.Loader[model.PathwayParam, []*model.Pathway]
	PathwaysByToStopID                      *dataloader.Loader[model.PathwayParam, []*model.Pathway]
	CalendarDatesByServiceID                *dataloader.Loader[model.CalendarDateParam, []*model.CalendarDate]
	CensusGeographiesByEntityID             *dataloader.Loader[model.CensusGeographyParam, []*model.CensusGeography]
	CensusValuesByGeographyID               *dataloader.Loader[model.CensusValueParam, []*model.CensusValue]
	StopObservationsByStopID                *dataloader.Loader[model.StopObservationParam, []*model.StopObservation]
	StopExternalReferencesByStopID          *dataloader.Loader[int, *model.StopExternalReference]
	StopsByLevelID                          *dataloader.Loader[model.StopParam, []*model.Stop]
	TargetStopsByStopID                     *dataloader.Loader[int, *model.Stop]
	RouteAttributesByRouteID                *dataloader.Loader[int, *model.RouteAttribute]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(dbf model.Finder) *Loaders {
	loaders := &Loaders{
		FeedStatesByFeedID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.FeedStatesByFeedID),
		AgenciesByID:                            withWaitAndCapacity(waitTime, maxBatch, dbf.AgenciesByID),
		CalendarsByID:                           withWaitAndCapacity(waitTime, maxBatch, dbf.CalendarsByID),
		FeedsByID:                               withWaitAndCapacity(waitTime, maxBatch, dbf.FeedsByID),
		RoutesByID:                              withWaitAndCapacity(waitTime, maxBatch, dbf.RoutesByID),
		ShapesByID:                              withWaitAndCapacity(waitTime, maxBatch, dbf.ShapesByID),
		StopsByID:                               withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByID),
		FeedVersionsByID:                        withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionsByID),
		LevelsByID:                              withWaitAndCapacity(waitTime, maxBatch, dbf.LevelsByID),
		TripsByID:                               withWaitAndCapacity(waitTime, maxBatch, dbf.TripsByID),
		OperatorsByCOIF:                         withWaitAndCapacity(waitTime, maxBatch, dbf.OperatorsByCOIF),
		OperatorsByOnestopID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.OperatorsByOnestopID),
		OperatorsByAgencyID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.OperatorsByAgencyID),
		FeedVersionGtfsImportsByFeedVersionID:   withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionGtfsImportsByFeedVersionID),
		FeedVersionsByFeedID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionsByFeedID),
		FeedFetchesByFeedID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.FeedFetchesByFeedID),
		OperatorsByFeedID:                       withWaitAndCapacity(waitTime, maxBatch, dbf.OperatorsByFeedID),
		AgenciesByOnestopID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.AgenciesByOnestopID),
		FeedVersionServiceLevelsByFeedVersionID: withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionServiceLevelsByFeedVersionID),
		FeedVersionFileInfosByFeedVersionID:     withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionFileInfosByFeedVersionID),
		AgenciesByFeedVersionID:                 withWaitAndCapacity(waitTime, maxBatch, dbf.AgenciesByFeedVersionID),
		RoutesByFeedVersionID:                   withWaitAndCapacity(waitTime, maxBatch, dbf.RoutesByFeedVersionID),
		StopsByFeedVersionID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByFeedVersionID),
		TripsByFeedVersionID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.TripsByFeedVersionID),
		FeedInfosByFeedVersionID:                withWaitAndCapacity(waitTime, maxBatch, dbf.FeedInfosByFeedVersionID),
		FeedsByOperatorOnestopID:                withWaitAndCapacity(waitTime, maxBatch, dbf.FeedsByOperatorOnestopID),
		StopsByRouteID:                          withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByRouteID),
		StopsByParentStopID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByParentStopID),
		AgencyPlacesByAgencyID:                  withWaitAndCapacity(waitTime, maxBatch, dbf.AgencyPlacesByAgencyID),
		RouteGeometriesByRouteID:                withWaitAndCapacity(waitTime, maxBatch, dbf.RouteGeometriesByRouteID),
		TripsByRouteID:                          withWaitAndCapacity(waitTime, maxBatch, dbf.TripsByRouteID),
		FrequenciesByTripID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.FrequenciesByTripID),
		StopTimesByTripID:                       withWaitAndCapacity(waitTime, maxBatch, dbf.StopTimesByTripID),
		StopTimesByStopID:                       withWaitAndCapacity(waitTime, maxBatch, dbf.StopTimesByStopID),
		RouteStopsByRouteID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.RouteStopsByRouteID),
		RouteStopPatternsByRouteID:              withWaitAndCapacity(waitTime, maxBatch, dbf.RouteStopPatternsByRouteID),
		RouteStopsByStopID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.RouteStopsByStopID),
		RouteHeadwaysByRouteID:                  withWaitAndCapacity(waitTime, maxBatch, dbf.RouteHeadwaysByRouteID),
		RoutesByAgencyID:                        withWaitAndCapacity(waitTime, maxBatch, dbf.RoutesByAgencyID),
		PathwaysByFromStopID:                    withWaitAndCapacity(waitTime, maxBatch, dbf.PathwaysByFromStopID),
		PathwaysByToStopID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.PathwaysByToStopID),
		CalendarDatesByServiceID:                withWaitAndCapacity(waitTime, maxBatch, dbf.CalendarDatesByServiceID),
		FeedVersionGeometryByID:                 withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionGeometryByID),
		CensusTableByID:                         withWaitAndCapacity(waitTime, maxBatch, dbf.CensusTableByID),
		CensusGeographiesByEntityID:             withWaitAndCapacity(waitTime, maxBatch, dbf.CensusGeographiesByEntityID),
		CensusValuesByGeographyID:               withWaitAndCapacity(waitTime, maxBatch, dbf.CensusValuesByGeographyID),
		StopObservationsByStopID:                withWaitAndCapacity(waitTime, maxBatch, dbf.StopObservationsByStopID),
		StopExternalReferencesByStopID:          withWaitAndCapacity(waitTime, maxBatch, dbf.StopExternalReferencesByStopID),
		StopsByLevelID:                          withWaitAndCapacity(waitTime, maxBatch, dbf.StopsByLevelID),
		TargetStopsByStopID:                     withWaitAndCapacity(waitTime, maxBatch, dbf.TargetStopsByStopID),
		RouteAttributesByRouteID:                withWaitAndCapacity(waitTime, maxBatch, dbf.RouteAttributesByRouteID),
		StopPlacesByStopID:                      withWaitAndCapacity(waitTime, maxBatch, dbf.StopPlacesByStopID),
	}
	return loaders
}

func loaderMiddleware(cfg model.Config, finder model.Finder, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is per request scoped loaders/cache
		// Is this OK to use as a long term cache?
		loaders := NewLoaders(finder)
		nextCtx := context.WithValue(r.Context(), loadersKey, loaders)
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
		a, err := cb(ctx, ps)
		if len(err) > 0 {
			log.Trace().Err(err[0]).Msg("error in dataloader")
			return nil
		}
		if len(a) != len(ps) {
			log.Trace().Msgf("error in dataloader, result len %d did not match param length %d", len(a), len(ps))
			return nil
		}
		ret := make([]*dataloader.Result[T], len(ps))
		for idx := range ps {
			ret[idx] = &dataloader.Result[T]{Data: a[idx]}
		}
		return ret
	}
	return x
}
