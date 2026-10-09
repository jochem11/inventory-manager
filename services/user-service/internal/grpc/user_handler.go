// Package grpc exposes the user service over gRPC. The handlers only convert
// between protobuf messages and the service's types; the service does the
// work.
package grpc

import (
	"context"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	userpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/user"
	"github.com/jochem11/inventory-manager/shared/errs"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	users domain.UserService
}

func NewUserHandler(users domain.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	user, err := h.users.FindByID(ctx, req.GetId())
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &userpb.GetUserResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) GetUsers(ctx context.Context, req *userpb.GetUsersRequest) (*userpb.GetUsersResponse, error) {
	users, err := h.users.FindByIDs(ctx, req.GetIds())
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &userpb.GetUsersResponse{Users: toProtoUsers(users)}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	page, err := h.users.FindAll(ctx, fromProtoListRequest(req))
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &userpb.ListUsersResponse{
		Users:      toProtoUsers(page.Nodes),
		TotalCount: int32(page.TotalCount),
	}, nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user, err := h.users.Create(ctx, fromProtoInput(req.GetUser()))
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &userpb.CreateUserResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	user, err := h.users.Update(ctx, req.GetId(), fromProtoInput(req.GetUser()))
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &userpb.UpdateUserResponse{User: toProtoUser(user)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	if err := h.users.Delete(ctx, req.GetId()); err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &userpb.DeleteUserResponse{}, nil
}
