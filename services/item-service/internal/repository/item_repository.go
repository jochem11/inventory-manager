package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/paging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ItemRepositoryImp struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) domain.ItemRepository {
	return &ItemRepositoryImp{db: db}
}


func (r *ItemRepositoryImp) Create(ctx context.Context, item *models.Item) error {
	if err := database.DB(ctx, r.db).Omit(clause.Associations).Create(item).Error; err != nil {
		return fmt.Errorf("create item: %w", translateItemError(err))
	}
	return nil
}

func (r *ItemRepositoryImp) FindByID(ctx context.Context, id string) (*models.Item, error) {
	var item models.Item
	if err := database.DB(ctx, r.db).Joins("Category").Joins("Status").First(&item, "items.id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find item %q: %w", id, translateItemError(err))
	}
	return &item, nil
}

func (r *ItemRepositoryImp) FindAllByIDs(ctx context.Context, ids []string) ([]*models.Item, error) {
	items := []*models.Item{}
	if len(ids) == 0 {
		return items, nil
	}
	if err := database.DB(ctx, r.db).Joins("Category").Joins("Status").Where("items.id IN ?", ids).Order("items.id").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("find items by ids: %w", err)
	}
	return items, nil
}

func (r *ItemRepositoryImp) FindAll(ctx context.Context, params types.ItemListParams) (*types.Page[*models.Item], error) {
	order, err := paging.OrderBy(itemSortColumns, params.OrderBy.Field, params.OrderBy.Direction)
	if err != nil {
		return nil, fmt.Errorf("find items: %w", err)
	}
	query := database.DB(ctx, r.db).Model(&models.Item{}).Joins("Category").Joins("Status").Scopes(itemFilterScope(params.Filter))
	page, err := paging.Find[models.Item](query, order, params.Params)
	if err != nil {
		return nil, fmt.Errorf("find items: %w", err)
	}
	return page, nil
}

func (r *ItemRepositoryImp) Update(ctx context.Context, item *models.Item) error {
	result := database.DB(ctx, r.db).
		Model(item).
		Select("*").
		Omit("id", "created_at", "deleted_at", clause.Associations).
		Updates(item)
	if result.Error != nil {
		return fmt.Errorf("update item %q: %w", item.ID, translateItemError(result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update item %q: %w", item.ID, domain.ErrItemNotFound)
	}
	return nil
}

func (r *ItemRepositoryImp) Delete(ctx context.Context, id string) error {
	result := database.DB(ctx, r.db).Delete(&models.Item{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete item %q: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete item %q: %w", id, domain.ErrItemNotFound)
	}
	return nil
}

var itemSortColumns = map[types.ItemSortField]string{
	types.ItemSortName:       "name",
	types.ItemSortCategoryID: "category_id",
	types.ItemSortStatusID:   "status_id",
	types.ItemSortCreatedAt:  "created_at",
	types.ItemSortUpdatedAt:  "updated_at",
}

func itemFilterScope(f types.ItemFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if f.Search != "" {
			pattern := paging.Contains(f.Search)
			db = db.Where("items.name LIKE ? OR items.description LIKE ?", pattern, pattern)
		}
		if f.Name != "" {
			db = db.Where("items.name LIKE ?", paging.Contains(f.Name))
		}
		if f.CategoryID != "" {
			db = db.Where("items.category_id = ?", f.CategoryID)
		}
		if f.StatusID != "" {
			db = db.Where("items.status_id = ?", f.StatusID)
		}
		return db
	}
}

func translateItemError(err error) error {
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return domain.ErrUnknownReference
	}
	return database.TranslateError(err, domain.ErrItemNotFound)
}
