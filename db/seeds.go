package db

import (
	"errors"

	"backend/models"
	config "backend/pkg"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedInitialData ensures a default company, a company super-admin, and a platform company admin exist.
func SeedInitialData(db *gorm.DB, cfg *config.Config, logger *zap.SugaredLogger) error {
	logger.Info("=== SeedInitialData: start ===")
	logger.Infof("  SuperAdminEmail:    %q", cfg.SuperAdminEmail)
	logger.Infof("  SuperAdminPassword: %q", cfg.SuperAdminPassword)
	logger.Infof("  DefaultCompanyID:    %q", cfg.DefaultCompanyID)

	var company models.Company

	// Determine default company: prefer configured DefaultCompanyID, otherwise use slug "default".
	if cfg.DefaultCompanyID != "" {
		logger.Infof("Step 1: looking up company by id=%q", cfg.DefaultCompanyID)
		if err := db.First(&company, "id = ?", cfg.DefaultCompanyID).Error; err != nil {
			// Treat any error (not found OR invalid UUID) as "doesn't exist" and fall through to slug lookup.
			logger.Infof("  company id lookup failed (%v), falling back to slug lookup", err)
		} else {
			logger.Infof("  found company by id: %s", company.ID.String())
		}
	}

	if company.ID == (models.BaseModel{}).ID {
		logger.Info("Step 2: company not resolved by id, looking up by slug=\"default\"")
		if err := db.First(&company, "slug = ?", "default").Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Info("  no company with slug=\"default\", creating one")
				company = models.Company{
					Name: "Default Company",
					Slug: "default",
				}
				if err := db.Create(&company).Error; err != nil {
					logger.Errorf("  failed to create default company: %v", err)
					return err
				}
				logger.Infof("  created default company id=%s", company.ID.String())
			} else {
				logger.Errorf("  slug lookup error: %v", err)
				return err
			}
		} else {
			logger.Infof("  found company by slug: id=%s", company.ID.String())
		}
	}

	logger.Infof("Step 3: using company id=%s", company.ID.String())

	// Seed company roles if missing
	if err := SeedDefaultRoles(db, company.ID.String(), logger); err != nil {
		logger.Warnf("Failed to seed default permissions: %v", err)
	}

	// Ensure standard roles exist
	roles := []models.Role{
		{CompanyID: company.ID.String(), Name: "Super Admin", Slug: "super_admin", IsSystem: true},
		{CompanyID: company.ID.String(), Name: "Admin", Slug: "admin", IsSystem: true},
		{CompanyID: company.ID.String(), Name: "User", Slug: "user", IsSystem: true},
	}
	for _, r := range roles {
		var existing models.Role
		if err := db.First(&existing, "company_id = ? AND slug = ?", company.ID.String(), r.Slug).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&r).Error; err != nil {
					logger.Warnf("Failed to create role %s: %v", r.Slug, err)
				}
			} else {
				return err
			}
		}
	}

	// Create a company super-admin user if none exists
	companyAdminEmail := cfg.SuperAdminEmail
	logger.Infof("Step 4: seeding super admin email=%q for company_id=%s", companyAdminEmail, company.ID.String())
	var tu models.User
	if err := db.Unscoped().Where("company_id = ? AND email = ?", company.ID.String(), companyAdminEmail).First(&tu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("  super admin not found, creating")
			pass := cfg.SuperAdminPassword
			hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			tu = models.User{
				CompanyBaseModel: models.CompanyBaseModel{CompanyID: company.ID.String()},
				Email:            companyAdminEmail,
				PasswordHash:     string(hashed),
				FirstName:        "Super",
				LastName:         "Admin",
				RoleSlug:         "super_admin",
				IsActive:         true,
			}
			if err := db.Create(&tu).Error; err != nil {
				logger.Errorf("  failed to create super admin: %v", err)
			} else {
				logger.Infof("  created super admin id=%s email=%s", tu.ID.String(), companyAdminEmail)
			}
		} else {
			logger.Errorf("  super admin lookup error: %v", err)
			return err
		}
	} else {
		logger.Infof("  super admin already exists id=%s deleted_at=%v", tu.ID.String(), tu.DeletedAt)
	}

	logger.Info("=== SeedInitialData: done ===")
	return nil
}
