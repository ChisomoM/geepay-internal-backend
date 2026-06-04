package models

// CompanyAdmin has been removed. Platform-level admin access is now handled
// by the super_admin role within the internal User model.

// type CompanyAdmin struct {
// 	BaseModel
// 	Email        string `json:"email" gorm:"unique;not null;index"`
// 	PasswordHash string `json:"-" gorm:"not null"`
// 	IsActive     bool   `json:"is_active" gorm:"default:true;index"`
// }
//
// func (CompanyAdmin) TableName() string {
// 	return "company_admins"
// }
