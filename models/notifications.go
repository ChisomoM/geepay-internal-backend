package models

import (
	"time"

	"github.com/google/uuid"
)

// Alert represents a system or domain-level alert.
type Alert struct {
	CompanyBaseModel
	Level   string `json:"level"`
	Message string `json:"message" gorm:"type:text"`
	IsRead  bool   `json:"is_read" gorm:"default:false"`
}

// AlertEmail stores email recipients for alerts.
type AlertEmail struct {
	CompanyBaseModel
	AlertID uuid.UUID `json:"alert_id" gorm:"type:uuid;index"`
	Email   string    `json:"email"`
}

// TaxonomyItem represents a due-date item (e.g., compliance task).
type TaxonomyItem struct {
	CompanyBaseModel
	Title   string     `json:"title"`
	DueDate *time.Time `json:"due_date"`
	Status  string     `json:"status" gorm:"index"`
}

// TaxonomyNotification represents an alert/notification for a taxonomy item.
type TaxonomyNotification struct {
	CompanyBaseModel
	TaxonomyItemID uuid.UUID  `json:"taxonomy_item_id" gorm:"type:uuid;index"`
	NotifiedAt     *time.Time `json:"notified_at"`
	Type           string     `json:"type"`
}

// UserNotificationPreference stores per-user notification settings.
type UserNotificationPreference struct {
	CompanyBaseModel
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Email  bool      `json:"email" gorm:"default:true"`
	SMS    bool      `json:"sms" gorm:"default:false"`
}
