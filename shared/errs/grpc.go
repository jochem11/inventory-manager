package errs

import (
	"context"
	"errors"
	"log/slog"
	"sort"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var kindCodes = map[Kind]codes.Code{
	Invalid:            codes.InvalidArgument,
	NotFound:           codes.NotFound,
	AlreadyExists:      codes.AlreadyExists,
	Unauthenticated:    codes.Unauthenticated,
	PermissionDenied:   codes.PermissionDenied,
	FailedPrecondition: codes.FailedPrecondition,
}

// ToStatus turns a service error into the gRPC status to answer with:
//   - an Error gets its kind's code and its own message, not the text of the
//     errors wrapping it
//   - a ValidationError is INVALID_ARGUMENT with a google.rpc.BadRequest detail
//     per field, so clients can show each message next to its field
//   - a cancelled or timed-out request keeps that code
//   - anything else is logged and returned as INTERNAL, without details
func ToStatus(ctx context.Context, err error) error {
	var invalid *ValidationError
	if errors.As(err, &invalid) {
		return invalidArgument(invalid)
	}
	var e *Error
	if errors.As(err, &e) && e.Kind != Internal {
		return status.Error(kindCodes[e.Kind], e.Message)
	}
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	}
	slog.ErrorContext(ctx, "request failed", "error", err)
	return status.Error(codes.Internal, "internal error")
}

func invalidArgument(invalid *ValidationError) error {
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

// FromStatus is the other way round, for a client of a service: it turns a
// gRPC error back into an Error or ValidationError, so the caller can use
// KindOf. Errors without a gRPC status, and codes without a kind (INTERNAL,
// CANCELED, UNAVAILABLE, …), are returned unchanged, so KindOf says Internal.
func FromStatus(err error) error {
	// errors.As finds the status anywhere in the chain; status.FromError would
	// use the text of the errors wrapping it as the message.
	var grpcErr interface{ GRPCStatus() *status.Status }
	if !errors.As(err, &grpcErr) {
		return err
	}
	st := grpcErr.GRPCStatus()
	if st.Code() == codes.InvalidArgument {
		fields := map[string]string{}
		for _, detail := range st.Details() {
			if br, ok := detail.(*errdetails.BadRequest); ok {
				for _, v := range br.GetFieldViolations() {
					fields[v.GetField()] = v.GetDescription()
				}
			}
		}
		if len(fields) > 0 {
			return &ValidationError{Fields: fields}
		}
		return New(Invalid, st.Message())
	}
	for kind, code := range kindCodes {
		if st.Code() == code {
			return New(kind, st.Message())
		}
	}
	return err
}
