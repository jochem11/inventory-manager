package models

import (
	"time"

	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

// User is cleaned up and validated by the validation package: `mod` tags
// normalize input, `validate` tags check it and `json` tags name the fields.
type User struct {
	// ID is a KSUID: 27 base62 chars that sort by creation time. ascii_bin keeps
	// comparisons case-sensitive so ordering and uniqueness match the KSUID spec.
	ID        string  `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey" validate:"omitempty,len=27,alphanum"`
	FirstName string  `json:"firstName" gorm:"size:100;not null" mod:"trim" validate:"required,max=100"`
	LastName  string  `json:"lastName" gorm:"size:100;not null" mod:"trim" validate:"required,max=100"`
	Email     string  `json:"email" gorm:"size:255;not null;uniqueIndex:idx_users_email" mod:"trim,lcase" validate:"required,email,max=255"`
	Phone     *string `json:"phone" gorm:"size:32" mod:"trim" validate:"omitempty,e164"`
	AvatarURL *string `json:"avatarUrl" gorm:"size:2048" mod:"trim" validate:"omitempty,http_url,max=2048"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a KSUID, so IDs are time-ordered and known before insert.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = ksuid.New().String()
	}
	return nil
}
