package taxonomy

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListTaxonomy(c echo.Context) error
	GetTaxonomyItem(c echo.Context) error
	CreateTaxonomyItem(c echo.Context) error
	UpdateTaxonomyItem(c echo.Context) error
	DeleteTaxonomyItem(c echo.Context) error
	ToggleComplete(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListTaxonomy(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListTaxonomy(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list taxonomy items"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetTaxonomyItem(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetTaxonomyItem(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("taxonomy item not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateTaxonomyItem(c echo.Context) error {
	var req CreateTaxonomyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateTaxonomyItem(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("taxonomy item created", item))
}

func (h *handler) UpdateTaxonomyItem(c echo.Context) error {
	id := c.Param("id")
	var req UpdateTaxonomyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateTaxonomyItem(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("taxonomy item updated", item))
}

func (h *handler) DeleteTaxonomyItem(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteTaxonomyItem(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) ToggleComplete(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.ToggleComplete(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("failed to toggle completion"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("status toggled", item))
}
