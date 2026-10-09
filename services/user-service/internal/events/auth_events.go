// Package events handles the Kafka events the user-service consumes.
package events

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	authpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/auth"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/errs"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/gorm"
)

// AuthEvents handles the auth.events topic.
type AuthEvents struct {
	db    *gorm.DB
	users domain.UserService
}

func NewAuthEvents(db *gorm.DB, users domain.UserService) *AuthEvents {
	return &AuthEvents{db: db, users: users}
}

// Handle is a kafka.Handler. Event types it doesn't know are skipped, so the
// auth-service can add new ones without breaking this consumer.
func (h *AuthEvents) Handle(ctx context.Context, record *kgo.Record) error {
	var event authpb.AuthEvent
	if err := kafka.Decode(record, &event); err != nil {
		return err
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
	return kafka.HandleOnce(ctx, h.db, eventID, func(ctx context.Context) error {
		_, err := h.users.CreateWithID(ctx, e.GetUserId(), types.UserInput{
			FirstName: e.GetFirstName(),
			LastName:  e.GetLastName(),
			Email:     e.GetEmail(),
			Phone:     e.Phone,
		})
		// Retrying can't fix invalid data or a taken email: skip the event.
		if errs.KindOf(err) == errs.Invalid || errors.Is(err, domain.ErrEmailTaken) {
			slog.ErrorContext(ctx, "can't create profile for registered user", "user_id", e.GetUserId(), "error", err)
			return kafka.Permanent(err)
		}
		return err
	})
}
