package auth

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

// The schema's guard directives (schema/schema.graphqls). gqlgen runs them
// before the field's resolver; an error stops the resolver from running.

// Authenticated is @authenticated: the field needs a login.
func Authenticated(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	if _, err := Require(ctx); err != nil {
		return nil, err
	}
	return next(ctx)
}

// HasPermission is @hasPermission(permission): the field needs a login with
// that permission.
func HasPermission(ctx context.Context, _ any, next graphql.Resolver, permission string) (any, error) {
	if err := RequirePermission(ctx, permission); err != nil {
		return nil, err
	}
	return next(ctx)
}

// HasRole is @hasRole(roles): the field needs a login with one of the roles.
func HasRole(ctx context.Context, _ any, next graphql.Resolver, roles []string) (any, error) {
	if err := RequireRole(ctx, roles...); err != nil {
		return nil, err
	}
	return next(ctx)
}
