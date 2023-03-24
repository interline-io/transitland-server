package resolvers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	generated "github.com/interline-io/transitland-server/generated/gqlgen"
	"github.com/interline-io/transitland-server/internal/fvsl"
	"github.com/interline-io/transitland-server/internal/meters"
	"github.com/interline-io/transitland-server/model"
)

func NewServer(cfg config.Config, dbfinder model.Finder, rtfinder model.RTFinder, gbfsFinder model.GbfsFinder) (http.Handler, error) {
	c := generated.Config{Resolvers: &Resolver{
		cfg:        cfg,
		finder:     dbfinder,
		rtfinder:   rtfinder,
		gbfsFinder: gbfsFinder,
		fvslCache:  fvsl.FVSLCache{Finder: dbfinder},
	}}
	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
		user := auth.ForContext(ctx)
		if user == nil || !user.HasRole(string(role)) {
			return nil, fmt.Errorf("access denied")
		}
		return next(ctx)
	}
	// Setup server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))
	graphqlServer := meterMiddleware(loaderMiddleware(cfg, dbfinder, srv))
	return graphqlServer, nil
}

func meterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if apiMeter := meters.ForContext(r.Context()); apiMeter != nil {
			// default to "unknown"
			apiMeter.Meter("graphql", 1.0, map[string]string{"resolver": "unknown"})
		}
	})
}
