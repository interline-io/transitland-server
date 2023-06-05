package resolvers

import (
	"context"
	"strconv"

	"github.com/interline-io/transitland-server/authz"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/generated/gqlgen"
	"github.com/interline-io/transitland-server/internal/fvsl"
	"github.com/interline-io/transitland-server/model"
)

func atoi(v string) int {
	a, _ := strconv.Atoi(v)
	return a
}

// Resolver .
type Resolver struct {
	cfg          config.Config
	rtfinder     model.RTFinder
	finder       model.Finder
	gbfsFinder   model.GbfsFinder
	fvslCache    *fvsl.FVSLCache
	authzChecker *authz.Checker
}

// Query .
func (r *Resolver) Query() gqlgen.QueryResolver { return &queryResolver{r} }

// Mutation .
func (r *Resolver) Mutation() gqlgen.MutationResolver { return &mutationResolver{r} }

// Agency .
func (r *Resolver) Agency() gqlgen.AgencyResolver { return &agencyResolver{r} }

// Feed .
func (r *Resolver) Feed() gqlgen.FeedResolver { return &feedResolver{r} }

// FeedState .
func (r *Resolver) FeedState() gqlgen.FeedStateResolver { return &feedStateResolver{r} }

// FeedVersion .
func (r *Resolver) FeedVersion() gqlgen.FeedVersionResolver { return &feedVersionResolver{r} }

// Route .
func (r *Resolver) Route() gqlgen.RouteResolver { return &routeResolver{r} }

// RouteStop .
func (r *Resolver) RouteStop() gqlgen.RouteStopResolver { return &routeStopResolver{r} }

// RouteHeadway .
func (r *Resolver) RouteHeadway() gqlgen.RouteHeadwayResolver { return &routeHeadwayResolver{r} }

// RouteStopPattern .
func (r *Resolver) RouteStopPattern() gqlgen.RouteStopPatternResolver {
	return &routePatternResolver{r}
}

// Stop .
func (r *Resolver) Stop() gqlgen.StopResolver { return &stopResolver{r} }

// Trip .
func (r *Resolver) Trip() gqlgen.TripResolver { return &tripResolver{r} }

// StopTime .
func (r *Resolver) StopTime() gqlgen.StopTimeResolver { return &stopTimeResolver{r} }

// Operator .
func (r *Resolver) Operator() gqlgen.OperatorResolver { return &operatorResolver{r} }

// FeedVersionGtfsImport .
func (r *Resolver) FeedVersionGtfsImport() gqlgen.FeedVersionGtfsImportResolver {
	return &feedVersionGtfsImportResolver{r}
}

func (r *Resolver) Level() gqlgen.LevelResolver {
	return &levelResolver{r}
}

// Calendar .
func (r *Resolver) Calendar() gqlgen.CalendarResolver {
	return &calendarResolver{r}
}

// CensusGeography .
func (r *Resolver) CensusGeography() gqlgen.CensusGeographyResolver {
	return &censusGeographyResolver{r}
}

// CensusValue .
func (r *Resolver) CensusValue() gqlgen.CensusValueResolver {
	return &censusValueResolver{r}
}

// Pathway .
func (r *Resolver) Pathway() gqlgen.PathwayResolver {
	return &pathwayResolver{r}
}

// StopExternalReference .
func (r *Resolver) StopExternalReference() gqlgen.StopExternalReferenceResolver {
	return &stopExternalReferenceResolver{r}
}

// Directions .
func (r *Resolver) Directions(ctx context.Context, where model.DirectionRequest) (*model.Directions, error) {
	dr := directionsResolver{r}
	return dr.Directions(ctx, where)
}

func (r *Resolver) Place() gqlgen.PlaceResolver {
	return &placeResolver{r}
}
