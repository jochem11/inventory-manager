package service

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/validation"
)

type ItemStatusServiceImp struct {
	statuses domain.ItemStatusRepository
}

func NewItemStatusService(statuses domain.ItemStatusRepository) domain.ItemStatusService {
	return &ItemStatusServiceImp{statuses: statuses}
}

func (s *ItemStatusServiceImp) Create(ctx context.Context, input types.ItemStatusInput) (*models.ItemStatus, error) {
	status := &models.ItemStatus{Name: input.Name}
	if err := validation.Clean(ctx, status); err != nil {
		return nil, err
	}
	if err := s.statuses.Create(ctx, status); err != nil {
		return nil, err
	}
	return status, nil
}

func (s *ItemStatusServiceImp) FindByID(ctx context.Context, id string) (*models.ItemStatus, error) {
	return s.statuses.FindByID(ctx, id)
}

func (s *ItemStatusServiceImp) FindAll(ctx context.Context, params types.ItemStatusListParams) (*types.Page[*models.ItemStatus], error) {
	if err := validation.Clean(ctx, &params); err != nil {
		return nil, err
	}
	return s.statuses.FindAll(ctx, params)
}

func (s *ItemStatusServiceImp) Update(ctx context.Context, id string, input types.ItemStatusInput) (*models.ItemStatus, error) {
	status, err := s.statuses.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	status.Name = input.Name
	if err := validation.Clean(ctx, status); err != nil {
		return nil, err
	}
	if err := s.statuses.Update(ctx, status); err != nil {
		return nil, err
	}
	return status, nil
}

func (s *ItemStatusServiceImp) Delete(ctx context.Context, id string) error {
	return s.statuses.Delete(ctx, id)
}
