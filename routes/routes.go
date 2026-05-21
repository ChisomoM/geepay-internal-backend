package routes

import (
	"backend/middleware"
	"backend/modules/auth"
	"backend/modules/controlhub"
	"backend/modules/rbac"
	"backend/modules/users"
	config "backend/pkg"

	"github.com/labstack/echo/v4"
)

// SetupAuthRoutes registers public authentication routes.
func SetupAuthRoutes(g *echo.Group, handler auth.Handler) {
	g.POST("/login", handler.Login)
	g.POST("/register", handler.Register)
	g.POST("/refresh", handler.RefreshToken)
	g.POST("/forgot-password", handler.RequestPasswordReset)
	g.POST("/reset-password", handler.ResetPassword)
}

// SetupControlHubRoutes registers ControlHub routes (platform-level, not tenant-scoped).
// These routes are separate from tenant routes and should be registered at the root Echo level.
func SetupControlHubRoutes(e *echo.Echo, handler controlhub.Handler, cfg *config.Config) {
	// ControlHub Auth — completely separate from tenant auth
	// Uses /controlhub/auth path (without /api/v1) to distinguish from tenant auth
	auth := e.Group("/controlhub/auth")
	{
		auth.POST("/login", handler.LoginCompanyAdmin)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", handler.Logout)
	}

	// Protected ControlHub routes — require JWT + ControlHubOnly middleware
	controlhubAPI := e.Group("/api/v1/controlhub")
	controlhubAPI.Use(middleware.JWTMiddleware(cfg))
	controlhubAPI.Use(middleware.ControlHubOnly())
	// Additional protected ControlHub routes can be registered here (tenants, clients, etc.)
}

// SetupUserRoutes registers user management routes (protected).
func SetupUserRoutes(g *echo.Group, handler users.Handler) {
	g.POST("/users", handler.Create)
	g.GET("/users", handler.List)
	g.GET("/users/:id", handler.Get)
	g.PUT("/users/:id", handler.Update)
	g.DELETE("/users/:id", handler.Delete)
}

// SetupRBACRoutes registers RBAC management routes (protected, admin-only).
func SetupRBACRoutes(g *echo.Group, handler rbac.Handler) {
	g.GET("/permissions", handler.ListPermissions)
	g.GET("/roles", handler.ListRoles)
	g.POST("/roles", handler.CreateRole)
}
