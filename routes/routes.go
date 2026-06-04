package routes

import (
	"backend/middleware"
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

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// SetupAuthRoutes registers public authentication routes.
func SetupAuthRoutes(g *echo.Group, handler auth.Handler) {
	g.POST("/login", handler.Login)
	g.POST("/register", handler.Register)
	g.POST("/refresh", handler.RefreshToken)
	g.POST("/forgot-password", handler.RequestPasswordReset)
	g.POST("/reset-password", handler.ResetPassword)
}

// SetupMerchantPortalRoutes registers merchant portal routes.
// Auth routes inject the global DB; protected routes use MerchantJWTMiddleware.
func SetupMerchantPortalRoutes(e *echo.Echo, handler merchantportal.Handler, cfg *config.Config, globalDB *gorm.DB) {
	dbInjector := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", globalDB)
			return next(c)
		}
	}

	// Public: merchant login
	merchantAuth := e.Group("/merchant/auth")
	merchantAuth.Use(dbInjector)
	merchantAuth.POST("/login", handler.Login)

	// Protected: merchant ticket operations
	merchantAPI := e.Group("/merchant/v1")
	merchantAPI.Use(middleware.MerchantJWTMiddleware(cfg))
	merchantAPI.GET("/tickets", handler.ListTickets)
	merchantAPI.POST("/tickets", handler.CreateTicket)
	merchantAPI.GET("/tickets/:id", handler.GetTicket)
	merchantAPI.PATCH("/tickets/:id/status", handler.UpdateTicketStatus)
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

// SetupFinanceRoutes registers finance-related endpoints.
func SetupFinanceRoutes(g *echo.Group, handler finance.Handler) {
	f := g.Group("/finance")

	// Budgets
	f.GET("/budget", handler.ListBudgets)
	f.POST("/budget", handler.CreateBudget)
	f.GET("/budget/:id", handler.GetBudget)
	f.PUT("/budget/:id", handler.UpdateBudget)
	f.DELETE("/budget/:id", handler.DeleteBudget)
	f.PATCH("/budget/:id/renew", handler.RenewBudget)
	f.GET("/budget/export-csv", handler.ExportBudgetsCSV)

	// Licenses
	f.GET("/licenses", handler.ListLicenses)
	f.POST("/licenses", handler.CreateLicense)
	f.GET("/licenses/:id", handler.GetLicense)
	f.PUT("/licenses/:id", handler.UpdateLicense)
	f.DELETE("/licenses/:id", handler.DeleteLicense)
	f.PATCH("/licenses/:id/renew", handler.RenewLicense)
	f.GET("/licenses/export-csv", handler.ExportLicensesCSV)

	// Statutories
	f.GET("/statutories", handler.ListStatutories)
	f.POST("/statutories", handler.CreateStatutory)
	f.GET("/statutories/:id", handler.GetStatutory)
	f.PUT("/statutories/:id", handler.UpdateStatutory)
	f.DELETE("/statutories/:id", handler.DeleteStatutory)
	f.PATCH("/statutories/:id/mark-paid", handler.MarkStatutoryPaid)

	// Salary advances
	f.GET("/salary-advance", handler.ListSalaryAdvances)
	f.POST("/salary-advance", handler.CreateSalaryAdvance)
	f.GET("/salary-advance/:id", handler.GetSalaryAdvance)
	f.PUT("/salary-advance/:id", handler.UpdateSalaryAdvance)
	f.DELETE("/salary-advance/:id", handler.DeleteSalaryAdvance)
	f.PATCH("/salary-advance/:id/status", handler.ApproveSalaryAdvance)
	f.POST("/salary-advance/:id/deduct", handler.DeductSalaryAdvance)
	f.GET("/salary-advance/export-csv", handler.ExportSalaryAdvanceCSV)
}

