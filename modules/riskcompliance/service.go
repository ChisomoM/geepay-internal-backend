package riskcompliance

import (
	"bytes"
	"encoding/csv"
	"errors"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListItems(db *gorm.DB, companyID string) ([]models.RiskAndCompliance, error)
	GetItem(db *gorm.DB, companyID, id string) (*models.RiskAndCompliance, error)
	CreateItem(db *gorm.DB, companyID string, req CreateRiskComplianceRequest) (*models.RiskAndCompliance, error)
	UpdateItem(db *gorm.DB, companyID, id string, req UpdateRiskComplianceRequest) (*models.RiskAndCompliance, error)
	DeleteItem(db *gorm.DB, companyID, id string) error
	ExportCSV(db *gorm.DB, companyID string) ([]byte, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateRiskComplianceRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Likelihood  string `json:"likelihood"`
	Impact      string `json:"impact"`
	Mitigation  string `json:"mitigation"`
	Status      string `json:"status"`
}

type UpdateRiskComplianceRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Likelihood  *string `json:"likelihood"`
	Impact      *string `json:"impact"`
	Mitigation  *string `json:"mitigation"`
	Status      *string `json:"status"`
}

func (s *service) ListItems(db *gorm.DB, companyID string) ([]models.RiskAndCompliance, error) {
	var items []models.RiskAndCompliance
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetItem(db *gorm.DB, companyID, id string) (*models.RiskAndCompliance, error) {
	var item models.RiskAndCompliance
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateItem(db *gorm.DB, companyID string, req CreateRiskComplianceRequest) (*models.RiskAndCompliance, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	status := req.Status
	if status == "" {
		status = "open"
	}
	item := models.RiskAndCompliance{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Title:            req.Title,
		Description:      req.Description,
		Likelihood:       req.Likelihood,
		Impact:           req.Impact,
		Mitigation:       req.Mitigation,
		Status:           status,
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) UpdateItem(db *gorm.DB, companyID, id string, req UpdateRiskComplianceRequest) (*models.RiskAndCompliance, error) {
	var item models.RiskAndCompliance
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Likelihood != nil {
		item.Likelihood = *req.Likelihood
	}
	if req.Impact != nil {
		item.Impact = *req.Impact
	}
	if req.Mitigation != nil {
		item.Mitigation = *req.Mitigation
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) DeleteItem(db *gorm.DB, companyID, id string) error {
	return db.Delete(&models.RiskAndCompliance{}, "id = ?", id).Error
}

func (s *service) ExportCSV(db *gorm.DB, companyID string) ([]byte, error) {
	items, err := s.ListItems(db, companyID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "Title", "Description", "Likelihood", "Impact", "Mitigation", "Status", "Created At"})
	for _, item := range items {
		_ = w.Write([]string{
			item.ID.String(),
			item.Title,
			item.Description,
			item.Likelihood,
			item.Impact,
			item.Mitigation,
			item.Status,
			item.CreatedAt.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
