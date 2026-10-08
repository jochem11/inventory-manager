// Package user resolves the user queries and mutations by calling the
// user-service. The gqlgen resolvers in package graph call into it.
package user

import (
	"context"

	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/auth"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/model"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/list"
	userpb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Resolver returns gRPC errors as they are; gqlerr.Present maps them.
type Resolver struct {
	client userpb.UserServiceClient
}

func NewResolver(client userpb.UserServiceClient) *Resolver {
	return &Resolver{client: client}
}

// The permissions these queries and mutations need are guarded in the schema
// (@hasPermission in schema/user.graphqls), before a resolver runs.

// Get returns nil when no user has this id.
func (r *Resolver) Get(ctx context.Context, id string) (*model.User, error) {
	return r.Profile(ctx, id)
}

// Me returns the logged-in user's profile. Everyone may read their own, so
// it needs no permission.
func (r *Resolver) Me(ctx context.Context) (*model.User, error) {
	claims, err := auth.Require(ctx)
	if err != nil {
		return nil, err
	}
	return r.Profile(ctx, claims.Subject)
}

// Profile returns a user without checking permissions, for the caller's own
// profile (Me, and the user in a login's AuthPayload). It returns nil when
// the user doesn't exist (yet).
func (r *Resolver) Profile(ctx context.Context, id string) (*model.User, error) {
	resp, err := r.client.GetUser(ctx, &userpb.GetUserRequest{Id: id})
	if status.Code(err) == codes.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUser(resp.GetUser()), nil
}

func (r *Resolver) List(ctx context.Context, offset, limit int, orderBy *model.UserOrder, filter *model.UserFilter) (*model.UserConnection, error) {
	if err := list.ValidatePage(offset, limit); err != nil {
		return nil, err
	}
	resp, err := r.client.ListUsers(ctx, toListUsersRequest(offset, limit, orderBy, filter))
	if err != nil {
		return nil, err
	}
	return &model.UserConnection{
		TotalCount: int(resp.GetTotalCount()),
		Nodes:      toUsers(resp.GetUsers()),
	}, nil
}

func (r *Resolver) Create(ctx context.Context, input model.UserInput) (*model.User, error) {
	resp, err := r.client.CreateUser(ctx, &userpb.CreateUserRequest{User: toProtoUserInput(input)})
	if err != nil {
		return nil, err
	}
	return toUser(resp.GetUser()), nil
}

func (r *Resolver) Update(ctx context.Context, id string, input model.UserInput) (*model.User, error) {
	resp, err := r.client.UpdateUser(ctx, &userpb.UpdateUserRequest{Id: id, User: toProtoUserInput(input)})
	if err != nil {
		return nil, err
	}
	return toUser(resp.GetUser()), nil
}

// Delete returns the id of the deleted user.
func (r *Resolver) Delete(ctx context.Context, id string) (string, error) {
	if _, err := r.client.DeleteUser(ctx, &userpb.DeleteUserRequest{Id: id}); err != nil {
		return "", err
	}
	return id, nil
}
