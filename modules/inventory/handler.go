package inventory

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListItems(c echo.Context) error
	ListItemsByCategory(c echo.Context) error
	GetItem(c echo.Context) error
	CreateItem(c echo.Context) error
	UpdateItem(c echo.Context) error
	DeleteItem(c echo.Context) error
	AssignItem(c echo.Context) error
	UnassignItem(c echo.Context) error

	ListCategories(c echo.Context) error
	GetCategory(c echo.Context) error
	CreateCategory(c echo.Context) error
	UpdateCategory(c echo.Context) error
	DeleteCategory(c echo.Context) error
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
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list items"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) ListItemsByCategory(c echo.Context) error {
	categoryID := c.Param("categoryId")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListItemsByCategory(db, companyID, categoryID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list items"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetItem(c echo.Context) error {
	id := c.Param("itemId")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetItem(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateItem(c echo.Context) error {
	categoryID := c.Param("categoryId")
	var req CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateItem(db, companyID, categoryID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("created", item))
}

func (h *handler) UpdateItem(c echo.Context) error {
	id := c.Param("itemId")
	var req UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateItem(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("updated", item))
}

func (h *handler) DeleteItem(c echo.Context) error {
	id := c.Param("itemId")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteItem(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) AssignItem(c echo.Context) error {
	id := c.Param("itemId")
	assignedTo := c.FormValue("assigned_to")
	assignedType := c.FormValue("assigned_type")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.AssignItem(db, companyID, id, assignedTo, assignedType)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("assigned", item))
}

func (h *handler) UnassignItem(c echo.Context) error {
	id := c.Param("itemId")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UnassignItem(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("unassigned", item))
}

// --- Categories ---
func (h *handler) ListCategories(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListCategories(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list categories"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetCategory(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetCategory(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateCategory(c echo.Context) error {
	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateCategory(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("created", item))
}

func (h *handler) UpdateCategory(c echo.Context) error {
	id := c.Param("id")
	var req UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateCategory(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("updated", item))
}

func (h *handler) DeleteCategory(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteCategory(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}
