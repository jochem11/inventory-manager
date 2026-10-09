// Package errs is the error model every service shares. An error has a Kind
// that says what went wrong in terms a client can act on (not found, invalid
// input, …). The kind decides the gRPC status code a service answers with
// (ToStatus), and the gateway turns it back into a GraphQL error code.
//
// Services define their own errors with New, so callers can still match them
// with errors.Is:
//
//	var ErrUserNotFound = errs.New(errs.NotFound, "user not found")
//
// Any other error counts as Internal: logged, and never shown to a client.
package errs

import (
	"errors"
	"sort"
	"strings"
)

// Kind is the category of an error.
type Kind int

const (
	// Internal is anything unexpected. Its message is never returned.
	Internal Kind = iota
	// Invalid is input that failed validation; see ValidationError.
	Invalid
	NotFound
	// AlreadyExists is a conflict with existing data, e.g. a taken email.
	AlreadyExists
	// Unauthenticated is a missing, wrong or expired login, token or link.
	Unauthenticated
	// PermissionDenied is a login without the needed permission or role.
	PermissionDenied
	// FailedPrecondition is a request that can't be done yet, e.g. logging
	// in before the email is verified.
	FailedPrecondition
)

var kindNames = map[Kind]string{
	Internal:           "internal",
	Invalid:            "invalid",
	NotFound:           "not found",
	AlreadyExists:      "already exists",
	Unauthenticated:    "unauthenticated",
	PermissionDenied:   "permission denied",
	FailedPrecondition: "failed precondition",
}

func (k Kind) String() string { return kindNames[k] }

// Error is an error with a kind and a message that is safe to show a client.
type Error struct {
	Kind    Kind
	Message string
}

// New returns an error of kind with a client-safe message. Create it once, as
// a package variable, so callers can match it with errors.Is.
func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

func (e *Error) Error() string { return e.Message }

// NotFoundError is the NotFound error for a kind of record, e.g.
// NotFoundError("item") says "item not found". Like New, create it once:
//
//	var ErrItemNotFound = errs.NotFoundError("item")
func NotFoundError(record string) *Error {
	return New(NotFound, record+" not found")
}

// AlreadyExistsError is the AlreadyExists error for a unique value, e.g.
// AlreadyExistsError("a category with this name") says "a category with this
// name already exists". Like New, create it once.
func AlreadyExistsError(what string) *Error {
	return New(AlreadyExists, what+" already exists")
}

// ValidationError is input that failed validation. Fields maps a field's name
// (camelCase, like the GraphQL inputs) to a readable message.
type ValidationError struct {
	Fields map[string]string
}

// Field returns a ValidationError for one field.
func Field(field, message string) *ValidationError {
	return &ValidationError{Fields: map[string]string{field: message}}
}

func (e *ValidationError) Error() string {
	names := make([]string, 0, len(e.Fields))
	for name := range e.Fields {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = e.Fields[name]
	}
	return "invalid input: " + strings.Join(parts, "; ")
}

// KindOf returns the kind of the first Error or ValidationError in err's
// chain, or Internal when there is none.
func KindOf(err error) Kind {
	var invalid *ValidationError
	if errors.As(err, &invalid) {
		return Invalid
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return Internal
}
