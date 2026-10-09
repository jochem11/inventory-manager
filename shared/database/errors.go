package database

import (
	"errors"

	"gorm.io/gorm"
)

// TranslateError turns GORM's errors into a repository's own domain errors:
// a missing row becomes notFound, and a duplicate unique key becomes
// duplicate. Other errors are returned as they are.
//
// duplicate is optional: leave it out for a table whose only unique column
// is its id. A duplicate key there is a bug, not something the caller can
// fix, so it stays an unexpected (internal) error.
//
//	database.TranslateError(err, ErrItemNotFound)                          // no unique columns
//	database.TranslateError(err, ErrCategoryNotFound, ErrCategoryNameTaken) // unique name
//
// Duplicate keys are recognised because Connect sets gorm.Config.TranslateError.
func TranslateError(err, notFound error, duplicate ...error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return notFound
	case errors.Is(err, gorm.ErrDuplicatedKey) && len(duplicate) > 0:
		return duplicate[0]
	}
	return err
}
