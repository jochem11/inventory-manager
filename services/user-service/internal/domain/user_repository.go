package domain

import (
	"context"

	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
)

// UserRepository persists users. The implementation lives in the repository
// package.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// FindByIDs returns the users with these ids, skipping unknown ones.
	FindByIDs(ctx context.Context, ids []string) ([]*models.User, error)
	// FindAll returns one page of users, with how many match the filter across
	// all pages.
	FindAll(ctx context.Context, params types.UserListParams) (*types.Page[*models.User], error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
}
