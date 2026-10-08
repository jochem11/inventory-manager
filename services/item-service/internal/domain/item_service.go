package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

type ItemService interface {
	Create(ctx context.Context, item *models.Item) error
	FindByID(ctx context.Context, id string) (*models.Item, error)
	FindAll(ctx context.Context, filter types.ItemFilter) ([]*models.Item, error)
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id string) error
}
