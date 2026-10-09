package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jochem11/inventory-manager/services/auth-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/models"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/token"
	authpb "github.com/jochem11/inventory-manager/services/auth-service/pkg/pb/auth"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/kafka"
	"github.com/jochem11/inventory-manager/shared/validation"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// An activation link works this long.
	verificationTTL = 24 * time.Hour
	// A login lasts this long; refreshing doesn't extend it.
	sessionTTL = 30 * 24 * time.Hour
)

// dummyHash is checked when a login's email is unknown, so a login takes
// equally long whether or not the account exists.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not a real password"), bcrypt.DefaultCost)

type AuthServiceImp struct {
	identities domain.IdentityRepository
	// tx makes several repository calls atomic.
	tx     database.Transactor
	tokens *token.Issuer
	// verifyURL is the web page that reads ?token= from the activation link
	// and calls VerifyEmail.
	verifyURL string
	now       func() time.Time
}

func NewAuthService(identities domain.IdentityRepository, tx database.Transactor, tokens *token.Issuer, verifyURL string) domain.AuthService {
	return &AuthServiceImp{identities: identities, tx: tx, tokens: tokens, verifyURL: verifyURL, now: time.Now}
}

func (s *AuthServiceImp) Register(ctx context.Context, input domain.RegisterInput) (string, error) {
	// Trims and lower-cases the email, then checks the input's tags.
	if err := validation.Clean(ctx, &input); err != nil {
		return "", err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	token, tokenHash, err := newToken()
	if err != nil {
		return "", err
	}

	identity := &models.Identity{
		// The user's id everywhere: the user-service creates the profile under it.
		UserID:       ksuid.New().String(),
		Email:        input.Email,
		PasswordHash: string(passwordHash),
	}
	err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.identities.CreateIdentity(ctx, identity); err != nil {
			return err
		}
		if err := s.identities.AssignRole(ctx, identity, domain.DefaultRole); err != nil {
			return err
		}
		if err := s.identities.CreateToken(ctx, s.verificationToken(identity, tokenHash)); err != nil {
			return err
		}
		eventID := kafka.NewEventID()
		return s.identities.Publish(ctx, kafka.TopicAuthEvents, identity.UserID, eventID, &authpb.AuthEvent{
			EventId:    eventID,
			OccurredAt: timestamppb.New(s.now()),
			Payload: &authpb.AuthEvent_IdentityRegistered{IdentityRegistered: &authpb.IdentityRegistered{
				UserId:    identity.UserID,
				Email:     identity.Email,
				FirstName: input.FirstName,
				LastName:  input.LastName,
				Phone:     input.Phone,
			}},
		})
	})
	if err != nil {
		return "", err
	}

	s.sendVerification(ctx, identity.Email, token)
	return identity.UserID, nil
}

func (s *AuthServiceImp) VerifyEmail(ctx context.Context, token string) error {
	return s.tx.Transaction(ctx, func(ctx context.Context) error {
		now := s.now()
		stored, err := s.identities.FindToken(ctx, models.PurposeVerifyEmail, hashToken(token))
		if err != nil {
			return err
		}
		if !stored.IsUsable(now) {
			return domain.ErrInvalidToken
		}
		identity, err := s.identities.FindIdentityByID(ctx, stored.IdentityID)
		if errors.Is(err, domain.ErrIdentityNotFound) {
			return domain.ErrInvalidToken
		}
		if err != nil {
			return err
		}
		// A link sent to an earlier email address doesn't verify the new one.
		if identity.Email != stored.Email {
			return domain.ErrInvalidToken
		}

		stored.UsedAt = &now
		if err := s.identities.SaveToken(ctx, stored); err != nil {
			return err
		}
		if identity.IsVerified() {
			return nil
		}
		identity.EmailVerifiedAt = &now
		return s.identities.SaveIdentity(ctx, identity)
	})
}

func (s *AuthServiceImp) ResendVerification(ctx context.Context, email string) error {
	identity, err := s.identities.FindIdentityByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, domain.ErrIdentityNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if identity.IsVerified() {
		return nil
	}

	token, tokenHash, err := newToken()
	if err != nil {
		return err
	}
	err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.identities.RetireTokens(ctx, identity.ID, models.PurposeVerifyEmail, s.now()); err != nil {
			return err
		}
		return s.identities.CreateToken(ctx, s.verificationToken(identity, tokenHash))
	})
	if err != nil {
		return err
	}
	s.sendVerification(ctx, identity.Email, token)
	return nil
}

