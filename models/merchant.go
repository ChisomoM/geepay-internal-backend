package models

import (
	"time"

	"github.com/google/uuid"
)

// Merchant represents a merchant/customer record and onboarding/integration state.
type Merchant struct {
	CompanyBaseModel
	Name               string              `json:"name"`
	Email              string              `json:"email"`
	ContactNumber      string              `json:"contact_number"`
	BusinessType       string              `json:"business_type"`
	Notes              string              `json:"notes" gorm:"type:text"`
	IntegrationStatus  string              `json:"integration_status" gorm:"index"`
	OnboardedOn        *time.Time          `json:"onboarded_on"`
	Statements         []MerchantStatement `json:"statements,omitempty" gorm:"foreignKey:MerchantID"`
	PortalEmail        string              `json:"portal_email" gorm:"index"`
	PortalPasswordHash string              `json:"-"`
	PortalEnabled      bool                `json:"portal_enabled" gorm:"default:false"`
}

// MerchantStatement links merchant statements stored externally (e.g., Google Drive).
type MerchantStatement struct {
	CompanyBaseModel
	MerchantID   uuid.UUID `json:"merchant_id" gorm:"type:uuid;index"`
	Merchant     *Merchant `json:"-" gorm:"foreignKey:MerchantID"`
	DriveLink    string    `json:"drive_link"`
	DocumentName string    `json:"document_name"`
}
