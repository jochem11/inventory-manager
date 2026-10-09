// Package service holds the item-service's business rules: validation,
// checks across tables, and loading related records. Storage is in the
// repositories.
package service

import (
	"context"
	"errors"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/errs"
	"github.com/jochem11/inventory-manager/shared/validation"
)

type ItemServiceImp struct {
	items      domain.ItemRepository
	categories domain.CategoryRepository
	statuses   domain.ItemStatusRepository
}

func NewItemService(items domain.ItemRepository, categories domain.CategoryRepository, statuses domain.ItemStatusRepository) domain.ItemService {
	return &ItemServiceImp{items: items, categories: categories, statuses: statuses}
}

func (s *ItemServiceImp) Create(ctx context.Context, input types.ItemInput) (*models.Item, error) {
	item := &models.Item{}
	applyItemInput(item, input)
	if err := s.check(ctx, item); err != nil {
		return nil, err
	}
	if err := s.items.Create(ctx, item); err != nil {
		return nil, err
	}
	return s.items.FindByID(ctx, item.ID)
}

func (s *ItemServiceImp) FindByID(ctx context.Context, id string) (*models.Item, error) {
	return s.items.FindByID(ctx, id)
}

// FindAll returns a page of items, each with its category and status (the
// repository joins them in the same query).
func (s *ItemServiceImp) FindAll(ctx context.Context, params types.ItemListParams) (*types.Page[*models.Item], error) {
	if err := validation.Clean(ctx, &params); err != nil {
		return nil, err
	}
	return s.items.FindAll(ctx, params)
}

func (s *ItemServiceImp) Update(ctx context.Context, id string, input types.ItemInput) (*models.Item, error) {
	item, err := s.items.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	applyItemInput(item, input)
	if err := s.check(ctx, item); err != nil {
		return nil, err
	}
	if err := s.items.Update(ctx, item); err != nil {
		return nil, err
	}
	return s.items.FindByID(ctx, item.ID)
}

func (s *ItemServiceImp) Delete(ctx context.Context, id string) error {
	return s.items.Delete(ctx, id)
}

// check validates the item, and that its category and status exist, so a
// wrong one is reported on its field rather than as a broken foreign key.
func (s *ItemServiceImp) check(ctx context.Context, item *models.Item) error {
	if err := validation.Clean(ctx, item); err != nil {
		return err
	}
	fields := map[string]string{}
	if _, err := s.categories.FindByID(ctx, item.CategoryID); errors.Is(err, domain.ErrCategoryNotFound) {
		fields["categoryId"] = "categoryId must be an existing category"
	} else if err != nil {
		return err
	}
	if _, err := s.statuses.FindByID(ctx, item.StatusID); errors.Is(err, domain.ErrItemStatusNotFound) {
		fields["statusId"] = "statusId must be an existing status"
	} else if err != nil {
		return err
	}
	if len(fields) > 0 {
		return &errs.ValidationError{Fields: fields}
	}
	return nil
}

// applyItemInput sets every editable field of item from input, including
// empty ones, so validation sees exactly what the client sent.
func applyItemInput(item *models.Item, input types.ItemInput) {
	item.Name = input.Name
	item.Description = input.Description
	item.ImageURL = input.ImageURL
	item.CategoryID = input.CategoryID
	item.StatusID = input.StatusID
}
