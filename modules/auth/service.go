package auth

import (
	"backend/global"
	"backend/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service defines authentication operations.
type Service interface {
	Login(db *gorm.DB, email, password string) (*LoginResponse, error)
	Register(db *gorm.DB, email, password, firstName, lastName string) (*LoginResponse, error)
	RefreshToken(db *gorm.DB, refreshToken string) (*LoginResponse, error)
	RequestPasswordReset(db *gorm.DB, email string) error
	ResetPassword(db *gorm.DB, token, newPassword string) error
}

// service implements Service.
type service struct {
	app *global.App
}

// NewService creates a new auth service.
func NewService(app *global.App) Service {
	return &service{app: app}
}

// Login authenticates a user by email and password within a company context.
// Retrieves user with role and permissions, verifies password, and generates JWT with company_id.
func (s *service) Login(db *gorm.DB, email, password string) (*LoginResponse, error) {
	s.app.Logger.Infof("Login attempt: %s", email)
	var user models.User

	// Query user by email
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.app.Logger.Warnf("User not found: %s", email)
			return nil, errors.New("invalid credentials")
		}
		s.app.Logger.Errorf("Database error during login: %v", err)
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.app.Logger.Warnf("Password mismatch for user: %s", email)
		return nil, errors.New("invalid credentials")
	}

	// Check if account is active
	if !user.IsActive {
		s.app.Logger.Warnf("Account inactive for user: %s", email)
		return nil, errors.New("account is inactive")
	}

	// Generate JWT with company_id
	token, err := s.generateJWT(&user, user.CompanyID)
	if err != nil {
		s.app.Logger.Errorf("Failed to generate JWT: %v", err)
		return nil, err
	}

	refreshToken, err := s.generateRefreshJWT(&user, user.CompanyID)
	if err != nil {
		s.app.Logger.Errorf("Failed to generate refresh JWT: %v", err)
		return nil, err
	}

	s.app.Logger.Infof("Login successful for user: %s", email)
	return &LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		User: UserDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			RoleSlug:  user.RoleSlug,
		},
	}, nil
}

// Register creates a new user in the company with default role.
func (s *service) Register(db *gorm.DB, email, password, firstName, lastName string) (*LoginResponse, error) {
	s.app.Logger.Infof("Register attempt: %s", email)

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		s.app.Logger.Warnf("User already exists: %s", email)
		return nil, errors.New("email already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.app.Logger.Errorf("Failed to hash password: %v", err)
		return nil, errors.New("registration failed")
	}

	// Create user with default role
	user := models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		RoleSlug:     "user", // Default role
		IsActive:     true,
	}

	if err := db.Create(&user).Error; err != nil {
		s.app.Logger.Errorf("Failed to create user: %v", err)
		return nil, errors.New("registration failed")
	}

	// Generate JWT (no need to preload role, just use RoleSlug)
	token, err := s.generateJWT(&user, user.CompanyID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshJWT(&user, user.CompanyID)
	if err != nil {
		return nil, err
	}

	s.app.Logger.Infof("Registration successful: %s", email)
	return &LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		User: UserDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			RoleSlug:  user.RoleSlug,
		},
	}, nil
}

// RefreshToken validates a refresh token and returns a new access token.
func (s *service) RefreshToken(db *gorm.DB, refreshToken string) (*LoginResponse, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.app.Config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if claims["type"] != "refresh" {
		return nil, errors.New("not a refresh token")
	}

	userID, _ := claims["sub"].(string)
	companyID, _ := claims["company_id"].(string)

	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	newAccessToken, err := s.generateJWT(&user, companyID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshJWT(&user, companyID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User: UserDTO{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			RoleSlug:  user.RoleSlug,
		},
	}, nil
}

// RequestPasswordReset initiates a password reset (stub).
func (s *service) RequestPasswordReset(db *gorm.DB, email string) error {
	s.app.Logger.Infof("Password reset requested: %s", email)
	// Stub: Generate reset token, send email via app.Email
	return nil
}

// ResetPassword completes a password reset (stub).
func (s *service) ResetPassword(db *gorm.DB, token, newPassword string) error {
	s.app.Logger.Infof("Password reset via token")
	// Stub: Validate token, update password
	return nil
}

// generateJWT creates an access JWT with company context (7-day expiry).
func (s *service) generateJWT(user *models.User, companyID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":        user.ID.String(),
		"email":      user.Email,
		"role_slug":  user.RoleSlug,
		"company_id": companyID,
		"type":       "access",
		"exp":        time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.app.Config.JWTSecret))
}

// generateRefreshJWT creates a refresh JWT (30-day expiry).
func (s *service) generateRefreshJWT(user *models.User, companyID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":        user.ID.String(),
		"company_id": companyID,
		"type":       "refresh",
		"exp":        time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.app.Config.JWTSecret))
}
