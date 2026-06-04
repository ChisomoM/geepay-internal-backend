package taxonomy

import (
	"errors"
	"time"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListTaxonomy(db *gorm.DB, companyID string) ([]models.TaxonomyItem, error)
	GetTaxonomyItem(db *gorm.DB, companyID, id string) (*models.TaxonomyItem, error)
	CreateTaxonomyItem(db *gorm.DB, companyID string, req CreateTaxonomyRequest) (*models.TaxonomyItem, error)
	UpdateTaxonomyItem(db *gorm.DB, companyID, id string, req UpdateTaxonomyRequest) (*models.TaxonomyItem, error)
	DeleteTaxonomyItem(db *gorm.DB, companyID, id string) error
	ToggleComplete(db *gorm.DB, companyID, id string) (*models.TaxonomyItem, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateTaxonomyRequest struct {
	Title   string     `json:"title"`
	DueDate *time.Time `json:"due_date"`
	Status  string     `json:"status"`
}

type UpdateTaxonomyRequest struct {
	Title   *string    `json:"title"`
	DueDate *time.Time `json:"due_date"`
	Status  *string    `json:"status"`
}

func (s *service) ListTaxonomy(db *gorm.DB, companyID string) ([]models.TaxonomyItem, error) {
	var items []models.TaxonomyItem
	if err := db.Order("due_date ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetTaxonomyItem(db *gorm.DB, companyID, id string) (*models.TaxonomyItem, error) {
	var item models.TaxonomyItem
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateTaxonomyItem(db *gorm.DB, companyID string, req CreateTaxonomyRequest) (*models.TaxonomyItem, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	status := req.Status
	if status == "" {
		status = "pending"
	}
	item := models.TaxonomyItem{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Title:            req.Title,
		DueDate:          req.DueDate,
		Status:           status,
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) UpdateTaxonomyItem(db *gorm.DB, companyID, id string, req UpdateTaxonomyRequest) (*models.TaxonomyItem, error) {
	var item models.TaxonomyItem
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.DueDate != nil {
		item.DueDate = req.DueDate
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) DeleteTaxonomyItem(db *gorm.DB, companyID, id string) error {
	return db.Delete(&models.TaxonomyItem{}, "id = ?", id).Error
}

func (s *service) ToggleComplete(db *gorm.DB, companyID, id string) (*models.TaxonomyItem, error) {
	var item models.TaxonomyItem
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if item.Status == "completed" {
		item.Status = "pending"
	} else {
		item.Status = "completed"
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
