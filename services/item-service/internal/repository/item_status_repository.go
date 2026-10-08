package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/paging"
	"gorm.io/gorm"
)

type ItemStatusRepositoryImp struct {
	db *gorm.DB
}

func NewItemStatusRepository(db *gorm.DB) domain.ItemStatusRepository {
	return &ItemStatusRepositoryImp{db: db}
}

func (r *ItemStatusRepositoryImp) Create(ctx context.Context, itemStatus *models.ItemStatus) error {
	if err := r.db.WithContext(ctx).Create(itemStatus).Error; err != nil {
		return fmt.Errorf("failed to create item status: %w", translateError(err, domain.ErrItemStatusNotFound, domain.ErrItemStatusNameTaken))
	}
	return nil
}

func (r *ItemStatusRepositoryImp) FindByID(ctx context.Context, id string) (*models.ItemStatus, error) {
	var itemStatus models.ItemStatus
	if err := r.db.WithContext(ctx).First(&itemStatus, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find item status by id %q: %w", id, translateError(err, domain.ErrItemStatusNotFound, domain.ErrItemStatusNameTaken))
	}

	return &itemStatus, nil
}

func (r *ItemStatusRepositoryImp) FindAllByIDs(ctx context.Context, ids []string) ([]*models.ItemStatus, error) {
	itemStatuses := []*models.ItemStatus{}
	if len(ids) == 0 {
		return itemStatuses, nil
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Order("id").Find(&itemStatuses).Error; err != nil {
		return nil, fmt.Errorf("find item statuses by ids: %w", err)
	}
	return itemStatuses, nil
}

func (r *ItemStatusRepositoryImp) FindAll(ctx context.Context, params types.ItemStatusListParams) (*types.Page[*models.ItemStatus], error) {
	order, err := paging.OrderBy(itemStatusSortColumns, params.OrderBy.Field, params.OrderBy.Direction)
	if err != nil {
		return nil, fmt.Errorf("find item statuses: %w", err)
	}
	query := r.db.WithContext(ctx).Model(&models.ItemStatus{}).Scopes(itemStatusFilterScope(params.Filter))
	page, err := paging.Find[models.ItemStatus](query, order, params.Offset, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("find item statuses: %w", err)
	}
	return page, nil
}

func (r *ItemStatusRepositoryImp) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

func (r *ItemStatusRepositoryImp) Update(ctx context.Context, itemStatus *models.ItemStatus) error {
	panic("unimplemented")
}

var itemStatusSortColumns = map[types.ItemStatusSortField]string{
	types.ItemStatusSortName:      "name",
	types.ItemStatusSortCreatedAt: "created_at",
	types.ItemStatusSortUpdatedAt: "updated_at",
}

func itemStatusFilterScope(f types.ItemStatusFilter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if f.Search != "" {
			db = db.Where("name LIKE ?", paging.Contains(f.Search))
		}
		for _, c := range []struct{ column, value string }{
			{"name", f.Name},
		} {
			if c.value != "" {
				db = db.Where(c.column+" LIKE ?", paging.Contains(c.value))
			}
		}
		return db
	}
}

// translateError turns GORM's errors into the domain errors the caller
// passes, so each repository reports its own "not found" and "taken".
func translateError(err, notFound, duplicate error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return notFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return duplicate
	}
	return err
}
