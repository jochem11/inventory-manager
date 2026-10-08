// Package kafka is the services' shared Kafka plumbing. Services publish
// events through a transactional outbox (Enqueue + RunRelay) and consume them
// with Consume, recording each event in processed_events (FirstDelivery) so a
// redelivered event is handled only once.
package kafka

import (
	"os"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Topics. Every message is keyed by user id, so one user's events keep their
// order.
const (
	TopicAuthEvents = "auth.events"
	TopicUserEvents = "user.events"
)

// BrokersFromEnv reads KAFKA_BROKERS (comma-separated), falling back to the
// Tilt port-forward for services run on your machine.
func BrokersFromEnv() []string {
	v := os.Getenv("KAFKA_BROKERS")
	if v == "" {
		v = "127.0.0.1:9094"
	}
	return strings.Split(v, ",")
}

// NewProducer returns a client for the outbox relay. franz-go producers are
// idempotent and wait for all in-sync replicas by default.
func NewProducer(brokers []string) (*kgo.Client, error) {
	return kgo.NewClient(kgo.SeedBrokers(brokers...))
}

// NewConsumer returns a client that reads topics as consumer group group.
// Offsets are committed by Consume after records are handled, and a new
// group starts at the oldest message so no event is missed.
func NewConsumer(brokers []string, group string, topics ...string) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
}

// NewEventID returns an id for an event envelope: a KSUID, so ids sort by
// creation time like the rest of the system's ids.
func NewEventID() string {
	return ksuid.New().String()
}

// Models are the tables this package needs; add them to AutoMigrate.
func Models() []any {
	return []any{&OutboxMessage{}, &ProcessedEvent{}}
}
