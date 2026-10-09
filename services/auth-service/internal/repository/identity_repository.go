package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jochem11/inventory-manager/services/auth-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/models"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IdentityRepositoryImp struct {
	db *gorm.DB
}

func NewIdentityRepository(db *gorm.DB) domain.IdentityRepository {
	return &IdentityRepositoryImp{db: db}
}

func (r *IdentityRepositoryImp) CreateIdentity(ctx context.Context, identity *models.Identity) error {
	if err := database.DB(ctx, r.db).Create(identity).Error; err != nil {
		return fmt.Errorf("create identity: %w", translateError(err))
	}
	return nil
}

func (r *IdentityRepositoryImp) FindIdentityByID(ctx context.Context, id string) (*models.Identity, error) {
	var identity models.Identity
	if err := database.DB(ctx, r.db).First(&identity, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find identity %q: %w", id, translateError(err))
	}
	return &identity, nil
}

func (r *IdentityRepositoryImp) FindIdentityByEmail(ctx context.Context, email string) (*models.Identity, error) {
	var identity models.Identity
	if err := database.DB(ctx, r.db).First(&identity, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("find identity by email %q: %w", email, translateError(err))
	}
	return &identity, nil
}

func (r *IdentityRepositoryImp) SaveIdentity(ctx context.Context, identity *models.Identity) error {
	if err := database.DB(ctx, r.db).Save(identity).Error; err != nil {
		return fmt.Errorf("save identity %q: %w", identity.ID, translateError(err))
	}
	return nil
}

func (r *IdentityRepositoryImp) FindIdentityWithRoles(ctx context.Context, id string) (*models.Identity, error) {
	var identity models.Identity
	if err := database.DB(ctx, r.db).Preload("Roles.Permissions").First(&identity, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("find identity %q: %w", id, translateError(err))
	}
	return &identity, nil
}

func (r *IdentityRepositoryImp) AssignRole(ctx context.Context, identity *models.Identity, roleName string) error {
	var role models.Role
	if err := database.DB(ctx, r.db).First(&role, "name = ?", roleName).Error; err != nil {
		return fmt.Errorf("find role %q: %w", roleName, err)
	}
	if err := database.DB(ctx, r.db).Model(identity).Association("Roles").Append(&role); err != nil {
		return fmt.Errorf("assign role %q: %w", roleName, err)
	}
	return nil
}

// DeleteIdentityByUserID deletes for real, not softly, so the email can be
// registered again. Sessions and tokens go with it through ON DELETE CASCADE;
// Select("Roles") removes the role assignments.
func (r *IdentityRepositoryImp) DeleteIdentityByUserID(ctx context.Context, userID string) error {
	var identity models.Identity
	if err := database.DB(ctx, r.db).First(&identity, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("find identity of user %q: %w", userID, translateError(err))
	}
	if err := database.DB(ctx, r.db).Unscoped().Select("Roles").Delete(&identity).Error; err != nil {
		return fmt.Errorf("delete identity of user %q: %w", userID, err)
	}
	return nil
}

func (r *IdentityRepositoryImp) CreateToken(ctx context.Context, token *models.IdentityToken) error {
	if err := database.DB(ctx, r.db).Create(token).Error; err != nil {
		return fmt.Errorf("create token: %w", err)
	}
	return nil
}

func (r *IdentityRepositoryImp) FindToken(ctx context.Context, purpose models.TokenPurpose, hash []byte) (*models.IdentityToken, error) {
	var token models.IdentityToken
	err := database.DB(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&token, "purpose = ? AND token_hash = ?", purpose, hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("find token: %w", err)
	}
	return &token, nil
}

func (r *IdentityRepositoryImp) SaveToken(ctx context.Context, token *models.IdentityToken) error {
	if err := database.DB(ctx, r.db).Save(token).Error; err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	return nil
}

func (r *IdentityRepositoryImp) RetireTokens(ctx context.Context, identityID string, purpose models.TokenPurpose, now time.Time) error {
	err := database.DB(ctx, r.db).Model(&models.IdentityToken{}).
		Where("identity_id = ? AND purpose = ? AND used_at IS NULL", identityID, purpose).
		Update("used_at", now).Error
	if err != nil {
		return fmt.Errorf("retire tokens: %w", err)
	}
	return nil
}

func (r *IdentityRepositoryImp) CreateSession(ctx context.Context, session *models.Session) error {
	if err := database.DB(ctx, r.db).Create(session).Error; err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *IdentityRepositoryImp) FindSessionByRefreshToken(ctx context.Context, hash []byte) (*models.Session, error) {
	var session models.Session
	err := database.DB(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&session, "refresh_token_hash = ?", hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	return &session, nil
}

func (r *IdentityRepositoryImp) SaveSession(ctx context.Context, session *models.Session) error {
	if err := database.DB(ctx, r.db).Save(session).Error; err != nil {
		return fmt.Errorf("save session %q: %w", session.ID, err)
	}
	return nil
}

func (r *IdentityRepositoryImp) Publish(ctx context.Context, topic, key, eventID string, event proto.Message) error {
	return kafka.Enqueue(ctx, r.db, topic, key, eventID, event)
}

// translateError maps GORM errors to the auth-service's errors. The email
// index is the only unique key an identity's input can hit.
func translateError(err error) error {
	return database.TranslateError(err, domain.ErrIdentityNotFound, domain.ErrEmailTaken)
}
