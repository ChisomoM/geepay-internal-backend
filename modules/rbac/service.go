package rbac

import (
	"backend/global"
	"backend/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateRoleRequest is the payload for creating a role.
type CreateRoleRequest struct {
	Name          string   `json:"name"`
	Slug          string   `json:"slug"`
	PermissionIDs []string `json:"permission_ids"`
}

// Service defines RBAC operations.
type Service interface {
	ListPermissions(db *gorm.DB, companyID string) ([]models.Permission, error)
	ListRoles(db *gorm.DB, companyID string) ([]models.Role, error)
	CreateRole(db *gorm.DB, companyID string, req CreateRoleRequest) (*models.Role, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

func (s *service) ListPermissions(db *gorm.DB, companyID string) ([]models.Permission, error) {
	var perms []models.Permission
	if err := db.Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *service) ListRoles(db *gorm.DB, companyID string) ([]models.Role, error) {
	var roles []models.Role
	if err := db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *service) CreateRole(db *gorm.DB, companyID string, req CreateRoleRequest) (*models.Role, error) {
	if req.Name == "" || req.Slug == "" {
		return nil, errors.New("name and slug required")
	}

	role := models.Role{
		CompanyID:   companyID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: "",
	}

	// If permission IDs provided, fetch them
	if len(req.PermissionIDs) > 0 {
		var perms []models.Permission
		for _, pid := range req.PermissionIDs {
			// Validate UUID
			if _, err := uuid.Parse(pid); err != nil {
				continue
			}
			var p models.Permission
			if err := db.First(&p, "id = ?", pid).Error; err == nil {
				perms = append(perms, p)
			}
		}
		role.Permissions = perms
	}

	if err := db.Create(&role).Error; err != nil {
		s.app.Logger.Errorf("Failed to create role: %v", err)
		return nil, err
	}
	return &role, nil
}
