package resolvers

// import graph gophers with your other imports
import (
	"context"
	"net/http"

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
	loaders := &Loaders{
		FeedStatesByFeedID:                      dataloader.NewBatchedLoader(Unwrap(dbf.FeedStatesByFeedID)),
		AgenciesByID:                            dataloader.NewBatchedLoader(Unwrap(dbf.AgenciesByID)),
		CalendarsByID:                           dataloader.NewBatchedLoader(Unwrap(dbf.CalendarsByID)),
		FeedsByID:                               dataloader.NewBatchedLoader(Unwrap(dbf.FeedsByID)),
		RoutesByID:                              dataloader.NewBatchedLoader(Unwrap(dbf.RoutesByID)),
		ShapesByID:                              dataloader.NewBatchedLoader(Unwrap(dbf.ShapesByID)),
		StopsByID:                               dataloader.NewBatchedLoader(Unwrap(dbf.StopsByID)),
		FeedVersionsByID:                        dataloader.NewBatchedLoader(Unwrap(dbf.FeedVersionsByID)),
		LevelsByID:                              dataloader.NewBatchedLoader(Unwrap(dbf.LevelsByID)),
		TripsByID:                               dataloader.NewBatchedLoader(Unwrap(dbf.TripsByID)),
		OperatorsByCOIF:                         dataloader.NewBatchedLoader(Unwrap(dbf.OperatorsByCOIF)),
		FeedVersionGtfsImportsByFeedVersionID:   dataloader.NewBatchedLoader(Unwrap(dbf.FeedVersionGtfsImportsByFeedVersionID)),
		FeedVersionsByFeedID:                    dataloader.NewBatchedLoader(Unwrap(dbf.FeedVersionsByFeedID)),
		FeedFetchesByFeedID:                     dataloader.NewBatchedLoader(Unwrap(dbf.FeedFetchesByFeedID)),
		OperatorsByFeedID:                       dataloader.NewBatchedLoader(Unwrap(dbf.OperatorsByFeedID)),
		AgenciesByOnestopID:                     dataloader.NewBatchedLoader(Unwrap(dbf.AgenciesByOnestopID)),
		FeedVersionServiceLevelsByFeedVersionID: dataloader.NewBatchedLoader(Unwrap(dbf.FeedVersionServiceLevelsByFeedVersionID)),
		FeedVersionFileInfosByFeedVersionID:     dataloader.NewBatchedLoader(Unwrap(dbf.FeedVersionFileInfosByFeedVersionID)),
		AgenciesByFeedVersionID:                 dataloader.NewBatchedLoader(Unwrap(dbf.AgenciesByFeedVersionID)),
		RoutesByFeedVersionID:                   dataloader.NewBatchedLoader(Unwrap(dbf.RoutesByFeedVersionID)),
		StopsByFeedVersionID:                    dataloader.NewBatchedLoader(Unwrap(dbf.StopsByFeedVersionID)),
		TripsByFeedVersionID:                    dataloader.NewBatchedLoader(Unwrap(dbf.TripsByFeedVersionID)),
		FeedInfosByFeedVersionID:                dataloader.NewBatchedLoader(Unwrap(dbf.FeedInfosByFeedVersionID)),
		FeedsByOperatorOnestopID:                dataloader.NewBatchedLoader(Unwrap(dbf.FeedsByOperatorOnestopID)),
		StopsByRouteID:                          dataloader.NewBatchedLoader(Unwrap(dbf.StopsByRouteID)),
		StopsByParentStopID:                     dataloader.NewBatchedLoader(Unwrap(dbf.StopsByParentStopID)),
		AgencyPlacesByAgencyID:                  dataloader.NewBatchedLoader(Unwrap(dbf.AgencyPlacesByAgencyID)),
		RouteGeometriesByRouteID:                dataloader.NewBatchedLoader(Unwrap(dbf.RouteGeometriesByRouteID)),
		TripsByRouteID:                          dataloader.NewBatchedLoader(Unwrap(dbf.TripsByRouteID)),
		FrequenciesByTripID:                     dataloader.NewBatchedLoader(Unwrap(dbf.FrequenciesByTripID)),
		StopTimesByTripID:                       dataloader.NewBatchedLoader(Unwrap(dbf.StopTimesByTripID)),
		StopTimesByStopID:                       dataloader.NewBatchedLoader(Unwrap(dbf.StopTimesByStopID)),
		RouteStopsByRouteID:                     dataloader.NewBatchedLoader(Unwrap(dbf.RouteStopsByRouteID)),
		RouteStopPatternsByRouteID:              dataloader.NewBatchedLoader(Unwrap(dbf.RouteStopPatternsByRouteID)),
		RouteStopsByStopID:                      dataloader.NewBatchedLoader(Unwrap(dbf.RouteStopsByStopID)),
		RouteHeadwaysByRouteID:                  dataloader.NewBatchedLoader(Unwrap(dbf.RouteHeadwaysByRouteID)),
		RoutesByAgencyID:                        dataloader.NewBatchedLoader(Unwrap(dbf.RoutesByAgencyID)),
		PathwaysByFromStopID:                    dataloader.NewBatchedLoader(Unwrap(dbf.PathwaysByFromStopID)),
		PathwaysByToStopID:                      dataloader.NewBatchedLoader(Unwrap(dbf.PathwaysByToStopID)),
		CalendarDatesByServiceID:                dataloader.NewBatchedLoader(Unwrap(dbf.CalendarDatesByServiceID)),
		CensusTableByID:                         dataloader.NewBatchedLoader(Unwrap(dbf.CensusTableByID)),
		CensusGeographiesByEntityID:             dataloader.NewBatchedLoader(Unwrap(dbf.CensusGeographiesByEntityID)),
		CensusValuesByGeographyID:               dataloader.NewBatchedLoader(Unwrap(dbf.CensusValuesByGeographyID)),
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

// This function adapts existing DBFinder methods to dataloader Result type
func UnwrapMulti[
	T any,
	AT []T,
	ParamT comparable,
](
	cb func(context.Context, []ParamT) ([]AT, []error),
) func(context.Context, []ParamT) []*dataloader.Result[AT] {
	x := func(ctx context.Context, ps []ParamT) []*dataloader.Result[AT] {
		a, _ := cb(ctx, ps)
		ret := make([]*dataloader.Result[AT], len(ps))
		for idx := range ps {
			ret[idx] = &dataloader.Result[AT]{Data: a[idx]}
		}
		return ret
	}
	return x
}
