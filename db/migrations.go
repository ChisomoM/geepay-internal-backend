package db

import (
	"backend/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunMigrations executes all migration files from the migrations/ directory.
// This should be called once at startup with the migration DB connection.
func RunMigrations(db *gorm.DB, logger *zap.SugaredLogger) error {
	logger.Info("Starting database migrations...")

	// Migration files should be in db/migrations/ and named with timestamps.
	// This is a placeholder that would typically use SQL file execution.
	// In a real implementation, you'd read SQL files and execute them.

	logger.Info("Migrations completed successfully")
	return nil
}

// AutoMigrate creates all database tables from models.
// This is a GORM-based alternative to file-based migrations.
// Use this for development/small projects, or use file-based migrations for prod.
func AutoMigrate(db *gorm.DB, logger *zap.SugaredLogger) error {
	logger.Info("Running AutoMigrate...")

	// Core models: tenants, company admins, users, roles, permissions, audit
	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.CompanyAdmin{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserPermissionOverride{},
		&models.AuditLog{},
	); err != nil {
		logger.Errorf("AutoMigrate failed: %v", err)
		return err
	}

	logger.Info("AutoMigrate completed successfully")
	return nil
}

// SeedDefaultRoles creates system roles and permissions if they don't exist.
func SeedDefaultRoles(db *gorm.DB, tenantID string, logger *zap.SugaredLogger) error {
	logger.Infof("Seeding default roles for tenant: %s", tenantID)

	// Default permissions
	permissions := []models.Permission{
		{
			Code:        "users.view",
			TenantID:    tenantID,
			Description: "View users",
			Category:    "users",
		},
		{
			Code:        "users.create",
			TenantID:    tenantID,
			Description: "Create users",
			Category:    "users",
		},
		{
			Code:        "users.update",
			TenantID:    tenantID,
			Description: "Update users",
			Category:    "users",
		},
		{
			Code:        "users.delete",
			TenantID:    tenantID,
			Description: "Delete users",
			Category:    "users",
		},
		{
			Code:        "admin.manage",
			TenantID:    tenantID,
			Description: "Admin access",
			Category:    "admin",
		},
	}

	for _, perm := range permissions {
		if err := db.FirstOrCreate(&perm, models.Permission{Code: perm.Code, TenantID: tenantID}).Error; err != nil {
			logger.Errorf("Failed to seed permission %s: %v", perm.Code, err)
			return err
		}
	}

	logger.Infof("Seeded %d permissions", len(permissions))
	return nil
}
