package models

import (
	"github.com/jochem11/inventory-manager/shared/database"
)

// Category is cleaned up and validated by its tags: `mod` tags normalize
// input, `validate` tags check it and `json` tags name the fields.
type Category struct {
	database.Model
	Name string `json:"name" gorm:"size:100;not null;uniqueIndex:idx_categories_name" mod:"trim" validate:"required,max=100"`

	database.SoftDelete
}
