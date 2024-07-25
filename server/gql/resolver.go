package gql

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tlxy"
	"github.com/interline-io/transitland-mw/meters"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/model"
)

// DEFAULTLIMIT is the default API limit
const DEFAULTLIMIT = 100

// MAXLIMIT is the API limit maximum
var MAXLIMIT = 1_000

// checkLimit checks the limit is positive and below the maximum limit.
func checkLimit(limit *int) *int {
	a := 0
	if limit != nil {
		a = *limit
	}
	if a <= 0 {
		a = DEFAULTLIMIT
	} else if a >= MAXLIMIT {
		a = MAXLIMIT
	}
	return &a
}

func checkCursor(after *int) *model.Cursor {
	var cursor *model.Cursor
	if after != nil {
		c := model.NewCursor(0, *after)
		cursor = &c
	}
	return cursor
}

func addMetric(ctx context.Context, resolverName string) {
	if apiMeter := meters.ForContext(ctx); apiMeter != nil {
		apiMeter.AddDimension("graphql", "resolver", resolverName)
	}
}

func checkGeo(near *model.PointRadius, bbox *model.BoundingBox) error {
	// We only want to enforce this check on top level resolvers
	if near != nil && near.Radius > MAX_RADIUS {
		return errors.New("radius too large")
	}
	if bbox != nil && !checkBbox(bbox, MAX_RADIUS*MAX_RADIUS) {
		return errors.New("bbox too large")
	}
	return nil
}

func checkBbox(bbox *model.BoundingBox, maxAreaM2 float64) bool {
	approxDiag := tlxy.DistanceHaversine(tlxy.Point{Lon: bbox.MinLon, Lat: bbox.MinLat}, tlxy.Point{Lon: bbox.MaxLon, Lat: bbox.MaxLat})
	// fmt.Println("approxDiag:", approxDiag)
	approxArea := 0.5 * (approxDiag * approxDiag)
	// fmt.Println("approxArea:", approxArea, "maxAreaM2:", maxAreaM2)
	return approxArea < maxAreaM2
}

func atoi(v string) int {
	a, _ := strconv.Atoi(v)
	return a
}

func nilOr[T any, PT *T](v PT, def T) T {
	if v == nil {
		return def
	}
	return *v
}

func ptr[T any, PT *T](v T) PT {
	a := v
	return &a
}

func kebabize(a string) string {
	return strings.ReplaceAll(strings.ToLower(a), "_", "-")
}

func tzTruncate(s time.Time, loc *time.Location) *tt.Date {
	return ptr(tt.NewDate(time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, loc)))
}

func checkFloat(v *float64, min float64, max float64) float64 {
	if v == nil || *v < min {
		return min
	} else if *v > max {
		return max
	}
	return *v
}

// Resolver .
type Resolver struct{}

// Query .
func (r *Resolver) Query() gqlout.QueryResolver { return &queryResolver{r} }

// Mutation .
func (r *Resolver) Mutation() gqlout.MutationResolver { return &mutationResolver{r} }

// Agency .
func (r *Resolver) Agency() gqlout.AgencyResolver { return &agencyResolver{r} }

// Feed .
func (r *Resolver) Feed() gqlout.FeedResolver { return &feedResolver{r} }

// FeedState .
func (r *Resolver) FeedState() gqlout.FeedStateResolver { return &feedStateResolver{r} }

// FeedVersion .
func (r *Resolver) FeedVersion() gqlout.FeedVersionResolver { return &feedVersionResolver{r} }

// Route .
func (r *Resolver) Route() gqlout.RouteResolver { return &routeResolver{r} }

// RouteStop .
func (r *Resolver) RouteStop() gqlout.RouteStopResolver { return &routeStopResolver{r} }

// RouteHeadway .
func (r *Resolver) RouteHeadway() gqlout.RouteHeadwayResolver { return &routeHeadwayResolver{r} }

// RouteStopPattern .
func (r *Resolver) RouteStopPattern() gqlout.RouteStopPatternResolver {
	return &routePatternResolver{r}
}

// Segment .
func (r *Resolver) Segment() gqlout.SegmentResolver { return &segmentResolver{r} }

// SegmentPattern .
func (r *Resolver) SegmentPattern() gqlout.SegmentPatternResolver { return &segmentPatternResolver{r} }

// Stop .
func (r *Resolver) Stop() gqlout.StopResolver { return &stopResolver{r} }

// Trip .
func (r *Resolver) Trip() gqlout.TripResolver { return &tripResolver{r} }

// StopTime .
func (r *Resolver) StopTime() gqlout.StopTimeResolver { return &stopTimeResolver{r} }

// Operator .
func (r *Resolver) Operator() gqlout.OperatorResolver { return &operatorResolver{r} }

// FeedVersionGtfsImport .
func (r *Resolver) FeedVersionGtfsImport() gqlout.FeedVersionGtfsImportResolver {
	return &feedVersionGtfsImportResolver{r}
}

func (r *Resolver) Level() gqlout.LevelResolver {
	return &levelResolver{r}
}

// Calendar .
func (r *Resolver) Calendar() gqlout.CalendarResolver {
	return &calendarResolver{r}
}

// CensusGeography .
func (r *Resolver) CensusGeography() gqlout.CensusGeographyResolver {
	return &censusGeographyResolver{r}
}

// Pathway .
func (r *Resolver) Pathway() gqlout.PathwayResolver {
	return &pathwayResolver{r}
}

// StopExternalReference .
func (r *Resolver) StopExternalReference() gqlout.StopExternalReferenceResolver {
	return &stopExternalReferenceResolver{r}
}

// Directions .
func (r *Resolver) Directions(ctx context.Context, where model.DirectionRequest) (*model.Directions, error) {
	dr := directionsResolver{r}
	return dr.Directions(ctx, where)
}

func (r *Resolver) Place() gqlout.PlaceResolver {
	return &placeResolver{r}
}

func (r *Resolver) ValidationReport() gqlout.ValidationReportResolver {
	return &validationReportResolver{r}
}

func (r *Resolver) ValidationReportErrorGroup() gqlout.ValidationReportErrorGroupResolver {
	return &validationReportErrorGroupResolver{r}
}
