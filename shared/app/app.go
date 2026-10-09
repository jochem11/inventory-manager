// Package app is the startup every gRPC service shares: tracing, the
// database, the Kafka outbox and consumers, and the gRPC server, all shut
// down cleanly on SIGINT or SIGTERM.
//
//	func main() {
//		app.Run("user-service", func(ctx context.Context, a *app.App) error {
//			db, err := a.Database("user_service", &models.User{})
//			if err != nil {
//				return err
//			}
//			…
//			return a.ServeGRPC(":50051", userpb.UserService_ServiceDesc.ServiceName, func(s *grpc.Server) {
//				userpb.RegisterUserServiceServer(s, handler)
//			})
//		})
//	}
package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"github.com/jochem11/inventory-manager/shared/telemetry"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// App holds what a service started, so Run can close it again.
type App struct {
	ctx     context.Context
	closers []func()
}

// Run starts a service: setup builds and runs it, and only returns when the
// service stops. Run gives setup a context that ends on SIGINT or SIGTERM,
// sets up tracing as service name, and afterwards closes everything setup
// opened through the App, in reverse order. An error from setup is fatal.
func Run(name string, setup func(ctx context.Context, a *App) error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := telemetry.Setup(ctx, name)
	if err != nil {
		log.Fatalf("set up tracing: %v", err)
	}
	a := &App{ctx: ctx}
	a.onClose(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownTracing(ctx)
	})

	err = setup(ctx, a)
	for i := len(a.closers) - 1; i >= 0; i-- {
		a.closers[i]()
	}
	if err != nil {
		log.Fatalf("%s: %v", name, err)
	}
}

func (a *App) onClose(fn func()) { a.closers = append(a.closers, fn) }

// Database connects to the service's MySQL database (see
// database.ConfigFromEnv; defaultName is its usual name), creating it when
// it's missing. It traces every statement, without the values (they hold
// personal data), and migrates models plus the Kafka outbox tables.
func (a *App) Database(defaultName string, models ...any) (*gorm.DB, error) {
	cfg := database.ConfigFromEnv(defaultName)
	if err := database.CreateIfMissing(cfg); err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithoutQueryVariables())); err != nil {
		return nil, fmt.Errorf("add tracing to database: %w", err)
	}
	if err := db.AutoMigrate(append(models, kafka.Models()...)...); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	log.Printf("connected to database %s", cfg.Name)
	return db, nil
}

// Outbox publishes the outbox table of db to Kafka (see kafka.Enqueue),
// until the service stops.
func (a *App) Outbox(db *gorm.DB) error {
	producer, err := kafka.NewProducer(kafka.BrokersFromEnv())
	if err != nil {
		return fmt.Errorf("create kafka producer: %w", err)
	}
	a.onClose(producer.Close)
	go kafka.RunRelay(a.ctx, db, producer)
	return nil
}

// Consume handles the records of topic in consumer group group, until the
// service stops.
func (a *App) Consume(group, topic string, handle kafka.Handler) error {
	consumer, err := kafka.NewConsumer(kafka.BrokersFromEnv(), group, topic)
	if err != nil {
		return fmt.Errorf("create kafka consumer for %s: %w", topic, err)
	}
	a.onClose(consumer.Close)
	go kafka.Consume(a.ctx, consumer, handle)
	return nil
}
