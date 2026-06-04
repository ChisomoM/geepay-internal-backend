package departments

import (
	"errors"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListDepartments(db *gorm.DB, companyID string) ([]models.Department, error)
	GetDepartment(db *gorm.DB, companyID, id string) (*models.Department, error)
	CreateDepartment(db *gorm.DB, companyID string, req CreateDepartmentRequest) (*models.Department, error)
	UpdateDepartment(db *gorm.DB, companyID, id string, req UpdateDepartmentRequest) (*models.Department, error)
	DeleteDepartment(db *gorm.DB, companyID, id string) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateDepartmentRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (s *service) ListDepartments(db *gorm.DB, companyID string) ([]models.Department, error) {
	var items []models.Department
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetDepartment(db *gorm.DB, companyID, id string) (*models.Department, error) {
	var d models.Department
	if err := db.Preload("Staff").First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *service) CreateDepartment(db *gorm.DB, companyID string, req CreateDepartmentRequest) (*models.Department, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	d := models.Department{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Name:             req.Name,
		Description:      req.Description,
	}
	if err := db.Create(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *service) UpdateDepartment(db *gorm.DB, companyID, id string, req UpdateDepartmentRequest) (*models.Department, error) {
	var d models.Department
	if err := db.First(&d, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Description != nil {
		d.Description = *req.Description
	}
	if err := db.Save(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *service) DeleteDepartment(db *gorm.DB, companyID, id string) error {
	return db.Delete(&models.Department{}, "id = ?", id).Error
}
