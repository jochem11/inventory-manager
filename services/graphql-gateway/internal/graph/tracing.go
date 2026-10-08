package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph")

// tracing adds a span per GraphQL operation and per resolver call, so a trace
// shows which fields a request resolved and which gRPC calls each one made.
type tracing struct{}

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = tracing{}

func (tracing) ExtensionName() string                          { return "Tracing" }
func (tracing) Validate(schema graphql.ExecutableSchema) error { return nil }

func (tracing) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	if !graphql.HasOperationContext(ctx) {
		return next(ctx)
	}
	oc := graphql.GetOperationContext(ctx)
	opType := "operation"
	if oc.Operation != nil {
		opType = string(oc.Operation.Operation)
	}
	name := oc.OperationName
	if name == "" {
		name = "anonymous"
	}

	ctx, span := tracer.Start(ctx, "graphql "+opType+" "+name)
	defer span.End()
	span.SetAttributes(
		attribute.String("graphql.operation.type", opType),
		attribute.String("graphql.operation.name", name),
		attribute.String("graphql.document", oc.RawQuery),
	)

	resp := next(ctx)
	if resp != nil && len(resp.Errors) > 0 {
		span.SetStatus(codes.Error, resp.Errors.Error())
	}
	return resp
}

func (tracing) InterceptField(ctx context.Context, next graphql.Resolver) (any, error) {
	fc := graphql.GetFieldContext(ctx)
	// Plain struct fields resolve instantly; only resolver methods do work.
	if fc == nil || !fc.IsResolver {
		return next(ctx)
	}

	ctx, span := tracer.Start(ctx, fc.Object+"."+fc.Field.Name)
	defer span.End()
	span.SetAttributes(attribute.String("graphql.field.path", fc.Path().String()))

	res, err := next(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}
