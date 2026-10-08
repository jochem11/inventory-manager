package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

type ItemStatusRepository interface {
	Create(ctx context.Context, itemStatus *models.ItemStatus) error
	FindByID(ctx context.Context, id string) (*models.ItemStatus, error)
	FindAll(ctx context.Context, params types.ItemStatusListParams) (*types.Page[*models.ItemStatus], error)
	FindAllByIDs(ctx context.Context, ids []string) ([]*models.ItemStatus, error)
	Update(ctx context.Context, itemStatus *models.ItemStatus) error
	Delete(ctx context.Context, id string) error
}
