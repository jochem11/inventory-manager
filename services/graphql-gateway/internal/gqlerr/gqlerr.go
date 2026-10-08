// Package gqlerr gives every GraphQL error an extensions.code, and maps the
// gRPC errors the services return to those codes.
package gqlerr

import (
	"context"
	"errors"
	"log/slog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error codes, returned in extensions.code.
const (
	CodeBadUserInput = "BAD_USER_INPUT"
	// CodeUnauthenticated: not logged in, or the access token is invalid or
	// expired. Clients refresh their token and retry.
	CodeUnauthenticated = "UNAUTHENTICATED"
	// CodeForbidden: logged in, but without the needed permission.
	CodeForbidden = "FORBIDDEN"
	// CodeFailedPrecondition: the request can't be done yet, e.g. logging in
	// before the email is verified.
	CodeFailedPrecondition = "FAILED_PRECONDITION"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeInternal           = "INTERNAL"
)

// BadUserInput is the error for arguments the gateway rejects itself.
func BadUserInput(field, message string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]any{
			"code":   CodeBadUserInput,
			"fields": map[string]string{field: message},
		},
	}
}

// Unauthenticated is the error for a request that needs a logged-in user.
func Unauthenticated(message string) error {
	return &gqlerror.Error{Message: message, Extensions: map[string]any{"code": CodeUnauthenticated}}
}

// Forbidden is the error for a user without the needed permission.
func Forbidden(message string) error {
	return &gqlerror.Error{Message: message, Extensions: map[string]any{"code": CodeForbidden}}
}

// Present is the server's error presenter, so resolvers can return gRPC
// errors as they are. It gives every error an extensions.code: gRPC statuses
// are mapped, and anything unexpected is logged and returned as INTERNAL
// without details, so internals don't leak.
func Present(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	if _, ok := gqlErr.Extensions["code"]; ok {
		return gqlErr
	}

	if st := grpcStatus(err); st != nil {
		switch st.Code() {
		case codes.InvalidArgument:
			gqlErr.Message = st.Message()
			gqlErr.Extensions = map[string]any{"code": CodeBadUserInput}
			if fields := fieldViolations(st); len(fields) > 0 {
				gqlErr.Extensions["fields"] = fields
			}
			return gqlErr
		case codes.NotFound:
			gqlErr.Message = st.Message()
			gqlErr.Extensions = map[string]any{"code": CodeNotFound}
			return gqlErr
		case codes.Unauthenticated:
			gqlErr.Message = st.Message()
			gqlErr.Extensions = map[string]any{"code": CodeUnauthenticated}
			return gqlErr
		case codes.PermissionDenied:
			gqlErr.Message = st.Message()
			gqlErr.Extensions = map[string]any{"code": CodeForbidden}
			return gqlErr
		case codes.FailedPrecondition:
			gqlErr.Message = st.Message()
			gqlErr.Extensions = map[string]any{"code": CodeFailedPrecondition}
			return gqlErr
		case codes.AlreadyExists:
			gqlErr.Message = st.Message()
			gqlErr.Extensions = map[string]any{"code": CodeConflict}
			return gqlErr
		}
	}

	// A client that hung up doesn't need logging; nobody reads the response.
	if !errors.Is(err, context.Canceled) && status.Code(err) != codes.Canceled {
		slog.ErrorContext(ctx, "resolver failed", "path", gqlErr.Path.String(), "error", err)
	}
	gqlErr.Message = "internal error"
	gqlErr.Extensions = map[string]any{"code": CodeInternal}
	return gqlErr
}

// grpcStatus finds the gRPC status in err's chain. Unlike status.FromError it
// keeps the service's own message, not the text of the errors wrapping it.
func grpcStatus(err error) *status.Status {
	var grpcErr interface{ GRPCStatus() *status.Status }
	if errors.As(err, &grpcErr) {
		return grpcErr.GRPCStatus()
	}
	return nil
}

// fieldViolations reads the google.rpc.BadRequest detail the services attach
// to INVALID_ARGUMENT, as field name → message. The services use the same
// camelCase field names as the GraphQL inputs.
func fieldViolations(st *status.Status) map[string]string {
	fields := map[string]string{}
	for _, detail := range st.Details() {
		if br, ok := detail.(*errdetails.BadRequest); ok {
			for _, v := range br.GetFieldViolations() {
				fields[v.GetField()] = v.GetDescription()
			}
		}
	}
	return fields
}
