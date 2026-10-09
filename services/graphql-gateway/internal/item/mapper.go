package item

import (
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/model"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/list"
	itempb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/item"
)

// itemSortFields maps the GraphQL sort enum to the fields ListItems accepts.
var itemSortFields = map[model.ItemSortField]string{
	model.ItemSortFieldName:      "name",
	model.ItemSortFieldCategory:  "category_id",
	model.ItemSortFieldStatus:    "status_id",
	model.ItemSortFieldCreatedAt: "created_at",
	model.ItemSortFieldUpdatedAt: "updated_at",
}

// nameSortFields does the same for categories and statuses.
var nameSortFields = map[model.NameSortField]string{
	model.NameSortFieldName:      "name",
	model.NameSortFieldCreatedAt: "created_at",
	model.NameSortFieldUpdatedAt: "updated_at",
}

func toCategory(c *itempb.Category) *model.Category {
	return &model.Category{ID: c.GetId(), Name: c.GetName(), CreatedAt: c.GetCreatedAt().AsTime(), UpdatedAt: c.GetUpdatedAt().AsTime()}
}

func toItemStatus(s *itempb.ItemStatus) *model.ItemStatus {
	return &model.ItemStatus{ID: s.GetId(), Name: s.GetName(), CreatedAt: s.GetCreatedAt().AsTime(), UpdatedAt: s.GetUpdatedAt().AsTime()}
}

func toItem(i *itempb.Item) *model.Item {
	return &model.Item{
		ID:          i.GetId(),
		Name:        i.GetName(),
		Description: i.Description,
		ImageURL:    i.ImageUrl,
		Category:    toCategory(i.GetCategory()),
		Status:      toItemStatus(i.GetStatus()),
		CreatedAt:   i.GetCreatedAt().AsTime(),
		UpdatedAt:   i.GetUpdatedAt().AsTime(),
	}
}

func mapAll[From, To any](in []From, convert func(From) To) []To {
	out := make([]To, len(in))
	for i, v := range in {
		out[i] = convert(v)
	}
	return out
}

func toProtoItemInput(in model.ItemInput) *itempb.ItemInput {
	return &itempb.ItemInput{
		Name:        in.Name,
		Description: in.Description,
		ImageUrl:    in.ImageURL,
		CategoryId:  in.CategoryID,
		StatusId:    in.StatusID,
	}
}

// toListItemsRequest converts the items query's arguments; they must already
// be validated. Unset filter fields stay nil, so "not set" differs from "".
func toListItemsRequest(offset, limit int, orderBy *model.ItemOrder, filter *model.ItemFilter) *itempb.ListItemsRequest {
	req := &itempb.ListItemsRequest{Offset: int32(offset), Limit: int32(limit)}
	if orderBy != nil {
		req.OrderBy = itemSortFields[orderBy.Field]
		req.SortDirection = list.SortDirection(orderBy.Direction)
	}
	if filter != nil {
		req.Filter = &itempb.ItemFilter{
			Search:     filter.Search,
			Name:       filter.Name,
			CategoryId: filter.CategoryID,
			StatusId:   filter.StatusID,
		}
	}
	return req
}

func nameOrder(orderBy *model.NameOrder) (field, direction string) {
	if orderBy == nil {
		return "", ""
	}
	return nameSortFields[orderBy.Field], list.SortDirection(orderBy.Direction)
}

func toProtoNameFilter(filter *model.NameFilter) *itempb.NameFilter {
	if filter == nil {
		return nil
	}
	return &itempb.NameFilter{Search: filter.Search, Name: filter.Name}
}
