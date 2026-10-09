// Package item resolves the item, category and status queries and mutations
// by calling the item-service. The gqlgen resolvers in package graph call
// into it; permissions are checked by the schema's directives.
package item

import (
	"context"

	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/model"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/list"
	itempb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/item"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Resolver returns gRPC errors as they are; gqlerr.Present maps them.
type Resolver struct {
	items      itempb.ItemServiceClient
	categories itempb.CategoryServiceClient
	statuses   itempb.ItemStatusServiceClient
}

// NewResolver uses one connection to the item-service for its three services.
func NewResolver(conn grpc.ClientConnInterface) *Resolver {
	return &Resolver{
		items:      itempb.NewItemServiceClient(conn),
		categories: itempb.NewCategoryServiceClient(conn),
		statuses:   itempb.NewItemStatusServiceClient(conn),
	}
}

// --- Items ----------------------------------------------------------------------

// Get returns nil when no item has this id.
func (r *Resolver) Get(ctx context.Context, id string) (*model.Item, error) {
	resp, err := r.items.GetItem(ctx, &itempb.GetItemRequest{Id: id})
	if status.Code(err) == codes.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toItem(resp.GetItem()), nil
}

func (r *Resolver) List(ctx context.Context, offset, limit int, orderBy *model.ItemOrder, filter *model.ItemFilter) (*model.ItemConnection, error) {
	if err := list.ValidatePage(offset, limit); err != nil {
		return nil, err
	}
	resp, err := r.items.ListItems(ctx, toListItemsRequest(offset, limit, orderBy, filter))
	if err != nil {
		return nil, err
	}
	return &model.ItemConnection{TotalCount: int(resp.GetTotalCount()), Nodes: mapAll(resp.GetItems(), toItem)}, nil
}

func (r *Resolver) Create(ctx context.Context, input model.ItemInput) (*model.Item, error) {
	resp, err := r.items.CreateItem(ctx, &itempb.CreateItemRequest{Item: toProtoItemInput(input)})
	if err != nil {
		return nil, err
	}
	return toItem(resp.GetItem()), nil
}

func (r *Resolver) Update(ctx context.Context, id string, input model.ItemInput) (*model.Item, error) {
	resp, err := r.items.UpdateItem(ctx, &itempb.UpdateItemRequest{Id: id, Item: toProtoItemInput(input)})
	if err != nil {
		return nil, err
	}
	return toItem(resp.GetItem()), nil
}

// Delete returns the id of the deleted item.
func (r *Resolver) Delete(ctx context.Context, id string) (string, error) {
	if _, err := r.items.DeleteItem(ctx, &itempb.DeleteItemRequest{Id: id}); err != nil {
		return "", err
	}
	return id, nil
}

// --- Categories -----------------------------------------------------------------

func (r *Resolver) ListCategories(ctx context.Context, offset, limit int, orderBy *model.NameOrder, filter *model.NameFilter) (*model.CategoryConnection, error) {
	if err := list.ValidatePage(offset, limit); err != nil {
		return nil, err
	}
	field, direction := nameOrder(orderBy)
	resp, err := r.categories.ListCategories(ctx, &itempb.ListCategoriesRequest{
		Offset: int32(offset), Limit: int32(limit), OrderBy: field, SortDirection: direction, Filter: toProtoNameFilter(filter),
	})
	if err != nil {
		return nil, err
	}
	return &model.CategoryConnection{TotalCount: int(resp.GetTotalCount()), Nodes: mapAll(resp.GetCategories(), toCategory)}, nil
}

func (r *Resolver) CreateCategory(ctx context.Context, name string) (*model.Category, error) {
	resp, err := r.categories.CreateCategory(ctx, &itempb.CreateCategoryRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return toCategory(resp.GetCategory()), nil
}

func (r *Resolver) UpdateCategory(ctx context.Context, id, name string) (*model.Category, error) {
	resp, err := r.categories.UpdateCategory(ctx, &itempb.UpdateCategoryRequest{Id: id, Name: name})
	if err != nil {
		return nil, err
	}
	return toCategory(resp.GetCategory()), nil
}

func (r *Resolver) DeleteCategory(ctx context.Context, id string) (string, error) {
	if _, err := r.categories.DeleteCategory(ctx, &itempb.DeleteCategoryRequest{Id: id}); err != nil {
		return "", err
	}
	return id, nil
}

// --- Statuses -------------------------------------------------------------------

func (r *Resolver) ListStatuses(ctx context.Context, offset, limit int, orderBy *model.NameOrder, filter *model.NameFilter) (*model.ItemStatusConnection, error) {
	if err := list.ValidatePage(offset, limit); err != nil {
		return nil, err
	}
	field, direction := nameOrder(orderBy)
	resp, err := r.statuses.ListItemStatuses(ctx, &itempb.ListItemStatusesRequest{
		Offset: int32(offset), Limit: int32(limit), OrderBy: field, SortDirection: direction, Filter: toProtoNameFilter(filter),
	})
	if err != nil {
		return nil, err
	}
	return &model.ItemStatusConnection{TotalCount: int(resp.GetTotalCount()), Nodes: mapAll(resp.GetStatuses(), toItemStatus)}, nil
}

func (r *Resolver) CreateStatus(ctx context.Context, name string) (*model.ItemStatus, error) {
	resp, err := r.statuses.CreateItemStatus(ctx, &itempb.CreateItemStatusRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return toItemStatus(resp.GetStatus()), nil
}

func (r *Resolver) UpdateStatus(ctx context.Context, id, name string) (*model.ItemStatus, error) {
	resp, err := r.statuses.UpdateItemStatus(ctx, &itempb.UpdateItemStatusRequest{Id: id, Name: name})
	if err != nil {
		return nil, err
	}
	return toItemStatus(resp.GetStatus()), nil
}

func (r *Resolver) DeleteStatus(ctx context.Context, id string) (string, error) {
	if _, err := r.statuses.DeleteItemStatus(ctx, &itempb.DeleteItemStatusRequest{Id: id}); err != nil {
		return "", err
	}
	return id, nil
}
