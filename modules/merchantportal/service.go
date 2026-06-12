package merchantportal

import (
	"errors"
	"time"

	"backend/global"
	"backend/models"
	"backend/modules/crm"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service defines merchant portal operations.
type Service interface {
	Login(db *gorm.DB, email, password string) (*MerchantLoginResponse, error)
	ListTickets(db *gorm.DB, merchantID string) ([]models.Ticket, error)
	GetTicket(db *gorm.DB, merchantID, id string) (*models.Ticket, error)
	CreateTicket(db *gorm.DB, merchantID string, req CreateTicketRequest) (*models.Ticket, error)
	UpdateTicketStatus(db *gorm.DB, merchantID, id, status string) (*models.Ticket, error)
	AddComment(db *gorm.DB, merchantID, ticketID, body string) (*models.TicketEvent, error)
}

type service struct {
	app    *global.App
	crmSvc crm.Service
}

// NewService creates a new merchant portal service.
// crmSvc is used for auto-routing newly created tickets.
func NewService(app *global.App, crmSvc crm.Service) Service {
	return &service{app: app, crmSvc: crmSvc}
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
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	CategoryID  string `json:"category_id"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status"`
}

type AddCommentRequest struct {
	Body string `json:"body"`
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

	return &MerchantLoginResponse{
		AccessToken:  token,
		MerchantID:   merchant.ID.String(),
		MerchantName: merchant.Name,
		Email:        merchant.PortalEmail,
	}, nil
}

func (s *service) ListTickets(db *gorm.DB, merchantID string) ([]models.Ticket, error) {
	var tickets []models.Ticket
	err := db.Preload("Category").
		Where("merchant_id = ? AND is_trashed = false", merchantID).
		Order("created_at DESC").
		Find(&tickets).Error
	return tickets, err
}

func (s *service) GetTicket(db *gorm.DB, merchantID, id string) (*models.Ticket, error) {
	var ticket models.Ticket
	err := db.Preload("Category").
		Preload("Events", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_internal = false").Order("created_at ASC")
		}).
		Where("merchant_id = ? AND id = ? AND is_trashed = false", merchantID, id).
		First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (s *service) CreateTicket(db *gorm.DB, merchantID string, req CreateTicketRequest) (*models.Ticket, error) {
	if req.Subject == "" {
		return nil, errors.New("subject is required")
	}

	merchantUID, err := uuid.Parse(merchantID)
	if err != nil {
		return nil, errors.New("invalid merchant_id")
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	ticket := models.Ticket{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: s.app.Config.DefaultCompanyID},
		Kind:             "support",
		Source:           "merchant_portal",
		MerchantID:       &merchantUID,
		Subject:          req.Subject,
		Description:      req.Description,
		Priority:         priority,
		Status:           "open",
	}

	if req.CategoryID != "" {
		if uid, err := uuid.Parse(req.CategoryID); err == nil {
			ticket.CategoryID = &uid
		}
	}

	// Assign a sequential ticket number within the company
	var maxNum int64
	db.Model(&models.Ticket{}).Where("company_id = ?", s.app.Config.DefaultCompanyID).
		Select("COALESCE(MAX(ticket_number), 0)").Scan(&maxNum)
	ticket.TicketNumber = maxNum + 1

	if err := db.Create(&ticket).Error; err != nil {
		return nil, err
	}

	// Apply routing rules and SLA deadline
	s.crmSvc.ApplyAutoRoute(db, &ticket)
	if ticket.AssignedToTeamID != nil || ticket.AssignedToUserID != nil || ticket.SLADeadline != nil {
		db.Save(&ticket)
	}

	// Write creation event (actor is the merchant)
	event := models.TicketEvent{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: ticket.CompanyID},
		TicketID:         ticket.ID,
		EventType:        "creation",
		ActorID:          merchantUID,
		ActorType:        "merchant",
		Body:             "Ticket submitted via merchant portal",
		IsInternal:       false,
	}
	db.Create(&event)

	return &ticket, nil
}

func (s *service) UpdateTicketStatus(db *gorm.DB, merchantID, id, status string) (*models.Ticket, error) {
	// Merchants may only close their own tickets
	allowedStatuses := map[string]bool{"closed": true}
	if !allowedStatuses[status] {
		return nil, errors.New("merchants may only set status to: closed")
	}

	var ticket models.Ticket
	if err := db.Where("merchant_id = ? AND id = ? AND is_trashed = false", merchantID, id).
		First(&ticket).Error; err != nil {
		return nil, err
	}

	ticket.Status = status
	if err := db.Save(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (s *service) AddComment(db *gorm.DB, merchantID, ticketID, body string) (*models.TicketEvent, error) {
	if body == "" {
		return nil, errors.New("comment body is required")
	}

	merchantUID, err := uuid.Parse(merchantID)
	if err != nil {
		return nil, errors.New("invalid merchant_id")
	}

	// Verify the ticket belongs to this merchant
	var ticket models.Ticket
	if err := db.Where("merchant_id = ? AND id = ? AND is_trashed = false", merchantID, ticketID).
		First(&ticket).Error; err != nil {
		return nil, err
	}

	ticketUID, err := uuid.Parse(ticketID)
	if err != nil {
		return nil, errors.New("invalid ticket_id")
	}

	event := models.TicketEvent{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: ticket.CompanyID},
		TicketID:         ticketUID,
		EventType:        "comment",
		ActorID:          merchantUID,
		ActorType:        "merchant",
		Body:             body,
		IsInternal:       false,
	}

	if err := db.Create(&event).Error; err != nil {
		return nil, err
	}

	// Reopen ticket if it was closed/resolved
	if ticket.Status == "resolved" || ticket.Status == "closed" {
		db.Model(&ticket).Update("status", "open")
	}

	return &event, nil
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
