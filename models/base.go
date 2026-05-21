package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel defines a standard set of fields with a UUID primary key.
// Use this as the base for models in single-tenant mode or shared infrastructure tables.
type BaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate hook to generate a new UUID for the primary key if not already set.
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

// TenantBaseModel extends BaseModel with tenant_id for multi-tenant applications.
// All domain models that belong to a specific tenant should embed this instead of BaseModel.
type TenantBaseModel struct {
	BaseModel
	TenantID string `json:"tenant_id" gorm:"index;not null"`
}
