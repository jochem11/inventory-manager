package domain

import (
	"context"
	"time"

	"github.com/jochem11/inventory-manager/shared/auth"
)

// DefaultRole is given to every identity at registration.
const DefaultRole = auth.RoleUser

// RegisterInput is what a new user fills in. The profile fields are passed
// on to the user-service in the IdentityRegistered event.
type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     *string
}

// Tokens are what Login and Refresh hand out.
type Tokens struct {
	UserID                string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

// PublicKey verifies access tokens whose JWT header has KeyID.
type PublicKey struct {
	KeyID     string
	Algorithm string
	// DER is the PKIX encoding.
	DER []byte
}

// AuthService holds the login logic. The implementation lives in the service
// package.
//
// Errors: invalid input is a *ValidationError, a registered email wraps
// ErrEmailTaken, a bad activation link or refresh token is ErrInvalidToken,
// a wrong email or password is ErrInvalidCredentials and logging in before
// activating is ErrEmailNotVerified.
type AuthService interface {
	// Register creates the identity and emails an activation link. It
	// returns the new user's id.
	Register(ctx context.Context, input RegisterInput) (string, error)
	// VerifyEmail activates the account of an activation link's token.
	VerifyEmail(ctx context.Context, token string) error
	// ResendVerification emails a new activation link. It returns nil for
	// unknown and already verified emails too, so it reveals nothing.
	ResendVerification(ctx context.Context, email string) error

	// Login starts a session.
	Login(ctx context.Context, email, password string) (*Tokens, error)
	// Refresh exchanges a refresh token for new tokens; the old refresh
	// token stops working.
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)
	// Logout ends the refresh token's session. An unknown token is not an
	// error: the result is the same, the session is gone.
	Logout(ctx context.Context, refreshToken string) error
	PublicKeys() ([]PublicKey, error)
}
