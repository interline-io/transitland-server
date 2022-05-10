package resolvers

import (
	"context"
	"net/http"
	"time"

	"github.com/interline-io/transitland-server/config"
	dl "github.com/interline-io/transitland-server/generated/dataloader"
	"github.com/interline-io/transitland-server/model"
)

// MAXBATCH is maximum batch size
const MAXBATCH = 1000

// WAIT is the time to wait
const WAIT = 2 * time.Millisecond

type loadersKey struct{}

// TODO: Use code generation!

// Loaders .
type Loaders struct {
	// ID Loaders
	AgenciesByID                            dl.AgencyLoader
	CalendarsByID                           dl.CalendarLoader
	FeedsByID                               dl.FeedLoader
	RoutesByID                              dl.RouteLoader
	ShapesByID                              dl.ShapeLoader
	StopsByID                               dl.StopLoader
	FeedVersionsByID                        dl.FeedVersionLoader
	LevelsByID                              dl.LevelLoader
	TripsByID                               dl.TripLoader
	FeedStatesByFeedID                      dl.FeedStateLoader
	OperatorsByCOIF                         dl.OperatorLoader
	FeedFetchesByFeedID                     dl.FeedFetchWhereLoader
	AgenciesByOnestopID                     dl.AgencyWhereLoader
	FeedVersionGtfsImportsByFeedVersionID   dl.FeedVersionGtfsImportLoader
	FeedVersionServiceLevelsByFeedVersionID dl.FeedVersionServiceLevelWhereLoader
	FeedVersionFileInfosByFeedVersionID     dl.FeedVersionFileInfoWhereLoader
	AgenciesByFeedVersionID                 dl.AgencyWhereLoader
	RoutesByFeedVersionID                   dl.RouteWhereLoader
	StopsByFeedVersionID                    dl.StopWhereLoader
	TripsByFeedVersionID                    dl.TripWhereLoader
	FeedInfosByFeedVersionID                dl.FeedInfoWhereLoader
	FeedsByOperatorOnestopID                dl.FeedWhereLoader
	StopsByRouteID                          dl.StopWhereLoader
	StopsByParentStopID                     dl.StopWhereLoader
	AgencyPlacesByAgencyID                  dl.AgencyPlaceWhereLoader
	RouteGeometriesByRouteID                dl.RouteGeometryWhereLoader
	TripsByRouteID                          dl.TripWhereLoader
	FrequenciesByTripID                     dl.FrequencyWhereLoader
	StopTimesByTripID                       dl.StopTimeWhereLoader
	StopTimesByStopID                       dl.StopTimeWhereLoader
	RouteStopsByRouteID                     dl.RouteStopWhereLoader
	RouteStopsByStopID                      dl.RouteStopWhereLoader
	RouteHeadwaysByRouteID                  dl.RouteHeadwayWhereLoader
	RoutesByAgencyID                        dl.RouteWhereLoader
	FeedVersionsByFeedID                    dl.FeedVersionWhereLoader
	OperatorsByFeedID                       dl.OperatorWhereLoader
	PathwaysByFromStopID                    dl.PathwayWhereLoader
	PathwaysByToStopID                      dl.PathwayWhereLoader
	CalendarDatesByServiceID                dl.CalendarDateWhereLoader
	CensusTableByID                         dl.CensusTableLoader
	CensusGeographiesByEntityID             dl.CensusGeographyWhereLoader
	CensusValuesByGeographyID               dl.CensusValueWhereLoader
}

