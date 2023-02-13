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
	StopExternalReferencesByStopID          dl.StopExternalReferenceLoader
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
	TargetStopsByStopID                     dl.StopLoader
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
	RouteStopPatternsByRouteID              dl.RouteStopPatternWhereLoader
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
		ctx := r.Context()
		rctx := context.WithValue(ctx, loadersKey{}, &Loaders{
			LevelsByID: *dl.NewLevelLoader(dl.LevelLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Level, []error) { return finder.LevelsByID(ctx, a) },
			}),
			TripsByID: *dl.NewTripLoader(dl.TripLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Trip, []error) { return finder.TripsByID(ctx, a) },
			}),
			CalendarsByID: *dl.NewCalendarLoader(dl.CalendarLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Calendar, []error) { return finder.CalendarsByID(ctx, a) },
			}),
			ShapesByID: *dl.NewShapeLoader(dl.ShapeLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Shape, []error) { return finder.ShapesByID(ctx, a) },
			}),
			FeedVersionsByID: *dl.NewFeedVersionLoader(dl.FeedVersionLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.FeedVersion, []error) { return finder.FeedVersionsByID(ctx, a) },
			}),
			StopExternalReferencesByStopID: *dl.NewStopExternalReferenceLoader(dl.StopExternalReferenceLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []int) ([]*model.StopExternalReference, []error) {
					return finder.StopExternalReferencesByStopID(ctx, a)
				},
			}),
			FeedsByID: *dl.NewFeedLoader(dl.FeedLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Feed, []error) { return finder.FeedsByID(ctx, a) },
			}),
			AgenciesByID: *dl.NewAgencyLoader(dl.AgencyLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Agency, []error) { return finder.AgenciesByID(ctx, a) },
			}),
			StopsByID: *dl.NewStopLoader(dl.StopLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Stop, []error) { return finder.StopsByID(ctx, a) },
			}),
			RoutesByID: *dl.NewRouteLoader(dl.RouteLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Route, []error) { return finder.RoutesByID(ctx, a) },
			}),
			// Other ID loaders
			FeedVersionGtfsImportsByFeedVersionID: *dl.NewFeedVersionGtfsImportLoader(dl.FeedVersionGtfsImportLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []int) ([]*model.FeedVersionGtfsImport, []error) {
					return finder.FeedVersionGtfsImportsByFeedVersionID(ctx, a)
				},
			}),
			FeedStatesByFeedID: *dl.NewFeedStateLoader(dl.FeedStateLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.FeedState, []error) { return finder.FeedStatesByFeedID(ctx, a) },
			}),
			FeedsByOperatorOnestopID: *dl.NewFeedWhereLoader(dl.FeedWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.FeedParam) ([][]*model.Feed, []error) { return finder.FeedsByOperatorOnestopID(ctx, a) },
			}),
			OperatorsByFeedID: *dl.NewOperatorWhereLoader(dl.OperatorWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.OperatorParam) ([][]*model.Operator, []error) { return finder.OperatorsByFeedID(ctx, a) },
			}),
			OperatorsByCOIF: *dl.NewOperatorLoader(dl.OperatorLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Operator, []error) { return finder.OperatorsByCOIF(ctx, a) },
			}),
			TargetStopsByStopID: *dl.NewStopLoader(dl.StopLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []int) ([]*model.Stop, []error) { return finder.TargetStopsByStopID(ctx, a) },
			}),
			// Where loaders
			FrequenciesByTripID: *dl.NewFrequencyWhereLoader(dl.FrequencyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.FrequencyParam) ([][]*model.Frequency, []error) {
					return finder.FrequenciesByTripID(ctx, a)
				},
			}),
			StopTimesByTripID: *dl.NewStopTimeWhereLoader(dl.StopTimeWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.StopTimeParam) ([][]*model.StopTime, []error) { return finder.StopTimesByTripID(ctx, a) },
			}),
			StopTimesByStopID: *dl.NewStopTimeWhereLoader(dl.StopTimeWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.StopTimeParam) ([][]*model.StopTime, []error) { return finder.StopTimesByStopID(ctx, a) },
			}),
			RouteStopsByStopID: *dl.NewRouteStopWhereLoader(dl.RouteStopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.RouteStopParam) ([][]*model.RouteStop, []error) {
					return finder.RouteStopsByStopID(ctx, a)
				},
			}),
			StopsByRouteID: *dl.NewStopWhereLoader(dl.StopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.StopParam) ([][]*model.Stop, []error) { return finder.StopsByRouteID(ctx, a) },
			}),
			RouteStopsByRouteID: *dl.NewRouteStopWhereLoader(dl.RouteStopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.RouteStopParam) ([][]*model.RouteStop, []error) {
					return finder.RouteStopsByRouteID(ctx, a)
				},
			}),
			RouteHeadwaysByRouteID: *dl.NewRouteHeadwayWhereLoader(dl.RouteHeadwayWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.RouteHeadwayParam) ([][]*model.RouteHeadway, []error) {
					return finder.RouteHeadwaysByRouteID(ctx, a)
				},
			}),
			RouteStopPatternsByRouteID: *dl.NewRouteStopPatternWhereLoader(dl.RouteStopPatternWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.RouteStopPatternParam) ([][]*model.RouteStopPattern, []error) {
					return finder.RouteStopPatternsByRouteID(ctx, a)
				},
			}),
			FeedVersionFileInfosByFeedVersionID: *dl.NewFeedVersionFileInfoWhereLoader(dl.FeedVersionFileInfoWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.FeedVersionFileInfoParam) ([][]*model.FeedVersionFileInfo, []error) {
					return finder.FeedVersionFileInfosByFeedVersionID(ctx, a)
				},
			}),
			// Has a select method
			StopsByParentStopID: *dl.NewStopWhereLoader(dl.StopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.StopParam) ([][]*model.Stop, []error) { return finder.StopsByParentStopID(ctx, a) },
			}),
			FeedVersionsByFeedID: *dl.NewFeedVersionWhereLoader(dl.FeedVersionWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.FeedVersionParam) ([][]*model.FeedVersion, []error) {
					return finder.FeedVersionsByFeedID(ctx, a)
				},
			}),
			AgencyPlacesByAgencyID: *dl.NewAgencyPlaceWhereLoader(dl.AgencyPlaceWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.AgencyPlaceParam) ([][]*model.AgencyPlace, []error) {
					return finder.AgencyPlacesByAgencyID(ctx, a)
				},
			}),
			RouteGeometriesByRouteID: *dl.NewRouteGeometryWhereLoader(dl.RouteGeometryWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.RouteGeometryParam) ([][]*model.RouteGeometry, []error) {
					return finder.RouteGeometriesByRouteID(ctx, a)
				},
			}),
			TripsByRouteID: *dl.NewTripWhereLoader(dl.TripWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.TripParam) ([][]*model.Trip, []error) { return finder.TripsByRouteID(ctx, a) },
			}),
			RoutesByAgencyID: *dl.NewRouteWhereLoader(dl.RouteWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.RouteParam) ([][]*model.Route, []error) { return finder.RoutesByAgencyID(ctx, a) },
			}),
			AgenciesByFeedVersionID: *dl.NewAgencyWhereLoader(dl.AgencyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.AgencyParam) ([][]*model.Agency, []error) {
					return finder.AgenciesByFeedVersionID(ctx, a)
				},
			}),
			FeedFetchesByFeedID: *dl.NewFeedFetchWhereLoader(dl.FeedFetchWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.FeedFetchParam) ([][]*model.FeedFetch, []error) {
					return finder.FeedFetchesByFeedID(ctx, a)
				},
			}),
			AgenciesByOnestopID: *dl.NewAgencyWhereLoader(dl.AgencyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.AgencyParam) ([][]*model.Agency, []error) { return finder.AgenciesByOnestopID(ctx, a) },
			}),
			StopsByFeedVersionID: *dl.NewStopWhereLoader(dl.StopWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.StopParam) ([][]*model.Stop, []error) { return finder.StopsByFeedVersionID(ctx, a) },
			}),
			TripsByFeedVersionID: *dl.NewTripWhereLoader(dl.TripWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.TripParam) ([][]*model.Trip, []error) { return finder.TripsByFeedVersionID(ctx, a) },
			}),
			FeedInfosByFeedVersionID: *dl.NewFeedInfoWhereLoader(dl.FeedInfoWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.FeedInfoParam) ([][]*model.FeedInfo, []error) {
					return finder.FeedInfosByFeedVersionID(ctx, a)
				},
			}),

			RoutesByFeedVersionID: *dl.NewRouteWhereLoader(dl.RouteWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.RouteParam) ([][]*model.Route, []error) { return finder.RoutesByFeedVersionID(ctx, a) },
			}),
			FeedVersionServiceLevelsByFeedVersionID: *dl.NewFeedVersionServiceLevelWhereLoader(dl.FeedVersionServiceLevelWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.FeedVersionServiceLevelParam) ([][]*model.FeedVersionServiceLevel, []error) {
					return finder.FeedVersionServiceLevelsByFeedVersionID(ctx, a)
				},
			}),
			PathwaysByFromStopID: *dl.NewPathwayWhereLoader(dl.PathwayWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.PathwayParam) ([][]*model.Pathway, []error) { return finder.PathwaysByFromStopID(ctx, a) },
			}),
			PathwaysByToStopID: *dl.NewPathwayWhereLoader(dl.PathwayWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch:    func(a []model.PathwayParam) ([][]*model.Pathway, []error) { return finder.PathwaysByToStopID(ctx, a) },
			}),
			CalendarDatesByServiceID: *dl.NewCalendarDateWhereLoader(dl.CalendarDateWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.CalendarDateParam) ([][]*model.CalendarDate, []error) {
					return finder.CalendarDatesByServiceID(ctx, a)
				},
			}),
			CensusGeographiesByEntityID: *dl.NewCensusGeographyWhereLoader(dl.CensusGeographyWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.CensusGeographyParam) ([][]*model.CensusGeography, []error) {
					return finder.CensusGeographiesByEntityID(ctx, a)
				},
			}),
			CensusValuesByGeographyID: *dl.NewCensusValueWhereLoader(dl.CensusValueWhereLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []model.CensusValueParam) ([][]*model.CensusValue, []error) {
					return finder.CensusValuesByGeographyID(ctx, a)
				},
			}),
			CensusTableByID: *dl.NewCensusTableLoader(dl.CensusTableLoaderConfig{
				MaxBatch: MAXBATCH,
				Wait:     WAIT,
				Fetch: func(a []int) ([]*model.CensusTable, []error) {
					return finder.CensusTableByID(ctx, a)
				},
			}),
		})
		r = r.WithContext(rctx)
		next.ServeHTTP(w, r)
	})
}

// For .
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey{}).(*Loaders)
}
