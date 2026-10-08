package domain

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email is already in use")
)

// ValidationError reports input that failed validation. Fields maps a field's
// JSON name (camelCase, like the web table's column ids) to a readable message,
// e.g. "email" → "email must be a valid email address".
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
