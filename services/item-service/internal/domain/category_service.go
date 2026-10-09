package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

// CategoryService manages categories. Errors: invalid input is an
// *errs.ValidationError, a taken name ErrCategoryNameTaken, an unknown
// category ErrCategoryNotFound, and deleting one that items still have
// ErrCategoryInUse.
type CategoryService interface {
	Create(ctx context.Context, input types.CategoryInput) (*models.Category, error)
	FindByID(ctx context.Context, id string) (*models.Category, error)
	FindAll(ctx context.Context, params types.CategoryListParams) (*types.Page[*models.Category], error)
	Update(ctx context.Context, id string, input types.CategoryInput) (*models.Category, error)
	Delete(ctx context.Context, id string) error
}
