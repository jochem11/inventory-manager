package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	userpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/user"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"github.com/jochem11/inventory-manager/shared/paging"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type UserRepositoryImp struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &UserRepositoryImp{db: db}
}

func (r *UserRepositoryImp) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", translateError(err))
	}
	return nil
}

func (r *UserRepositoryImp) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find user by id %q: %w", id, translateError(err))
	}
	return &user, nil
}

func (r *UserRepositoryImp) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("find user by email %q: %w", email, translateError(err))
	}
	return &user, nil
}

func (r *UserRepositoryImp) FindByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	users := []*models.User{}
	if len(ids) == 0 {
		return users, nil
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Order("id").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("find users by ids: %w", err)
	}
	return users, nil
}

func (r *UserRepositoryImp) FindAll(ctx context.Context, params types.UserListParams) (*types.Page[*models.User], error) {
	order, err := paging.OrderBy(userSortColumns, params.OrderBy.Field, params.OrderBy.Direction)
	if err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	query := r.db.WithContext(ctx).Model(&models.User{}).Scopes(userFilterScope(params.Filter))
	page, err := paging.Find[models.User](query, order, params.Offset, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}
	return page, nil
}

// userSortColumns maps sort fields to columns, see paging.OrderBy.
var userSortColumns = map[types.UserSortField]string{
	types.UserSortFirstName: "first_name",
	types.UserSortLastName:  "last_name",
	types.UserSortEmail:     "email",
	types.UserSortPhone:     "phone",
	types.UserSortCreatedAt: "created_at",
	types.UserSortUpdatedAt: "updated_at",
}

// userFilterScope applies a UserFilter. The name columns use the table's
// default case-insensitive collation, so LIKE matches regardless of case.
func userFilterScope(f types.UserFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if f.Search != "" {
			pattern := paging.Contains(f.Search)
			db = db.Where(
				"first_name LIKE ? OR last_name LIKE ? OR CONCAT(first_name, ' ', last_name) LIKE ? OR email LIKE ? OR phone LIKE ?",
				pattern, pattern, pattern, pattern, pattern,
			)
		}
		for _, c := range []struct{ column, value string }{
			{"first_name", f.FirstName},
			{"last_name", f.LastName},
			{"email", f.Email},
			{"phone", f.Phone},
		} {
			if c.value != "" {
				db = db.Where(c.column+" LIKE ?", paging.Contains(c.value))
			}
		}
		return db
	}
}

// Update writes every field of user except the ID and creation time.
func (r *UserRepositoryImp) Update(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).
		Model(user).
		Select("*").
		Omit("id", "created_at", "deleted_at").
		Updates(user)
	if result.Error != nil {
		return fmt.Errorf("update user %q: %w", user.ID, translateError(result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update user %q: %w", user.ID, domain.ErrUserNotFound)
	}
	return nil
}

// Delete removes the user and, in the same transaction, queues a UserDeleted
// event in the outbox, so the event is published exactly when the delete
// commits.
func (r *UserRepositoryImp) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&models.User{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("delete user %q: %w", id, result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("delete user %q: %w", id, domain.ErrUserNotFound)
		}
		eventID := kafka.NewEventID()
		return kafka.Enqueue(ctx, tx, kafka.TopicUserEvents, id, eventID, &userpb.UserEvent{
			EventId:    eventID,
			OccurredAt: timestamppb.Now(),
			Payload:    &userpb.UserEvent_UserDeleted{UserDeleted: &userpb.UserDeleted{UserId: id}},
		})
	})
}

// translateError maps GORM errors to domain errors. The email index is the
// only unique key besides the (generated) ID, so a duplicate means the email
// is taken. It relies on gorm.Config.TranslateError.
func translateError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return domain.ErrUserNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return domain.ErrEmailTaken
	}
	return err
}
