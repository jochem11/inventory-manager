package models

import (
	"time"

	"gorm.io/gorm"
)

// Role groups permissions, e.g. "admin" or "warehouse". An identity can have
// several roles; its permissions are the union of theirs.
type Role struct {
	ID          string       `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey"`
	Name        string       `json:"name" gorm:"size:64;not null;uniqueIndex"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = newID()
	}
	return nil
}
