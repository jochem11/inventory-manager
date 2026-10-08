package models

import (
	"time"

	"gorm.io/gorm"
)

// ItemStatus is a state an item can be in, e.g. "New", "Used", "Damaged"
type ItemStatus struct {
	ID   string `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey" validate:"omitempty,len=27,alphanum"`
	Name string `json:"name" gorm:"size:50;not null;uniqueIndex:idx_item_statuses_name" mod:"trim" validate:"required,max=50"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate assigns a KSUID, so IDs are time-ordered and known before insert.
func (s *ItemStatus) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = newID()
	}
	return nil
}
