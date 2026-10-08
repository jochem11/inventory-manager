package kafka

import (
	"time"

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
// it is seen. Call it first inside the transaction that handles the event:
// if that transaction rolls back, the record goes with it and the event is
// handled again on the retry.
func FirstDelivery(tx *gorm.DB, eventID string) (bool, error) {
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ProcessedEvent{EventID: eventID, ProcessedAt: time.Now()})
	return result.RowsAffected == 1, result.Error
}
