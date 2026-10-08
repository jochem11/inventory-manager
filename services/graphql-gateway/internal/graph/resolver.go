// Package graph wires the gqlgen schema to the per-service resolvers. The
// resolver methods here are one-liners: the work happens in the service's own
// package (internal/user, …).
package graph

import (
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/auth"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/user"
)

// Resolver holds one resolver per service.
type Resolver struct {
	UserResolver *user.Resolver
	AuthResolver *auth.Resolver
}