func (s *AuthServiceImp) Login(ctx context.Context, email, password string) (*domain.Tokens, error) {
	identity, err := s.identities.FindIdentityByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, domain.ErrIdentityNotFound) {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(identity.PasswordHash), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}
	// Checked after the password, so it only tells the account's owner.
	if !identity.IsVerified() {
		return nil, domain.ErrEmailNotVerified
	}

	refreshToken, refreshHash, err := newToken()
	if err != nil {
		return nil, err
	}
	session := &models.Session{
		IdentityID:       identity.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        s.now().Add(sessionTTL),
	}
	var tokens *domain.Tokens
	err = s.tx.Transaction(ctx, func(ctx context.Context) error {
		if err := s.identities.CreateSession(ctx, session); err != nil {
			return err
		}
		tokens, err = s.issue(ctx, session, refreshToken)
		return err
	})
	return tokens, err
}

func (s *AuthServiceImp) Refresh(ctx context.Context, refreshToken string) (*domain.Tokens, error) {
	var tokens *domain.Tokens
	err := s.tx.Transaction(ctx, func(ctx context.Context) error {
		session, err := s.identities.FindSessionByRefreshToken(ctx, hashToken(refreshToken))
		if err != nil {
			return err
		}
		if !session.IsActive(s.now()) {
			return domain.ErrInvalidToken
		}

		// Rotate: the token just used stops working.
		next, nextHash, err := newToken()
		if err != nil {
			return err
		}
		session.RefreshTokenHash = nextHash
		if err := s.identities.SaveSession(ctx, session); err != nil {
			return err
		}
		tokens, err = s.issue(ctx, session, next)
		return err
	})
	return tokens, err
}

func (s *AuthServiceImp) Logout(ctx context.Context, refreshToken string) error {
	err := s.tx.Transaction(ctx, func(ctx context.Context) error {
		session, err := s.identities.FindSessionByRefreshToken(ctx, hashToken(refreshToken))
		if err != nil {
			return err
		}
		if session.RevokedAt != nil {
			return nil
		}
		now := s.now()
		session.RevokedAt = &now
		return s.identities.SaveSession(ctx, session)
	})
	if errors.Is(err, domain.ErrInvalidToken) {
		return nil
	}
	return err
}

func (s *AuthServiceImp) PublicKeys() ([]domain.PublicKey, error) {
	keyID, der, err := s.tokens.PublicKey()
	if err != nil {
		return nil, err
	}
	return []domain.PublicKey{{KeyID: keyID, Algorithm: token.Algorithm, DER: der}}, nil
}

// issue signs an access token for session with the identity's current roles,
// so a role change takes effect at the next refresh.
func (s *AuthServiceImp) issue(ctx context.Context, session *models.Session, refreshToken string) (*domain.Tokens, error) {
	identity, err := s.identities.FindIdentityWithRoles(ctx, session.IdentityID)
	if errors.Is(err, domain.ErrIdentityNotFound) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	if !identity.IsVerified() {
		return nil, domain.ErrInvalidToken
	}

	roles, permissions := rolesAndPermissions(identity)
	access, expiresAt, err := s.tokens.Issue(identity.UserID, session.ID, identity.Email, roles, permissions)
	if err != nil {
		return nil, err
	}
	return &domain.Tokens{
		UserID:                identity.UserID,
		AccessToken:           access,
		AccessTokenExpiresAt:  expiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: session.ExpiresAt,
	}, nil
}

// rolesAndPermissions lists the identity's role names and the union of their
// permissions, both sorted.
func rolesAndPermissions(identity *models.Identity) (roles, permissions []string) {
	seen := map[string]bool{}
	for _, role := range identity.Roles {
		roles = append(roles, role.Name)
		for _, p := range role.Permissions {
			if !seen[p.Name] {
				seen[p.Name] = true
				permissions = append(permissions, p.Name)
			}
		}
	}
	slices.Sort(roles)
	slices.Sort(permissions)
	return roles, permissions
}

func (s *AuthServiceImp) verificationToken(identity *models.Identity, tokenHash []byte) *models.IdentityToken {
	return &models.IdentityToken{
		IdentityID: identity.ID,
		Purpose:    models.PurposeVerifyEmail,
		Email:      identity.Email,
		TokenHash:  tokenHash,
		ExpiresAt:  s.now().Add(verificationTTL),
	}
}

// sendVerification logs the activation link. TODO: email it (SMTP, with
// Mailpit in development) instead.
func (s *AuthServiceImp) sendVerification(ctx context.Context, email, token string) {
	link := s.verifyURL + "?token=" + url.QueryEscape(token)
	// The link goes in the message, not an attribute: attribute values with
	// "=" or "?" get quoted, and a clicked link would include the quote.
	slog.InfoContext(ctx, "activation link for "+email+" (email sending isn't built yet): "+link)
}

// newToken returns a random token for an email link and the hash to store.
func newToken() (token string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

// hashToken is what the database stores and looks up. Tokens are long and
// random, so a plain SHA-256 is enough; passwords need bcrypt because people
// pick guessable ones.
func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
