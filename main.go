package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/db"
	"backend/global"
	"backend/middleware"
	"backend/modules/auth"
	"backend/modules/controlhub"
	"backend/modules/rbac"
	"backend/modules/users"
	config "backend/pkg"
	"backend/routes"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		sugar.Fatalf("Failed to load configuration: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		sugar.Fatalf("Configuration validation failed: %v", err)
	}

	sugar.Infof("Starting application in %s environment", cfg.Env)
	sugar.Infof("Multi-tenancy: %v", cfg.MultiTenantEnabled)

	// Initialize database connection.
	if err := db.Initialize(cfg, sugar); err != nil {
		sugar.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run AutoMigrate (alternatively, implement file-based SQL migrations in db/migrations/)
	if err := db.AutoMigrate(db.GetDB(), sugar); err != nil {
		sugar.Fatalf("AutoMigrate failed: %v", err)
	}

	// Seed default roles for the default tenant (if multi-tenancy enabled)
	if cfg.MultiTenantEnabled && cfg.DefaultTenantID != "" {
		if err := db.SeedDefaultRoles(db.GetDB(), cfg.DefaultTenantID, sugar); err != nil {
			sugar.Warnf("Failed to seed default roles: %v", err)
		}
	}

	// Initialize infrastructure services
	// In this template: Email and Upload are optional. Provide nil or implement interfaces as needed.
	var emailService global.EmailSender
	var uploadService global.FileUploader

	// Initialize the shared App infrastructure (passed to all modules)
	app := global.New(cfg, sugar, emailService, uploadService)
	sugar.Info("App infrastructure initialized")

	// Initialize modules
	authService := auth.NewService(app)
	authHandler := auth.NewHandler(authService)

	usersService := users.NewService(app)
	usersHandler := users.NewHandler(usersService)

	rbacService := rbac.NewService(app)
	rbacHandler := rbac.NewHandler(rbacService)

	controlhubService := controlhub.NewService(app)
	controlhubHandler := controlhub.NewHandler(controlhubService)

	// Create Echo router
	e := echo.New()

	// Global middleware
	e.Use(middleware.LoggerMiddleware(sugar))
	e.Use(middleware.CORSMiddleware(middleware.CORSConfig{
		AllowedOrigins: parseCORSOrigins(cfg.CORSAllowedOrigins),
		Logger:         sugar,
	}))

	// Health checks (no auth required)
	e.GET("/api/v1/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Public routes (auth group, no TenantMiddleware)
	publicAPI := e.Group("/api/v1/auth")
	routes.SetupAuthRoutes(publicAPI, authHandler)

	// ControlHub routes (platform-level, separate from tenant auth)
	routes.SetupControlHubRoutes(e, controlhubHandler, cfg)

	// Protected routes (with JWT + Tenant middleware)
	api := e.Group("/api/v1", middleware.JWTMiddleware(cfg))

	// Attach tenant middleware
	tenantCfg := middleware.TenantConfig{
		DB:                 db.GetDB(),
		Logger:             sugar,
		MultiTenantEnabled: cfg.MultiTenantEnabled,
	}
	api.Use(middleware.TenantMiddleware(tenantCfg))

	// Attach audit middleware for state-changing requests
	api.Use(middleware.AuditMiddleware(db.GetDB(), sugar))

	// Register protected routes
	routes.SetupUserRoutes(api, usersHandler)
	routes.SetupRBACRoutes(api, rbacHandler)

	// Start server with graceful shutdown
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err.Error() != "http: Server closed" {
			sugar.Errorf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		sugar.Errorf("Shutdown error: %v", err)
	}

	sugar.Info("Server stopped")
}

// parseCORSOrigins parses comma-separated CORS origins.
func parseCORSOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{"http://localhost:3000", "http://localhost:5173"}
	}
	var origins []string
	for _, o := range splitString(originsStr, ",") {
		origins = append(origins, o)
	}
	return origins
}

// splitString splits a string by delimiter and trims whitespace.
func splitString(s, sep string) []string {
	var result []string
	for _, part := range splitBy(s, sep) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// splitBy is a helper for string splitting.
func splitBy(s, sep string) []string {
	var result []string
	var current string
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}
