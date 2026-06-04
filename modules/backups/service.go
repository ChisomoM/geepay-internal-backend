package backups

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListBackups(db *gorm.DB, companyID string) ([]models.Backup, error)
	GetBackup(db *gorm.DB, companyID, id string) (*models.Backup, error)
	CreateBackup(db *gorm.DB, companyID string, req CreateBackupRequest) (*models.Backup, error)
	UpdateBackup(db *gorm.DB, companyID, id string, req UpdateBackupRequest) (*models.Backup, error)
	DeleteBackup(db *gorm.DB, companyID, id string) error
	ExportBackupsCSV(db *gorm.DB, companyID string) ([]byte, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateBackupRequest struct {
	FilePath    string     `json:"file_path"`
	Success     bool       `json:"success"`
	CompletedAt *time.Time `json:"completed_at"`
}

type UpdateBackupRequest struct {
	FilePath    *string    `json:"file_path"`
	Success     *bool      `json:"success"`
	CompletedAt *time.Time `json:"completed_at"`
}

func (s *service) ListBackups(db *gorm.DB, companyID string) ([]models.Backup, error) {
	var items []models.Backup
	if err := db.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetBackup(db *gorm.DB, companyID, id string) (*models.Backup, error) {
	var item models.Backup
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateBackup(db *gorm.DB, companyID string, req CreateBackupRequest) (*models.Backup, error) {
	if req.FilePath == "" {
		return nil, errors.New("file_path is required")
	}
	item := models.Backup{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		FilePath:         req.FilePath,
		Success:          req.Success,
		CompletedAt:      req.CompletedAt,
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) UpdateBackup(db *gorm.DB, companyID, id string, req UpdateBackupRequest) (*models.Backup, error) {
	var item models.Backup
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.FilePath != nil {
		item.FilePath = *req.FilePath
	}
	if req.Success != nil {
		item.Success = *req.Success
	}
	if req.CompletedAt != nil {
		item.CompletedAt = req.CompletedAt
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) DeleteBackup(db *gorm.DB, companyID, id string) error {
	return db.Delete(&models.Backup{}, "id = ?", id).Error
}

func (s *service) ExportBackupsCSV(db *gorm.DB, companyID string) ([]byte, error) {
	items, err := s.ListBackups(db, companyID)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "File Path", "Success", "Completed At", "Created At"})
	for _, item := range items {
		completedAt := ""
		if item.CompletedAt != nil {
			completedAt = item.CompletedAt.String()
		}
		_ = w.Write([]string{
			item.ID.String(),
			item.FilePath,
			fmt.Sprintf("%v", item.Success),
			completedAt,
			item.CreatedAt.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
