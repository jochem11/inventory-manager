package models

import (
	"github.com/jochem11/inventory-manager/shared/database"
)

// User is cleaned up and validated by the validation package: `mod` tags
// normalize input, `validate` tags check it and `json` tags name the fields.
type User struct {
	database.Model
	FirstName string  `json:"firstName" gorm:"size:100;not null" mod:"trim" validate:"required,max=100"`
	LastName  string  `json:"lastName" gorm:"size:100;not null" mod:"trim" validate:"required,max=100"`
	Email     string  `json:"email" gorm:"size:255;not null;uniqueIndex:idx_users_email" mod:"trim,lcase" validate:"required,email,max=255"`
	Phone     *string `json:"phone" gorm:"size:32" mod:"trim" validate:"omitempty,e164"`
	AvatarURL *string `json:"avatarUrl" gorm:"size:2048" mod:"trim" validate:"omitempty,http_url,max=2048"`

	database.SoftDelete
}
