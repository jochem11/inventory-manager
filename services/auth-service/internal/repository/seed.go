package repository

import (
	"context"
	"fmt"

	"github.com/jochem11/inventory-manager/services/auth-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/auth-service/internal/models"
	"github.com/jochem11/inventory-manager/shared/auth"
	"gorm.io/gorm"
)

// Seed makes the roles and their permissions match auth.RolePermissions:
// missing ones are created, and a role loses permissions no longer listed for
// it. It gives the default role to identities without a role (e.g. registered
// before roles existed), and the admin role to the identities with an email in
// adminEmails.
func Seed(ctx context.Context, db *gorm.DB, adminEmails []string) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for name, permissionNames := range auth.RolePermissions {
			role := models.Role{Name: name}
			if err := tx.Where(models.Role{Name: name}).FirstOrCreate(&role).Error; err != nil {
				return fmt.Errorf("seed role %q: %w", name, err)
			}
			permissions := make([]models.Permission, len(permissionNames))
			for i, permissionName := range permissionNames {
				permissions[i] = models.Permission{Name: permissionName}
				if err := tx.Where(models.Permission{Name: permissionName}).FirstOrCreate(&permissions[i]).Error; err != nil {
					return fmt.Errorf("seed permission %q: %w", permissionName, err)
				}
			}
			if err := tx.Model(&role).Association("Permissions").Replace(permissions); err != nil {
				return fmt.Errorf("set the permissions of %q: %w", name, err)
			}
		}

		var defaultRole models.Role
		if err := tx.First(&defaultRole, "name = ?", domain.DefaultRole).Error; err != nil {
			return fmt.Errorf("find default role: %w", err)
		}
		err := tx.Exec(`INSERT INTO identity_roles (identity_id, role_id)
			SELECT id, ? FROM identities
			WHERE deleted_at IS NULL AND id NOT IN (SELECT identity_id FROM identity_roles)`, defaultRole.ID).Error
		if err != nil {
			return fmt.Errorf("assign default role: %w", err)
		}

		if len(adminEmails) == 0 {
			return nil
		}
		var adminRole models.Role
		if err := tx.First(&adminRole, "name = ?", auth.RoleAdmin).Error; err != nil {
			return fmt.Errorf("find admin role: %w", err)
		}
		err = tx.Exec(`INSERT IGNORE INTO identity_roles (identity_id, role_id)
			SELECT id, ? FROM identities WHERE deleted_at IS NULL AND email IN ?`, adminRole.ID, adminEmails).Error
		if err != nil {
			return fmt.Errorf("assign admin role: %w", err)
		}
		return nil
	})
}
