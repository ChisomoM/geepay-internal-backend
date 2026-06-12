package models

import (
	"time"

	"github.com/google/uuid"
)

// Ticket is the unified CRM entity for both support requests and incident reports.
// Kind distinguishes between the two: "support" or "incident".
type Ticket struct {
	CompanyBaseModel
	TicketNumber     int64          `json:"ticket_number" gorm:"uniqueIndex:idx_company_ticket_number"`
	Kind             string         `json:"kind" gorm:"index;not null"`
	Subject          string         `json:"subject" gorm:"not null"`
	Description      string         `json:"description" gorm:"type:text"`
	Priority         string         `json:"priority" gorm:"index;default:'medium'"`
	Status           string         `json:"status" gorm:"index;default:'open'"`
	Source           string         `json:"source" gorm:"default:'admin'"`
	CategoryID       *uuid.UUID     `json:"category_id" gorm:"type:uuid;index"`
	Category         *TicketCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	MerchantID       *uuid.UUID     `json:"merchant_id" gorm:"type:uuid;index"`
	Merchant         *Merchant      `json:"merchant,omitempty" gorm:"foreignKey:MerchantID"`
	CreatedByUserID  *uuid.UUID     `json:"created_by_user_id" gorm:"type:uuid;index"`
	AssignedToUserID *uuid.UUID     `json:"assigned_to_user_id" gorm:"type:uuid;index"`
	AssignedToTeamID *uuid.UUID     `json:"assigned_to_team_id" gorm:"type:uuid;index"`
	SLADeadline      *time.Time     `json:"sla_deadline"`
	ResolvedAt       *time.Time     `json:"resolved_at"`
	FirstResponseAt  *time.Time     `json:"first_response_at"`
	BreachNotifiedAt *time.Time     `json:"breach_notified_at"`
	IsTrashed        bool           `json:"is_trashed" gorm:"default:false;index"`
	Events           []TicketEvent  `json:"events,omitempty" gorm:"foreignKey:TicketID"`
}

// TicketEvent records every action taken on a Ticket: comments, status changes, assignments, escalations.
// IsInternal hides staff-only notes from the merchant-facing view.
type TicketEvent struct {
	CompanyBaseModel
	TicketID   uuid.UUID `json:"ticket_id" gorm:"type:uuid;index;not null"`
	EventType  string    `json:"event_type" gorm:"not null"`
	ActorID    uuid.UUID `json:"actor_id" gorm:"type:uuid;not null"`
	ActorType  string    `json:"actor_type" gorm:"not null"`
	Body       string    `json:"body" gorm:"type:text"`
	IsInternal bool      `json:"is_internal" gorm:"default:false"`
	Metadata   string    `json:"metadata" gorm:"type:text"`
}

// TicketCategory is the taxonomy used to classify tickets and drive auto-routing.
// DepartmentID is the default department assigned when this category is selected.
type TicketCategory struct {
	CompanyBaseModel
	Name         string     `json:"name" gorm:"not null"`
	DepartmentID *uuid.UUID `json:"department_id" gorm:"type:uuid;index"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
}

// CRMRoutingRule defines auto-assignment logic: when a ticket matches a given
// category and priority, it is routed to the specified department (and optionally a user).
type CRMRoutingRule struct {
	CompanyBaseModel
	CategoryID     *uuid.UUID `json:"category_id" gorm:"type:uuid;index"`
	Priority       string     `json:"priority" gorm:"not null"`
	DepartmentID   *uuid.UUID `json:"department_id" gorm:"type:uuid"`
	AssignedUserID *uuid.UUID `json:"assigned_user_id" gorm:"type:uuid"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
}

// SLAPolicy defines response and resolution time targets per priority level.
type SLAPolicy struct {
	CompanyBaseModel
	Priority        string `json:"priority" gorm:"not null"`
	ResponseHours   int    `json:"response_hours" gorm:"not null"`
	ResolutionHours int    `json:"resolution_hours" gorm:"not null"`
	IsActive        bool   `json:"is_active" gorm:"default:true"`
}
