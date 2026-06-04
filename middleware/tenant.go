package middleware

import (
	"backend/models"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CompanyMiddleware resolves the company from the validated JWT claims and attaches
// a WHERE-scoped *gorm.DB to the Echo context for automatic company isolation.
// This middleware must run after JWTMiddleware so that "company_id" is already in context.
//
// When MULTI_COMPANY_ENABLED=false, this is a no-op and handlers work with unscoped DB.
// When true, all handler queries are automatically filtered by company_id.
func CompanyMiddleware(cfg CompanyConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If multi-tenancy is disabled, skip scoping
			if !cfg.MultiCompanyEnabled {
				return next(c)
			}

			// company_id is set in context by JWTMiddleware after validating the token.
			companyID, ok := c.Get("company_id").(string)
			if !ok || companyID == "" {
				cfg.Logger.Warnf("No company_id in JWT context for request: %s", c.Request().URL.Path)
				return echo.NewHTTPError(401, "Authentication required")
			}

			c.Set("companyID", companyID)

			// Verify company exists in database
			var count int64
			if err := cfg.DB.Model(&models.Company{}).
				Where("id = ?", companyID).
				Count(&count).Error; err != nil || count == 0 {
				cfg.Logger.Warnf("Company not found: %s", companyID)
				return echo.NewHTTPError(404, "Company not found")
			}

			// Attach a WHERE-scoped session so every handler query is automatically
			// company-scoped without any per-query effort.
			// Session(NewDB:false) preserves the WHERE clause across Preload().
			scopedDB := cfg.DB.Where("company_id = ?", companyID).Session(&gorm.Session{NewDB: false})
			c.Set("db", scopedDB)

			return next(c)
		}
	}
}

// CompanyConfig holds configuration for CompanyMiddleware.
type CompanyConfig struct {
	DB                  *gorm.DB
	Logger              *zap.SugaredLogger
	MultiCompanyEnabled bool
}
