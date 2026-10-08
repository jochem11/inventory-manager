package models

import (
	"time"

	"gorm.io/gorm"
)

// Item is cleaned up and validated by its tags: `mod` tags normalize input,
// `validate` tags check it and `json` tags name the fields.
type Item struct {
	ID          string  `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey" validate:"omitempty,len=27,alphanum"`
	Name        string  `json:"name" gorm:"size:200;not null;index" mod:"trim" validate:"required,max=200"`
	Description *string `json:"description" gorm:"type:text" mod:"trim" validate:"omitempty,max=2000"`
	ImageURL    *string `json:"imageUrl" gorm:"size:2048" mod:"trim" validate:"omitempty,http_url,max=2048"`

	// A category or status in use can't be deleted (RESTRICT); remove or move
	// its items first.
	CategoryID string   `json:"categoryId" gorm:"type:char(27) character set ascii collate ascii_bin;not null;index" validate:"required,len=27,alphanum"`
	Category   Category `json:"category" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" mod:"-" validate:"-"`

	StatusID string     `json:"statusId" gorm:"type:char(27) character set ascii collate ascii_bin;not null;index" validate:"required,len=27,alphanum"`
	Status   ItemStatus `json:"status" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" mod:"-" validate:"-"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a KSUID, so IDs are time-ordered and known before insert.
func (i *Item) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = newID()
	}
	return nil
}
