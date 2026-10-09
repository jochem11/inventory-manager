package repository

import (
	"context"
	"fmt"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/paging"
	"gorm.io/gorm"
)

type ItemStatusRepositoryImp struct {
	db *gorm.DB
}

func NewItemStatusRepository(db *gorm.DB) domain.ItemStatusRepository {
	return &ItemStatusRepositoryImp{db: db}
}

// Create inserts the status, with a new id when it has none.
func (r *ItemStatusRepositoryImp) Create(ctx context.Context, status *models.ItemStatus) error {
	if err := database.DB(ctx, r.db).Create(status).Error; err != nil {
		return fmt.Errorf("create item status: %w", translateItemStatusError(err))
	}
	return nil
}

func (r *ItemStatusRepositoryImp) FindByID(ctx context.Context, id string) (*models.ItemStatus, error) {
	var status models.ItemStatus
	if err := database.DB(ctx, r.db).First(&status, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find item status %q: %w", id, translateItemStatusError(err))
	}
	return &status, nil
}

// FindAllByIDs returns the statuses it finds, in id order; unknown ids are
// skipped.
func (r *ItemStatusRepositoryImp) FindAllByIDs(ctx context.Context, ids []string) ([]*models.ItemStatus, error) {
	statuses := []*models.ItemStatus{}
	if len(ids) == 0 {
		return statuses, nil
	}
	if err := database.DB(ctx, r.db).Where("id IN ?", ids).Order("id").Find(&statuses).Error; err != nil {
		return nil, fmt.Errorf("find item statuses by ids: %w", err)
	}
	return statuses, nil
}

// FindAll returns one page of statuses plus the total number of matches.
func (r *ItemStatusRepositoryImp) FindAll(ctx context.Context, params types.ItemStatusListParams) (*types.Page[*models.ItemStatus], error) {
	order, err := paging.OrderBy(itemStatusSortColumns, params.OrderBy.Field, params.OrderBy.Direction)
	if err != nil {
		return nil, fmt.Errorf("find item statuses: %w", err)
	}
	query := database.DB(ctx, r.db).Model(&models.ItemStatus{}).Scopes(itemStatusFilterScope(params.Filter))
	page, err := paging.Find[models.ItemStatus](query, order, params.Params)
	if err != nil {
		return nil, fmt.Errorf("find item statuses: %w", err)
	}
	return page, nil
}

// Update writes the status's name.
func (r *ItemStatusRepositoryImp) Update(ctx context.Context, status *models.ItemStatus) error {
	result := database.DB(ctx, r.db).
		Model(status).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(status)
	if result.Error != nil {
		return fmt.Errorf("update item status %q: %w", status.ID, translateItemStatusError(result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update item status %q: %w", status.ID, domain.ErrItemStatusNotFound)
	}
	return nil
}

// Delete soft-deletes the status, unless items still have it (see
// CategoryRepositoryImp.Delete for why that's checked here).
func (r *ItemStatusRepositoryImp) Delete(ctx context.Context, id string) error {
	return database.Transaction(ctx, r.db, func(ctx context.Context) error {
		db := database.DB(ctx, r.db)
		var items int64
		if err := db.Model(&models.Item{}).Where("status_id = ?", id).Count(&items).Error; err != nil {
			return fmt.Errorf("delete item status %q: count its items: %w", id, err)
		}
		if items > 0 {
			return fmt.Errorf("delete item status %q: %w", id, domain.ErrItemStatusInUse)
		}
		result := db.Delete(&models.ItemStatus{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("delete item status %q: %w", id, result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("delete item status %q: %w", id, domain.ErrItemStatusNotFound)
		}
		return nil
	})
}

// itemStatusSortColumns maps sort fields to columns, see paging.OrderBy.
var itemStatusSortColumns = map[types.ItemStatusSortField]string{
	types.ItemStatusSortName:      "name",
	types.ItemStatusSortCreatedAt: "created_at",
	types.ItemStatusSortUpdatedAt: "updated_at",
}

// itemStatusFilterScope applies an ItemStatusFilter: substrings, ignoring case.
func itemStatusFilterScope(f types.ItemStatusFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if f.Search != "" {
			db = db.Where("name LIKE ?", paging.Contains(f.Search))
		}
		if f.Name != "" {
			db = db.Where("name LIKE ?", paging.Contains(f.Name))
		}
		return db
	}
}

// translateItemStatusError maps GORM errors to the status errors. The name is
// the only unique column besides the id.
func translateItemStatusError(err error) error {
	return database.TranslateError(err, domain.ErrItemStatusNotFound, domain.ErrItemStatusNameTaken)
}
