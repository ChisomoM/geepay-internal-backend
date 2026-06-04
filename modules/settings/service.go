package settings

import (
	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListSettings(db *gorm.DB) ([]models.Setting, error)
	UpdateSettings(db *gorm.DB, req UpdateSettingsRequest) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings"`
}

func (s *service) ListSettings(db *gorm.DB) ([]models.Setting, error) {
	var items []models.Setting
	// Settings are global (not company-scoped), use unscoped query on the session DB
	if err := db.Unscoped().Where("deleted_at IS NULL").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) UpdateSettings(db *gorm.DB, req UpdateSettingsRequest) error {
	for k, v := range req.Settings {
		setting := models.Setting{Key: k, Value: v}
		if err := db.Where(models.Setting{Key: k}).Assign(models.Setting{Value: v}).FirstOrCreate(&setting).Error; err != nil {
			return err
		}
		// Update value in case it already existed
		if err := db.Model(&setting).Where("key = ?", k).Update("value", v).Error; err != nil {
			return err
		}
	}
	return nil
}
