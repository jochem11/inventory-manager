// Package grpc exposes the item-service over gRPC. The handlers only convert
// between protobuf messages and the services' types; the services do the
// work, and errs.ToStatus turns their errors into status codes.
package grpc

import (
	"context"

	"github.com/jochem11/inventory-manager/services/item-service/internal/domain"
	itempb "github.com/jochem11/inventory-manager/services/item-service/pkg/pb/item"
	"github.com/jochem11/inventory-manager/shared/errs"
)

type ItemHandler struct {
	itempb.UnimplementedItemServiceServer
	items domain.ItemService
}

func NewItemHandler(items domain.ItemService) *ItemHandler {
	return &ItemHandler{items: items}
}

func (h *ItemHandler) GetItem(ctx context.Context, req *itempb.GetItemRequest) (*itempb.GetItemResponse, error) {
	item, err := h.items.FindByID(ctx, req.GetId())
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.GetItemResponse{Item: toProtoItem(item)}, nil
}

func (h *ItemHandler) ListItems(ctx context.Context, req *itempb.ListItemsRequest) (*itempb.ListItemsResponse, error) {
	page, err := h.items.FindAll(ctx, fromListItemsRequest(req))
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.ListItemsResponse{Items: mapAll(page.Nodes, toProtoItem), TotalCount: int32(page.TotalCount)}, nil
}

func (h *ItemHandler) CreateItem(ctx context.Context, req *itempb.CreateItemRequest) (*itempb.CreateItemResponse, error) {
	item, err := h.items.Create(ctx, fromProtoItemInput(req.GetItem()))
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.CreateItemResponse{Item: toProtoItem(item)}, nil
}

func (h *ItemHandler) UpdateItem(ctx context.Context, req *itempb.UpdateItemRequest) (*itempb.UpdateItemResponse, error) {
	item, err := h.items.Update(ctx, req.GetId(), fromProtoItemInput(req.GetItem()))
	if err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.UpdateItemResponse{Item: toProtoItem(item)}, nil
}

func (h *ItemHandler) DeleteItem(ctx context.Context, req *itempb.DeleteItemRequest) (*itempb.DeleteItemResponse, error) {
	if err := h.items.Delete(ctx, req.GetId()); err != nil {
		return nil, errs.ToStatus(ctx, err)
	}
	return &itempb.DeleteItemResponse{}, nil
}
