package models

import (
	"github.com/jochem11/inventory-manager/shared/database"
)

// Permission is one thing an identity may do, named "resource:action", e.g.
// "items:write". Services check these names in the access token's claims.
type Permission struct {
	database.Model
	Name string `json:"name" gorm:"size:128;not null;uniqueIndex"`

	database.SoftDelete
}
