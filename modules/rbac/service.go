package rbac

import (
	"backend/global"

	"gorm.io/gorm"
)

// Service defines RBAC operations.
type Service interface {
	ListPermissions(db *gorm.DB) (interface{}, error)
	ListRoles(db *gorm.DB) (interface{}, error)
	CreateRole(db *gorm.DB) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

func (s *service) ListPermissions(db *gorm.DB) (interface{}, error) {
	return nil, nil
}

func (s *service) ListRoles(db *gorm.DB) (interface{}, error) {
	return nil, nil
}

func (s *service) CreateRole(db *gorm.DB) error {
	return nil
}
