package domain

import "github.com/jochem11/inventory-manager/shared/errs"

// The user-service's errors. Their kind decides the gRPC status code
// (errs.ToStatus); invalid input is an *errs.ValidationError.
var (
	ErrUserNotFound = errs.NotFoundError("user")
	ErrEmailTaken   = errs.New(errs.AlreadyExists, "email is already in use")
)
