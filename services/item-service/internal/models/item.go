package models

import (
	"github.com/jochem11/inventory-manager/shared/database"
)

// Item is cleaned up and validated by its tags: `mod` tags normalize input,
// `validate` tags check it and `json` tags name the fields.
type Item struct {
	database.Model
	Name        string  `json:"name" gorm:"size:200;not null;index" mod:"trim" validate:"required,max=200"`
	Description *string `json:"description" gorm:"type:text" mod:"trim" validate:"omitempty,max=2000"`
	ImageURL    *string `json:"imageUrl" gorm:"size:2048" mod:"trim" validate:"omitempty,http_url,max=2048"`

	// A category or status in use can't be deleted (RESTRICT); remove or move
	// its items first.
	CategoryID string   `json:"categoryId" gorm:"type:char(27) character set ascii collate ascii_bin;not null;index" validate:"required,len=27,alphanum"`
	Category   Category `json:"category" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" mod:"-" validate:"-"`

	StatusID string     `json:"statusId" gorm:"type:char(27) character set ascii collate ascii_bin;not null;index" validate:"required,len=27,alphanum"`
	Status   ItemStatus `json:"status" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" mod:"-" validate:"-"`

	database.SoftDelete
}
