package domain

import "github.com/jochem11/inventory-manager/shared/errs"

// The item-service's errors. Their kind decides the gRPC status code
// (errs.ToStatus); invalid input is an *errs.ValidationError.
var (
	// Items have no unique columns besides their id, so no "taken" error.
	ErrItemNotFound = errs.NotFoundError("item")
	// ErrUnknownReference is an item whose category or status doesn't exist.
	ErrUnknownReference = errs.New(errs.Invalid, "the category or status doesn't exist")

	ErrCategoryNotFound  = errs.NotFoundError("category")
	ErrCategoryNameTaken = errs.AlreadyExistsError("a category with this name")
	// ErrCategoryInUse is deleting a category that items still have: move or
	// delete those items first.
	ErrCategoryInUse = errs.New(errs.FailedPrecondition, "the category still has items")

	ErrItemStatusNotFound  = errs.NotFoundError("item status")
	ErrItemStatusNameTaken = errs.AlreadyExistsError("an item status with this name")
	// ErrItemStatusInUse is deleting a status that items still have.
	ErrItemStatusInUse = errs.New(errs.FailedPrecondition, "the status still has items")
)
