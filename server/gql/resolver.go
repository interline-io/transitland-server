package gql

import (
	"context"
	"strconv"

	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/internal/fvsl"
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
	authzChecker model.Checker
	fvslCache    *fvsl.FVSLCache
}

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

// CensusValue .
func (r *Resolver) CensusValue() gqlout.CensusValueResolver {
	return &censusValueResolver{r}
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
