package models

import (
	"github.com/jochem11/inventory-manager/shared/database"
	"time"
)

// Identity is a login: the credentials and roles of one user. The profile
// (name, phone, …) lives in the user-service under the same UserID.
type Identity struct {
	database.Model
	UserID          string     `json:"user_id" gorm:"type:char(27) character set ascii collate ascii_bin;not null;uniqueIndex"`
	Email           string     `json:"email" gorm:"size:255;not null;uniqueIndex"`
	PasswordHash    string     `json:"-" gorm:"size:255;not null"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`

	Roles    []Role          `json:"roles" gorm:"many2many:identity_roles"`
	Sessions []Session       `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	Tokens   []IdentityToken `json:"-" gorm:"constraint:OnDelete:CASCADE"`

	database.SoftDelete
}

// IsVerified reports whether the email address has been confirmed.
func (i *Identity) IsVerified() bool {
	return i.EmailVerifiedAt != nil
}
