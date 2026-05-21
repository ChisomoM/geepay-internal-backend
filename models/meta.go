package models

// Tenant represents a multi-tenant organization or workspace.
// In single-tenant mode, only one Tenant record exists.
type Tenant struct {
	BaseModel
	Name        string `json:"name"`
	Slug        string `json:"slug" gorm:"uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Metadata    string `json:"metadata" gorm:"type:jsonb"` // Custom tenant config
}
