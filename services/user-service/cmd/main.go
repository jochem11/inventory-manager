package main

import (
	"context"

	"github.com/jochem11/inventory-manager/services/user-service/internal/events"
	grpchandler "github.com/jochem11/inventory-manager/services/user-service/internal/grpc"
	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	"github.com/jochem11/inventory-manager/services/user-service/internal/repository"
	userpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/user"
	"github.com/jochem11/inventory-manager/services/user-service/service"
	"github.com/jochem11/inventory-manager/shared/app"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"google.golang.org/grpc"
)

func main() {
	app.Run("user-service", func(ctx context.Context, a *app.App) error {
		db, err := a.Database("user_service", &models.User{})
		if err != nil {
			return err
		}
		users := service.NewUserService(repository.NewUserRepository(db))

		// Publish the outbox (UserDeleted, …) and consume auth events
		// (IdentityRegistered creates the profile).
		if err := a.Outbox(db); err != nil {
			return err
		}
		if err := a.Consume("user-service", kafka.TopicAuthEvents, events.NewAuthEvents(db, users).Handle); err != nil {
			return err
		}

		return a.ServeGRPC(":50051", userpb.UserService_ServiceDesc.ServiceName, func(s *grpc.Server) {
			userpb.RegisterUserServiceServer(s, grpchandler.NewUserHandler(users))
		})
	})
}
