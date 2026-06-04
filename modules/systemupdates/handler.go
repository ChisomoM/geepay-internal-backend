package systemupdates

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListSystemUpdates(c echo.Context) error
	GetSystemUpdate(c echo.Context) error
	CreateSystemUpdate(c echo.Context) error
	UpdateSystemUpdate(c echo.Context) error
	DeleteSystemUpdate(c echo.Context) error
	ExportSystemUpdatesCSV(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListSystemUpdates(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	items, err := h.svc.ListSystemUpdates(db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list system updates"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetSystemUpdate(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	item, err := h.svc.GetSystemUpdate(db, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("system update not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateSystemUpdate(c echo.Context) error {
	var req CreateSystemUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	item, err := h.svc.CreateSystemUpdate(db, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("system update created", item))
}

func (h *handler) UpdateSystemUpdate(c echo.Context) error {
	id := c.Param("id")
	var req UpdateSystemUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	item, err := h.svc.UpdateSystemUpdate(db, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("system update updated", item))
}

func (h *handler) DeleteSystemUpdate(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	if err := h.svc.DeleteSystemUpdate(db, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) ExportSystemUpdatesCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	data, err := h.svc.ExportSystemUpdatesCSV(db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=system_updates.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}
