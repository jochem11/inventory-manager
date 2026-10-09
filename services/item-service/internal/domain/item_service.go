package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

// ItemService manages items. Items come back with their category and status.
// Errors: invalid input is an *errs.ValidationError (also for a category or
// status that doesn't exist), and an unknown item wraps ErrItemNotFound.
type ItemService interface {
	Create(ctx context.Context, input types.ItemInput) (*models.Item, error)
	FindByID(ctx context.Context, id string) (*models.Item, error)
	FindAll(ctx context.Context, params types.ItemListParams) (*types.Page[*models.Item], error)
	// Update replaces every editable field of the item.
	Update(ctx context.Context, id string, input types.ItemInput) (*models.Item, error)
	Delete(ctx context.Context, id string) error
}
