// Package events handles the Kafka events the auth-service consumes.
package events

import (
	"context"
	"errors"
	"fmt"

	"github.com/jochem11/inventory-manager/services/auth-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/repository"
	userpb "github.com/jochem11/inventory-manager/services/auth-service/pkg/pb/user"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// UserEvents handles the user.events topic.
type UserEvents struct {
	db *gorm.DB
}

func NewUserEvents(db *gorm.DB) *UserEvents {
	return &UserEvents{db: db}
}

// Handle is a kafka.Handler. Event types it doesn't know are skipped, so the
// user-service can add new ones without breaking this consumer.
func (h *UserEvents) Handle(ctx context.Context, record *kgo.Record) error {
	var event userpb.UserEvent
	if err := proto.Unmarshal(record.Value, &event); err != nil {
		return kafka.Permanent(fmt.Errorf("decode user event: %w", err))
	}
	switch payload := event.GetPayload().(type) {
	case *userpb.UserEvent_UserDeleted:
		return h.userDeleted(ctx, event.GetEventId(), payload.UserDeleted)
	}
	return nil
}

// userDeleted removes the deleted user's login and sessions.
func (h *UserEvents) userDeleted(ctx context.Context, eventID string, e *userpb.UserDeleted) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		first, err := kafka.FirstDelivery(tx, eventID)
		if err != nil || !first {
			return err
		}
		err = repository.NewIdentityRepository(tx).DeleteIdentityByUserID(ctx, e.GetUserId())
		// Users created without registering (e.g. createUser) have no login.
		if errors.Is(err, domain.ErrIdentityNotFound) {
			return nil
		}
		return err
	})
}
