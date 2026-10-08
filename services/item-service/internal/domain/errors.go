package domain

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrItemNotFound        = errors.New("item not found")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryNameTaken   = errors.New("a category with this name already exists")
	ErrItemStatusNotFound  = errors.New("item status not found")
	ErrItemStatusNameTaken = errors.New("an item status with this name already exists")
)

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
