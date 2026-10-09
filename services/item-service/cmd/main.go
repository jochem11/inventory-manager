package main

import (
	"context"

	grpchandler "github.com/jochem11/inventory-manager/services/item-service/internal/grpc"
	"github.com/jochem11/inventory-manager/services/item-service/internal/models"
	"github.com/jochem11/inventory-manager/services/item-service/internal/repository"
	itempb "github.com/jochem11/inventory-manager/services/item-service/pkg/pb/item"
	"github.com/jochem11/inventory-manager/services/item-service/service"
	"github.com/jochem11/inventory-manager/shared/app"
	"google.golang.org/grpc"
)

func main() {
	app.Run("item-service", func(ctx context.Context, a *app.App) error {
		db, err := a.Database("item_service", &models.Category{}, &models.ItemStatus{}, &models.Item{})
		if err != nil {
			return err
		}
		items := repository.NewItemRepository(db)
		categories := repository.NewCategoryRepository(db)
		statuses := repository.NewItemStatusRepository(db)

		return a.ServeGRPC(":50053", itempb.ItemService_ServiceDesc.ServiceName, func(s *grpc.Server) {
			itempb.RegisterItemServiceServer(s, grpchandler.NewItemHandler(service.NewItemService(items, categories, statuses)))
			itempb.RegisterCategoryServiceServer(s, grpchandler.NewCategoryHandler(service.NewCategoryService(categories)))
			itempb.RegisterItemStatusServiceServer(s, grpchandler.NewItemStatusHandler(service.NewItemStatusService(statuses)))
		})
	})
}
