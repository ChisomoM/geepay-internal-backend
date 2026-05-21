package controlhub

import (
	"backend/pkg/response"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines ControlHub HTTP handlers.
type Handler interface {
	LoginCompanyAdmin(c echo.Context) error
	RefreshToken(c echo.Context) error
	Logout(c echo.Context) error
}

// handler implements Handler.
type handler struct {
	service Service
}

// NewHandler creates a new ControlHub handler.
func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// LoginCompanyAdmin handles POST /controlhub/auth/login
// Expects unscoped DB access (platform-level authentication).
func (h *handler) LoginCompanyAdmin(c echo.Context) error {
	var req CompanyAdminLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
	}

	// Get unscoped DB (no tenant middleware here)
	db := c.Get("db").(*gorm.DB)
	if db == nil {
		// Fallback: should not happen, but get fresh connection if needed
		db = c.Get("db").(*gorm.DB)
	}

	// Normalize email to lower-case for case-insensitive matching
	email := strings.ToLower(req.Email)

	resp, err := h.service.LoginCompanyAdmin(db, email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Error(err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(resp))
}

// RefreshToken handles POST /controlhub/auth/refresh
func (h *handler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
	}

	db := c.Get("db").(*gorm.DB)
	resp, err := h.service.RefreshToken(db, req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Error("Token refresh failed"))
	}

	return c.JSON(http.StatusOK, response.Success(resp))
}

// Logout handles POST /controlhub/auth/logout
// For stateless JWT, this is typically client-side. Stub for now.
func (h *handler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, response.Success(map[string]string{
		"message": "Logged out successfully",
	}))
}
