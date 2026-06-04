package MODULENAME

import (
	"backend/global"

	"gorm.io/gorm"
)

// Service defines the business logic for this module.
// All methods receive *gorm.DB as the first parameter (company-scoped by middleware).
type Service interface {
	Create(db *gorm.DB, req CreateRequest) (*ResponseDTO, error)
	Get(db *gorm.DB, id string) (*ResponseDTO, error)
	List(db *gorm.DB) ([]ResponseDTO, error)
	Update(db *gorm.DB, id string, req UpdateRequest) (*ResponseDTO, error)
	Delete(db *gorm.DB, id string) error
}

// service is the unexported implementation of Service.
type service struct {
	app *global.App
}

// NewService creates a new service instance.
// Receives *global.App which holds infrastructure (config, logger, email, upload).
func NewService(app *global.App) Service {
	return &service{
		app: app,
	}
}

// Create implements Service.Create
func (s *service) Create(db *gorm.DB, req CreateRequest) (*ResponseDTO, error) {
	// Business logic here
	// The db parameter is already company-scoped by CompanyMiddleware
	s.app.Logger.Infof("Creating entity: %v", req)
	return &ResponseDTO{}, nil
}

// Get implements Service.Get
func (s *service) Get(db *gorm.DB, id string) (*ResponseDTO, error) {
	s.app.Logger.Infof("Getting entity: %s", id)
	return &ResponseDTO{}, nil
}

// List implements Service.List
func (s *service) List(db *gorm.DB) ([]ResponseDTO, error) {
	s.app.Logger.Info("Listing entities")
	return []ResponseDTO{}, nil
}

// Update implements Service.Update
func (s *service) Update(db *gorm.DB, id string, req UpdateRequest) (*ResponseDTO, error) {
	s.app.Logger.Infof("Updating entity: %s", id)
	return &ResponseDTO{}, nil
}

// Delete implements Service.Delete
func (s *service) Delete(db *gorm.DB, id string) error {
	s.app.Logger.Infof("Deleting entity: %s", id)
	return nil
}
