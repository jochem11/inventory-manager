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

type CategorySortField string

const (
	CategorySortName      CategorySortField = "name"
	CategorySortCreatedAt CategorySortField = "created_at"
	CategorySortUpdatedAt CategorySortField = "updated_at"
)

// CategoryOrder sorts the category list. The zero value means no sort was asked for.
type CategoryOrder struct {
	Field     CategorySortField `json:"field" validate:"omitempty,oneof=name created_at updated_at"`
	Direction SortDirection     `json:"direction" mod:"ucase" validate:"omitempty,oneof=ASC DESC"`
}

// CategoryListParams selects one page of categories, filtered and sorted.
type CategoryListParams struct {
	paging.Params
	OrderBy CategoryOrder  `json:"orderBy"`
	Filter  CategoryFilter `json:"filter"`
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
	paging.Params
	OrderBy ItemStatusOrder  `json:"orderBy"`
	Filter  ItemStatusFilter `json:"filter"`
}

type ItemListParams struct {
	paging.Params
	OrderBy ItemOrder  `json:"orderBy"`
	Filter  ItemFilter `json:"filter"`
}

// CategoryInput is a category's editable fields, to create or rename one.
type CategoryInput struct {
	Name string `json:"name"`
}

// ItemStatusInput is a status's editable fields, to create or rename one.
type ItemStatusInput struct {
	Name string `json:"name"`
}

// ItemInput is an item's full set of editable fields, used to create an item
// and to replace one on update. Leaving Description or ImageURL nil or empty
// means "not set" (and clears it on update).
type ItemInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ImageURL    *string `json:"imageUrl"`

	CategoryID string `json:"categoryId"`
	StatusID   string `json:"statusId"`
}
