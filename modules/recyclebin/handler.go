package recyclebin

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListDeletedItems(c echo.Context) error
	RestoreItem(c echo.Context) error
	ForceDeleteItem(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListDeletedItems(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListDeletedItems(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list deleted items"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) RestoreItem(c echo.Context) error {
	table := c.Param("table")
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.RestoreItem(db, companyID, table, id); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("item restored", nil))
}

func (h *handler) ForceDeleteItem(c echo.Context) error {
	table := c.Param("table")
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.ForceDeleteItem(db, companyID, table, id); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("item permanently deleted", nil))
}
