package errs

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errThingNotFound = New(NotFound, "thing not found")

func TestKindOf(t *testing.T) {
	tests := []struct {
		err  error
		want Kind
	}{
		{errThingNotFound, NotFound},
		{fmt.Errorf("find thing %q: %w", "x", errThingNotFound), NotFound},
		{Field("name", "name is required"), Invalid},
		{fmt.Errorf("save: %w", Field("name", "name is required")), Invalid},
		{errors.New("connection refused"), Internal},
		{nil, Internal},
	}
	for _, tt := range tests {
		if got := KindOf(tt.err); got != tt.want {
			t.Errorf("KindOf(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
	if !errors.Is(fmt.Errorf("wrapped: %w", errThingNotFound), errThingNotFound) {
		t.Error("errors.Is doesn't find a wrapped error")
	}
}

func TestToStatus(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
		wantMsg  string
	}{
		{"kind, own message", fmt.Errorf("find thing %q: %w", "x", errThingNotFound), codes.NotFound, "thing not found"},
		{"validation", Field("name", "name is required"), codes.InvalidArgument, "invalid input: name is required"},
		{"unexpected", errors.New("dial tcp: refused"), codes.Internal, "internal error"},
		{"cancelled", fmt.Errorf("query: %w", context.Canceled), codes.Canceled, "query: context canceled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := status.Convert(ToStatus(ctx, tt.err))
			if st.Code() != tt.wantCode || st.Message() != tt.wantMsg {
				t.Errorf("got %v %q, want %v %q", st.Code(), st.Message(), tt.wantCode, tt.wantMsg)
			}
		})
	}
}

// What a service sends comes back as the same kind, message and fields.
func TestFromStatusRoundTrip(t *testing.T) {
	ctx := context.Background()

	back := FromStatus(fmt.Errorf("call: %w", ToStatus(ctx, errThingNotFound)))
	if KindOf(back) != NotFound || back.Error() != "thing not found" {
		t.Errorf("not found: got %v %q", KindOf(back), back.Error())
	}

	back = FromStatus(ToStatus(ctx, &ValidationError{Fields: map[string]string{"name": "required", "email": "invalid"}}))
	var invalid *ValidationError
	if !errors.As(back, &invalid) || invalid.Fields["name"] != "required" || invalid.Fields["email"] != "invalid" {
		t.Errorf("validation: got %#v", back)
	}

	internal := ToStatus(ctx, errors.New("boom"))
	if back := FromStatus(internal); KindOf(back) != Internal {
		t.Errorf("internal: got kind %v", KindOf(back))
	}
	plain := errors.New("no status")
	if FromStatus(plain) != plain {
		t.Error("an error without a status should come back unchanged")
	}
}

func TestRecordErrors(t *testing.T) {
	notFound := NotFoundError("item")
	taken := AlreadyExistsError("a category with this name")
	if KindOf(notFound) != NotFound || notFound.Error() != "item not found" {
		t.Errorf("NotFoundError: %v %q", KindOf(notFound), notFound.Error())
	}
	if KindOf(taken) != AlreadyExists || taken.Error() != "a category with this name already exists" {
		t.Errorf("AlreadyExistsError: %v %q", KindOf(taken), taken.Error())
	}
}
