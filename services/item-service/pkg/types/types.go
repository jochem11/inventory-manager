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

type ItemSortField string

const (
	ItemSortName       ItemSortField = "name"
	ItemSortCategoryID ItemSortField = "category_id"
	ItemSortStatusID   ItemSortField = "status_id"
	ItemSortCreatedAt  ItemSortField = "created_at"
	ItemSortUpdatedAt  ItemSortField = "updated_at"
)

type ItemStatusSortField string

const (
	ItemStatusSortName      ItemStatusSortField = "name"
	ItemStatusSortCreatedAt ItemStatusSortField = "created_at"
	ItemStatusSortUpdatedAt ItemStatusSortField = "updated_at"
)

type ItemOrder struct {
	Field     ItemSortField `json:"field" validate:"omitempty,oneof=name category_id status_id created_at updated_at"`
	Direction SortDirection `json:"direction" mod:"ucase" validate:"omitempty,oneof=ASC DESC"`
}

type ItemStatusOrder struct {
	Field     ItemStatusSortField `json:"field" validate:"omitempty,oneof=name created_at updated_at"`
	Direction SortDirection       `json:"direction" mod:"ucase" validate:"omitempty,oneof=ASC DESC"`
}

type ItemFilter struct {
	Search     string `json:"search" mod:"trim"`
	Name       string `json:"name" mod:"trim"`
	CategoryID string `json:"categoryId"`
	StatusID   string `json:"statusId"`
}

type CategoryFilter struct {
	Search string `json:"search" mod:"trim"`
	Name   string `json:"name" mod:"trim"`
}

type ItemStatusFilter struct {
	Search string `json:"search" mod:"trim"`
	Name   string `json:"name" mod:"trim"`
}

type ItemStatusListParams struct {
	Offset  int              `json:"offset" validate:"min=0"`
	Limit   int              `json:"limit" mod:"default=10" validate:"min=1,max=100"`
	OrderBy ItemStatusOrder  `json:"orderBy"`
	Filter  ItemStatusFilter `json:"filter"`
}

type ItemListParams struct {
	Offset  int        `json:"offset" validate:"min=0"`
	Limit   int        `json:"limit" mod:"default=10" validate:"min=1,max=100"`
	OrderBy ItemOrder  `json:"orderBy"`
	Filter  ItemFilter `json:"filter"`
}

type ItemInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ImageURL    *string `json:"imageUrl"`

	CategoryID string `json:"categoryId"`
	StatusID   string `json:"statusId"`
}
