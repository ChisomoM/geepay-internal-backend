package models

import (
	"time"

	"github.com/google/uuid"
)

// BudgetAndLicense represents recurring budget items and associated license info.
type BudgetAndLicense struct {
	CompanyBaseModel
	Name              string     `json:"name"`
	RenewalDate       *time.Time `json:"renewal_date"`
	ActualAmount      float64    `json:"actual_amount" gorm:"type:numeric"`
	PurchaseFrequency string     `json:"purchase_frequency"`
	LicenseID         uuid.UUID  `json:"license_id" gorm:"type:uuid;index"`
	License           *License   `json:"license,omitempty" gorm:"foreignKey:LicenseID"`
}

// License represents a software/hardware license with renewal metadata.
type License struct {
	CompanyBaseModel
	Name        string     `json:"name"`
	Provider    string     `json:"provider"`
	RenewalDate *time.Time `json:"renewal_date"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
}

// Statutory represents statutory obligations (taxes, filings) that can be marked paid.
type Statutory struct {
	CompanyBaseModel
	Name    string     `json:"name"`
	DueDate *time.Time `json:"due_date"`
	Amount  float64    `json:"amount" gorm:"type:numeric"`
	IsPaid  bool       `json:"is_paid" gorm:"default:false"`
	PaidAt  *time.Time `json:"paid_at"`
}

// SalaryAdvance represents an employee salary advance and its approval/deduction state.
type SalaryAdvance struct {
	CompanyBaseModel
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;index"`
	Amount          float64    `json:"amount" gorm:"type:numeric"`
	Status          string     `json:"status" gorm:"index"`
	DeductionMonths int        `json:"deduction_months"`
	ApprovedBy      uuid.UUID  `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt      *time.Time `json:"approved_at"`
}
