package models

import (
	"github.com/google/uuid"
)

// InventoryCategory groups inventory items and holds thresholds.
type InventoryCategory struct {
	CompanyBaseModel
	Name                string      `json:"name"`
	LowStockThreshold   int         `json:"low_stock_threshold"`
	OutOfStockThreshold int         `json:"out_of_stock_threshold"`
	Items               []Inventory `json:"items,omitempty" gorm:"foreignKey:CategoryID"`
}

// Inventory represents an item that can be assigned to users or departments.
type Inventory struct {
	CompanyBaseModel
	CategoryID   uuid.UUID          `json:"category_id" gorm:"type:uuid;index"`
	Category     *InventoryCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Name         string             `json:"name"`
	Quantity     int                `json:"quantity"`
	Status       string             `json:"status" gorm:"index"` // assigned/available/assignable
	AssignedTo   uuid.UUID          `json:"assigned_to" gorm:"type:uuid;index"`
	AssignedType string             `json:"assigned_type"` // e.g., "user", "department"
}
