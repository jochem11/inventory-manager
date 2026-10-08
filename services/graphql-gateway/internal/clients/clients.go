// Package clients creates the gateway's connections to the gRPC services.
package clients

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Dial returns a connection to the service at addr (host:port, no scheme).
// It doesn't connect until the first call, so the gateway starts even while
// a service is still down. Every call gets a client span, and the trace
// context travels along in the gRPC metadata.
func Dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(logCalls),
	)
}

// logCalls logs every outgoing gRPC call, so you can see how many calls one
// GraphQL request makes.
func logCalls(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	slog.InfoContext(ctx, "grpc call", "method", method, "code", status.Code(err).String(), "duration", time.Since(start),
		"trace_id", trace.SpanContextFromContext(ctx).TraceID().String())
	return err
}
