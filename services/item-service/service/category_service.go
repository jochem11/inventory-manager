package service

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/validation"
)

type CategoryServiceImp struct {
	categories domain.CategoryRepository
}

func NewCategoryService(categories domain.CategoryRepository) domain.CategoryService {
	return &CategoryServiceImp{categories: categories}
}

func (s *CategoryServiceImp) Create(ctx context.Context, input types.CategoryInput) (*models.Category, error) {
	category := &models.Category{Name: input.Name}
	if err := validation.Clean(ctx, category); err != nil {
		return nil, err
	}
	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryServiceImp) FindByID(ctx context.Context, id string) (*models.Category, error) {
	return s.categories.FindByID(ctx, id)
}

func (s *CategoryServiceImp) FindAll(ctx context.Context, params types.CategoryListParams) (*types.Page[*models.Category], error) {
	if err := validation.Clean(ctx, &params); err != nil {
		return nil, err
	}
	return s.categories.FindAll(ctx, params)
}

func (s *CategoryServiceImp) Update(ctx context.Context, id string, input types.CategoryInput) (*models.Category, error) {
	category, err := s.categories.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	category.Name = input.Name
	if err := validation.Clean(ctx, category); err != nil {
		return nil, err
	}
	if err := s.categories.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryServiceImp) Delete(ctx context.Context, id string) error {
	return s.categories.Delete(ctx, id)
}
