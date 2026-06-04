package stafflisting

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListStaff(c echo.Context) error
	GetStaff(c echo.Context) error
	CreateStaff(c echo.Context) error
	UpdateStaff(c echo.Context) error
	DeleteStaff(c echo.Context) error
	ExportStaffCSV(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListStaff(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListStaff(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list staff"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetStaff(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetStaff(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("staff member not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateStaff(c echo.Context) error {
	var req CreateStaffRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateStaff(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("staff member created", item))
}

func (h *handler) UpdateStaff(c echo.Context) error {
	id := c.Param("id")
	var req UpdateStaffRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateStaff(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("staff member updated", item))
}

func (h *handler) DeleteStaff(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteStaff(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) ExportStaffCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.svc.ExportStaffCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=staff.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}
