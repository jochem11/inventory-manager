package models

import (
	"time"

	"gorm.io/gorm"
)

// TokenPurpose says what an IdentityToken may be used for.
type TokenPurpose string

// More purposes (e.g. password reset) can reuse this table later.
const PurposeVerifyEmail TokenPurpose = "verify_email"

// IdentityToken is a one-time token sent by email to activate an account.
// The link carries a random 32-byte token; only its SHA-256 is stored, so a
// leaked database can't be used to take over accounts.
//
// A token works once (UsedAt), until ExpiresAt, and only for Email: after an
// email change, links sent to the old address stop working.
type IdentityToken struct {
	ID         string       `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey"`
	IdentityID string       `json:"identity_id" gorm:"type:char(27) character set ascii collate ascii_bin;not null;index"`
	Purpose    TokenPurpose `json:"purpose" gorm:"size:32;not null"`
	Email      string       `json:"email" gorm:"size:255;not null"`
	TokenHash  []byte       `json:"-" gorm:"type:binary(32);not null;uniqueIndex"`
	ExpiresAt  time.Time    `json:"expires_at" gorm:"not null"`
	// UsedAt is set when the token is redeemed, or when a newer token of the
	// same purpose replaces it.
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at" gorm:"index"`
}

// IsUsable reports whether the token can still be redeemed at now.
func (t *IdentityToken) IsUsable(now time.Time) bool {
	return t.UsedAt == nil && now.Before(t.ExpiresAt)
}

func (t *IdentityToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = newID()
	}
	return nil
}
