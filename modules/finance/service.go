package finance

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"time"

	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service defines finance-related business operations.
type Service interface {
	// Budgets
	ListBudgets(db *gorm.DB, companyID string) ([]models.BudgetAndLicense, error)
	GetBudget(db *gorm.DB, companyID, id string) (*models.BudgetAndLicense, error)
	CreateBudget(db *gorm.DB, companyID string, req CreateBudgetRequest) (*models.BudgetAndLicense, error)
	UpdateBudget(db *gorm.DB, companyID, id string, req UpdateBudgetRequest) (*models.BudgetAndLicense, error)
	DeleteBudget(db *gorm.DB, companyID, id string) error
	RenewBudget(db *gorm.DB, companyID, id string) (*models.BudgetAndLicense, error)
	ExportBudgetsCSV(db *gorm.DB, companyID string) ([]byte, error)

	// Licenses
	ListLicenses(db *gorm.DB, companyID string) ([]models.License, error)
	GetLicense(db *gorm.DB, companyID, id string) (*models.License, error)
	CreateLicense(db *gorm.DB, companyID string, req CreateLicenseRequest) (*models.License, error)
	UpdateLicense(db *gorm.DB, companyID, id string, req UpdateLicenseRequest) (*models.License, error)
	DeleteLicense(db *gorm.DB, companyID, id string) error
	RenewLicense(db *gorm.DB, companyID, id string) (*models.License, error)
	ExportLicensesCSV(db *gorm.DB, companyID string) ([]byte, error)

	// Statutories
	ListStatutories(db *gorm.DB, companyID string) ([]models.Statutory, error)
	GetStatutory(db *gorm.DB, companyID, id string) (*models.Statutory, error)
	CreateStatutory(db *gorm.DB, companyID string, req CreateStatutoryRequest) (*models.Statutory, error)
	UpdateStatutory(db *gorm.DB, companyID, id string, req UpdateStatutoryRequest) (*models.Statutory, error)
	DeleteStatutory(db *gorm.DB, companyID, id string) error
	MarkStatutoryPaid(db *gorm.DB, companyID, id string) (*models.Statutory, error)

	// Salary advances
	ListSalaryAdvances(db *gorm.DB, companyID string) ([]models.SalaryAdvance, error)
	GetSalaryAdvance(db *gorm.DB, companyID, id string) (*models.SalaryAdvance, error)
	CreateSalaryAdvance(db *gorm.DB, companyID string, req CreateSalaryAdvanceRequest) (*models.SalaryAdvance, error)
	UpdateSalaryAdvance(db *gorm.DB, companyID, id string, req UpdateSalaryAdvanceRequest) (*models.SalaryAdvance, error)
	DeleteSalaryAdvance(db *gorm.DB, companyID, id string) error
	ApproveSalaryAdvance(db *gorm.DB, companyID, id string) (*models.SalaryAdvance, error)
	DeductSalaryAdvance(db *gorm.DB, companyID, id string, months int) (*models.SalaryAdvance, error)
	ExportSalaryAdvancesCSV(db *gorm.DB, companyID string) ([]byte, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

// Request/Update DTOs
type CreateBudgetRequest struct {
	Name              string     `json:"name"`
	RenewalDate       *time.Time `json:"renewal_date"`
	ActualAmount      float64    `json:"actual_amount"`
	PurchaseFrequency string     `json:"purchase_frequency"`
	LicenseID         string     `json:"license_id"`
}

type UpdateBudgetRequest struct {
	Name              *string    `json:"name"`
	RenewalDate       *time.Time `json:"renewal_date"`
	ActualAmount      *float64   `json:"actual_amount"`
	PurchaseFrequency *string    `json:"purchase_frequency"`
	LicenseID         *string    `json:"license_id"`
}

type CreateLicenseRequest struct {
	Name        string     `json:"name"`
	Provider    string     `json:"provider"`
	RenewalDate *time.Time `json:"renewal_date"`
}

type UpdateLicenseRequest struct {
	Name        *string    `json:"name"`
	Provider    *string    `json:"provider"`
	RenewalDate *time.Time `json:"renewal_date"`
	IsActive    *bool      `json:"is_active"`
}

type CreateStatutoryRequest struct {
	Name    string     `json:"name"`
	DueDate *time.Time `json:"due_date"`
	Amount  float64    `json:"amount"`
}

type UpdateStatutoryRequest struct {
	Name    *string    `json:"name"`
	DueDate *time.Time `json:"due_date"`
	Amount  *float64   `json:"amount"`
	IsPaid  *bool      `json:"is_paid"`
}

type CreateSalaryAdvanceRequest struct {
	UserID          string  `json:"user_id"`
	Amount          float64 `json:"amount"`
	DeductionMonths int     `json:"deduction_months"`
}

type UpdateSalaryAdvanceRequest struct {
	Amount          *float64 `json:"amount"`
	Status          *string  `json:"status"`
	DeductionMonths *int     `json:"deduction_months"`
}

// --- Budgets ---
func (s *service) ListBudgets(db *gorm.DB, companyID string) ([]models.BudgetAndLicense, error) {
	var items []models.BudgetAndLicense
	if err := db.Preload("License").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetBudget(db *gorm.DB, companyID, id string) (*models.BudgetAndLicense, error) {
	var item models.BudgetAndLicense
	if err := db.Preload("License").First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateBudget(db *gorm.DB, companyID string, req CreateBudgetRequest) (*models.BudgetAndLicense, error) {
	b := models.BudgetAndLicense{
		CompanyBaseModel:  models.CompanyBaseModel{CompanyID: companyID},
		Name:              req.Name,
		RenewalDate:       req.RenewalDate,
		ActualAmount:      req.ActualAmount,
		PurchaseFrequency: req.PurchaseFrequency,
	}
	if req.LicenseID != "" {
		if uid, err := uuid.Parse(req.LicenseID); err == nil {
			b.LicenseID = uid
		}
	}
	if err := db.Create(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *service) UpdateBudget(db *gorm.DB, companyID, id string, req UpdateBudgetRequest) (*models.BudgetAndLicense, error) {
	var b models.BudgetAndLicense
	if err := db.First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		b.Name = *req.Name
	}
	if req.RenewalDate != nil {
		b.RenewalDate = req.RenewalDate
	}
	if req.ActualAmount != nil {
		b.ActualAmount = *req.ActualAmount
	}
	if req.PurchaseFrequency != nil {
		b.PurchaseFrequency = *req.PurchaseFrequency
	}
	if req.LicenseID != nil {
		if uid, err := uuid.Parse(*req.LicenseID); err == nil {
			b.LicenseID = uid
		}
	}
	if err := db.Save(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *service) DeleteBudget(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.BudgetAndLicense{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) RenewBudget(db *gorm.DB, companyID, id string) (*models.BudgetAndLicense, error) {
	var b models.BudgetAndLicense
	if err := db.First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	b.RenewalDate = &now
	if err := db.Save(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *service) ExportBudgetsCSV(db *gorm.DB, companyID string) ([]byte, error) {
	items, err := s.ListBudgets(db, companyID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"id", "name", "renewal_date", "actual_amount", "purchase_frequency"})
	for _, it := range items {
		rd := ""
		if it.RenewalDate != nil {
			rd = it.RenewalDate.Format(time.RFC3339)
		}
		_ = w.Write([]string{it.ID.String(), it.Name, rd, strconv.FormatFloat(it.ActualAmount, 'f', 2, 64), it.PurchaseFrequency})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- Licenses ---
func (s *service) ListLicenses(db *gorm.DB, companyID string) ([]models.License, error) {
	var items []models.License
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetLicense(db *gorm.DB, companyID, id string) (*models.License, error) {
	var l models.License
	if err := db.First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *service) CreateLicense(db *gorm.DB, companyID string, req CreateLicenseRequest) (*models.License, error) {
	l := models.License{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Name:             req.Name,
		Provider:         req.Provider,
		RenewalDate:      req.RenewalDate,
		IsActive:         true,
	}
	if err := db.Create(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *service) UpdateLicense(db *gorm.DB, companyID, id string, req UpdateLicenseRequest) (*models.License, error) {
	var l models.License
	if err := db.First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		l.Name = *req.Name
	}
	if req.Provider != nil {
		l.Provider = *req.Provider
	}
	if req.RenewalDate != nil {
		l.RenewalDate = req.RenewalDate
	}
	if req.IsActive != nil {
		l.IsActive = *req.IsActive
	}
	if err := db.Save(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *service) DeleteLicense(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.License{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) RenewLicense(db *gorm.DB, companyID, id string) (*models.License, error) {
	var l models.License
	if err := db.First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	l.RenewalDate = &now
	if err := db.Save(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *service) ExportLicensesCSV(db *gorm.DB, companyID string) ([]byte, error) {
	items, err := s.ListLicenses(db, companyID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"id", "name", "provider", "renewal_date", "is_active"})
	for _, it := range items {
		rd := ""
		if it.RenewalDate != nil {
			rd = it.RenewalDate.Format(time.RFC3339)
		}
		_ = w.Write([]string{it.ID.String(), it.Name, it.Provider, rd, strconv.FormatBool(it.IsActive)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- Statutories ---
func (s *service) ListStatutories(db *gorm.DB, companyID string) ([]models.Statutory, error) {
	var items []models.Statutory
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetStatutory(db *gorm.DB, companyID, id string) (*models.Statutory, error) {
	var st models.Statutory
	if err := db.First(&st, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *service) CreateStatutory(db *gorm.DB, companyID string, req CreateStatutoryRequest) (*models.Statutory, error) {
	st := models.Statutory{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Name:             req.Name,
		DueDate:          req.DueDate,
		Amount:           req.Amount,
		IsPaid:           false,
	}
	if err := db.Create(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *service) UpdateStatutory(db *gorm.DB, companyID, id string, req UpdateStatutoryRequest) (*models.Statutory, error) {
	var st models.Statutory
	if err := db.First(&st, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		st.Name = *req.Name
	}
	if req.DueDate != nil {
		st.DueDate = req.DueDate
	}
	if req.Amount != nil {
		st.Amount = *req.Amount
	}
	if req.IsPaid != nil {
		st.IsPaid = *req.IsPaid
		if st.IsPaid {
			now := time.Now()
			st.PaidAt = &now
		}
	}
	if err := db.Save(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *service) DeleteStatutory(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.Statutory{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) MarkStatutoryPaid(db *gorm.DB, companyID, id string) (*models.Statutory, error) {
	var st models.Statutory
	if err := db.First(&st, "id = ?", id).Error; err != nil {
		return nil, err
	}
	st.IsPaid = true
	now := time.Now()
	st.PaidAt = &now
	if err := db.Save(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

// --- Salary Advances ---
func (s *service) ListSalaryAdvances(db *gorm.DB, companyID string) ([]models.SalaryAdvance, error) {
	var items []models.SalaryAdvance
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetSalaryAdvance(db *gorm.DB, companyID, id string) (*models.SalaryAdvance, error) {
	var sa models.SalaryAdvance
	if err := db.First(&sa, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *service) CreateSalaryAdvance(db *gorm.DB, companyID string, req CreateSalaryAdvanceRequest) (*models.SalaryAdvance, error) {
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}
	sa := models.SalaryAdvance{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		UserID:           uid,
		Amount:           req.Amount,
		Status:           "pending",
		DeductionMonths:  req.DeductionMonths,
	}
	if err := db.Create(&sa).Error; err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *service) UpdateSalaryAdvance(db *gorm.DB, companyID, id string, req UpdateSalaryAdvanceRequest) (*models.SalaryAdvance, error) {
	var sa models.SalaryAdvance
	if err := db.First(&sa, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Amount != nil {
		sa.Amount = *req.Amount
	}
	if req.Status != nil {
		sa.Status = *req.Status
	}
	if req.DeductionMonths != nil {
		sa.DeductionMonths = *req.DeductionMonths
	}
	if err := db.Save(&sa).Error; err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *service) DeleteSalaryAdvance(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.SalaryAdvance{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) ApproveSalaryAdvance(db *gorm.DB, companyID, id string) (*models.SalaryAdvance, error) {
	var sa models.SalaryAdvance
	if err := db.First(&sa, "id = ?", id).Error; err != nil {
		return nil, err
	}
	sa.Status = "approved"
	now := time.Now()
	sa.ApprovedAt = &now
	if err := db.Save(&sa).Error; err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *service) DeductSalaryAdvance(db *gorm.DB, companyID, id string, months int) (*models.SalaryAdvance, error) {
	var sa models.SalaryAdvance
	if err := db.First(&sa, "id = ?", id).Error; err != nil {
		return nil, err
	}
	sa.DeductionMonths = months
	if err := db.Save(&sa).Error; err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *service) ExportSalaryAdvancesCSV(db *gorm.DB, companyID string) ([]byte, error) {
	items, err := s.ListSalaryAdvances(db, companyID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"id", "user_id", "amount", "status", "deduction_months"})
	for _, it := range items {
		_ = w.Write([]string{it.ID.String(), it.UserID.String(), strconv.FormatFloat(it.Amount, 'f', 2, 64), it.Status, strconv.Itoa(it.DeductionMonths)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
