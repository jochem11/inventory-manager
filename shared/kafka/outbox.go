package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	relayInterval  = 500 * time.Millisecond
	relayBatchSize = 100
	// Published rows are kept this long, for debugging, then deleted.
	outboxRetention = 7 * 24 * time.Hour
)

var tracer = otel.Tracer("github.com/jochem11/inventory-manager/shared/kafka")

// OutboxMessage is an event waiting to be published. It is written in the
// same database transaction as the change it describes, so the event exists
// if and only if the change was committed; RunRelay then publishes it.
type OutboxMessage struct {
	// ID orders the messages: the relay publishes them in this order.
	ID      uint64 `gorm:"primaryKey;autoIncrement"`
	EventID string `gorm:"size:64;not null;uniqueIndex"`
	Topic   string `gorm:"size:255;not null"`
	Key     string `gorm:"size:255;not null"`
	Payload []byte `gorm:"type:mediumblob;not null"`
	// Headers carry the trace context of the request that caused the event.
	Headers     map[string]string `gorm:"serializer:json"`
	CreatedAt   time.Time
	PublishedAt *time.Time `gorm:"index"`
}

func (OutboxMessage) TableName() string { return "outbox" }

// Enqueue adds event to the outbox. Call it inside the transaction that makes
// the change (database.Transaction), so both commit or neither does. key is
// the user id the event is about.
func Enqueue(ctx context.Context, db *gorm.DB, topic, key, eventID string, event proto.Message) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	headers := map[string]string{}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

	msg := &OutboxMessage{EventID: eventID, Topic: topic, Key: key, Payload: payload, Headers: headers}
	if err := database.DB(ctx, db).Create(msg).Error; err != nil {
		return fmt.Errorf("add event to outbox: %w", err)
	}
	return nil
}

// RunRelay publishes outbox messages in order until ctx is done. A message
// is marked published only after Kafka acknowledged it, so a crash in between
// publishes it again: consumers must expect duplicates (see FirstDelivery).
func RunRelay(ctx context.Context, db *gorm.DB, client *kgo.Client) {
	ticker := time.NewTicker(relayInterval)
	defer ticker.Stop()
	lastCleanup := time.Time{}
	for {
		n, err := publishBatch(ctx, db, client)
		if err != nil && ctx.Err() == nil {
			slog.ErrorContext(ctx, "outbox relay", "error", err)
		}
		if time.Since(lastCleanup) > time.Hour {
			deletePublished(ctx, db)
			lastCleanup = time.Now()
		}
		if n == relayBatchSize {
			continue // more are waiting
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// publishBatch publishes up to relayBatchSize unpublished messages. SKIP
// LOCKED lets several replicas of a service run the relay without
// publishing the same rows.
func publishBatch(ctx context.Context, db *gorm.DB, client *kgo.Client) (int, error) {
	var published int
	err := database.Transaction(ctx, db, func(ctx context.Context) error {
		tx := database.DB(ctx, db)
		var msgs []OutboxMessage
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("published_at IS NULL").
			Order("id").
			Limit(relayBatchSize).
			Find(&msgs).Error
		if err != nil || len(msgs) == 0 {
			return err
		}

		records := make([]*kgo.Record, len(msgs))
		spans := make([]trace.Span, len(msgs))
		for i, m := range msgs {
			records[i] = &kgo.Record{Topic: m.Topic, Key: []byte(m.Key), Value: m.Payload}
			// The publish span continues the trace of the request that
			// caused the event, and the consumer continues from it.
			parent := otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(m.Headers))
			spanCtx, span := tracer.Start(parent, "publish "+m.Topic,
				trace.WithSpanKind(trace.SpanKindProducer),
				trace.WithAttributes(
					attribute.String("messaging.system", "kafka"),
					attribute.String("messaging.destination.name", m.Topic),
					attribute.String("messaging.message.id", m.EventID),
				))
			otel.GetTextMapPropagator().Inject(spanCtx, recordHeaders{records[i]})
			spans[i] = span
		}
		err = client.ProduceSync(ctx, records...).FirstErr()
		for _, span := range spans {
			if err != nil {
				span.RecordError(err)
			}
			span.End()
		}
		if err != nil {
			return fmt.Errorf("produce: %w", err)
		}

		ids := make([]uint64, len(msgs))
		for i, m := range msgs {
			ids[i] = m.ID
		}
		if err := tx.Model(&OutboxMessage{}).Where("id IN ?", ids).Update("published_at", time.Now()).Error; err != nil {
			return fmt.Errorf("mark published: %w", err)
		}
		published = len(msgs)
		return nil
	})
	return published, err
}

func deletePublished(ctx context.Context, db *gorm.DB) {
	err := db.WithContext(ctx).
		Where("published_at < ?", time.Now().Add(-outboxRetention)).
		Delete(&OutboxMessage{}).Error
	if err != nil && ctx.Err() == nil {
		slog.ErrorContext(ctx, "clean up outbox", "error", err)
	}
}
