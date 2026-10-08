// Package events handles the Kafka events the user-service consumes.
package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/user-service/internal/repository"
	authpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/auth"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
	"github.com/jochem11/inventory-manager/services/user-service/service"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// AuthEvents handles the auth.events topic.
type AuthEvents struct {
	db *gorm.DB
}

func NewAuthEvents(db *gorm.DB) *AuthEvents {
	return &AuthEvents{db: db}
}

// Handle is a kafka.Handler. Event types it doesn't know are skipped, so the
// auth-service can add new ones without breaking this consumer.
func (h *AuthEvents) Handle(ctx context.Context, record *kgo.Record) error {
	var event authpb.AuthEvent
	if err := proto.Unmarshal(record.Value, &event); err != nil {
		return kafka.Permanent(fmt.Errorf("decode auth event: %w", err))
	}
	switch payload := event.GetPayload().(type) {
	case *authpb.AuthEvent_IdentityRegistered:
		return h.identityRegistered(ctx, event.GetEventId(), payload.IdentityRegistered)
	}
	return nil
}

// identityRegistered creates the profile of a newly registered user, under
// the id the auth-service generated.
func (h *AuthEvents) identityRegistered(ctx context.Context, eventID string, e *authpb.IdentityRegistered) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		first, err := kafka.FirstDelivery(tx, eventID)
		if err != nil || !first {
			return err
		}

		users := service.NewUserService(repository.NewUserRepository(tx))
		_, err = users.CreateWithID(ctx, e.GetUserId(), types.UserInput{
			FirstName: e.GetFirstName(),
			LastName:  e.GetLastName(),
			Email:     e.GetEmail(),
			Phone:     e.Phone,
		})
		// Retrying can't fix invalid data or a taken email: skip the event.
		var invalid *domain.ValidationError
		if errors.As(err, &invalid) || errors.Is(err, domain.ErrEmailTaken) {
			slog.ErrorContext(ctx, "can't create profile for registered user", "user_id", e.GetUserId(), "error", err)
			return kafka.Permanent(err)
		}
		return err
	})
}
