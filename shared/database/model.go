package database

import (
	"time"

	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

// NewID returns a KSUID: 27 base62 characters that sort by creation time.
// Every table's id is one. Store ids (and the foreign keys that point at
// them) as `char(27) character set ascii collate ascii_bin`, so ordering and
// uniqueness are case-sensitive, as the KSUID spec requires.
func NewID() string {
	return ksuid.New().String()
}

// Model is the id and timestamps every table has. Embed it first:
//
//	type User struct {
//		database.Model
//		Name string
//		database.SoftDelete
//	}
//
// The id is set on create when it's empty, so it's known before the insert.
type Model struct {
	ID        string    `json:"id" gorm:"type:char(27) character set ascii collate ascii_bin;primaryKey" validate:"omitempty,len=27,alphanum"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeCreate is a GORM hook: it assigns a new id when there is none.
func (m *Model) BeforeCreate(*gorm.DB) error {
	if m.ID == "" {
		m.ID = NewID()
	}
	return nil
}

// SoftDelete makes deletes only set deleted_at; GORM then leaves those rows
// out of every query. Embed it in a model that should keep deleted rows.
type SoftDelete struct {
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
