package resolvers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gorilla/mux"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	generated "github.com/interline-io/transitland-server/generated/gqlgen"
	"github.com/interline-io/transitland-server/internal/fvsl"
	"github.com/interline-io/transitland-server/model"
)

func NewServer(cfg config.Config, dbfinder model.Finder, rtfinder model.RTFinder) (http.Handler, error) {
	c := generated.Config{Resolvers: &Resolver{
		cfg:       cfg,
		finder:    dbfinder,
		rtfinder:  rtfinder,
		fvslCache: fvsl.FVSLCache{Finder: dbfinder},
	}}
	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
		user := auth.ForContext(ctx)
		if user == nil {
			user = &auth.User{}
		}
		if !user.HasRole(string(role)) {
			return nil, fmt.Errorf("access denied")
		}
		return next(ctx)
	}
	// Setup server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))
	graphqlServer := Middleware(cfg, dbfinder, srv)
	root := mux.NewRouter()
	root.Handle("/", graphqlServer).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	return root, nil
}
