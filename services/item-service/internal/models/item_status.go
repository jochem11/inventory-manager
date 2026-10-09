package models

import (
	"github.com/jochem11/inventory-manager/shared/database"
)

// ItemStatus is a state an item can be in, e.g. "New", "Used", "Damaged"
type ItemStatus struct {
	database.Model
	Name string `json:"name" gorm:"size:50;not null;uniqueIndex:idx_item_statuses_name" mod:"trim" validate:"required,max=50"`

	database.SoftDelete
}
