package gql

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/interline-io/transitland-mw/auth/authn"
	"github.com/interline-io/transitland-server/internal/generated/gqlout"
	"github.com/interline-io/transitland-server/model"
)

func NewServer(te model.Config) (http.Handler, error) {
	c := gqlout.Config{Resolvers: &Resolver{
		fvslCache: newFvslCache(te.Finder),
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
	graphqlServer := loaderMiddleware(te, srv)
	graphqlServer = model.AddConfig(te)(graphqlServer)
	return graphqlServer, nil
}
