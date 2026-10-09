package kafka

import (
	"context"
	"time"

	"github.com/jochem11/inventory-manager/shared/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProcessedEvent records an event a consumer has handled, so a redelivered
// copy is skipped.
type ProcessedEvent struct {
	EventID     string `gorm:"size:64;primaryKey"`
	ProcessedAt time.Time
}

// FirstDelivery records eventID and reports whether this is the first time
// it is seen. Call it first inside the transaction that handles the event
// (database.Transaction): if that transaction rolls back, the record goes
// with it and the event is handled again on the retry.
func FirstDelivery(ctx context.Context, db *gorm.DB, eventID string) (bool, error) {
	result := database.DB(ctx, db).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ProcessedEvent{EventID: eventID, ProcessedAt: time.Now()})
	return result.RowsAffected == 1, result.Error
}

// HandleOnce runs fn for an event at most once, however often Kafka delivers
// it: in one transaction it records eventID (FirstDelivery) and runs fn with
// the transaction's ctx. A redelivered event is skipped. If fn fails, the
// record rolls back with it, so the retry runs fn again.
func HandleOnce(ctx context.Context, db *gorm.DB, eventID string, fn func(ctx context.Context) error) error {
	return database.Transaction(ctx, db, func(ctx context.Context) error {
		first, err := FirstDelivery(ctx, db, eventID)
		if err != nil || !first {
			return err
		}
		return fn(ctx)
	})
}
