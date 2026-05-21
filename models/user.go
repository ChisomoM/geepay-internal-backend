package models

import (
	"database/sql"

	"github.com/google/uuid"
)

// User represents a system user with authentication and role assignment.
type User struct {
	TenantBaseModel
	Email               string                   `json:"email" gorm:"uniqueIndex:idx_tenant_email"`
	PasswordHash        string                   `json:"-"`
	FirstName           string                   `json:"first_name"`
	LastName            string                   `json:"last_name"`
	IsActive            bool                     `json:"is_active" gorm:"default:true"`
	RoleSlug            string                   `json:"role_slug" gorm:"index"` // e.g., "admin", "user", "viewer"
	PermissionOverrides []UserPermissionOverride `json:"permission_overrides,omitempty" gorm:"foreignKey:UserID"`
}

// Role represents a collection of permissions that can be assigned to users.
type Role struct {
	BaseModel
	TenantID    string       `json:"tenant_id" gorm:"index;not null"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug" gorm:"uniqueIndex:idx_tenant_role_slug"`
	Description string       `json:"description" gorm:"type:text"`
	IsSystem    bool         `json:"is_system" gorm:"default:false"` // Read-only system role
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

// Permission represents a granular action that can be performed in the system.
// Code format: "resource.action" (e.g., "documents.create", "users.export")
type Permission struct {
	BaseModel
	Code        string `json:"code" gorm:"uniqueIndex:idx_tenant_permission_code"`
	TenantID    string `json:"tenant_id" gorm:"index;not null"`
	Description string `json:"description" gorm:"type:text"`
	Category    string `json:"category"` // e.g., "documents", "users", "admin"
}

// UserPermissionOverride allows fine-grained permission grants/denials per user without changing their role.
// If Granted=true, the permission is granted; if false, it's denied (revoked from their role).
type UserPermissionOverride struct {
	BaseModel
	UserID       uuid.UUID    `json:"user_id" gorm:"type:uuid;index"`
	User         *User        `json:"-" gorm:"foreignKey:UserID"`
	PermissionID uuid.UUID    `json:"permission_id" gorm:"type:uuid;index"`
	Permission   *Permission  `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
	Granted      bool         `json:"granted"` // true = grant, false = deny
	GrantedBy    uuid.UUID    `json:"granted_by" gorm:"type:uuid"`
	GrantedAt    sql.NullTime `json:"granted_at"`
	Reason       string       `json:"reason" gorm:"type:text"`
}
