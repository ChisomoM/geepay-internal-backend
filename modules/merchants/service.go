package merchants

import (
	"bytes"
	"encoding/csv"
	"errors"
	"time"

	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	// Merchants
	ListMerchants(db *gorm.DB, companyID string) ([]models.Merchant, error)
	GetMerchant(db *gorm.DB, companyID, id string) (*models.Merchant, error)
	CreateMerchant(db *gorm.DB, companyID string, req CreateMerchantRequest) (*models.Merchant, error)
	UpdateMerchant(db *gorm.DB, companyID, id string, req UpdateMerchantRequest) (*models.Merchant, error)
	DeleteMerchant(db *gorm.DB, companyID, id string) error
	ExportMerchantsCSV(db *gorm.DB, companyID string) ([]byte, error)
	EnableMerchantPortal(db *gorm.DB, merchantID string, req EnableMerchantPortalRequest) (*models.Merchant, error)

	// Statements
	ListStatements(db *gorm.DB, companyID, merchantID string) ([]models.MerchantStatement, error)
	GetStatement(db *gorm.DB, companyID, id string) (*models.MerchantStatement, error)
	CreateStatement(db *gorm.DB, companyID string, req CreateStatementRequest) (*models.MerchantStatement, error)
	UpdateStatement(db *gorm.DB, companyID, id string, req UpdateStatementRequest) (*models.MerchantStatement, error)
	DeleteStatement(db *gorm.DB, companyID, id string) error
	ExportStatementsCSV(db *gorm.DB, companyID string) ([]byte, error)
}

// EnableMerchantPortalRequest holds the credentials and toggle for merchant portal access.
type EnableMerchantPortalRequest struct {
	PortalEmail    string `json:"portal_email"`
	PortalPassword string `json:"portal_password"`
	PortalEnabled  bool   `json:"portal_enabled"`
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateMerchantRequest struct {
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	ContactNumber     string     `json:"contact_number"`
	BusinessType      string     `json:"business_type"`
	Notes             string     `json:"notes"`
	IntegrationStatus string     `json:"integration_status"`
	OnboardedOn       *time.Time `json:"onboarded_on"`
}

type UpdateMerchantRequest struct {
	Name              *string    `json:"name"`
	Email             *string    `json:"email"`
	ContactNumber     *string    `json:"contact_number"`
	BusinessType      *string    `json:"business_type"`
	Notes             *string    `json:"notes"`
	IntegrationStatus *string    `json:"integration_status"`
	OnboardedOn       *time.Time `json:"onboarded_on"`
}

type CreateStatementRequest struct {
	MerchantID   string `json:"merchant_id"`
	DriveLink    string `json:"drive_link"`
	DocumentName string `json:"document_name"`
}

type UpdateStatementRequest struct {
	DriveLink    *string `json:"drive_link"`
	DocumentName *string `json:"document_name"`
}

// --- Merchants ---

func (s *service) ListMerchants(db *gorm.DB, companyID string) ([]models.Merchant, error) {
	var items []models.Merchant
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetMerchant(db *gorm.DB, companyID, id string) (*models.Merchant, error) {
	var item models.Merchant
	if err := db.Preload("Statements").First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateMerchant(db *gorm.DB, companyID string, req CreateMerchantRequest) (*models.Merchant, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	item := models.Merchant{
		CompanyBaseModel:  models.CompanyBaseModel{CompanyID: companyID},
		Name:              req.Name,
		Email:             req.Email,
		ContactNumber:     req.ContactNumber,
		BusinessType:      req.BusinessType,
		Notes:             req.Notes,
		IntegrationStatus: req.IntegrationStatus,
		OnboardedOn:       req.OnboardedOn,
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) UpdateMerchant(db *gorm.DB, companyID, id string, req UpdateMerchantRequest) (*models.Merchant, error) {
	var item models.Merchant
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Email != nil {
		item.Email = *req.Email
	}
	if req.ContactNumber != nil {
		item.ContactNumber = *req.ContactNumber
	}
	if req.BusinessType != nil {
		item.BusinessType = *req.BusinessType
	}
	if req.Notes != nil {
		item.Notes = *req.Notes
	}
	if req.IntegrationStatus != nil {
		item.IntegrationStatus = *req.IntegrationStatus
	}
	if req.OnboardedOn != nil {
		item.OnboardedOn = req.OnboardedOn
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) DeleteMerchant(db *gorm.DB, companyID, id string) error {
	return db.Delete(&models.Merchant{}, "id = ?", id).Error
}

func (s *service) EnableMerchantPortal(db *gorm.DB, merchantID string, req EnableMerchantPortalRequest) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := db.First(&merchant, "id = ?", merchantID).Error; err != nil {
		return nil, errors.New("merchant not found")
	}

	merchant.PortalEnabled = req.PortalEnabled

	if req.PortalEmail != "" {
		merchant.PortalEmail = req.PortalEmail
	}

	if req.PortalPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.PortalPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		merchant.PortalPasswordHash = string(hash)
	}

	if err := db.Save(&merchant).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (s *service) ExportMerchantsCSV(db *gorm.DB, companyID string) ([]byte, error) {
	items, err := s.ListMerchants(db, companyID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "Name", "Email", "Contact Number", "Business Type", "Integration Status", "Onboarded On", "Notes", "Created At"})
	for _, item := range items {
		onboardedOn := ""
		if item.OnboardedOn != nil {
			onboardedOn = item.OnboardedOn.Format("2006-01-02")
		}
		_ = w.Write([]string{
			item.ID.String(),
			item.Name,
			item.Email,
			item.ContactNumber,
			item.BusinessType,
			item.IntegrationStatus,
			onboardedOn,
			item.Notes,
			item.CreatedAt.Format("2006-01-02"),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- Merchant Statements ---

func (s *service) ListStatements(db *gorm.DB, companyID, merchantID string) ([]models.MerchantStatement, error) {
	var items []models.MerchantStatement
	if err := db.Where("merchant_id = ?", merchantID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetStatement(db *gorm.DB, companyID, id string) (*models.MerchantStatement, error) {
	var item models.MerchantStatement
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateStatement(db *gorm.DB, companyID string, req CreateStatementRequest) (*models.MerchantStatement, error) {
	if req.DriveLink == "" {
		return nil, errors.New("drive_link is required")
	}
	merchantUID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		return nil, errors.New("invalid merchant_id")
	}
	item := models.MerchantStatement{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		MerchantID:       merchantUID,
		DriveLink:        req.DriveLink,
		DocumentName:     req.DocumentName,
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) UpdateStatement(db *gorm.DB, companyID, id string, req UpdateStatementRequest) (*models.MerchantStatement, error) {
	var item models.MerchantStatement
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.DriveLink != nil {
		item.DriveLink = *req.DriveLink
	}
	if req.DocumentName != nil {
		item.DocumentName = *req.DocumentName
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) DeleteStatement(db *gorm.DB, companyID, id string) error {
	return db.Delete(&models.MerchantStatement{}, "id = ?", id).Error
}

func (s *service) ExportStatementsCSV(db *gorm.DB, companyID string) ([]byte, error) {
	var items []models.MerchantStatement
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "Merchant ID", "Document Name", "Drive Link", "Created At"})
	for _, item := range items {
		_ = w.Write([]string{
			item.ID.String(),
			item.MerchantID.String(),
			item.DocumentName,
			item.DriveLink,
			item.CreatedAt.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