// SetupInventoryRoutes registers inventory endpoints.
func SetupInventoryRoutes(g *echo.Group, handler inventory.Handler) {
	inv := g.Group("/inventory")

	// Public/company inventory listings
	inv.GET("", handler.ListItems)
	inv.GET("/category/:categoryId/items", handler.ListItemsByCategory)
	inv.GET("/item/:itemId", handler.GetItem)

	// Item operations
	inv.POST("/category/:categoryId/item", handler.CreateItem)
	inv.PUT("/item/:itemId", handler.UpdateItem)
	inv.DELETE("/item/:itemId", handler.DeleteItem)
	inv.POST("/item/:itemId/assign", handler.AssignItem)
	inv.POST("/item/:itemId/unassign", handler.UnassignItem)

	// Categories (admin)
	inv.GET("/category", handler.ListCategories)
	inv.POST("/category", handler.CreateCategory)
	inv.GET("/category/:id", handler.GetCategory)
	inv.PUT("/category/:id", handler.UpdateCategory)
	inv.DELETE("/category/:id", handler.DeleteCategory)
}

// SetupIncidentsRoutes registers incidents and support ticket endpoints.
func SetupIncidentsRoutes(g *echo.Group, handler incidents.Handler) {
	it := g.Group("/incidents")

	// Incidents
	it.GET("", handler.ListIncidents)
	it.POST("", handler.CreateIncident)
	it.GET("/:id", handler.GetIncident)
	it.PUT("/:id", handler.UpdateIncident)
	it.POST("/:id/status", handler.ChangeIncidentStatus)
	it.POST("/:id/assign", handler.AssignIncident)
	it.POST("/:id/notify", handler.NotifyIncident)

	// Support tickets under /api/v1/tickets
	t := g.Group("/tickets")
	t.GET("", handler.ListTickets)
	t.POST("", handler.CreateTicket)
	t.GET("/:id", handler.GetTicket)
	t.PUT("/:id", handler.UpdateTicket)
	t.POST("/:id/trash", handler.TrashTicket)
	t.POST("/:id/restore", handler.RestoreTicket)
}

// SetupProductRoutes registers product catalog endpoints.
func SetupProductRoutes(g *echo.Group, handler products.Handler) {
	p := g.Group("/products")
	p.GET("", handler.ListProducts)
	p.POST("", handler.CreateProduct)
	p.GET("/:id", handler.GetProduct)
	p.PUT("/:id", handler.UpdateProduct)
	p.DELETE("/:id", handler.DeleteProduct)
}

// SetupSimRoutes registers SIM card management endpoints.
func SetupSimRoutes(g *echo.Group, handler sims.Handler) {
	s := g.Group("/sims")
	s.GET("", handler.ListSims)
	s.POST("", handler.CreateSim)
	s.GET("/:id", handler.GetSim)
	s.PUT("/:id", handler.UpdateSim)
	s.DELETE("/:id", handler.DeleteSim)
	s.POST("/:id/assign", handler.AssignSim)
	s.POST("/:id/unassign", handler.UnassignSim)
}

// SetupDepartmentRoutes registers department management endpoints.
func SetupDepartmentRoutes(g *echo.Group, handler departments.Handler) {
	d := g.Group("/departments")
	d.GET("", handler.ListDepartments)
	d.POST("", handler.CreateDepartment)
	d.GET("/:id", handler.GetDepartment)
	d.PUT("/:id", handler.UpdateDepartment)
	d.DELETE("/:id", handler.DeleteDepartment)
}

// SetupStaffRoutes registers staff listing endpoints.
func SetupStaffRoutes(g *echo.Group, handler stafflisting.Handler) {
	s := g.Group("/staff")
	s.GET("", handler.ListStaff)
	s.POST("", handler.CreateStaff)
	s.GET("/export-csv", handler.ExportStaffCSV)
	s.GET("/:id", handler.GetStaff)
	s.PUT("/:id", handler.UpdateStaff)
	s.DELETE("/:id", handler.DeleteStaff)
}

// SetupSystemUpdateRoutes registers system update endpoints.
func SetupSystemUpdateRoutes(g *echo.Group, handler systemupdates.Handler) {
	su := g.Group("/system-updates")
	su.GET("", handler.ListSystemUpdates)
	su.POST("", handler.CreateSystemUpdate)
	su.GET("/export-csv", handler.ExportSystemUpdatesCSV)
	su.GET("/:id", handler.GetSystemUpdate)
	su.PUT("/:id", handler.UpdateSystemUpdate)
	su.DELETE("/:id", handler.DeleteSystemUpdate)
}

