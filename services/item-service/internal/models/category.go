package models

import (
	"time"

	"gorm.io/gorm"
)

// Category is cleaned up and validated by its tags: `mod` tags normalize
// input, `validate` tags check it and `json` tags name the fields.
type Category struct {
	ID   string `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey" validate:"omitempty,len=27,alphanum"`
	Name string `json:"name" gorm:"size:100;not null;uniqueIndex:idx_categories_name" mod:"trim" validate:"required,max=100"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a KSUID, so IDs are time-ordered and known before insert.
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = newID()
	}
	return nil
}
