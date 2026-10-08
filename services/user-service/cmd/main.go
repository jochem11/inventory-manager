package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jochem11/inventory-manager/services/user-service/internal/events"
	grpchandler "github.com/jochem11/inventory-manager/services/user-service/internal/grpc"
	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	"github.com/jochem11/inventory-manager/services/user-service/internal/repository"
	userpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/user"
	"github.com/jochem11/inventory-manager/services/user-service/service"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"github.com/jochem11/inventory-manager/shared/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
	"gorm.io/plugin/opentelemetry/tracing"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := telemetry.Setup(ctx, "user-service")
	if err != nil {
		log.Fatalf("set up tracing: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownTracing(ctx)
	}()

	db, err := database.Connect(database.ConfigFromEnv("user_service"))
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	// A span per SQL statement, without the values (they hold personal data).
	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithoutQueryVariables())); err != nil {
		log.Fatalf("add tracing to database: %v", err)
	}

	if err := db.AutoMigrate(append([]any{&models.User{}}, kafka.Models()...)...); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	log.Println("connected to database")

	users := service.NewUserService(repository.NewUserRepository(db))

	// Kafka: publish the outbox (UserDeleted, …) and consume auth events
	// (IdentityRegistered creates the profile).
	brokers := kafka.BrokersFromEnv()
	producer, err := kafka.NewProducer(brokers)
	if err != nil {
		log.Fatalf("create kafka producer: %v", err)
	}
	defer producer.Close()
	consumer, err := kafka.NewConsumer(brokers, "user-service", kafka.TopicAuthEvents)
	if err != nil {
		log.Fatalf("create kafka consumer: %v", err)
	}
	defer consumer.Close()
	go kafka.RunRelay(ctx, db, producer)
	go kafka.Consume(ctx, consumer, events.NewAuthEvents(db).Handle)

	// A span per call, continuing the caller's trace. Health probes are left out.
	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler(
		otelgrpc.WithFilter(func(info *stats.RPCTagInfo) bool {
			return !strings.HasPrefix(info.FullMethodName, "/grpc.health.v1.Health/")
		}),
	)))
	userpb.RegisterUserServiceServer(server, grpchandler.NewUserHandler(users))

	// Health checks for Kubernetes probes, and reflection so grpcurl works
	// without the .proto files.
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus(userpb.UserService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	reflection.Register(server)

	addr := ":50051"
	if v, ok := os.LookupEnv("GRPC_ADDR"); ok {
		addr = v
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen on %s: %v", addr, err)
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down")
		healthServer.Shutdown()
		server.GracefulStop()
	}()

	log.Printf("gRPC server listening on %s", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
