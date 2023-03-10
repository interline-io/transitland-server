package resolvers

// import graph gophers with your other imports
import (
	"context"
	"net/http"
	"time"

	dataloader "github.com/graph-gophers/dataloader/v7"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
	waitTime   = 2 * time.Millisecond
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
	FeedVersionGtfsImportsByFeedVersionID   *dataloader.Loader[int, *model.FeedVersionGtfsImport]
	CensusTableByID                         *dataloader.Loader[int, *model.CensusTable]
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
	StopTimesByTripID                       *dataloader.Loader[model.StopTimeParam, []*model.StopTime]
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
		FeedStatesByFeedID:                      withWait(waitTime, dbf.FeedStatesByFeedID),
		AgenciesByID:                            withWait(waitTime, dbf.AgenciesByID),
		CalendarsByID:                           withWait(waitTime, dbf.CalendarsByID),
		FeedsByID:                               withWait(waitTime, dbf.FeedsByID),
		RoutesByID:                              withWait(waitTime, dbf.RoutesByID),
		ShapesByID:                              withWait(waitTime, dbf.ShapesByID),
		StopsByID:                               withWait(waitTime, dbf.StopsByID),
		FeedVersionsByID:                        withWait(waitTime, dbf.FeedVersionsByID),
		LevelsByID:                              withWait(waitTime, dbf.LevelsByID),
		TripsByID:                               withWait(waitTime, dbf.TripsByID),
		OperatorsByCOIF:                         withWait(waitTime, dbf.OperatorsByCOIF),
		FeedVersionGtfsImportsByFeedVersionID:   withWait(waitTime, dbf.FeedVersionGtfsImportsByFeedVersionID),
		FeedVersionsByFeedID:                    withWait(waitTime, dbf.FeedVersionsByFeedID),
		FeedFetchesByFeedID:                     withWait(waitTime, dbf.FeedFetchesByFeedID),
		OperatorsByFeedID:                       withWait(waitTime, dbf.OperatorsByFeedID),
		AgenciesByOnestopID:                     withWait(waitTime, dbf.AgenciesByOnestopID),
		FeedVersionServiceLevelsByFeedVersionID: withWait(waitTime, dbf.FeedVersionServiceLevelsByFeedVersionID),
		FeedVersionFileInfosByFeedVersionID:     withWait(waitTime, dbf.FeedVersionFileInfosByFeedVersionID),
		AgenciesByFeedVersionID:                 withWait(waitTime, dbf.AgenciesByFeedVersionID),
		RoutesByFeedVersionID:                   withWait(waitTime, dbf.RoutesByFeedVersionID),
		StopsByFeedVersionID:                    withWait(waitTime, dbf.StopsByFeedVersionID),
		TripsByFeedVersionID:                    withWait(waitTime, dbf.TripsByFeedVersionID),
		FeedInfosByFeedVersionID:                withWait(waitTime, dbf.FeedInfosByFeedVersionID),
		FeedsByOperatorOnestopID:                withWait(waitTime, dbf.FeedsByOperatorOnestopID),
		StopsByRouteID:                          withWait(waitTime, dbf.StopsByRouteID),
		StopsByParentStopID:                     withWait(waitTime, dbf.StopsByParentStopID),
		AgencyPlacesByAgencyID:                  withWait(waitTime, dbf.AgencyPlacesByAgencyID),
		RouteGeometriesByRouteID:                withWait(waitTime, dbf.RouteGeometriesByRouteID),
		TripsByRouteID:                          withWait(waitTime, dbf.TripsByRouteID),
		FrequenciesByTripID:                     withWait(waitTime, dbf.FrequenciesByTripID),
		StopTimesByTripID:                       withWait(waitTime, dbf.StopTimesByTripID),
		StopTimesByStopID:                       withWait(waitTime, dbf.StopTimesByStopID),
		RouteStopsByRouteID:                     withWait(waitTime, dbf.RouteStopsByRouteID),
		RouteStopPatternsByRouteID:              withWait(waitTime, dbf.RouteStopPatternsByRouteID),
		RouteStopsByStopID:                      withWait(waitTime, dbf.RouteStopsByStopID),
		RouteHeadwaysByRouteID:                  withWait(waitTime, dbf.RouteHeadwaysByRouteID),
		RoutesByAgencyID:                        withWait(waitTime, dbf.RoutesByAgencyID),
		PathwaysByFromStopID:                    withWait(waitTime, dbf.PathwaysByFromStopID),
		PathwaysByToStopID:                      withWait(waitTime, dbf.PathwaysByToStopID),
		CalendarDatesByServiceID:                withWait(waitTime, dbf.CalendarDatesByServiceID),
		CensusTableByID:                         withWait(waitTime, dbf.CensusTableByID),
		CensusGeographiesByEntityID:             withWait(waitTime, dbf.CensusGeographiesByEntityID),
		CensusValuesByGeographyID:               withWait(waitTime, dbf.CensusValuesByGeographyID),
		StopObservationsByStopID:                withWait(waitTime, dbf.StopObservationsByStopID),
		StopExternalReferencesByStopID:          withWait(waitTime, dbf.StopExternalReferencesByStopID),
		StopsByLevelID:                          withWait(waitTime, dbf.StopsByLevelID),
		TargetStopsByStopID:                     withWait(waitTime, dbf.TargetStopsByStopID),
		RouteAttributesByRouteID:                withWait(waitTime, dbf.RouteAttributesByRouteID),
	}
	return loaders
}

func Middleware(cfg config.Config, finder model.Finder, next http.Handler) http.Handler {
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
func withWait[
	T any,
	ParamT comparable,
](d time.Duration, cb func(context.Context, []ParamT) ([]T, []error)) *dataloader.Loader[ParamT, T] {
	return dataloader.NewBatchedLoader(unwrapResult(cb), dataloader.WithWait[ParamT, T](d))
}

// unwrap function adapts existing DBFinder methods to dataloader Result type
func unwrapResult[
	T any,
	ParamT comparable,
](
	cb func(context.Context, []ParamT) ([]T, []error),
) func(context.Context, []ParamT) []*dataloader.Result[T] {
	x := func(ctx context.Context, ps []ParamT) []*dataloader.Result[T] {
		a, _ := cb(ctx, ps)
		ret := make([]*dataloader.Result[T], len(ps))
		for idx := range ps {
			ret[idx] = &dataloader.Result[T]{Data: a[idx]}
		}
		return ret
	}
	return x
}
