package incidents

import (
	"errors"
	"time"

	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	// Incidents
	ListIncidents(db *gorm.DB, companyID string) ([]models.Incident, error)
	GetIncident(db *gorm.DB, companyID, id string) (*models.Incident, error)
	CreateIncident(db *gorm.DB, companyID string, req CreateIncidentRequest) (*models.Incident, error)
	UpdateIncident(db *gorm.DB, companyID, id string, req UpdateIncidentRequest) (*models.Incident, error)
	ChangeIncidentStatus(db *gorm.DB, companyID, id, status string) (*models.Incident, error)
	AssignIncident(db *gorm.DB, companyID, id, assignee string) (*models.Incident, error)
	NotifyIncident(db *gorm.DB, companyID, id string) (*models.Incident, error)

	// Support tickets
	ListTickets(db *gorm.DB, companyID string) ([]models.SupportTicket, error)
	GetTicket(db *gorm.DB, companyID, id string) (*models.SupportTicket, error)
	CreateTicket(db *gorm.DB, companyID string, req CreateTicketRequest) (*models.SupportTicket, error)
	UpdateTicket(db *gorm.DB, companyID, id string, req UpdateTicketRequest) (*models.SupportTicket, error)
	TrashTicket(db *gorm.DB, companyID, id string) error
	RestoreTicket(db *gorm.DB, companyID, id string) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

// DTOs
type CreateIncidentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ReportedBy  string `json:"reported_by"`
}

type UpdateIncidentRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
}

type CreateTicketRequest struct {
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	CreatedBy string `json:"created_by"`
}

type UpdateTicketRequest struct {
	Subject *string `json:"subject"`
	Body    *string `json:"body"`
}

// Incidents implementation
func (s *service) ListIncidents(db *gorm.DB, companyID string) ([]models.Incident, error) {
	var list []models.Incident
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *service) GetIncident(db *gorm.DB, companyID, id string) (*models.Incident, error) {
	var it models.Incident
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) CreateIncident(db *gorm.DB, companyID string, req CreateIncidentRequest) (*models.Incident, error) {
	reportedBy, _ := uuid.Parse(req.ReportedBy)
	now := time.Now()
	it := models.Incident{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Title:            req.Title,
		Description:      req.Description,
		Type:             req.Type,
		Status:           "new",
		ReportedBy:       reportedBy,
		ReportedAt:       &now,
	}
	if err := db.Create(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) UpdateIncident(db *gorm.DB, companyID, id string, req UpdateIncidentRequest) (*models.Incident, error) {
	var it models.Incident
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Title != nil {
		it.Title = *req.Title
	}
	if req.Description != nil {
		it.Description = *req.Description
	}
	if req.Type != nil {
		it.Type = *req.Type
	}
	if err := db.Save(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) ChangeIncidentStatus(db *gorm.DB, companyID, id, status string) (*models.Incident, error) {
	var it models.Incident
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	it.Status = status
	if err := db.Save(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) AssignIncident(db *gorm.DB, companyID, id, assignee string) (*models.Incident, error) {
	var it models.Incident
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(assignee)
	if err != nil {
		return nil, errors.New("invalid assignee id")
	}
	it.AssignedTo = uid
	it.Status = "assigned"
	if err := db.Save(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

func (s *service) NotifyIncident(db *gorm.DB, companyID, id string) (*models.Incident, error) {
	var it models.Incident
	if err := db.First(&it, "id = ?", id).Error; err != nil {
		return nil, err
	}
	it.Notified = true
	if err := db.Save(&it).Error; err != nil {
		return nil, err
	}
	return &it, nil
}

// Tickets implementation
func (s *service) ListTickets(db *gorm.DB, companyID string) ([]models.SupportTicket, error) {
	var list []models.SupportTicket
	if err := db.Where("is_trashed = false").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *service) GetTicket(db *gorm.DB, companyID, id string) (*models.SupportTicket, error) {
	var t models.SupportTicket
	if err := db.First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *service) CreateTicket(db *gorm.DB, companyID string, req CreateTicketRequest) (*models.SupportTicket, error) {
	cb, _ := uuid.Parse(req.CreatedBy)
	t := models.SupportTicket{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Subject:          req.Subject,
		Body:             req.Body,
		Status:           "open",
		CreatedBy:        cb,
	}
	if err := db.Create(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *service) UpdateTicket(db *gorm.DB, companyID, id string, req UpdateTicketRequest) (*models.SupportTicket, error) {
	var t models.SupportTicket
	if err := db.First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Subject != nil {
		t.Subject = *req.Subject
	}
	if req.Body != nil {
		t.Body = *req.Body
	}
	if err := db.Save(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *service) TrashTicket(db *gorm.DB, companyID, id string) error {
	if err := db.Model(&models.SupportTicket{}).Where("id = ?", id).Update("is_trashed", true).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) RestoreTicket(db *gorm.DB, companyID, id string) error {
	if err := db.Model(&models.SupportTicket{}).Where("id = ?", id).Update("is_trashed", false).Error; err != nil {
		return err
	}
	return nil
}
