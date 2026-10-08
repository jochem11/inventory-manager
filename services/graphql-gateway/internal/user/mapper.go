package user

import (
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/model"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/list"
	userpb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/user"
)

// userSortFields maps the GraphQL sort enum to the field names ListUsers
// accepts.
var userSortFields = map[model.UserSortField]string{
	model.UserSortFieldFirstName: "first_name",
	model.UserSortFieldLastName:  "last_name",
	model.UserSortFieldEmail:     "email",
	model.UserSortFieldPhone:     "phone",
	model.UserSortFieldCreatedAt: "created_at",
	model.UserSortFieldUpdatedAt: "updated_at",
}

func toUser(u *userpb.User) *model.User {
	return &model.User{
		ID:        u.GetId(),
		FirstName: u.GetFirstName(),
		LastName:  u.GetLastName(),
		Email:     u.GetEmail(),
		Phone:     u.Phone,
		AvatarURL: u.AvatarUrl,
		CreatedAt: u.GetCreatedAt().AsTime(),
		UpdatedAt: u.GetUpdatedAt().AsTime(),
	}
}

func toUsers(users []*userpb.User) []*model.User {
	out := make([]*model.User, len(users))
	for i, u := range users {
		out[i] = toUser(u)
	}
	return out
}

func toProtoUserInput(in model.UserInput) *userpb.UserInput {
	return &userpb.UserInput{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Email:     in.Email,
		Phone:     in.Phone,
		AvatarUrl: in.AvatarURL,
	}
}

// toListUsersRequest converts the users query's arguments; they must already
// be validated. Unset filter fields stay nil, so "not set" differs from "".
func toListUsersRequest(offset, limit int, orderBy *model.UserOrder, filter *model.UserFilter) *userpb.ListUsersRequest {
	req := &userpb.ListUsersRequest{
		Offset: int32(offset),
		Limit:  int32(limit),
	}
	if orderBy != nil {
		req.OrderBy = userSortFields[orderBy.Field]
		req.SortDirection = list.SortDirection(orderBy.Direction)
	}
	if filter != nil {
		req.Filter = &userpb.UserFilter{
			Search:    filter.Search,
			FirstName: filter.FirstName,
			LastName:  filter.LastName,
			Email:     filter.Email,
			Phone:     filter.Phone,
		}
	}
	return req
}
