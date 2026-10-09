package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jochem11/inventory-manager/services/auth-service/internal/events"
	grpchandler "github.com/jochem11/inventory-manager/services/auth-service/internal/grpc"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/models"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/repository"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/token"
	authpb "github.com/jochem11/inventory-manager/services/auth-service/pkg/pb/auth"
	"github.com/jochem11/inventory-manager/services/auth-service/service"
	"github.com/jochem11/inventory-manager/shared/app"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/env"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"google.golang.org/grpc"
)

func main() {
	app.Run("auth-service", func(ctx context.Context, a *app.App) error {
		db, err := a.Database("auth_service",
			&models.Permission{}, &models.Role{}, &models.Identity{}, &models.IdentityToken{}, &models.Session{})
		if err != nil {
			return err
		}
		// ADMIN_EMAILS (comma-separated) get the admin role, at every start.
		if err := repository.Seed(ctx, db, adminEmails(env.Get("ADMIN_EMAILS", ""))); err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}

		// JWT_PRIVATE_KEY is a PKCS#8 PEM Ed25519 key; in .env it's one line,
		// with \n for the line breaks. Without it a new key is generated at
		// every start.
		tokens, generated, err := token.NewIssuer(strings.ReplaceAll(env.Get("JWT_PRIVATE_KEY", ""), `\n`, "\n"))
		if err != nil {
			return fmt.Errorf("load signing key: %w", err)
		}
		if generated {
			log.Println("JWT_PRIVATE_KEY not set: signing with a key generated for this run")
		}

		identities := repository.NewIdentityRepository(db)
		auth := service.NewAuthService(identities, database.NewTransactor(db), tokens,
			env.Get("VERIFY_EMAIL_URL", "http://localhost:3000/verify-email"))

		// Publish the outbox (IdentityRegistered, …) and consume user events
		// (UserDeleted removes the login).
		if err := a.Outbox(db); err != nil {
			return err
		}
		if err := a.Consume("auth-service", kafka.TopicUserEvents, events.NewUserEvents(db, identities).Handle); err != nil {
			return err
		}

		return a.ServeGRPC(":50052", authpb.AuthService_ServiceDesc.ServiceName, func(s *grpc.Server) {
			authpb.RegisterAuthServiceServer(s, grpchandler.NewAuthHandler(auth))
		})
	})
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
