package db

import (
	"backend/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunMigrations executes all .sql files in db/migrations/ in lexical order.
// Each file is executed as a single SQL batch using the provided GORM DB.
// func RunMigrations(db *gorm.DB, logger *zap.SugaredLogger) error {
// 	logger.Info("Starting database migrations from db/migrations/")

// 	dir := "db/migrations"
// 	entries, err := os.ReadDir(dir)
// 	if err != nil {
// 		logger.Errorf("Failed to read migrations directory: %v", err)
// 		return err
// 	}

// 	var files []string
// 	for _, e := range entries {
// 		if e.IsDir() {
// 			continue
// 		}
// 		name := e.Name()
// 		if strings.HasSuffix(name, ".sql") {
// 			files = append(files, filepath.Join(dir, name))
// 		}
// 	}

// 	sort.Strings(files)

// 	for _, f := range files {
// 		logger.Infof("Applying migration: %s", f)
// 		b, err := os.ReadFile(f)
// 		if err != nil {
// 			logger.Errorf("Failed to read migration file %s: %v", f, err)
// 			return err
// 		}
// 		sql := string(b)
// 		if strings.TrimSpace(sql) == "" {
// 			continue
// 		}
// 		if err := db.Exec(sql).Error; err != nil {
// 			logger.Errorf("Failed to execute migration %s: %v", f, err)
// 			return err
// 		}
// 	}

// 	logger.Info("Migrations completed successfully")
// 	return nil
// }

// AutoMigrate creates all database tables from models using GORM's AutoMigrate.
// This is useful during development; production deployments should prefer file-based migrations.
func AutoMigrate(db *gorm.DB, logger *zap.SugaredLogger) error {
	logger.Info("Running AutoMigrate...")

	// Core models + domain models: companys, users, rbac, audit + domains
	if err := db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserPermissionOverride{},
		&models.AuditLog{},

		// Finance
		&models.BudgetAndLicense{},
		&models.License{},
		&models.Statutory{},
		&models.SalaryAdvance{},

		// Inventory
		&models.InventoryCategory{},
		&models.Inventory{},

		// Merchants
		&models.Merchant{},
		&models.MerchantStatement{},

		// Incidents & Support
		&models.Incident{},
		&models.SupportTicket{},

		// Products & Backups
		&models.ProductCatalog{},
		&models.Backup{},
		&models.SimCard{},

		// Staff + Profile
		&models.Department{},
		&models.StaffListing{},
		&models.Profile{},

		// Notifications & Taxonomy
		&models.Alert{},
		&models.AlertEmail{},
		&models.TaxonomyItem{},
		&models.TaxonomyNotification{},
		&models.UserNotificationPreference{},

		// Reporting + misc
		&models.Report{},
		&models.Setting{},
		&models.RecycleBin{},
		&models.RiskAndCompliance{},
		&models.SystemUpdate{},
	); err != nil {
		logger.Errorf("AutoMigrate failed: %v", err)
		return err
	}

	logger.Info("AutoMigrate completed successfully")
	return nil
}

// SeedDefaultRoles creates system roles and permissions if they don't exist.
func SeedDefaultRoles(db *gorm.DB, companyID string, logger *zap.SugaredLogger) error {
	logger.Infof("Seeding default roles for company: %s", companyID)

	// Default permissions
	permissions := []models.Permission{
		{
			Code:        "users.view",
			CompanyID:   companyID,
			Description: "View users",
			Category:    "users",
		},
		{
			Code:        "users.create",
			CompanyID:   companyID,
			Description: "Create users",
			Category:    "users",
		},
		{
			Code:        "users.update",
			CompanyID:   companyID,
			Description: "Update users",
			Category:    "users",
		},
		{
			Code:        "users.delete",
			CompanyID:   companyID,
			Description: "Delete users",
			Category:    "users",
		},
		{
			Code:        "admin.manage",
			CompanyID:   companyID,
			Description: "Admin access",
			Category:    "admin",
		},
	}

	for _, perm := range permissions {
		if err := db.FirstOrCreate(&perm, models.Permission{Code: perm.Code, CompanyID: companyID}).Error; err != nil {
			logger.Errorf("Failed to seed permission %s: %v", perm.Code, err)
			return err
		}
	}

	logger.Infof("Seeded %d permissions", len(permissions))
	return nil
}
