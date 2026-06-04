package dashboard

import (
	"time"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	GetDashboardStats(db *gorm.DB, companyID string) (*DashboardStats, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CountPair struct {
	Total int64 `json:"total"`
	Other int64 `json:"other,omitempty"`
}

type DashboardStats struct {
	Merchants   MerchantStats    `json:"merchants"`
	Staff       StaffStats       `json:"staff"`
	Incidents   IncidentStats    `json:"incidents"`
	Tickets     TicketStats      `json:"tickets"`
	BudgetItems BudgetStats      `json:"budget_items"`
	Inventory   InventoryStats   `json:"inventory"`
	SimCards    SimCardStats     `json:"sim_cards"`
	Finance     FinanceSnapshot  `json:"finance"`
	Compliance  ComplianceAlerts `json:"compliance"`
}

type FinanceSnapshot struct {
	PendingAdvances    int64 `json:"pending_advances"`
	ApprovedAdvances   int64 `json:"approved_advances"`
	MerchantStatements int64 `json:"merchant_statements"`
}

type ComplianceAlerts struct {
	DueSoon int64 `json:"due_soon"`
	Overdue int64 `json:"overdue"`
}

type MerchantStats struct {
	Total int64 `json:"total"`
}

type StaffStats struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type IncidentStats struct {
	Total int64 `json:"total"`
	Open  int64 `json:"open"`
}

type TicketStats struct {
	Total int64 `json:"total"`
	Open  int64 `json:"open"`
}

type BudgetStats struct {
	Total        int64 `json:"total"`
	ExpiringSoon int64 `json:"expiring_soon"`
}

type InventoryStats struct {
	Total    int64 `json:"total"`
	LowStock int64 `json:"low_stock"`
}

type SimCardStats struct {
	Total    int64 `json:"total"`
	Assigned int64 `json:"assigned"`
}

func (s *service) GetDashboardStats(db *gorm.DB, companyID string) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Merchants
	db.Model(&models.Merchant{}).Count(&stats.Merchants.Total)

	// Staff
	db.Model(&models.StaffListing{}).Count(&stats.Staff.Total)
	db.Model(&models.StaffListing{}).Where("is_active = ?", true).Count(&stats.Staff.Active)

	// Incidents
	db.Model(&models.Incident{}).Count(&stats.Incidents.Total)
	db.Model(&models.Incident{}).Where("status NOT IN (?)", []string{"resolved", "closed"}).Count(&stats.Incidents.Open)

	// Support Tickets (is_trashed = false = active tickets)
	db.Model(&models.SupportTicket{}).Where("is_trashed = ?", false).Count(&stats.Tickets.Total)
	db.Model(&models.SupportTicket{}).Where("is_trashed = ? AND status NOT IN (?)", false, []string{"closed", "resolved"}).Count(&stats.Tickets.Open)

	// Budget items
	db.Model(&models.BudgetAndLicense{}).Count(&stats.BudgetItems.Total)
	thirtyDaysOut := time.Now().AddDate(0, 0, 30)
	db.Model(&models.BudgetAndLicense{}).
		Where("renewal_date IS NOT NULL AND renewal_date <= ? AND renewal_date >= ?", thirtyDaysOut, time.Now()).
		Count(&stats.BudgetItems.ExpiringSoon)

	// Inventory
	db.Model(&models.Inventory{}).Count(&stats.Inventory.Total)
	// Low stock: items where quantity <= category low_stock_threshold — simplified count of status "low_stock"
	db.Model(&models.Inventory{}).Where("status = ?", "low_stock").Count(&stats.Inventory.LowStock)

	// SIM cards
	db.Model(&models.SimCard{}).Count(&stats.SimCards.Total)
	db.Model(&models.SimCard{}).Where("status = ?", "assigned").Count(&stats.SimCards.Assigned)

	// Finance snapshot
	db.Model(&models.SalaryAdvance{}).Where("status = ?", "pending").Count(&stats.Finance.PendingAdvances)
	db.Model(&models.SalaryAdvance{}).Where("status = ?", "approved").Count(&stats.Finance.ApprovedAdvances)
	db.Model(&models.MerchantStatement{}).Count(&stats.Finance.MerchantStatements)

	// Compliance alerts (taxonomy items not yet completed)
	sevenDaysOut := time.Now().AddDate(0, 0, 7)
	db.Model(&models.TaxonomyItem{}).
		Where("status != ? AND due_date BETWEEN ? AND ?", "completed", time.Now(), sevenDaysOut).
		Count(&stats.Compliance.DueSoon)
	db.Model(&models.TaxonomyItem{}).
		Where("status != ? AND due_date < ?", "completed", time.Now()).
		Count(&stats.Compliance.Overdue)

	return stats, nil
}
