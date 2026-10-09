package domain

import (
	"context"
	"time"

	"github.com/jochem11/inventory-manager/services/auth-service/internal/models"
	"google.golang.org/protobuf/proto"
)

// IdentityRepository persists identities, their email tokens and the events
// about them. The implementation lives in the repository package. Calls made
// with the ctx of a database.Transaction run in that transaction.
type IdentityRepository interface {
	// CreateIdentity returns ErrEmailTaken for an email that is registered.
	CreateIdentity(ctx context.Context, identity *models.Identity) error
	FindIdentityByID(ctx context.Context, id string) (*models.Identity, error)
	FindIdentityByEmail(ctx context.Context, email string) (*models.Identity, error)
	SaveIdentity(ctx context.Context, identity *models.Identity) error
	// FindIdentityWithRoles loads the identity with its roles and their
	// permissions, which go into its access tokens.
	FindIdentityWithRoles(ctx context.Context, id string) (*models.Identity, error)
	// DeleteIdentityByUserID removes the identity with its sessions, tokens
	// and role assignments.
	DeleteIdentityByUserID(ctx context.Context, userID string) error
	// AssignRole gives the identity the role with this name.
	AssignRole(ctx context.Context, identity *models.Identity, roleName string) error

	CreateToken(ctx context.Context, token *models.IdentityToken) error
	// FindToken locks the token until the transaction ends, so two requests
	// can't redeem it at once. It returns ErrInvalidToken when there is none.
	FindToken(ctx context.Context, purpose models.TokenPurpose, hash []byte) (*models.IdentityToken, error)
	SaveToken(ctx context.Context, token *models.IdentityToken) error
	// RetireTokens marks the identity's unused tokens for purpose as used, so
	// only the newest email link works.
	RetireTokens(ctx context.Context, identityID string, purpose models.TokenPurpose, now time.Time) error

	CreateSession(ctx context.Context, session *models.Session) error
	// FindSessionByRefreshToken locks the session until the transaction
	// ends, so a refresh token can't be rotated twice at once. It returns
	// ErrInvalidToken when there is none.
	FindSessionByRefreshToken(ctx context.Context, hash []byte) (*models.Session, error)
	SaveSession(ctx context.Context, session *models.Session) error

	// Publish adds an event to the outbox; it is sent to Kafka once the
	// surrounding transaction commits.
	Publish(ctx context.Context, topic, key, eventID string, event proto.Message) error
}
