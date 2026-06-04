package sims

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListSims(c echo.Context) error
	GetSim(c echo.Context) error
	CreateSim(c echo.Context) error
	UpdateSim(c echo.Context) error
	DeleteSim(c echo.Context) error
	AssignSim(c echo.Context) error
	UnassignSim(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListSims(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	list, err := h.svc.ListSims(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list sims"))
	}
	return c.JSON(http.StatusOK, response.Success(list))
}

func (h *handler) GetSim(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	s, err := h.svc.GetSim(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(s))
}

func (h *handler) CreateSim(c echo.Context) error {
	var req CreateSimRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	s, err := h.svc.CreateSim(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("created", s))
}

func (h *handler) UpdateSim(c echo.Context) error {
	id := c.Param("id")
	var req UpdateSimRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	s, err := h.svc.UpdateSim(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("updated", s))
}

func (h *handler) DeleteSim(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteSim(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) AssignSim(c echo.Context) error {
	id := c.Param("id")
	assignedTo := c.FormValue("assigned_to")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	s, err := h.svc.AssignSim(db, companyID, id, assignedTo)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("assigned", s))
}

func (h *handler) UnassignSim(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	s, err := h.svc.UnassignSim(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("unassigned", s))
}
