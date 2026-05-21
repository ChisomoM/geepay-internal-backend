package middleware

import (
	"backend/models"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditMiddleware logs state-changing operations (POST, PUT, PATCH, DELETE).
// Attach this to route groups to automatically log all mutations.
func AuditMiddleware(db *gorm.DB, logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Skip audit for non-mutating requests
			if req.Method == "GET" || req.Method == "HEAD" || req.Method == "OPTIONS" {
				return next(c)
			}

			userID, _ := c.Get("user_id").(string)
			tenantID, _ := c.Get("tenant_id").(string)

			if userID == "" || tenantID == "" {
				return next(c)
			}

			// Extract action from request context or route
			action := req.Method
			resource := extractResourceFromPath(req.URL.Path)
			resourceID := extractIDFromPath(req.URL.Path)

			// Log the audit entry
			auditLog := models.AuditLog{
				TenantBaseModel: models.TenantBaseModel{
					TenantID: tenantID,
				},
				UserID:     uuid.MustParse(userID),
				Action:     action,
				Resource:   resource,
				ResourceID: resourceID,
				IPAddress:  c.RealIP(),
				UserAgent:  req.Header.Get("User-Agent"),
				Details: func() string {
					body := map[string]interface{}{
						"path":   req.URL.Path,
						"method": req.Method,
					}
					b, _ := json.Marshal(body)
					return string(b)
				}(),
			}

			if err := db.Create(&auditLog).Error; err != nil {
				logger.Warnf("Failed to log audit: %v", err)
			}

			return next(c)
		}
	}
}

// extractResourceFromPath extracts the resource type from the URL path.
// e.g., "/api/v1/users/123" -> "users"
func extractResourceFromPath(path string) string {
	parts := split(path, "/")
	for i, part := range parts {
		if part == "api" && i+2 < len(parts) {
			return parts[i+2]
		}
	}
	return ""
}

// extractIDFromPath extracts the resource ID from the URL path.
// e.g., "/api/v1/users/123" -> "123"
func extractIDFromPath(path string) string {
	parts := split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// split is a helper for path splitting.
func split(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
