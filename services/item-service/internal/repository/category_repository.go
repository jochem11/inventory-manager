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

type CategoryRepositoryImp struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &CategoryRepositoryImp{db: db}
}

func (r *CategoryRepositoryImp) Create(ctx context.Context, category *models.Category) error {
	if err := database.DB(ctx, r.db).Create(category).Error; err != nil {
		return fmt.Errorf("create category: %w", translateCategoryError(err))
	}
	return nil
}

func (r *CategoryRepositoryImp) FindByID(ctx context.Context, id string) (*models.Category, error) {
	var category models.Category
	if err := database.DB(ctx, r.db).First(&category, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find category %q: %w", id, translateCategoryError(err))
	}
	return &category, nil
}

func (r *CategoryRepositoryImp) FindAllByIDs(ctx context.Context, ids []string) ([]*models.Category, error) {
	categories := []*models.Category{}
	if len(ids) == 0 {
		return categories, nil
	}
	if err := database.DB(ctx, r.db).Where("id IN ?", ids).Order("id").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("find categories by ids: %w", err)
	}
	return categories, nil
}

func (r *CategoryRepositoryImp) FindAll(ctx context.Context, params types.CategoryListParams) (*types.Page[*models.Category], error) {
	order, err := paging.OrderBy(categorySortColumns, params.OrderBy.Field, params.OrderBy.Direction)
	if err != nil {
		return nil, fmt.Errorf("find categories: %w", err)
	}
	query := database.DB(ctx, r.db).Model(&models.Category{}).Scopes(categoryFilterScope(params.Filter))
	page, err := paging.Find[models.Category](query, order, params.Params)
	if err != nil {
		return nil, fmt.Errorf("find categories: %w", err)
	}
	return page, nil
}

func (r *CategoryRepositoryImp) Update(ctx context.Context, category *models.Category) error {
	result := database.DB(ctx, r.db).
		Model(category).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(category)
	if result.Error != nil {
		return fmt.Errorf("update category %q: %w", category.ID, translateCategoryError(result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update category %q: %w", category.ID, domain.ErrCategoryNotFound)
	}
	return nil
}

func (r *CategoryRepositoryImp) Delete(ctx context.Context, id string) error {
	return database.Transaction(ctx, r.db, func(ctx context.Context) error {
		db := database.DB(ctx, r.db)
		var items int64
		if err := db.Model(&models.Item{}).Where("category_id = ?", id).Count(&items).Error; err != nil {
			return fmt.Errorf("delete category %q: count its items: %w", id, err)
		}
		if items > 0 {
			return fmt.Errorf("delete category %q: %w", id, domain.ErrCategoryInUse)
		}
		result := db.Delete(&models.Category{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("delete category %q: %w", id, result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("delete category %q: %w", id, domain.ErrCategoryNotFound)
		}
		return nil
	})
}

var categorySortColumns = map[types.CategorySortField]string{
	types.CategorySortName:      "name",
	types.CategorySortCreatedAt: "created_at",
	types.CategorySortUpdatedAt: "updated_at",
}

func categoryFilterScope(f types.CategoryFilter) func(*gorm.DB) *gorm.DB {
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

func translateCategoryError(err error) error {
	return database.TranslateError(err, domain.ErrCategoryNotFound, domain.ErrCategoryNameTaken)
}
