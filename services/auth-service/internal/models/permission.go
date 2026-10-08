package models

import (
	"time"

	"gorm.io/gorm"
)

// Permission is one thing an identity may do, named "resource:action", e.g.
// "items:write". Services check these names in the access token's claims.
type Permission struct {
	ID   string `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey"`
	Name string `json:"name" gorm:"size:128;not null;uniqueIndex"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = newID()
	}
	return nil
}
