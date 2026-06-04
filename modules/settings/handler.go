package settings

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	ListSettings(c echo.Context) error
	UpdateSettings(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) ListSettings(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	items, err := h.svc.ListSettings(db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list settings"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) UpdateSettings(c echo.Context) error {
	var req UpdateSettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	if len(req.Settings) == 0 {
		return c.JSON(http.StatusBadRequest, response.Error("no settings provided"))
	}
	db := c.Get("db").(*gorm.DB)
	if err := h.svc.UpdateSettings(db, req); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to update settings"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("settings updated", nil))
}
