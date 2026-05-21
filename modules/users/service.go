package users

import (
	"backend/global"

	"gorm.io/gorm"
)

// Service defines user management operations.
type Service interface {
	Create(db *gorm.DB, email, password string) error
	Get(db *gorm.DB, id string) error
	List(db *gorm.DB) error
	Update(db *gorm.DB, id string) error
	Delete(db *gorm.DB, id string) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

func (s *service) Create(db *gorm.DB, email, password string) error {
	return nil
}

func (s *service) Get(db *gorm.DB, id string) error {
	return nil
}

func (s *service) List(db *gorm.DB) error {
	return nil
}

func (s *service) Update(db *gorm.DB, id string) error {
	return nil
}

func (s *service) Delete(db *gorm.DB, id string) error {
	return nil
}
