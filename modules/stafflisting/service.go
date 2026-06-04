package stafflisting

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"

	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	ListStaff(db *gorm.DB, companyID string) ([]models.StaffListing, error)
	GetStaff(db *gorm.DB, companyID, id string) (*models.StaffListing, error)
	CreateStaff(db *gorm.DB, companyID string, req CreateStaffRequest) (*models.StaffListing, error)
	UpdateStaff(db *gorm.DB, companyID, id string, req UpdateStaffRequest) (*models.StaffListing, error)
	DeleteStaff(db *gorm.DB, companyID, id string) error
	ExportStaffCSV(db *gorm.DB, companyID string) ([]byte, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateStaffRequest struct {
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Designation  string  `json:"designation"`
	Salary       float64 `json:"salary"`
	DepartmentID string  `json:"department_id"`
	IsActive     bool    `json:"is_active"`
}

type UpdateStaffRequest struct {
	FirstName    *string  `json:"first_name"`
	LastName     *string  `json:"last_name"`
	Designation  *string  `json:"designation"`
	Salary       *float64 `json:"salary"`
	DepartmentID *string  `json:"department_id"`
	IsActive     *bool    `json:"is_active"`
}

func (s *service) ListStaff(db *gorm.DB, companyID string) ([]models.StaffListing, error) {
	var items []models.StaffListing
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetStaff(db *gorm.DB, companyID, id string) (*models.StaffListing, error) {
	var item models.StaffListing
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateStaff(db *gorm.DB, companyID string, req CreateStaffRequest) (*models.StaffListing, error) {
	if req.FirstName == "" || req.LastName == "" {
		return nil, errors.New("first_name and last_name are required")
	}
	item := models.StaffListing{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Designation:      req.Designation,
		Salary:           req.Salary,
		IsActive:         req.IsActive,
	}
	if req.DepartmentID != "" {
		uid, err := uuid.Parse(req.DepartmentID)
		if err != nil {
			return nil, errors.New("invalid department_id")
		}
		item.DepartmentID = uid
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) UpdateStaff(db *gorm.DB, companyID, id string, req UpdateStaffRequest) (*models.StaffListing, error) {
	var item models.StaffListing
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.FirstName != nil {
		item.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		item.LastName = *req.LastName
	}
	if req.Designation != nil {
		item.Designation = *req.Designation
	}
	if req.Salary != nil {
		item.Salary = *req.Salary
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}
	if req.DepartmentID != nil {
		uid, err := uuid.Parse(*req.DepartmentID)
		if err != nil {
			return nil, errors.New("invalid department_id")
		}
		item.DepartmentID = uid
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) DeleteStaff(db *gorm.DB, companyID, id string) error {
	return db.Delete(&models.StaffListing{}, "id = ?", id).Error
}

func (s *service) ExportStaffCSV(db *gorm.DB, companyID string) ([]byte, error) {
	items, err := s.ListStaff(db, companyID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "First Name", "Last Name", "Designation", "Salary", "Department ID", "Is Active", "Created At"})
	for _, item := range items {
		_ = w.Write([]string{
			item.ID.String(),
			item.FirstName,
			item.LastName,
			item.Designation,
			fmt.Sprintf("%.2f", item.Salary),
			item.DepartmentID.String(),
			fmt.Sprintf("%v", item.IsActive),
			item.CreatedAt.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
