package models

import (
	"time"

	"gorm.io/gorm"
)

// Session is one login on one device. Access tokens (short-lived JWTs that
// are never stored) carry its ID in the "sid" claim.
//
// The session holds the hash of its current refresh token: every refresh
// looks the session up by that hash and replaces it with a new one, so each
// refresh token works once. Logging out sets RevokedAt, after which the
// refresh token stops working and the device is logged out once its current
// access token expires.
type Session struct {
	ID         string `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey"`
	IdentityID string `json:"identity_id" gorm:"type:char(27) character set ascii collate ascii_bin;not null;index"`
	// RefreshTokenHash is the SHA-256 of the current refresh token; the token
	// itself is only ever in the client's cookie.
	RefreshTokenHash []byte `json:"-" gorm:"type:binary(32);not null;uniqueIndex"`
	// ExpiresAt is the end of the session (e.g. 30 days after login);
	// refreshing doesn't extend it, so the user logs in again after that.
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt *time.Time `json:"revoked_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsActive reports whether the session can still be refreshed at now.
func (s *Session) IsActive(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = newID()
	}
	return nil
}
