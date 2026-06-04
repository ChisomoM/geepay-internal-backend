package users

import (
	"errors"

	"backend/global"
	"backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateUserRequest is the payload for creating users.
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleSlug  string `json:"role_slug"`
}

// UpdateUserRequest is the payload for updating users.
type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Password  *string `json:"password"`
	RoleSlug  *string `json:"role_slug"`
	IsActive  *bool   `json:"is_active"`
}

// Service defines user management operations.
type Service interface {
	Create(db *gorm.DB, companyID string, req CreateUserRequest) (*models.User, error)
	Get(db *gorm.DB, companyID, id string) (*models.User, error)
	List(db *gorm.DB, companyID string) ([]models.User, error)
	Update(db *gorm.DB, companyID, id string, req UpdateUserRequest) (*models.User, error)
	Delete(db *gorm.DB, companyID, id string) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

func (s *service) Create(db *gorm.DB, companyID string, req CreateUserRequest) (*models.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.app.Logger.Errorf("Failed to hash password: %v", err)
		return nil, err
	}

	user := models.User{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Email:            req.Email,
		PasswordHash:     string(hashed),
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		RoleSlug:         req.RoleSlug,
		IsActive:         true,
	}

	if err := db.Create(&user).Error; err != nil {
		s.app.Logger.Errorf("Failed to create user: %v", err)
		return nil, err
	}
	return &user, nil
}

func (s *service) Get(db *gorm.DB, companyID, id string) (*models.User, error) {
	var user models.User
	if err := db.Preload("PermissionOverrides.Permission").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *service) List(db *gorm.DB, companyID string) ([]models.User, error) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *service) Update(db *gorm.DB, companyID, id string, req UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.RoleSlug != nil {
		user.RoleSlug = *req.RoleSlug
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hashed)
	}

	if err := db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *service) Delete(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
