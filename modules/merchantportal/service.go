package merchantportal

import (
	"backend/global"
	"backend/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service defines merchant portal operations.
type Service interface {
	Login(db *gorm.DB, email, password string) (*MerchantLoginResponse, error)
	ListTickets(db *gorm.DB, merchantID string) ([]models.SupportTicket, error)
	GetTicket(db *gorm.DB, merchantID, id string) (*models.SupportTicket, error)
	CreateTicket(db *gorm.DB, merchantID string, req CreateTicketRequest) (*models.SupportTicket, error)
	UpdateTicketStatus(db *gorm.DB, merchantID, id, status string) (*models.SupportTicket, error)
}

type service struct {
	app *global.App
}

// NewService creates a new merchant portal service.
func NewService(app *global.App) Service {
	return &service{app: app}
}

// --- DTOs ---

type MerchantLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MerchantLoginResponse struct {
	AccessToken  string `json:"access_token"`
	MerchantID   string `json:"merchant_id"`
	MerchantName string `json:"merchant_name"`
	Email        string `json:"email"`
}

type CreateTicketRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status"`
}

// --- Implementations ---

func (s *service) Login(db *gorm.DB, email, password string) (*MerchantLoginResponse, error) {
	s.app.Logger.Infof("Merchant portal login attempt: %s", email)

	var merchant models.Merchant
	if err := db.Where("portal_email = ? AND portal_enabled = true", email).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(merchant.PortalPasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateMerchantJWT(&merchant)
	if err != nil {
		s.app.Logger.Errorf("Failed to generate merchant JWT: %v", err)
		return nil, err
	}

	s.app.Logger.Infof("Merchant portal login successful: %s", email)
	return &MerchantLoginResponse{
		AccessToken:  token,
		MerchantID:   merchant.ID.String(),
		MerchantName: merchant.Name,
		Email:        merchant.PortalEmail,
	}, nil
}

func (s *service) ListTickets(db *gorm.DB, merchantID string) ([]models.SupportTicket, error) {
	var tickets []models.SupportTicket
	if err := db.Where("merchant_id = ? AND is_trashed = false", merchantID).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (s *service) GetTicket(db *gorm.DB, merchantID, id string) (*models.SupportTicket, error) {
	var ticket models.SupportTicket
	if err := db.Where("merchant_id = ? AND id = ?", merchantID, id).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (s *service) CreateTicket(db *gorm.DB, merchantID string, req CreateTicketRequest) (*models.SupportTicket, error) {
	if req.Subject == "" {
		return nil, errors.New("subject is required")
	}

	merchantUID, err := uuid.Parse(merchantID)
	if err != nil {
		return nil, errors.New("invalid merchant_id")
	}

	ticket := models.SupportTicket{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: s.app.Config.DefaultCompanyID},
		MerchantID:       &merchantUID,
		Subject:          req.Subject,
		Body:             req.Body,
		Status:           "open",
		CreatedBy:        merchantUID,
	}

	if err := db.Create(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (s *service) UpdateTicketStatus(db *gorm.DB, merchantID, id, status string) (*models.SupportTicket, error) {
	var ticket models.SupportTicket
	if err := db.Where("merchant_id = ? AND id = ?", merchantID, id).First(&ticket).Error; err != nil {
		return nil, err
	}
	ticket.Status = status
	if err := db.Save(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (s *service) generateMerchantJWT(merchant *models.Merchant) (string, error) {
	claims := jwt.MapClaims{
		"sub":         merchant.ID.String(),
		"email":       merchant.PortalEmail,
		"merchant_id": merchant.ID.String(),
		"user_type":   "merchant",
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.app.Config.JWTSecret))
}
