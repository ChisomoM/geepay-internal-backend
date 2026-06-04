package sims

import (
	"errors"

	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	ListSims(db *gorm.DB, companyID string) ([]models.SimCard, error)
	GetSim(db *gorm.DB, companyID, id string) (*models.SimCard, error)
	CreateSim(db *gorm.DB, companyID string, req CreateSimRequest) (*models.SimCard, error)
	UpdateSim(db *gorm.DB, companyID, id string, req UpdateSimRequest) (*models.SimCard, error)
	DeleteSim(db *gorm.DB, companyID, id string) error
	AssignSim(db *gorm.DB, companyID, id string, assignedTo string) (*models.SimCard, error)
	UnassignSim(db *gorm.DB, companyID, id string) (*models.SimCard, error)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

type CreateSimRequest struct {
	ICCID      string  `json:"iccid"`
	Network    string  `json:"network"`
	Status     string  `json:"status"`
	AssignedTo *string `json:"assigned_to,omitempty"`
}

type UpdateSimRequest struct {
	ICCID      *string `json:"iccid"`
	Network    *string `json:"network"`
	Status     *string `json:"status"`
	AssignedTo *string `json:"assigned_to"`
}

func (s *service) ListSims(db *gorm.DB, companyID string) ([]models.SimCard, error) {
	var list []models.SimCard
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *service) GetSim(db *gorm.DB, companyID, id string) (*models.SimCard, error) {
	var sim models.SimCard
	if err := db.First(&sim, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sim, nil
}

func (s *service) CreateSim(db *gorm.DB, companyID string, req CreateSimRequest) (*models.SimCard, error) {
	if req.ICCID == "" {
		return nil, errors.New("iccid is required")
	}
	sim := models.SimCard{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		ICCID:            req.ICCID,
		Network:          req.Network,
		Status:           req.Status,
	}
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		uid, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			return nil, errors.New("invalid assigned_to id")
		}
		sim.AssignedTo = uid
	}
	if err := db.Create(&sim).Error; err != nil {
		return nil, err
	}
	return &sim, nil
}

func (s *service) UpdateSim(db *gorm.DB, companyID, id string, req UpdateSimRequest) (*models.SimCard, error) {
	var sim models.SimCard
	if err := db.First(&sim, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if req.ICCID != nil {
		sim.ICCID = *req.ICCID
	}
	if req.Network != nil {
		sim.Network = *req.Network
	}
	if req.Status != nil {
		sim.Status = *req.Status
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == "" {
			sim.AssignedTo = uuid.Nil
			sim.Status = "available"
		} else {
			uid, err := uuid.Parse(*req.AssignedTo)
			if err != nil {
				return nil, errors.New("invalid assigned_to id")
			}
			sim.AssignedTo = uid
			sim.Status = "assigned"
		}
	}
	if err := db.Save(&sim).Error; err != nil {
		return nil, err
	}
	return &sim, nil
}

func (s *service) DeleteSim(db *gorm.DB, companyID, id string) error {
	if err := db.Delete(&models.SimCard{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (s *service) AssignSim(db *gorm.DB, companyID, id string, assignedTo string) (*models.SimCard, error) {
	var sim models.SimCard
	if err := db.First(&sim, "id = ?", id).Error; err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(assignedTo)
	if err != nil {
		return nil, errors.New("invalid assigned_to id")
	}
	sim.AssignedTo = uid
	sim.Status = "assigned"
	if err := db.Save(&sim).Error; err != nil {
		return nil, err
	}
	return &sim, nil
}

func (s *service) UnassignSim(db *gorm.DB, companyID, id string) (*models.SimCard, error) {
	var sim models.SimCard
	if err := db.First(&sim, "id = ?", id).Error; err != nil {
		return nil, err
	}
	sim.AssignedTo = uuid.Nil
	sim.Status = "available"
	if err := db.Save(&sim).Error; err != nil {
		return nil, err
	}
	return &sim, nil
}
