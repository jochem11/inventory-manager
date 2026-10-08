package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

type CategoryService interface {
	Create(ctx context.Context, category *models.Category) error
	FindByID(ctx context.Context, id string) (*models.Category, error)
	FindAll(ctx context.Context, filter types.CategoryFilter) ([]*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id string) error
}