// Middleware provides context local request batching
func Middleware(cfg config.Config, finder model.Finder, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loadersKey{}, &Loaders{
			LevelsByID: *dl.NewLevelLoader(dl.LevelLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.LevelsByID,
			}),
			TripsByID: *dl.NewTripLoader(dl.TripLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.TripsByID,
			}),
			CalendarsByID: *dl.NewCalendarLoader(dl.CalendarLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.CalendarsByID,
			}),
			ShapesByID: *dl.NewShapeLoader(dl.ShapeLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.ShapesByID,
			}),
			FeedVersionsByID: *dl.NewFeedVersionLoader(dl.FeedVersionLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedVersionsByID,
			}),
			FeedsByID: *dl.NewFeedLoader(dl.FeedLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedsByID,
			}),
			AgenciesByID: *dl.NewAgencyLoader(dl.AgencyLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.AgenciesByID,
			}),
			StopsByID: *dl.NewStopLoader(dl.StopLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.StopsByID,
			}),
			RoutesByID: *dl.NewRouteLoader(dl.RouteLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.RoutesByID,
			}),
			// Other ID loaders
			FeedVersionGtfsImportsByFeedVersionID: *dl.NewFeedVersionGtfsImportLoader(dl.FeedVersionGtfsImportLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedVersionGtfsImportsByFeedVersionID,
			}),
			FeedStatesByFeedID: *dl.NewFeedStateLoader(dl.FeedStateLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedStatesByFeedID,
			}),
			FeedsByOperatorOnestopID: *dl.NewFeedWhereLoader(dl.FeedWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedsByOperatorOnestopID,
			}),
			OperatorsByFeedID: *dl.NewOperatorWhereLoader(dl.OperatorWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.OperatorsByFeedID,
			}),
			OperatorsByCOIF: *dl.NewOperatorLoader(dl.OperatorLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.OperatorsByCOIF,
			}),
			// Where loaders
			FrequenciesByTripID: *dl.NewFrequencyWhereLoader(dl.FrequencyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FrequenciesByTripID,
			}),
			StopTimesByTripID: *dl.NewStopTimeWhereLoader(dl.StopTimeWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.StopTimesByTripID,
			}),
			StopTimesByStopID: *dl.NewStopTimeWhereLoader(dl.StopTimeWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.StopTimesByStopID,
			}),
			RouteStopsByStopID: *dl.NewRouteStopWhereLoader(dl.RouteStopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.RouteStopsByStopID,
			}),
			StopsByRouteID: *dl.NewStopWhereLoader(dl.StopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.StopsByRouteID,
			}),
			RouteStopsByRouteID: *dl.NewRouteStopWhereLoader(dl.RouteStopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.RouteStopsByRouteID,
			}),
			RouteHeadwaysByRouteID: *dl.NewRouteHeadwayWhereLoader(dl.RouteHeadwayWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.RouteHeadwaysByRouteID,
			}),
			FeedVersionFileInfosByFeedVersionID: *dl.NewFeedVersionFileInfoWhereLoader(dl.FeedVersionFileInfoWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedVersionFileInfosByFeedVersionID,
			}),
			// Has a select method
			StopsByParentStopID: *dl.NewStopWhereLoader(dl.StopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.StopsByParentStopID,
			}),
			FeedVersionsByFeedID: *dl.NewFeedVersionWhereLoader(dl.FeedVersionWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedVersionsByFeedID,
			}),
			AgencyPlacesByAgencyID: *dl.NewAgencyPlaceWhereLoader(dl.AgencyPlaceWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.AgencyPlacesByAgencyID,
			}),
			RouteGeometriesByRouteID: *dl.NewRouteGeometryWhereLoader(dl.RouteGeometryWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.RouteGeometriesByRouteID,
			}),
			TripsByRouteID: *dl.NewTripWhereLoader(dl.TripWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.TripsByRouteID,
			}),
			RoutesByAgencyID: *dl.NewRouteWhereLoader(dl.RouteWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.RoutesByAgencyID,
			}),
			AgenciesByFeedVersionID: *dl.NewAgencyWhereLoader(dl.AgencyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.AgenciesByFeedVersionID,
			}),
			FeedFetchesByFeedID: *dl.NewFeedFetchWhereLoader(dl.FeedFetchWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedFetchesByFeedID,
			}),
			AgenciesByOnestopID: *dl.NewAgencyWhereLoader(dl.AgencyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.AgenciesByOnestopID,
			}),
			StopsByFeedVersionID: *dl.NewStopWhereLoader(dl.StopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.StopsByFeedVersionID,
			}),
			TripsByFeedVersionID: *dl.NewTripWhereLoader(dl.TripWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.TripsByFeedVersionID,
			}),
			FeedInfosByFeedVersionID: *dl.NewFeedInfoWhereLoader(dl.FeedInfoWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedInfosByFeedVersionID,
			}),

			RoutesByFeedVersionID: *dl.NewRouteWhereLoader(dl.RouteWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.RoutesByFeedVersionID,
			}),
			FeedVersionServiceLevelsByFeedVersionID: *dl.NewFeedVersionServiceLevelWhereLoader(dl.FeedVersionServiceLevelWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.FeedVersionServiceLevelsByFeedVersionID,
			}),
			PathwaysByFromStopID: *dl.NewPathwayWhereLoader(dl.PathwayWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.PathwaysByFromStopID,
			}),
			PathwaysByToStopID: *dl.NewPathwayWhereLoader(dl.PathwayWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.PathwaysByToStopID,
			}),
			CalendarDatesByServiceID: *dl.NewCalendarDateWhereLoader(dl.CalendarDateWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.CalendarDatesByServiceID,
			}),
			CensusGeographiesByEntityID: *dl.NewCensusGeographyWhereLoader(dl.CensusGeographyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.CensusGeographiesByEntityID,
			}),
			CensusValuesByGeographyID: *dl.NewCensusValueWhereLoader(dl.CensusValueWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.CensusValuesByGeographyID,
			}),
			CensusTableByID: *dl.NewCensusTableLoader(dl.CensusTableLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    finder.CensusTableByID,
			}),
		})
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// For .
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey{}).(*Loaders)
}
