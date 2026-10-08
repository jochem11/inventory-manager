package domain

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrEmailTaken       = errors.New("email is already registered")
	ErrIdentityNotFound = errors.New("identity not found")
	// ErrInvalidToken covers unknown, used and expired tokens alike, so a
	// caller can't tell which tokens exist.
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrInvalidCredentials doesn't say whether the email or the password
	// was wrong, so it can't be used to find out who has an account.
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailNotVerified   = errors.New("email is not verified")
)

// ValidationError reports input that failed validation. Fields maps a field's
// name (camelCase, like the GraphQL inputs) to a readable message.
type ValidationError struct {
	Fields map[string]string
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
