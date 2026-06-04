package MODULENAME

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines all HTTP handlers for this module.
type Handler interface {
	Create(c echo.Context) error
	Get(c echo.Context) error
	List(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
}

// handler is the unexported implementation of Handler.
type handler struct {
	service Service
}

// NewHandler creates a new handler instance.
func NewHandler(service Service) Handler {
	return &handler{
		service: service,
	}
}

// Create handles POST /entities
func (h *handler) Create(c echo.Context) error {
	// Extract company-scoped DB from context (set by CompanyMiddleware)
	db := c.Get("db").(*gorm.DB)

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	dto, err := h.service.Create(db, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to create entity"))
	}

	return c.JSON(http.StatusCreated, response.SuccessWithMessage("Entity created", dto))
}

// Get handles GET /entities/:id
func (h *handler) Get(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")

	dto, err := h.service.Get(db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Entity not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to get entity"))
	}

	return c.JSON(http.StatusOK, response.Success(dto))
}

// List handles GET /entities
func (h *handler) List(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)

	dtos, err := h.service.List(db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to list entities"))
	}

	return c.JSON(http.StatusOK, response.Success(dtos))
}

// Update handles PUT /entities/:id
func (h *handler) Update(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	dto, err := h.service.Update(db, id, req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Entity not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to update entity"))
	}

	return c.JSON(http.StatusOK, response.SuccessWithMessage("Entity updated", dto))
}

// Delete handles DELETE /entities/:id
func (h *handler) Delete(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	id := c.Param("id")

	err := h.service.Delete(db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Entity not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to delete entity"))
	}

	return c.JSON(http.StatusOK, response.SuccessWithMessage("Entity deleted", nil))
}