// SetupBackupRoutes registers backup management endpoints.
func SetupBackupRoutes(g *echo.Group, handler backups.Handler) {
	b := g.Group("/backups")
	b.GET("", handler.ListBackups)
	b.POST("", handler.CreateBackup)
	b.GET("/export-csv", handler.ExportBackupsCSV)
	b.GET("/:id", handler.GetBackup)
	b.PUT("/:id", handler.UpdateBackup)
	b.DELETE("/:id", handler.DeleteBackup)
	b.GET("/:id/download", handler.DownloadBackup)
}

// SetupMerchantRoutes registers merchant and merchant statement endpoints.
func SetupMerchantRoutes(g *echo.Group, handler merchants.Handler) {
	m := g.Group("/merchants")
	m.GET("", handler.ListMerchants)
	m.POST("", handler.CreateMerchant)
	m.GET("/export-csv", handler.ExportMerchantsCSV)
	m.GET("/statements/export-csv", handler.ExportStatementsCSV)
	m.GET("/:id", handler.GetMerchant)
	m.PUT("/:id", handler.UpdateMerchant)
	m.DELETE("/:id", handler.DeleteMerchant)
	m.POST("/:id/portal", handler.EnableMerchantPortal)
	m.GET("/:merchantId/statements", handler.ListStatements)
	m.POST("/:merchantId/statements", handler.CreateStatement)
	m.GET("/:merchantId/statements/:id", handler.GetStatement)
	m.PUT("/:merchantId/statements/:id", handler.UpdateStatement)
	m.DELETE("/:merchantId/statements/:id", handler.DeleteStatement)
}

// SetupTaxonomyRoutes registers compliance taxonomy endpoints.
func SetupTaxonomyRoutes(g *echo.Group, handler taxonomy.Handler) {
	t := g.Group("/taxonomy")
	t.GET("", handler.ListTaxonomy)
	t.POST("", handler.CreateTaxonomyItem)
	t.GET("/:id", handler.GetTaxonomyItem)
	t.PUT("/:id", handler.UpdateTaxonomyItem)
	t.DELETE("/:id", handler.DeleteTaxonomyItem)
	t.POST("/:id/complete", handler.ToggleComplete)
}

// SetupSettingsRoutes registers platform settings endpoints.
func SetupSettingsRoutes(g *echo.Group, handler settings.Handler) {
	s := g.Group("/settings")
	s.GET("", handler.ListSettings)
	s.PUT("", handler.UpdateSettings)
}

// SetupProfileRoutes registers user profile endpoints.
func SetupProfileRoutes(g *echo.Group, handler profile.Handler) {
	p := g.Group("/profile")
	p.GET("", handler.GetMyProfile)
	p.PUT("", handler.UpdateMyProfile)
}

// SetupRiskComplianceRoutes registers risk & compliance endpoints.
func SetupRiskComplianceRoutes(g *echo.Group, handler riskcompliance.Handler) {
	r := g.Group("/risk-compliance")
	r.GET("", handler.ListItems)
	r.POST("", handler.CreateItem)
	r.GET("/export-csv", handler.ExportCSV)
	r.GET("/:id", handler.GetItem)
	r.PUT("/:id", handler.UpdateItem)
	r.DELETE("/:id", handler.DeleteItem)
}

// SetupRecycleBinRoutes registers recycle bin (soft-delete recovery) endpoints.
func SetupRecycleBinRoutes(g *echo.Group, handler recyclebin.Handler) {
	rb := g.Group("/recycle-bin")
	rb.GET("", handler.ListDeletedItems)
	rb.POST("/:table/:id/restore", handler.RestoreItem)
	rb.DELETE("/:table/:id", handler.ForceDeleteItem)
}

// SetupDashboardRoutes registers the dashboard KPI endpoint.
func SetupDashboardRoutes(g *echo.Group, handler dashboard.Handler) {
	g.GET("/dashboard", handler.GetDashboard)
}
