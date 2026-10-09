package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
)

// UserService holds the user business logic: normalizing and validating
// input, and rules such as unique emails. The implementation lives in the
// service package.
//
// Errors: invalid input is an *errs.ValidationError, a missing user wraps
// ErrUserNotFound and a duplicate email wraps ErrEmailTaken.
type UserService interface {
	Create(ctx context.Context, input types.UserInput) (*models.User, error)
	// CreateWithID creates a user under an id chosen elsewhere: the
	// auth-service picks it at registration (see the IdentityRegistered event).
	CreateWithID(ctx context.Context, id string, input types.UserInput) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// FindByIDs returns the users with these ids, skipping unknown ones.
	FindByIDs(ctx context.Context, ids []string) ([]*models.User, error)
	// FindAll returns one page of users. It trims the filters, defaults the
	// page size and rejects out-of-range paging and unknown sort fields.
	FindAll(ctx context.Context, params types.UserListParams) (*types.Page[*models.User], error)
	// Update replaces all of the user's editable fields with input, so a
	// missing required field is a validation error rather than left unchanged.
	Update(ctx context.Context, id string, input types.UserInput) (*models.User, error)
	Delete(ctx context.Context, id string) error
}
