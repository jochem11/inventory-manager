package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/validation"
)

type UserServiceImp struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) domain.UserService {
	return &UserServiceImp{users: users}
}

func (s *UserServiceImp) Create(ctx context.Context, input types.UserInput) (*models.User, error) {
	return s.CreateWithID(ctx, "", input)
}

// CreateWithID validates id like any other field (a KSUID); an empty id gets
// a new one.
func (s *UserServiceImp) CreateWithID(ctx context.Context, id string, input types.UserInput) (*models.User, error) {
	user := &models.User{Model: database.Model{ID: id}}
	applyInput(user, input)
	if err := validation.Clean(ctx, user); err != nil {
		return nil, err
	}
	if err := s.ensureEmailFree(ctx, user.Email, ""); err != nil {
		return nil, err
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserServiceImp) FindByID(ctx context.Context, id string) (*models.User, error) {
	return s.users.FindByID(ctx, id)
}

// FindByEmail relies on the email column's case-insensitive collation, so the
// email needn't be normalized first.
func (s *UserServiceImp) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.users.FindByEmail(ctx, email)
}

func (s *UserServiceImp) FindByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	return s.users.FindByIDs(ctx, ids)
}

func (s *UserServiceImp) FindAll(ctx context.Context, params types.UserListParams) (*types.Page[*models.User], error) {
	if err := validation.Clean(ctx, &params); err != nil {
		return nil, err
	}
	return s.users.FindAll(ctx, params)
}

func (s *UserServiceImp) Update(ctx context.Context, id string, input types.UserInput) (*models.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	oldEmail := user.Email

	applyInput(user, input)
	if err := validation.Clean(ctx, user); err != nil {
		return nil, err
	}
	if user.Email != oldEmail {
		if err := s.ensureEmailFree(ctx, user.Email, user.ID); err != nil {
			return nil, err
		}
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserServiceImp) Delete(ctx context.Context, id string) error {
	return s.users.Delete(ctx, id)
}

// ensureEmailFree returns ErrEmailTaken when another user than exceptID has
// the email. The unique index still backs this up against concurrent writes
// and soft-deleted users, which the repository also reports as ErrEmailTaken.
func (s *UserServiceImp) ensureEmailFree(ctx context.Context, email, exceptID string) error {
	existing, err := s.users.FindByEmail(ctx, email)
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return nil
	case err != nil:
		return err
	case existing.ID != exceptID:
		return fmt.Errorf("%q: %w", email, domain.ErrEmailTaken)
	}
	return nil
}

// applyInput sets every editable field of user from input, including empty
// ones, so validation sees exactly what the client sent.
func applyInput(user *models.User, input types.UserInput) {
	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.Email = input.Email
	user.Phone = input.Phone
	user.AvatarURL = input.AvatarURL
}
