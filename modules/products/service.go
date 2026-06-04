package products

import (
	"errors"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListProducts(db *gorm.DB, companyID string) ([]models.ProductCatalog, error)
	GetProduct(db *gorm.DB, companyID, id string) (*models.ProductCatalog, error)
	CreateProduct(db *gorm.DB, companyID string, req CreateProductRequest) (*models.ProductCatalog, error)
	UpdateProduct(db *gorm.DB, companyID, id string, req UpdateProductRequest) (*models.ProductCatalog, error)
	DeleteProduct(db *gorm.DB, companyID, id string) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

// DTOs
type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ImageURL    string `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	ImageURL    *string `json:"image_url"`
}

func (s *service) ListProducts(db *gorm.DB, companyID string) ([]models.ProductCatalog, error) {
	var items []models.ProductCatalog
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetProduct(db *gorm.DB, companyID, id string) (*models.ProductCatalog, error) {
	var p models.ProductCatalog
	if err := db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *service) CreateProduct(db *gorm.DB, companyID string, req CreateProductRequest) (*models.ProductCatalog, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	p := models.ProductCatalog{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Name:             req.Name,
		Description:      req.Description,
		Status:           req.Status,
		ImageURL:         req.ImageURL,
	}
	if err := db.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *service) UpdateProduct(db *gorm.DB, companyID, id string, req UpdateProductRequest) (*models.ProductCatalog, error) {
	var p models.ProductCatalog
	if err := db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	if req.ImageURL != nil {
		p.ImageURL = *req.ImageURL
	}
	if err := db.Save(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *service) DeleteProduct(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.ProductCatalog{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
