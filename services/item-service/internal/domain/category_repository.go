package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

// CategoryRepository stores categories. Create and Update return
// ErrCategoryNameTaken for a name that's in use (ignoring case); Update and
// Delete return ErrCategoryNotFound for an unknown category, and Delete
// returns ErrCategoryInUse while items still have the category.
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	FindByID(ctx context.Context, id string) (*models.Category, error)
	// FindAll returns one page of categories plus the total number of matches.
	FindAll(ctx context.Context, params types.CategoryListParams) (*types.Page[*models.Category], error)
	// FindAllByIDs returns the categories it finds, e.g. those of a page of
	// items; unknown ids are skipped.
	FindAllByIDs(ctx context.Context, ids []string) ([]*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id string) error
}
