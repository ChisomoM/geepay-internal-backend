package inventory

import (
	"errors"

	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	// Items
	ListItems(db *gorm.DB, companyID string) ([]models.Inventory, error)
	ListItemsByCategory(db *gorm.DB, companyID, categoryID string) ([]models.Inventory, error)
	GetItem(db *gorm.DB, companyID, id string) (*models.Inventory, error)
	CreateItem(db *gorm.DB, companyID, categoryID string, req CreateItemRequest) (*models.Inventory, error)
	UpdateItem(db *gorm.DB, companyID, id string, req UpdateItemRequest) (*models.Inventory, error)
	DeleteItem(db *gorm.DB, companyID, id string) error
	AssignItem(db *gorm.DB, companyID, id string, assignedTo string, assignedType string) (*models.Inventory, error)
	UnassignItem(db *gorm.DB, companyID, id string) (*models.Inventory, error)

	// Categories
	ListCategories(db *gorm.DB, companyID string) ([]models.InventoryCategory, error)
	GetCategory(db *gorm.DB, companyID, id string) (*models.InventoryCategory, error)
	CreateCategory(db *gorm.DB, companyID string, req CreateCategoryRequest) (*models.InventoryCategory, error)
	UpdateCategory(db *gorm.DB, companyID, id string, req UpdateCategoryRequest) (*models.InventoryCategory, error)
	DeleteCategory(db *gorm.DB, companyID, id string) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

// DTOs
type CreateItemRequest struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Status   string `json:"status"`
}

type UpdateItemRequest struct {
	Name     *string `json:"name"`
	Quantity *int    `json:"quantity"`
	Status   *string `json:"status"`
}

type CreateCategoryRequest struct {
	Name                string `json:"name"`
	LowStockThreshold   int    `json:"low_stock_threshold"`
	OutOfStockThreshold int    `json:"out_of_stock_threshold"`
}

type UpdateCategoryRequest struct {
	Name                *string `json:"name"`
	LowStockThreshold   *int    `json:"low_stock_threshold"`
	OutOfStockThreshold *int    `json:"out_of_stock_threshold"`
}

// Items implementation
func (s *service) ListItems(db *gorm.DB, companyID string) ([]models.Inventory, error) {
	var items []models.Inventory
	if err := db.Preload("Category").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) ListItemsByCategory(db *gorm.DB, companyID, categoryID string) ([]models.Inventory, error) {
	var items []models.Inventory
	if err := db.Where("category_id = ?", categoryID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetItem(db *gorm.DB, companyID, id string) (*models.Inventory, error) {
	var it models.Inventory
	if err := db.Preload("Category").First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) CreateItem(db *gorm.DB, companyID, categoryID string, req CreateItemRequest) (*models.Inventory, error) {
	cid, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, errors.New("invalid category id")
	}
	it := models.Inventory{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		CategoryID:       cid,
		Name:             req.Name,
		Quantity:         req.Quantity,
		Status:           req.Status,
	}
	if err := db.Create(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) UpdateItem(db *gorm.DB, companyID, id string, req UpdateItemRequest) (*models.Inventory, error) {
	var it models.Inventory
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		it.Name = *req.Name
	}
	if req.Quantity != nil {
		it.Quantity = *req.Quantity
	}
	if req.Status != nil {
		it.Status = *req.Status
	}
	if err := db.Save(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) DeleteItem(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.Inventory{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) AssignItem(db *gorm.DB, companyID, id string, assignedTo string, assignedType string) (*models.Inventory, error) {
	var it models.Inventory
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(assignedTo)
	if err != nil {
		return nil, errors.New("invalid assigned_to id")
	}
	it.AssignedTo = uid
	it.AssignedType = assignedType
	it.Status = "assigned"
	if err := db.Save(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) UnassignItem(db *gorm.DB, companyID, id string) (*models.Inventory, error) {
	var it models.Inventory
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	it.AssignedTo = uuid.Nil
	it.AssignedType = ""
	it.Status = "available"
	if err := db.Save(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

// Categories implementation
func (s *service) ListCategories(db *gorm.DB, companyID string) ([]models.InventoryCategory, error) {
	var cats []models.InventoryCategory
	if err := db.Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

func (s *service) GetCategory(db *gorm.DB, companyID, id string) (*models.InventoryCategory, error) {
	var c models.InventoryCategory
	if err := db.First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *service) CreateCategory(db *gorm.DB, companyID string, req CreateCategoryRequest) (*models.InventoryCategory, error) {
	c := models.InventoryCategory{
		CompanyBaseModel:    models.CompanyBaseModel{CompanyID: companyID},
		Name:                req.Name,
		LowStockThreshold:   req.LowStockThreshold,
		OutOfStockThreshold: req.OutOfStockThreshold,
	}
	if err := db.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *service) UpdateCategory(db *gorm.DB, companyID, id string, req UpdateCategoryRequest) (*models.InventoryCategory, error) {
	var c models.InventoryCategory
	if err := db.First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.LowStockThreshold != nil {
		c.LowStockThreshold = *req.LowStockThreshold
	}
	if req.OutOfStockThreshold != nil {
		c.OutOfStockThreshold = *req.OutOfStockThreshold
	}
	if err := db.Save(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *service) DeleteCategory(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.InventoryCategory{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
