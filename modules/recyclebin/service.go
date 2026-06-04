package recyclebin

import (
	"errors"

	"backend/global"
	"backend/models"

	"gorm.io/gorm"
)

type Service interface {
	ListDeletedItems(db *gorm.DB, companyID string) ([]models.RecycleBin, error)
	RestoreItem(db *gorm.DB, companyID, table, id string) error
	ForceDeleteItem(db *gorm.DB, companyID, table, id string) error
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

// tableModelMap maps URL-safe table names to their GORM model pointers.
var tableModelMap = map[string]func() interface{}{
	"merchants":           func() interface{} { return &models.Merchant{} },
	"merchant_statements": func() interface{} { return &models.MerchantStatement{} },
	"incidents":           func() interface{} { return &models.Incident{} },
	"support_tickets":     func() interface{} { return &models.SupportTicket{} },
	"risk_and_compliance": func() interface{} { return &models.RiskAndCompliance{} },
	"backups":             func() interface{} { return &models.Backup{} },
	"inventory":           func() interface{} { return &models.Inventory{} },
	"product_catalogs":    func() interface{} { return &models.ProductCatalog{} },
	"sim_cards":           func() interface{} { return &models.SimCard{} },
	"staff_listings":      func() interface{} { return &models.StaffListing{} },
	"departments":         func() interface{} { return &models.Department{} },
	"taxonomy_items":      func() interface{} { return &models.TaxonomyItem{} },
	"system_updates":      func() interface{} { return &models.SystemUpdate{} },
}

func (s *service) ListDeletedItems(db *gorm.DB, companyID string) ([]models.RecycleBin, error) {
	var items []models.RecycleBin
	if err := db.Order("deleted_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) RestoreItem(db *gorm.DB, companyID, table, id string) error {
	modelFn, ok := tableModelMap[table]
	if !ok {
		return errors.New("unknown table: " + table)
	}
	// Unscoped so GORM sees soft-deleted rows; clear deleted_at to restore
	result := db.Unscoped().Model(modelFn()).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	// Remove the recycle bin entry
	db.Where("resource = ? AND resource_id = ?", table, id).Delete(&models.RecycleBin{})
	return nil
}

func (s *service) ForceDeleteItem(db *gorm.DB, companyID, table, id string) error {
	modelFn, ok := tableModelMap[table]
	if !ok {
		return errors.New("unknown table: " + table)
	}
	result := db.Unscoped().Where("id = ?", id).Delete(modelFn())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	// Remove the recycle bin entry
	db.Where("resource = ? AND resource_id = ?", table, id).Delete(&models.RecycleBin{})
	return nil
}
