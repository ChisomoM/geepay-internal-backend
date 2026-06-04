package dashboard

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	GetDashboard(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

func (h *handler) GetDashboard(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	stats, err := h.svc.GetDashboardStats(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to load dashboard"))
	}
	return c.JSON(http.StatusOK, response.Success(stats))
}
