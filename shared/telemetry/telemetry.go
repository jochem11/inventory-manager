// Package telemetry sets up OpenTelemetry tracing. Spans go over OTLP/gRPC to
// OTEL_EXPORTER_OTLP_ENDPOINT (e.g. http://jaeger:4317); when that isn't set,
// nothing is exported, so running without Jaeger works the same.
package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Setup installs the global tracer provider and returns a function that
// flushes the remaining spans on shutdown. OTEL_SERVICE_NAME overrides
// serviceName.
func Setup(ctx context.Context, serviceName string) (shutdown func(context.Context) error, err error) {
	// W3C trace context in gRPC metadata and HTTP headers is what joins the
	// spans of different services into one trace.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", serviceName)),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	return provider.Shutdown, nil
}
