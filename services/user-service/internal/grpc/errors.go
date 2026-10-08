package grpc

import (
	"context"
	"errors"
	"log/slog"
	"sort"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toStatus maps a service error to a gRPC status. Unexpected errors are logged
// and returned as INTERNAL without details, so internals don't leak.
func toStatus(ctx context.Context, err error) error {
	var invalid *domain.ValidationError
	switch {
	case errors.As(err, &invalid):
		return invalidArgument(invalid)
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, domain.ErrUserNotFound.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, domain.ErrEmailTaken.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	slog.ErrorContext(ctx, "request failed", "error", err)
	return status.Error(codes.Internal, "internal error")
}

// invalidArgument returns INVALID_ARGUMENT with a google.rpc.BadRequest
// detail per field, so clients can show each message next to its field.
func invalidArgument(invalid *domain.ValidationError) error {
	fields := make([]string, 0, len(invalid.Fields))
	for field := range invalid.Fields {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	details := &errdetails.BadRequest{}
	for _, field := range fields {
		details.FieldViolations = append(details.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: invalid.Fields[field],
		})
	}
	st, err := status.New(codes.InvalidArgument, invalid.Error()).WithDetails(details)
	if err != nil {
		return status.Error(codes.InvalidArgument, invalid.Error())
	}
	return st.Err()
}
