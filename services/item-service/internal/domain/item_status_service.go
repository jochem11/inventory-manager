package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

type ItemStatusService interface {
	Create(ctx context.Context, status *models.ItemStatus) error
	FindByID(ctx context.Context, id string) (*models.ItemStatus, error)
	FindAll(ctx context.Context, filter types.ItemStatusFilter) ([]*models.ItemStatus, error)
	Update(ctx context.Context, status *models.ItemStatus) error
	Delete(ctx context.Context, id string) error
}