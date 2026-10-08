package types

import "github.com/jochem11/inventory-manager/shared/paging"

// The list types every service shares, see shared/paging.
type (
	SortDirection = paging.SortDirection
	Page[T any]   = paging.Page[T]
)

const (
	SortAsc  = paging.SortAsc
	SortDesc = paging.SortDesc
)

// UserSortField is a field the user list can be ordered by.
type UserSortField string

const (
	UserSortFirstName UserSortField = "first_name"
	UserSortLastName  UserSortField = "last_name"
	UserSortEmail     UserSortField = "email"
	UserSortPhone     UserSortField = "phone"
	UserSortCreatedAt UserSortField = "created_at"
	UserSortUpdatedAt UserSortField = "updated_at"
)

// UserOrder sorts the user list. The zero value means no sort was asked for.
type UserOrder struct {
	Field     UserSortField `json:"field" validate:"omitempty,oneof=first_name last_name email phone created_at updated_at"`
	Direction SortDirection `json:"direction" mod:"ucase" validate:"omitempty,oneof=ASC DESC"`
}

// UserFilter narrows the user list. Every non-empty field must match as a
// case-insensitive substring.
type UserFilter struct {
	// Search matches any of the name (including "first last"), email or phone.
	Search    string `json:"search" mod:"trim"`
	FirstName string `json:"firstName" mod:"trim"`
	LastName  string `json:"lastName" mod:"trim"`
	Email     string `json:"email" mod:"trim"`
	Phone     string `json:"phone" mod:"trim"`
}

// UserListParams selects one page of users, filtered and sorted. Limit
// defaults to 10; the web DataTable offers 10, 25 and 50.
type UserListParams struct {
	Offset  int        `json:"offset" validate:"min=0"`
	Limit   int        `json:"limit" mod:"default=10" validate:"min=1,max=100"`
	OrderBy UserOrder  `json:"orderBy"`
	Filter  UserFilter `json:"filter"`
}

// UserInput is a user's full set of editable fields, used to create a user
// and to replace one on update. The ID and timestamps are generated. Leaving
// Phone or AvatarURL nil or empty means "not set" (and clears it on update).
type UserInput struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone"`
	AvatarURL *string `json:"avatarUrl"`
}
