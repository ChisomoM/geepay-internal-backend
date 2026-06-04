package backups

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListBackups(c echo.Context) error
	GetBackup(c echo.Context) error
	CreateBackup(c echo.Context) error
	UpdateBackup(c echo.Context) error
	DeleteBackup(c echo.Context) error
	ExportBackupsCSV(c echo.Context) error
	DownloadBackup(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListBackups(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListBackups(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list backups"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetBackup(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetBackup(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("backup not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateBackup(c echo.Context) error {
	var req CreateBackupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateBackup(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("backup created", item))
}

func (h *handler) UpdateBackup(c echo.Context) error {
	id := c.Param("id")
	var req UpdateBackupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateBackup(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("backup updated", item))
}

func (h *handler) DeleteBackup(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteBackup(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) ExportBackupsCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.svc.ExportBackupsCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=backups.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}

func (h *handler) DownloadBackup(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetBackup(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("backup not found"))
	}
	if item.FilePath == "" {
		return c.JSON(http.StatusNotFound, response.Error("no file associated with this backup"))
	}
	return c.Redirect(http.StatusFound, item.FilePath)
}
