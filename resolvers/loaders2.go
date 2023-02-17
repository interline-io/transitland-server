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
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(dbf model.Finder) *Loaders {
	d := 2 * time.Millisecond
	loaders := &Loaders{
		FeedStatesByFeedID:                      withWait(d, dbf.FeedStatesByFeedID),
		AgenciesByID:                            withWait(d, dbf.AgenciesByID),
		CalendarsByID:                           withWait(d, dbf.CalendarsByID),
		FeedsByID:                               withWait(d, dbf.FeedsByID),
		RoutesByID:                              withWait(d, dbf.RoutesByID),
		ShapesByID:                              withWait(d, dbf.ShapesByID),
		StopsByID:                               withWait(d, dbf.StopsByID),
		FeedVersionsByID:                        withWait(d, dbf.FeedVersionsByID),
		LevelsByID:                              withWait(d, dbf.LevelsByID),
		TripsByID:                               withWait(d, dbf.TripsByID),
		OperatorsByCOIF:                         withWait(d, dbf.OperatorsByCOIF),
		FeedVersionGtfsImportsByFeedVersionID:   withWait(d, dbf.FeedVersionGtfsImportsByFeedVersionID),
		FeedVersionsByFeedID:                    withWait(d, dbf.FeedVersionsByFeedID),
		FeedFetchesByFeedID:                     withWait(d, dbf.FeedFetchesByFeedID),
		OperatorsByFeedID:                       withWait(d, dbf.OperatorsByFeedID),
		AgenciesByOnestopID:                     withWait(d, dbf.AgenciesByOnestopID),
		FeedVersionServiceLevelsByFeedVersionID: withWait(d, dbf.FeedVersionServiceLevelsByFeedVersionID),
		FeedVersionFileInfosByFeedVersionID:     withWait(d, dbf.FeedVersionFileInfosByFeedVersionID),
		AgenciesByFeedVersionID:                 withWait(d, dbf.AgenciesByFeedVersionID),
		RoutesByFeedVersionID:                   withWait(d, dbf.RoutesByFeedVersionID),
		StopsByFeedVersionID:                    withWait(d, dbf.StopsByFeedVersionID),
		TripsByFeedVersionID:                    withWait(d, dbf.TripsByFeedVersionID),
		FeedInfosByFeedVersionID:                withWait(d, dbf.FeedInfosByFeedVersionID),
		FeedsByOperatorOnestopID:                withWait(d, dbf.FeedsByOperatorOnestopID),
		StopsByRouteID:                          withWait(d, dbf.StopsByRouteID),
		StopsByParentStopID:                     withWait(d, dbf.StopsByParentStopID),
		AgencyPlacesByAgencyID:                  withWait(d, dbf.AgencyPlacesByAgencyID),
		RouteGeometriesByRouteID:                withWait(d, dbf.RouteGeometriesByRouteID),
		TripsByRouteID:                          withWait(d, dbf.TripsByRouteID),
		FrequenciesByTripID:                     withWait(d, dbf.FrequenciesByTripID),
		StopTimesByTripID:                       withWait(d, dbf.StopTimesByTripID),
		StopTimesByStopID:                       withWait(d, dbf.StopTimesByStopID),
		RouteStopsByRouteID:                     withWait(d, dbf.RouteStopsByRouteID),
		RouteStopPatternsByRouteID:              withWait(d, dbf.RouteStopPatternsByRouteID),
		RouteStopsByStopID:                      withWait(d, dbf.RouteStopsByStopID),
		RouteHeadwaysByRouteID:                  withWait(d, dbf.RouteHeadwaysByRouteID),
		RoutesByAgencyID:                        withWait(d, dbf.RoutesByAgencyID),
		PathwaysByFromStopID:                    withWait(d, dbf.PathwaysByFromStopID),
		PathwaysByToStopID:                      withWait(d, dbf.PathwaysByToStopID),
		CalendarDatesByServiceID:                withWait(d, dbf.CalendarDatesByServiceID),
		CensusTableByID:                         withWait(d, dbf.CensusTableByID),
		CensusGeographiesByEntityID:             withWait(d, dbf.CensusGeographiesByEntityID),
		CensusValuesByGeographyID:               withWait(d, dbf.CensusValuesByGeographyID),
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

// This function adapts existing DBFinder methods to dataloader Result type
func Unwrap[
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

func withWait[
	T any,
	ParamT comparable,
](d time.Duration, cb func(context.Context, []ParamT) ([]T, []error)) *dataloader.Loader[ParamT, T] {
	return dataloader.NewBatchedLoader(Unwrap(cb), dataloader.WithWait[ParamT, T](d))
}
