package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

const (
	retryMinBackoff = time.Second
	retryMaxBackoff = 30 * time.Second
)

// Handler handles one record. A returned error is retried with backoff, so
// a database outage pauses consuming instead of losing events; wrap errors
// that retrying can't fix with Permanent.
type Handler func(ctx context.Context, record *kgo.Record) error

type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

// Permanent marks err as one that retrying won't fix, such as an event that
// can't be decoded. Consume logs it and moves on to the next record.
func Permanent(err error) error {
	return permanentError{err}
}

// Decode unmarshals a record's value into msg. A record that can't be decoded
// never will be, so the error is Permanent: Consume skips the record.
func Decode(record *kgo.Record, msg proto.Message) error {
	if err := proto.Unmarshal(record.Value, msg); err != nil {
		return Permanent(fmt.Errorf("decode %s: %w", msg.ProtoReflect().Descriptor().FullName(), err))
	}
	return nil
}

// Consume handles records until ctx is done, in order per partition. Offsets
// are committed after the records are handled, so a crash redelivers them
// rather than losing them; handlers make that safe with FirstDelivery.
func Consume(ctx context.Context, client *kgo.Client, handle Handler) {
	for {
		fetches := client.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			slog.ErrorContext(ctx, "kafka fetch", "topic", topic, "partition", partition, "error", err)
		})
		fetches.EachRecord(func(r *kgo.Record) {
			handleWithRetry(ctx, r, handle)
		})
		if ctx.Err() != nil {
			return
		}
		if err := client.CommitUncommittedOffsets(ctx); err != nil && ctx.Err() == nil {
			slog.ErrorContext(ctx, "kafka commit", "error", err)
		}
	}
}

func handleWithRetry(ctx context.Context, r *kgo.Record, handle Handler) {
	backoff := retryMinBackoff
	for {
		err := handleTraced(ctx, r, handle)
		if err == nil {
			return
		}
		if errors.As(err, new(permanentError)) {
			slog.ErrorContext(ctx, "skipping kafka record", "topic", r.Topic, "offset", r.Offset, "error", err)
			return
		}
		slog.WarnContext(ctx, "kafka record failed, retrying", "topic", r.Topic, "offset", r.Offset, "in", backoff, "error", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, retryMaxBackoff)
	}
}

// handleTraced runs handle in a span that continues the producer's trace.
func handleTraced(ctx context.Context, r *kgo.Record, handle Handler) error {
	ctx = otel.GetTextMapPropagator().Extract(ctx, recordHeaders{r})
	ctx, span := tracer.Start(ctx, "consume "+r.Topic,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", r.Topic),
			attribute.Int64("messaging.kafka.offset", r.Offset),
			attribute.Int("messaging.kafka.destination.partition", int(r.Partition)),
		))
	defer span.End()

	err := handle(ctx, r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
