package models

import (
	"github.com/google/uuid"
)

// Department represents organizational departments for assignments.
type Department struct {
	CompanyBaseModel
	Name        string         `json:"name"`
	Description string         `json:"description" gorm:"type:text"`
	Staff       []StaffListing `json:"staff,omitempty" gorm:"foreignKey:DepartmentID"`
}

// StaffListing represents a staff member with basic payroll info.
type StaffListing struct {
	CompanyBaseModel
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Designation  string    `json:"designation"`
	Salary       float64   `json:"salary" gorm:"type:numeric"`
	DepartmentID uuid.UUID `json:"department_id" gorm:"type:uuid;index"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
}

// Profile stores user profile details beyond the main User model.
type Profile struct {
	CompanyBaseModel
	UserID  uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Bio     string    `json:"bio" gorm:"type:text"`
	Phone   string    `json:"phone"`
	Address string    `json:"address" gorm:"type:text"`
}
