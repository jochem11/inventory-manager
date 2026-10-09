package grpc

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	itempb "github.com/jochem11/inventory-manager/services/item-service/pkg/pb/item"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/errs"
)

type CategoryHandler struct {
	itempb.UnimplementedCategoryServiceServer
	categories domain.CategoryService
}

func NewCategoryHandler(categories domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (h *CategoryHandler) ListCategories(ctx context.Context, req *itempb.ListCategoriesRequest) (*itempb.ListCategoriesResponse, error) {
	f := req.GetFilter()
	params := types.CategoryListParams{
		Params: pageParams(req.GetOffset(), req.GetLimit()),
		Filter: types.CategoryFilter{Search: f.GetSearch(), Name: f.GetName()},
	}
	if req.GetOrderBy() != "" {
		params.OrderBy = types.CategoryOrder{Field: types.CategorySortField(req.GetOrderBy()), Direction: direction(req.GetSortDirection())}
	}
	page, err := h.categories.FindAll(ctx, params)
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.ListCategoriesResponse{Categories: mapAll(page.Nodes, toProtoCategory), TotalCount: int32(page.TotalCount)}, nil
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *itempb.CreateCategoryRequest) (*itempb.CreateCategoryResponse, error) {
	category, err := h.categories.Create(ctx, types.CategoryInput{Name: req.GetName()})
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.CreateCategoryResponse{Category: toProtoCategory(category)}, nil
}

func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *itempb.UpdateCategoryRequest) (*itempb.UpdateCategoryResponse, error) {
	category, err := h.categories.Update(ctx, req.GetId(), types.CategoryInput{Name: req.GetName()})
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.UpdateCategoryResponse{Category: toProtoCategory(category)}, nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *itempb.DeleteCategoryRequest) (*itempb.DeleteCategoryResponse, error) {
	if err := h.categories.Delete(ctx, req.GetId()); err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.DeleteCategoryResponse{}, nil
}
