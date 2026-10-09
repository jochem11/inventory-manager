package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
)

// ItemStatusService manages item statuses. Errors: invalid input is an
// *errs.ValidationError, a taken name ErrItemStatusNameTaken, an unknown
// status ErrItemStatusNotFound, and deleting one that items still have
// ErrItemStatusInUse.
type ItemStatusService interface {
	Create(ctx context.Context, input types.ItemStatusInput) (*models.ItemStatus, error)
	FindByID(ctx context.Context, id string) (*models.ItemStatus, error)
	FindAll(ctx context.Context, params types.ItemStatusListParams) (*types.Page[*models.ItemStatus], error)
	Update(ctx context.Context, id string, input types.ItemStatusInput) (*models.ItemStatus, error)
	Delete(ctx context.Context, id string) error
}
