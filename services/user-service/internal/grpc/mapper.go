package grpc

import (
	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	userpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/user"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/paging"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoUser(u *models.User) *userpb.User {
	return &userpb.User{
		Id:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		AvatarUrl: u.AvatarURL,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func toProtoUsers(users []*models.User) []*userpb.User {
	out := make([]*userpb.User, len(users))
	for i, u := range users {
		out[i] = toProtoUser(u)
	}
	return out
}

// fromProtoInput accepts a nil input (the field left out of the request),
// which then fails validation on the required fields.
func fromProtoInput(in *userpb.UserInput) types.UserInput {
	if in == nil {
		return types.UserInput{}
	}
	return types.UserInput{
		FirstName: in.GetFirstName(),
		LastName:  in.GetLastName(),
		Email:     in.GetEmail(),
		Phone:     in.Phone,
		AvatarURL: in.AvatarUrl,
	}
}

func fromProtoListRequest(req *userpb.ListUsersRequest) types.UserListParams {
	f := req.GetFilter()
	params := types.UserListParams{
		Params: paging.Params{Offset: int(req.GetOffset()), Limit: int(req.GetLimit())},
		Filter: types.UserFilter{
			Search:    f.GetSearch(),
			FirstName: f.GetFirstName(),
			LastName:  f.GetLastName(),
			Email:     f.GetEmail(),
			Phone:     f.GetPhone(),
		},
	}
	if req.GetOrderBy() != "" {
		// Unknown fields or directions fail validation in the service.
		params.OrderBy = types.UserOrder{
			Field:     types.UserSortField(req.GetOrderBy()),
			Direction: types.SortAsc,
		}
		if req.GetSortDirection() != "" {
			params.OrderBy.Direction = types.SortDirection(req.GetSortDirection())
		}
	}
	return params
}
