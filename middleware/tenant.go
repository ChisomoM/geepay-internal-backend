package middleware

import (
	"backend/models"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TenantMiddleware resolves the tenant from the validated JWT claims and attaches
// a WHERE-scoped *gorm.DB to the Echo context for automatic tenant isolation.
// This middleware must run after JWTMiddleware so that "tenant_id" is already in context.
//
// When MULTI_TENANT_ENABLED=false, this is a no-op and handlers work with unscoped DB.
// When true, all handler queries are automatically filtered by tenant_id.
func TenantMiddleware(cfg TenantConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If multi-tenancy is disabled, skip scoping
			if !cfg.MultiTenantEnabled {
				return next(c)
			}

			// tenant_id is set in context by JWTMiddleware after validating the token.
			tenantID, ok := c.Get("tenant_id").(string)
			if !ok || tenantID == "" {
				cfg.Logger.Warnf("No tenant_id in JWT context for request: %s", c.Request().URL.Path)
				return echo.NewHTTPError(401, "Authentication required")
			}

			c.Set("tenantID", tenantID)

			// Verify tenant exists in database
			var count int64
			if err := cfg.DB.Model(&models.Tenant{}).
				Where("id = ?", tenantID).
				Count(&count).Error; err != nil || count == 0 {
				cfg.Logger.Warnf("Tenant not found: %s", tenantID)
				return echo.NewHTTPError(404, "Tenant not found")
			}

			// Attach a WHERE-scoped session so every handler query is automatically
			// tenant-scoped without any per-query effort.
			// Session(NewDB:false) preserves the WHERE clause across Preload().
			scopedDB := cfg.DB.Where("tenant_id = ?", tenantID).Session(&gorm.Session{NewDB: false})
			c.Set("db", scopedDB)

			return next(c)
		}
	}
}

// TenantConfig holds configuration for TenantMiddleware.
type TenantConfig struct {
	DB                 *gorm.DB
	Logger             *zap.SugaredLogger
	MultiTenantEnabled bool
}
