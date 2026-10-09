package models

import (
	"github.com/jochem11/inventory-manager/shared/database"
)

// Role groups permissions, e.g. "admin" or "warehouse". An identity can have
// several roles; its permissions are the union of theirs.
type Role struct {
	database.Model
	Name        string       `json:"name" gorm:"size:64;not null;uniqueIndex"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions"`

	database.SoftDelete
}
