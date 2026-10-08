package graph

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/auth"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/gqlerr"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/generated"
	"github.com/vektah/gqlparser/v2/ast"
)

// NewServer returns the GraphQL HTTP handler for POST requests.
func NewServer(r *Resolver) *handler.Server {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: r,
		// The guards in schema/schema.graphqls.
		Directives: generated.DirectiveRoot{
			Authenticated: auth.Authenticated,
			HasPermission: auth.HasPermission,
			HasRole:       auth.HasRole,
		},
	}))
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.SetErrorPresenter(gqlerr.Present)
	// Introspection powers the playground and editor autocompletion.
	srv.Use(extension.Introspection{})
	srv.Use(tracing{})
	return srv
}
