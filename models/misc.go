package models

import (
	"time"

	"github.com/google/uuid"
)

// Report represents a generated report file (CSV/PDF) for auditing or export.
type Report struct {
	CompanyBaseModel
	Name        string     `json:"name"`
	GeneratedBy uuid.UUID  `json:"generated_by" gorm:"type:uuid;index"`
	GeneratedAt *time.Time `json:"generated_at"`
	FilePath    string     `json:"file_path"`
}

// Setting holds system-wide configuration entries (non-company or platform-scoped).
type Setting struct {
	BaseModel
	Key   string `json:"key" gorm:"uniqueIndex"`
	Value string `json:"value" gorm:"type:text"`
}

// RecycleBin stores metadata about soft-deleted resources for recovery.
type RecycleBin struct {
	CompanyBaseModel
	Resource   string     `json:"resource"`
	ResourceID string     `json:"resource_id"`
	DeletedBy  uuid.UUID  `json:"deleted_by" gorm:"type:uuid;index"`
	DeletedAt  *time.Time `json:"deleted_at"`
}

// RiskAndCompliance represents an item in the risk register.
type RiskAndCompliance struct {
	CompanyBaseModel
	Title       string `json:"title"`
	Description string `json:"description" gorm:"type:text"`
	Likelihood  string `json:"likelihood"`
	Impact      string `json:"impact"`
	Mitigation  string `json:"mitigation" gorm:"type:text"`
	Status      string `json:"status" gorm:"index"`
}

// SystemUpdate stores changelog/version records for the platform.
type SystemUpdate struct {
	BaseModel
	Version    string     `json:"version"`
	Changelog  string     `json:"changelog" gorm:"type:text"`
	ReleasedAt *time.Time `json:"released_at"`
}
