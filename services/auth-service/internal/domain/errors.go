package domain

import "github.com/jochem11/inventory-manager/shared/errs"

// The auth-service's errors. Their kind decides the gRPC status code
// (errs.ToStatus); invalid input is an *errs.ValidationError.
var (
	ErrEmailTaken       = errs.New(errs.AlreadyExists, "email is already registered")
	ErrIdentityNotFound = errs.NotFoundError("identity")
	// ErrInvalidToken covers unknown, used and expired tokens alike, so a
	// caller can't tell which tokens exist.
	ErrInvalidToken = errs.New(errs.Unauthenticated, "invalid or expired token")
	// ErrInvalidCredentials doesn't say whether the email or the password
	// was wrong, so it can't be used to find out who has an account.
	ErrInvalidCredentials = errs.New(errs.Unauthenticated, "invalid email or password")
	ErrEmailNotVerified   = errs.New(errs.FailedPrecondition, "email is not verified")
)
