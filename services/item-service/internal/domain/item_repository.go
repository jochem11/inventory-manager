package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

// ItemRepository stores items. Create and Update return ErrUnknownReference
// when the category or status doesn't exist; Update and Delete return
// ErrItemNotFound for an unknown (or deleted) item.
type ItemRepository interface {
	Create(ctx context.Context, item *models.Item) error
	FindByID(ctx context.Context, id string) (*models.Item, error)
	FindAll(ctx context.Context, params types.ItemListParams) (*types.Page[*models.Item], error)
	FindAllByIDs(ctx context.Context, ids []string) ([]*models.Item, error)
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id string) error
}
