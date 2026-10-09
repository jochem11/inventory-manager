package grpc

import (
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	itempb "github.com/jochem11/inventory-manager/services/item-service/pkg/pb/item"
	"github.com/jochem11/inventory-manager/services/item-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/paging"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoCategory(c *models.Category) *itempb.Category {
	return &itempb.Category{
		Id:        c.ID,
		Name:      c.Name,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func toProtoItemStatus(s *models.ItemStatus) *itempb.ItemStatus {
	return &itempb.ItemStatus{
		Id:        s.ID,
		Name:      s.Name,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}
}

func toProtoItem(i *models.Item) *itempb.Item {
	return &itempb.Item{
		Id:          i.ID,
		Name:        i.Name,
		Description: i.Description,
		ImageUrl:    i.ImageURL,
		CategoryId:  i.CategoryID,
		StatusId:    i.StatusID,
		Category:    toProtoCategory(&i.Category),
		Status:      toProtoItemStatus(&i.Status),
		CreatedAt:   timestamppb.New(i.CreatedAt),
		UpdatedAt:   timestamppb.New(i.UpdatedAt),
	}
}

func mapAll[From, To any](in []From, convert func(From) To) []To {
	out := make([]To, len(in))
	for i, v := range in {
		out[i] = convert(v)
	}
	return out
}

// fromProtoItemInput accepts a nil input (the field left out of the request),
// which then fails validation on the required fields.
func fromProtoItemInput(in *itempb.ItemInput) types.ItemInput {
	if in == nil {
		return types.ItemInput{}
	}
	return types.ItemInput{
		Name:        in.GetName(),
		Description: in.Description,
		ImageURL:    in.ImageUrl,
		CategoryID:  in.GetCategoryId(),
		StatusID:    in.GetStatusId(),
	}
}

func fromListItemsRequest(req *itempb.ListItemsRequest) types.ItemListParams {
	f := req.GetFilter()
	params := types.ItemListParams{
		Params: pageParams(req.GetOffset(), req.GetLimit()),
		Filter: types.ItemFilter{
			Search:     f.GetSearch(),
			Name:       f.GetName(),
			CategoryID: f.GetCategoryId(),
			StatusID:   f.GetStatusId(),
		},
	}
	if req.GetOrderBy() != "" {
		// Unknown fields or directions fail validation in the service.
		params.OrderBy = types.ItemOrder{Field: types.ItemSortField(req.GetOrderBy()), Direction: direction(req.GetSortDirection())}
	}
	return params
}

func pageParams(offset, limit int32) paging.Params {
	return paging.Params{Offset: int(offset), Limit: int(limit)}
}

// direction defaults to ascending; validation rejects anything else than
// asc or desc (in either case).
func direction(d string) types.SortDirection {
	if d == "" {
		return types.SortAsc
	}
	return types.SortDirection(d)
}
