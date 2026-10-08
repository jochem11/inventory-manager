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

	"github.com/jochem11/inventory-manager/services/auth-service/internal/events"
	grpchandler "github.com/jochem11/inventory-manager/services/auth-service/internal/grpc"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/models"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/repository"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/token"
	authpb "github.com/jochem11/inventory-manager/services/auth-service/pkg/pb/auth"
	"github.com/jochem11/inventory-manager/services/auth-service/service"
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

	shutdownTracing, err := telemetry.Setup(ctx, "auth-service")
	if err != nil {
		log.Fatalf("set up tracing: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownTracing(ctx)
	}()

	dbConfig := database.ConfigFromEnv("auth_service")
	if err := database.CreateIfMissing(dbConfig); err != nil {
		log.Fatalf("create database: %v", err)
	}
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	// A span per SQL statement, without the values (they hold personal data).
	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithoutQueryVariables())); err != nil {
		log.Fatalf("add tracing to database: %v", err)
	}
	tables := append([]any{
		&models.Permission{}, &models.Role{}, &models.Identity{}, &models.IdentityToken{}, &models.Session{},
	}, kafka.Models()...)
	if err := db.AutoMigrate(tables...); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	// ADMIN_EMAILS (comma-separated) get the admin role, at every start.
	if err := repository.Seed(ctx, db, adminEmails(env("ADMIN_EMAILS", ""))); err != nil {
		log.Fatalf("seed roles: %v", err)
	}
	log.Println("connected to database")

	// JWT_PRIVATE_KEY is a PKCS#8 PEM Ed25519 key; without it a new key is
	// generated at every start.
	// In .env it's one line, with \n for the line breaks.
	tokens, generated, err := token.NewIssuer(strings.ReplaceAll(os.Getenv("JWT_PRIVATE_KEY"), `\n`, "\n"))
	if err != nil {
		log.Fatalf("load signing key: %v", err)
	}
	if generated {
		log.Println("JWT_PRIVATE_KEY not set: signing with a key generated for this run")
	}

	auth := service.NewAuthService(repository.NewIdentityRepository(db), tokens,
		env("VERIFY_EMAIL_URL", "http://localhost:3000/verify-email"))

	// Kafka: publish the outbox (IdentityRegistered, …) and consume user
	// events (UserDeleted removes the login).
	brokers := kafka.BrokersFromEnv()
	producer, err := kafka.NewProducer(brokers)
	if err != nil {
		log.Fatalf("create kafka producer: %v", err)
	}
	defer producer.Close()
	consumer, err := kafka.NewConsumer(brokers, "auth-service", kafka.TopicUserEvents)
	if err != nil {
		log.Fatalf("create kafka consumer: %v", err)
	}
	defer consumer.Close()
	go kafka.RunRelay(ctx, db, producer)
	go kafka.Consume(ctx, consumer, events.NewUserEvents(db).Handle)

	// A span per call, continuing the caller's trace. Health probes are left out.
	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler(
		otelgrpc.WithFilter(func(info *stats.RPCTagInfo) bool {
			return !strings.HasPrefix(info.FullMethodName, "/grpc.health.v1.Health/")
		}),
	)))
	authpb.RegisterAuthServiceServer(server, grpchandler.NewAuthHandler(auth))

	// Health checks for Kubernetes probes, and reflection so grpcurl works
	// without the .proto files.
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus(authpb.AuthService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	reflection.Register(server)

	addr := env("GRPC_ADDR", ":50052")
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

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// adminEmails splits a comma-separated list, lower-cased like stored emails.
func adminEmails(list string) []string {
	var emails []string
	for _, email := range strings.Split(list, ",") {
		if email = strings.ToLower(strings.TrimSpace(email)); email != "" {
			emails = append(emails, email)
		}
	}
	return emails
}
