package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductCatalog represents products or services offered by the company.
type ProductCatalog struct {
	CompanyBaseModel
	Name        string `json:"name"`
	Description string `json:"description" gorm:"type:text"`
	Status      string `json:"status" gorm:"index"`
	ImageURL    string `json:"image_url"`
}

// Backup tracks backup files and their completion status.
type Backup struct {
	CompanyBaseModel
	FilePath    string     `json:"file_path"`
	Success     bool       `json:"success" gorm:"default:false"`
	CompletedAt *time.Time `json:"completed_at"`
}

// SimCard represents a SIM card asset.
type SimCard struct {
	CompanyBaseModel
	ICCID      string    `json:"iccid" gorm:"uniqueIndex"`
	Network    string    `json:"network"`
	AssignedTo uuid.UUID `json:"assigned_to" gorm:"type:uuid;index"`
	Status     string    `json:"status" gorm:"index"`
}
