package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Company represents a multi-company organization or workspace.
// In single-company mode, only one Company record exists.
type Company struct {
	BaseModel
	Name        string `json:"name"`
	Slug        string `json:"slug" gorm:"uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Metadata    string `json:"metadata" gorm:"type:jsonb"` // Custom company config
}

// BeforeCreate hook to generate UUID if missing and set Metadata to "{}" if empty.
// Note: Because Company has this hook, BaseModel's BeforeCreate is not called automatically.
func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Metadata == "" {
		c.Metadata = "{}"
	}
	return nil
}
