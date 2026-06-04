package models

import (
	"github.com/google/uuid"
)

// AuditLog represents a log of user activity for compliance and debugging.
// All state-changing operations should be logged here via AuditMiddleware.
type AuditLog struct {
	CompanyBaseModel
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action     string    `json:"action" gorm:"index"`      // e.g., "LOGIN", "UPDATE_PROFILE", "EXPORT"
	Resource   string    `json:"resource" gorm:"index"`    // e.g., "users", "documents"
	ResourceID string    `json:"resource_id" gorm:"index"` // ID of the affected resource
	Details    string    `json:"details" gorm:"type:text"` // JSON or text details about the change
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
}
