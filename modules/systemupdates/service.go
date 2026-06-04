package systemupdates

import (
	"bytes"
	"encoding/csv"
	"errors"
	"time"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListSystemUpdates(db *gorm.DB) ([]models.SystemUpdate, error)
	GetSystemUpdate(db *gorm.DB, id string) (*models.SystemUpdate, error)
	CreateSystemUpdate(db *gorm.DB, req CreateSystemUpdateRequest) (*models.SystemUpdate, error)
	UpdateSystemUpdate(db *gorm.DB, id string, req UpdateSystemUpdateRequest) (*models.SystemUpdate, error)
	DeleteSystemUpdate(db *gorm.DB, id string) error
	ExportSystemUpdatesCSV(db *gorm.DB) ([]byte, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateSystemUpdateRequest struct {
	Version    string     `json:"version"`
	Changelog  string     `json:"changelog"`
	ReleasedAt *time.Time `json:"released_at"`
}

type UpdateSystemUpdateRequest struct {
	Version    *string    `json:"version"`
	Changelog  *string    `json:"changelog"`
	ReleasedAt *time.Time `json:"released_at"`
}

func (s *service) ListSystemUpdates(db *gorm.DB) ([]models.SystemUpdate, error) {
	var items []models.SystemUpdate
	if err := db.Order("released_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) GetSystemUpdate(db *gorm.DB, id string) (*models.SystemUpdate, error) {
	var item models.SystemUpdate
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) CreateSystemUpdate(db *gorm.DB, req CreateSystemUpdateRequest) (*models.SystemUpdate, error) {
	if req.Version == "" {
		return nil, errors.New("version is required")
	}
	item := models.SystemUpdate{
		Version:    req.Version,
		Changelog:  req.Changelog,
		ReleasedAt: req.ReleasedAt,
	}
	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) UpdateSystemUpdate(db *gorm.DB, id string, req UpdateSystemUpdateRequest) (*models.SystemUpdate, error) {
	var item models.SystemUpdate
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.Version != nil {
		item.Version = *req.Version
	}
	if req.Changelog != nil {
		item.Changelog = *req.Changelog
	}
	if req.ReleasedAt != nil {
		item.ReleasedAt = req.ReleasedAt
	}
	if err := db.Save(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *service) DeleteSystemUpdate(db *gorm.DB, id string) error {
	return db.Delete(&models.SystemUpdate{}, "id = ?", id).Error
}

func (s *service) ExportSystemUpdatesCSV(db *gorm.DB) ([]byte, error) {
	items, err := s.ListSystemUpdates(db)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{"ID", "Version", "Changelog", "Released At", "Created At"})
	for _, item := range items {
		releasedAt := ""
		if item.ReleasedAt != nil {
			releasedAt = item.ReleasedAt.String()
		}
		_ = w.Write([]string{
			item.ID.String(),
			item.Version,
			item.Changelog,
			releasedAt,
			item.CreatedAt.String(),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
