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
	miniouploader "backend/pkg/minio"

	"backend/modules/auth"
	"backend/modules/backups"
	"backend/modules/dashboard"
	"backend/modules/departments"
	"backend/modules/finance"
	"backend/modules/incidents"
	"backend/modules/inventory"
	"backend/modules/merchantportal"
	"backend/modules/merchants"
	"backend/modules/products"
	"backend/modules/profile"
	"backend/modules/rbac"
	"backend/modules/recyclebin"
	"backend/modules/riskcompliance"
	"backend/modules/settings"
	"backend/modules/sims"
	"backend/modules/stafflisting"
	"backend/modules/systemupdates"
	"backend/modules/taxonomy"
	"backend/modules/users"
	config "backend/pkg"
	"backend/routes"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"strings"
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
	sugar.Infof("Multi-tenancy: %v", cfg.MultiCompanyEnabled)

	// Initialize database connection.
	if err := db.Initialize(cfg, sugar); err != nil {
		sugar.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run AutoMigrate to create/update all tables from models
	if err := db.AutoMigrate(db.GetDB(), sugar); err != nil {
		sugar.Fatalf("AutoMigrate failed: %v", err)
	}

	// Seed initial data: default company, company admin, and platform company admin
	if err := db.SeedInitialData(db.GetDB(), cfg, sugar); err != nil {
		sugar.Warnf("Failed to seed initial data: %v", err)
	}

	// Initialize infrastructure services
	var emailService global.EmailSender
	var uploadService global.FileUploader

	if cfg.FileStorageType == "minio" && cfg.MinioURL != "" {
		useSSL := strings.HasPrefix(cfg.MinioURL, "https://")
		endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.MinioURL, "https://"), "http://")
		uploader, err := miniouploader.New(endpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioPublicURL, useSSL)
		if err != nil {
			sugar.Warnf("MinIO initialization failed: %v — file upload disabled", err)
		} else {
			uploadService = uploader
			sugar.Info("MinIO upload service initialized")
		}
	}

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

	financeService := finance.NewService(app)
	financeHandler := finance.NewHandler(financeService)

	inventoryService := inventory.NewService(app)
	inventoryHandler := inventory.NewHandler(inventoryService)

	productsService := products.NewService(app)
	productsHandler := products.NewHandler(productsService)

	incidentsService := incidents.NewService(app)
	incidentsHandler := incidents.NewHandler(incidentsService)

	simsService := sims.NewService(app)
	simsHandler := sims.NewHandler(simsService)

	departmentsService := departments.NewService(app)
	departmentsHandler := departments.NewHandler(departmentsService)

	staffService := stafflisting.NewService(app)
	staffHandler := stafflisting.NewHandler(staffService)

	systemUpdatesService := systemupdates.NewService(app)
	systemUpdatesHandler := systemupdates.NewHandler(systemUpdatesService)

	backupsService := backups.NewService(app)
	backupsHandler := backups.NewHandler(backupsService)

	merchantsService := merchants.NewService(app)
	merchantsHandler := merchants.NewHandler(merchantsService)

	taxonomyService := taxonomy.NewService(app)
	taxonomyHandler := taxonomy.NewHandler(taxonomyService)

	settingsService := settings.NewService(app)
	settingsHandler := settings.NewHandler(settingsService)

	profileService := profile.NewService(app)
	profileHandler := profile.NewHandler(profileService)

	riskComplianceService := riskcompliance.NewService(app)
	riskComplianceHandler := riskcompliance.NewHandler(riskComplianceService)

	recycleBinService := recyclebin.NewService(app)
	recycleBinHandler := recyclebin.NewHandler(recycleBinService)

	dashboardService := dashboard.NewService(app)
	dashboardHandler := dashboard.NewHandler(dashboardService)

	merchantPortalService := merchantportal.NewService(app)
	merchantPortalHandler := merchantportal.NewHandler(merchantPortalService)

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

	// Public routes (auth group, no CompanyMiddleware)
	// Auth handlers call c.Get("db"), so inject the global unscoped DB here
	// since CompanyMiddleware (which normally sets "db") is not applied to this group.
	publicAPI := e.Group("/api/v1/auth")
	publicAPI.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db.GetDB())
			return next(c)
		}
	})
	routes.SetupAuthRoutes(publicAPI, authHandler)

	// Merchant portal routes (merchant-facing, separate auth)
	routes.SetupMerchantPortalRoutes(e, merchantPortalHandler, cfg, db.GetDB())

	// Protected routes (with JWT + Company middleware)
	api := e.Group("/api/v1", middleware.JWTMiddleware(cfg))

	// Attach company middleware
	companyCfg := middleware.CompanyConfig{
		DB:                  db.GetDB(),
		Logger:              sugar,
		MultiCompanyEnabled: cfg.MultiCompanyEnabled,
	}
	api.Use(middleware.CompanyMiddleware(companyCfg))

	// Attach audit middleware for state-changing requests
	api.Use(middleware.AuditMiddleware(db.GetDB(), sugar))

	// Register protected routes
	routes.SetupUserRoutes(api, usersHandler)
	routes.SetupRBACRoutes(api, rbacHandler)
	routes.SetupFinanceRoutes(api, financeHandler)
	routes.SetupInventoryRoutes(api, inventoryHandler)
	routes.SetupProductRoutes(api, productsHandler)
	routes.SetupSimRoutes(api, simsHandler)
	routes.SetupIncidentsRoutes(api, incidentsHandler)
	routes.SetupDepartmentRoutes(api, departmentsHandler)
	routes.SetupStaffRoutes(api, staffHandler)
	routes.SetupSystemUpdateRoutes(api, systemUpdatesHandler)
	routes.SetupBackupRoutes(api, backupsHandler)
	routes.SetupMerchantRoutes(api, merchantsHandler)
	routes.SetupTaxonomyRoutes(api, taxonomyHandler)
	routes.SetupSettingsRoutes(api, settingsHandler)
	routes.SetupProfileRoutes(api, profileHandler)
	routes.SetupRiskComplianceRoutes(api, riskComplianceHandler)
	routes.SetupRecycleBinRoutes(api, recycleBinHandler)
	routes.SetupDashboardRoutes(api, dashboardHandler)

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
