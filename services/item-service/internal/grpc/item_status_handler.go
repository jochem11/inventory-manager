package grpc

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	itempb "github.com/jochem11/inventory-manager/services/item-service/pkg/pb/item"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/errs"
)

type ItemStatusHandler struct {
	itempb.UnimplementedItemStatusServiceServer
	statuses domain.ItemStatusService
}

func NewItemStatusHandler(statuses domain.ItemStatusService) *ItemStatusHandler {
	return &ItemStatusHandler{statuses: statuses}
}

func (h *ItemStatusHandler) ListItemStatuses(ctx context.Context, req *itempb.ListItemStatusesRequest) (*itempb.ListItemStatusesResponse, error) {
	f := req.GetFilter()
	params := types.ItemStatusListParams{
		Params: pageParams(req.GetOffset(), req.GetLimit()),
		Filter: types.ItemStatusFilter{Search: f.GetSearch(), Name: f.GetName()},
	}
	if req.GetOrderBy() != "" {
		params.OrderBy = types.ItemStatusOrder{Field: types.ItemStatusSortField(req.GetOrderBy()), Direction: direction(req.GetSortDirection())}
	}
	page, err := h.statuses.FindAll(ctx, params)
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.ListItemStatusesResponse{Statuses: mapAll(page.Nodes, toProtoItemStatus), TotalCount: int32(page.TotalCount)}, nil
}

func (h *ItemStatusHandler) CreateItemStatus(ctx context.Context, req *itempb.CreateItemStatusRequest) (*itempb.CreateItemStatusResponse, error) {
	status, err := h.statuses.Create(ctx, types.ItemStatusInput{Name: req.GetName()})
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.CreateItemStatusResponse{Status: toProtoItemStatus(status)}, nil
}

func (h *ItemStatusHandler) UpdateItemStatus(ctx context.Context, req *itempb.UpdateItemStatusRequest) (*itempb.UpdateItemStatusResponse, error) {
	status, err := h.statuses.Update(ctx, req.GetId(), types.ItemStatusInput{Name: req.GetName()})
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.UpdateItemStatusResponse{Status: toProtoItemStatus(status)}, nil
}

func (h *ItemStatusHandler) DeleteItemStatus(ctx context.Context, req *itempb.DeleteItemStatusRequest) (*itempb.DeleteItemStatusResponse, error) {
	if err := h.statuses.Delete(ctx, req.GetId()); err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.DeleteItemStatusResponse{}, nil
}
