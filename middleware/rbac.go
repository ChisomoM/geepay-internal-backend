package middleware

import (
	"backend/models"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RBACMiddleware checks if the current user has one of the required permissions.
// Use as a route middleware: e.g., api.POST("/users", handler.Create, RBACMiddleware(db, logger, "users.create", "admin.manage"))
func RBACMiddleware(db *gorm.DB, logger *zap.SugaredLogger, requiredCodes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(string)
			if !ok || userID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No user in context"})
			}

			roleSlug, ok := c.Get("role_slug").(string)
			if !ok || roleSlug == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No role in context"})
			}

			// Super admin always has permission
			if roleSlug == "super_admin" {
				return next(c)
			}

			// Load user with role and permissions
			companyDB := c.Get("db").(*gorm.DB)
			var user models.User
			if err := companyDB.
				Preload("PermissionOverrides.Permission").
				First(&user, "id = ?", userID).Error; err != nil {
				logger.Warnf("Failed to load user: %v", err)
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
			}

			// Load role with permissions
			var role models.Role
			if err := companyDB.
				Preload("Permissions").
				First(&role, "slug = ?", roleSlug).Error; err != nil {
				logger.Warnf("Failed to load role: %v", err)
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
			}

			// Build permission map from role
			permMap := make(map[string]bool)
			for _, p := range role.Permissions {
				permMap[p.Code] = true
			}

			// Apply permission overrides
			for _, override := range user.PermissionOverrides {
				if override.Permission != nil {
					if override.Granted {
						permMap[override.Permission.Code] = true
					} else {
						delete(permMap, override.Permission.Code)
					}
				}
			}

			// Check if user has any required permission
			hasPerm := false
			for _, code := range requiredCodes {
				if permMap[code] {
					hasPerm = true
					break
				}
			}

			if !hasPerm {
				logger.Warnf("User %s denied access (role=%s, required=%v)", userID, roleSlug, requiredCodes)
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
			}

			return next(c)
		}
	}
}
