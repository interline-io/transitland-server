package gql

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/internal/fvsl"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/model"
)

func NewServer(cfg config.Config, dbfinder model.Finder, rtfinder model.RTFinder, gbfsFinder model.GbfsFinder, checker model.Checker) (http.Handler, error) {
	c := gqlout.Config{Resolvers: &Resolver{
		cfg:          cfg,
		finder:       dbfinder,
		rtfinder:     rtfinder,
		gbfsFinder:   gbfsFinder,
		fvslCache:    fvsl.NewFVSLCache(dbfinder),
		authzChecker: checker,
	}}
	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
		user := authn.ForContext(ctx)
		if user == nil || !user.HasRole(string(role)) {
			return nil, fmt.Errorf("access denied")
		}
		return next(ctx)
	}
	// Setup server
	srv := handler.NewDefaultServer(gqlout.NewExecutableSchema(c))
	graphqlServer := loaderMiddleware(cfg, dbfinder, srv)
	return graphqlServer, nil
}
