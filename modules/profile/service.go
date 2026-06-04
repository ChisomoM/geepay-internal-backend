package profile

import (
	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	GetProfile(db *gorm.DB, companyID, userID string) (*models.Profile, error)
	UpdateProfile(db *gorm.DB, companyID, userID string, req UpdateProfileRequest) (*models.Profile, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type UpdateProfileRequest struct {
	Bio     *string `json:"bio"`
	Phone   *string `json:"phone"`
	Address *string `json:"address"`
}

func (s *service) GetProfile(db *gorm.DB, companyID, userID string) (*models.Profile, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	var p models.Profile
	// Use FirstOrCreate so every user always has a profile record
	result := db.Where(models.Profile{UserID: uid}).FirstOrCreate(&p, models.Profile{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		UserID:           uid,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return &p, nil
}

func (s *service) UpdateProfile(db *gorm.DB, companyID, userID string, req UpdateProfileRequest) (*models.Profile, error) {
	p, err := s.GetProfile(db, companyID, userID)
	if err != nil {
		return nil, err
	}
	if req.Bio != nil {
		p.Bio = *req.Bio
	}
	if req.Phone != nil {
		p.Phone = *req.Phone
	}
	if req.Address != nil {
		p.Address = *req.Address
	}
	if err := db.Save(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}
