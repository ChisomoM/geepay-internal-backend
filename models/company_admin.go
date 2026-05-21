package models

// CompanyAdmin represents a platform-level administrator for the company/organization.
// Unlike User (which is tenant-scoped), CompanyAdmin is platform-scoped and stored in the meta database.
// CompanyAdmins have access to ControlHub for managing multiple tenants.
type CompanyAdmin struct {
	BaseModel
	Email        string `json:"email" gorm:"unique;not null;index"`
	PasswordHash string `json:"-" gorm:"not null"`
	IsActive     bool   `json:"is_active" gorm:"default:true;index"`
}

// TableName specifies the table name for CompanyAdmin.
// This should be stored in the meta database, not tenant-scoped databases.
func (CompanyAdmin) TableName() string {
	return "company_admins"
}
