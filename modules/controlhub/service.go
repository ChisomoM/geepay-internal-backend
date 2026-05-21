package controlhub

import (
	"backend/global"
	"backend/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service defines ControlHub operations.
type Service interface {
	LoginCompanyAdmin(db *gorm.DB, email, password string) (*CompanyAdminLoginResponse, error)
	RefreshToken(db *gorm.DB, refreshToken string) (*CompanyAdminLoginResponse, error)
}

// service implements Service.
type service struct {
	app *global.App
}

// NewService creates a new ControlHub service.
func NewService(app *global.App) Service {
	return &service{app: app}
}

// LoginCompanyAdmin authenticates a company admin (platform-level) and returns a JWT.
// This is separate from tenant-level user authentication.
func (s *service) LoginCompanyAdmin(db *gorm.DB, email, password string) (*CompanyAdminLoginResponse, error) {
	s.app.Logger.Infof("ControlHub login attempt: %s", email)
	var admin models.CompanyAdmin

	// Query unscoped DB (platform-level, no tenant scope)
	err := db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.app.Logger.Warnf("CompanyAdmin not found: %s", email)
			return nil, errors.New("invalid credentials")
		}
		s.app.Logger.Errorf("Database error during ControlHub login: %v", err)
		return nil, err
	}

	// Check if admin is active
	if !admin.IsActive {
		s.app.Logger.Warnf("CompanyAdmin account inactive: %s", email)
		return nil, errors.New("account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		s.app.Logger.Warnf("Password mismatch for CompanyAdmin: %s", email)
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT with controlhub claims
	token, err := s.generateControlHubJWT(&admin)
	if err != nil {
		s.app.Logger.Errorf("Failed to generate ControlHub JWT: %v", err)
		return nil, err
	}

	s.app.Logger.Infof("ControlHub login successful: %s", email)
	return &CompanyAdminLoginResponse{
		Token:    token,
		Email:    admin.Email,
		Role:     "controlhub",
		UserType: "controlhub_admin",
	}, nil
}

// RefreshToken validates a refresh token and returns a new access token (stub).
func (s *service) RefreshToken(db *gorm.DB, refreshToken string) (*CompanyAdminLoginResponse, error) {
	s.app.Logger.Infof("ControlHub token refresh requested")
	// Stub: Parse refresh token, validate, regenerate access token
	return &CompanyAdminLoginResponse{}, nil
}

// generateControlHubJWT creates a JWT for platform-level access.
func (s *service) generateControlHubJWT(admin *models.CompanyAdmin) (string, error) {
	claims := jwt.MapClaims{
		"sub":       admin.ID.String(),
		"email":     admin.Email,
		"role":      "controlhub",                         // Platform-level role
		"level":     "company_admin",                      // Access level
		"user_type": "controlhub_admin",                   // Explicit type for frontend
		"exp":       time.Now().Add(4 * time.Hour).Unix(), // 4 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.app.Config.JWTSecret))
}
