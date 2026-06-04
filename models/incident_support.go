package models

import (
	"time"

	"github.com/google/uuid"
)

// Incident represents an incident record with workflows and notifications.
type Incident struct {
	CompanyBaseModel
	Title       string     `json:"title"`
	Description string     `json:"description" gorm:"type:text"`
	Type        string     `json:"type" gorm:"index"`
	Status      string     `json:"status" gorm:"index"`
	ReportedBy  uuid.UUID  `json:"reported_by" gorm:"type:uuid;index"`
	AssignedTo  uuid.UUID  `json:"assigned_to" gorm:"type:uuid;index"`
	Notified    bool       `json:"notified" gorm:"default:false"`
	ReportedAt  *time.Time `json:"reported_at"`
}

// SupportTicket represents a customer/internal support ticket with soft-delete support.
// MerchantID is null for internal Geepay tickets; set when submitted via the merchant portal.
type SupportTicket struct {
	CompanyBaseModel
	MerchantID *uuid.UUID `json:"merchant_id,omitempty" gorm:"type:uuid;index"`
	Subject    string     `json:"subject"`
	Body       string     `json:"body" gorm:"type:text"`
	Status     string     `json:"status" gorm:"index"`
	IsTrashed  bool       `json:"is_trashed" gorm:"default:false;index"`
	CreatedBy  uuid.UUID  `json:"created_by" gorm:"type:uuid;index"`
}
