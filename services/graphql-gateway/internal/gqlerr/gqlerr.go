// Package gqlerr gives every GraphQL error an extensions.code. Errors are
// classified with the shared error model (shared/errs): a resolver returns a
// service's gRPC error as it is, or an errs error of its own, and the error
// presenter turns the kind into a code.
package gqlerr

import (
	"context"
	"errors"
	"log/slog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jochem11/inventory-manager/shared/errs"
	"github.com/vektah/gqlparser/v2/gqlerror"
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

var kindCodes = map[errs.Kind]string{
	errs.Invalid:            CodeBadUserInput,
	errs.NotFound:           CodeNotFound,
	errs.AlreadyExists:      CodeConflict,
	errs.Unauthenticated:    CodeUnauthenticated,
	errs.PermissionDenied:   CodeForbidden,
	errs.FailedPrecondition: CodeFailedPrecondition,
}

// Present is the server's error presenter. Every error gets an
// extensions.code from its kind; BAD_USER_INPUT also gets extensions.fields,
// a message per field. Unexpected errors are logged and returned as INTERNAL
// without details, so internals don't leak.
func Present(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	// Errors gqlgen made itself (e.g. an invalid query) already have a code.
	if _, ok := gqlErr.Extensions["code"]; ok {
		return gqlErr
	}

	err = errs.FromStatus(err)
	if code, ok := kindCodes[errs.KindOf(err)]; ok {
		gqlErr.Extensions = map[string]any{"code": code}
		var invalid *errs.ValidationError
		var e *errs.Error
		switch {
		case errors.As(err, &invalid):
			gqlErr.Message = invalid.Error()
			gqlErr.Extensions["fields"] = invalid.Fields
		case errors.As(err, &e):
			gqlErr.Message = e.Message
		}
		return gqlErr
	}

	// A client that hung up doesn't need logging; nobody reads the response.
	if !errors.Is(err, context.Canceled) && status.Code(err) != codes.Canceled {
		slog.ErrorContext(ctx, "resolver failed", "path", gqlErr.Path.String(), "error", err)
	}
	gqlErr.Message = "internal error"
	gqlErr.Extensions = map[string]any{"code": CodeInternal}
	return gqlErr
}
