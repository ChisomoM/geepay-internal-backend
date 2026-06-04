package riskcompliance

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListItems(c echo.Context) error
	GetItem(c echo.Context) error
	CreateItem(c echo.Context) error
	UpdateItem(c echo.Context) error
	DeleteItem(c echo.Context) error
	ExportCSV(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListItems(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListItems(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list risk & compliance items"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetItem(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetItem(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("item not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateItem(c echo.Context) error {
	var req CreateRiskComplianceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateItem(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("item created", item))
}

func (h *handler) UpdateItem(c echo.Context) error {
	id := c.Param("id")
	var req UpdateRiskComplianceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateItem(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("item updated", item))
}

func (h *handler) DeleteItem(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteItem(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) ExportCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.svc.ExportCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=risk_compliance.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}
