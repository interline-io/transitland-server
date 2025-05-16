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
	AgenciesByID                                                 *dataloader.Loader[int, *model.Agency]
	AgenciesByOnestopIDs                                         *dataloader.Loader[model.AgencyParam, []*model.Agency]
	AgencyPlacesByAgencyIDs                                      *dataloader.Loader[model.AgencyPlaceParam, []*model.AgencyPlace]
	CalendarDatesByServiceID                                     *dataloader.Loader[model.CalendarDateParam, []*model.CalendarDate]
	CalendarsByID                                                *dataloader.Loader[int, *model.Calendar]
	CensusDatasetLayersByDatasetID                               *dataloader.Loader[int, []string]
	CensusSourceLayersBySourceID                                 *dataloader.Loader[int, []string]
	CensusFieldsByTableID                                        *dataloader.Loader[model.CensusFieldParam, []*model.CensusField]
	CensusGeographiesByDatasetID                                 *dataloader.Loader[model.CensusDatasetGeographyParam, []*model.CensusGeography]
	CensusGeographiesByEntityID                                  *dataloader.Loader[model.CensusGeographyParam, []*model.CensusGeography]
	CensusSourcesByDatasetID                                     *dataloader.Loader[model.CensusSourceParam, []*model.CensusSource]
	CensusTableByID                                              *dataloader.Loader[int, *model.CensusTable]
	CensusValuesByGeographyID                                    *dataloader.Loader[model.CensusValueParam, []*model.CensusValue]
	FeedFetchesByFeedID                                          *dataloader.Loader[model.FeedFetchParam, []*model.FeedFetch]
	FeedInfosByFeedVersionID                                     *dataloader.Loader[model.FeedInfoParam, []*model.FeedInfo]
	FeedsByID                                                    *dataloader.Loader[int, *model.Feed]
	FeedsByOperatorOnestopID                                     *dataloader.Loader[model.FeedParam, []*model.Feed]
	FeedStatesByFeedID                                           *dataloader.Loader[int, *model.FeedState]
	FeedVersionFileInfosByFeedVersionID                          *dataloader.Loader[model.FeedVersionFileInfoParam, []*model.FeedVersionFileInfo]
	FeedVersionGeometryByID                                      *dataloader.Loader[int, *tt.Polygon]
	FeedVersionGtfsImportByFeedVersionID                         *dataloader.Loader[int, *model.FeedVersionGtfsImport]
	FeedVersionsByFeedID                                         *dataloader.Loader[model.FeedVersionParam, []*model.FeedVersion]
	FeedVersionsByID                                             *dataloader.Loader[int, *model.FeedVersion]
	FeedVersionServiceLevelsByFeedVersionID                      *dataloader.Loader[model.FeedVersionServiceLevelParam, []*model.FeedVersionServiceLevel]
	FeedVersionServiceWindowByFeedVersionID                      *dataloader.Loader[int, *model.FeedVersionServiceWindow]
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
	SegmentsByFeedVersionID                                      *dataloader.Loader[model.SegmentParam, []*model.Segment]
	SegmentsByID                                                 *dataloader.Loader[int, *model.Segment]
	SegmentsByRouteID                                            *dataloader.Loader[model.SegmentParam, []*model.Segment]
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
		AgenciesByID: withWaitAndCapacity(waitTime, batchSize, dbf.AgenciesByID),
		AgenciesByOnestopIDs: withWaitAndCapacity(waitTime, batchSize, func(ctx context.Context, params []model.AgencyParam) ([][]*model.Agency, []error) {
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
		AgencyPlacesByAgencyIDs: withWaitAndCapacity(waitTime, batchSize, func(ctx context.Context, params []model.AgencyPlaceParam) ([][]*model.AgencyPlace, []error) {
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
		CalendarDatesByServiceID:                withWaitAndCapacity(waitTime, batchSize, dbf.CalendarDatesByServiceID),
		CalendarsByID:                           withWaitAndCapacity(waitTime, batchSize, dbf.CalendarsByID),
		CensusDatasetLayersByDatasetID:          withWaitAndCapacity(waitTime, batchSize, dbf.CensusDatasetLayersByDatasetID),
		CensusSourceLayersBySourceID:            withWaitAndCapacity(waitTime, batchSize, dbf.CensusSourceLayersBySourceID),
		CensusFieldsByTableID:                   withWaitAndCapacity(waitTime, batchSize, dbf.CensusFieldsByTableID),
		CensusGeographiesByDatasetID:            withWaitAndCapacity(waitTime, batchSize, dbf.CensusGeographiesByDatasetID),
		CensusGeographiesByEntityID:             withWaitAndCapacity(waitTime, batchSize, dbf.CensusGeographiesByEntityID),
		CensusSourcesByDatasetID:                withWaitAndCapacity(waitTime, batchSize, dbf.CensusSourcesByDatasetID),
		CensusTableByID:                         withWaitAndCapacity(waitTime, batchSize, dbf.CensusTableByID),
		CensusValuesByGeographyID:               withWaitAndCapacity(waitTime, batchSize, dbf.CensusValuesByGeographyID),
		FeedFetchesByFeedID:                     withWaitAndCapacity(waitTime, batchSize, dbf.FeedFetchesByFeedID),
		FeedInfosByFeedVersionID:                withWaitAndCapacity(waitTime, batchSize, dbf.FeedInfosByFeedVersionID),
		FeedsByID:                               withWaitAndCapacity(waitTime, batchSize, dbf.FeedsByID),
		FeedsByOperatorOnestopID:                withWaitAndCapacity(waitTime, batchSize, dbf.FeedsByOperatorOnestopID),
		FeedStatesByFeedID:                      withWaitAndCapacity(waitTime, batchSize, dbf.FeedStatesByFeedID),
		FeedVersionFileInfosByFeedVersionID:     withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionFileInfosByFeedVersionID),
		FeedVersionGeometryByID:                 withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionGeometryByID),
		FeedVersionGtfsImportByFeedVersionID:    withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionGtfsImportByFeedVersionID),
		FeedVersionsByFeedID:                    withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionsByFeedID),
		FeedVersionsByID:                        withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionsByID),
		FeedVersionServiceLevelsByFeedVersionID: withWaitAndCapacity(waitTime, batchSize, dbf.FeedVersionServiceLevelsByFeedVersionID),
		FeedVersionServiceWindowByFeedVersionID: withWaitAndCapacity(waitTime, maxBatch, dbf.FeedVersionServiceWindowByFeedVersionID),
		FrequenciesByTripID:                     withWaitAndCapacity(waitTime, batchSize, dbf.FrequenciesByTripID),
		LevelsByID:                              withWaitAndCapacity(waitTime, batchSize, dbf.LevelsByID),
		LevelsByParentStationID:                 withWaitAndCapacity(waitTime, batchSize, dbf.LevelsByParentStationID),
		OperatorsByAgencyID:                     withWaitAndCapacity(waitTime, batchSize, dbf.OperatorsByAgencyID),
		OperatorsByCOIF:                         withWaitAndCapacity(waitTime, batchSize, dbf.OperatorsByCOIF),
		OperatorsByFeedID:                       withWaitAndCapacity(waitTime, batchSize, dbf.OperatorsByFeedID),
		PathwaysByFromStopID:                    withWaitAndCapacity(waitTime, batchSize, dbf.PathwaysByFromStopID),
		PathwaysByID:                            withWaitAndCapacity(waitTime, batchSize, dbf.PathwaysByID),
		PathwaysByToStopID:                      withWaitAndCapacity(waitTime, batchSize, dbf.PathwaysByToStopID),
		RouteAttributesByRouteID:                withWaitAndCapacity(waitTime, batchSize, dbf.RouteAttributesByRouteID),
		RouteGeometriesByRouteID:                withWaitAndCapacity(waitTime, batchSize, dbf.RouteGeometriesByRouteID),
		RouteHeadwaysByRouteID:                  withWaitAndCapacity(waitTime, batchSize, dbf.RouteHeadwaysByRouteID),
		RoutesByAgencyID:                        withWaitAndCapacity(waitTime, batchSize, dbf.RoutesByAgencyID),
		RoutesByFeedVersionID:                   withWaitAndCapacity(waitTime, batchSize, dbf.RoutesByFeedVersionID),
		RoutesByID:                              withWaitAndCapacity(waitTime, batchSize, dbf.RoutesByID),
		RouteStopPatternsByRouteID:              withWaitAndCapacity(waitTime, batchSize, dbf.RouteStopPatternsByRouteID),
		RouteStopsByRouteID:                     withWaitAndCapacity(waitTime, batchSize, dbf.RouteStopsByRouteID),
		RouteStopsByStopID:                      withWaitAndCapacity(waitTime, batchSize, dbf.RouteStopsByStopID),
		SegmentPatternsByRouteID:                withWaitAndCapacity(waitTime, batchSize, dbf.SegmentPatternsByRouteID),
		SegmentPatternsBySegmentID:              withWaitAndCapacity(waitTime, batchSize, dbf.SegmentPatternsBySegmentID),
		SegmentsByFeedVersionID:                 withWaitAndCapacity(waitTime, batchSize, dbf.SegmentsByFeedVersionID),
		SegmentsByID:                            withWaitAndCapacity(waitTime, batchSize, dbf.SegmentsByID),
		SegmentsByRouteID:                       withWaitAndCapacity(waitTime, batchSize, dbf.SegmentsByRouteID),
		ShapesByID:                              withWaitAndCapacity(waitTime, batchSize, dbf.ShapesByID),
		StopExternalReferencesByStopID:          withWaitAndCapacity(waitTime, batchSize, dbf.StopExternalReferencesByStopID),
		StopObservationsByStopID:                withWaitAndCapacity(waitTime, batchSize, dbf.StopObservationsByStopID),
		StopPlacesByStopID:                      withWaitAndCapacity(waitTime, batchSize, dbf.StopPlacesByStopID),
		StopsByFeedVersionID:                    withWaitAndCapacity(waitTime, batchSize, dbf.StopsByFeedVersionID),
		StopsByID:                               withWaitAndCapacity(waitTime, batchSize, dbf.StopsByID),
		StopsByLevelID:                          withWaitAndCapacity(waitTime, batchSize, dbf.StopsByLevelID),
		StopsByParentStopID:                     withWaitAndCapacity(waitTime, batchSize, dbf.StopsByParentStopID),
		StopsByRouteID:                          withWaitAndCapacity(waitTime, batchSize, dbf.StopsByRouteID),
		StopTimesByStopID:                       withWaitAndCapacity(waitTime, stopTimeBatchSize, dbf.StopTimesByStopID),
		StopTimesByTripID:                       withWaitAndCapacity(waitTime, batchSize, dbf.StopTimesByTripID),
		TargetStopsByStopID:                     withWaitAndCapacity(waitTime, batchSize, dbf.TargetStopsByStopID),
		TripsByFeedVersionID:                    withWaitAndCapacity(waitTime, batchSize, dbf.TripsByFeedVersionID),
		TripsByID:                               withWaitAndCapacity(waitTime, batchSize, dbf.TripsByID),
		TripsByRouteID:                          withWaitAndCapacity(waitTime, batchSize, dbf.TripsByRouteID),
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
