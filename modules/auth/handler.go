package auth

import (
	"net/http"
	"strings"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines auth HTTP handlers.
type Handler interface {
	Login(c echo.Context) error
	Register(c echo.Context) error
	RefreshToken(c echo.Context) error
	RequestPasswordReset(c echo.Context) error
	ResetPassword(c echo.Context) error
}

// handler implements Handler.
type handler struct {
	service Service
}

// NewHandler creates a new auth handler.
func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// Login handles POST /api/v1/auth/login
func (h *handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
	}

	db := c.Get("db").(*gorm.DB)

	// Normalize email to lower-case to ensure case-insensitive matching
	email := strings.ToLower(req.Email)

	resp, err := h.service.Login(db, email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Error("Login failed"))
	}

	return c.JSON(http.StatusOK, response.Success(resp))
}

// Register handles POST /api/v1/auth/register
func (h *handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
	}

	db := c.Get("db").(*gorm.DB)

	// Normalize email to lower-case before creating account
	email := strings.ToLower(req.Email)

	resp, err := h.service.Register(db, email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Registration failed"))
	}

	return c.JSON(http.StatusCreated, response.Success(resp))
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *handler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
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

// RequestPasswordReset handles POST /api/v1/auth/forgot-password
func (h *handler) RequestPasswordReset(c echo.Context) error {
	var req PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
	}

	db := c.Get("db").(*gorm.DB)

	// Normalize email for password reset
	email := strings.ToLower(req.Email)

	if err := h.service.RequestPasswordReset(db, email); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Request failed"))
	}

	return c.JSON(http.StatusOK, response.SuccessWithMessage("Password reset email sent", nil))
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *handler) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
	}

	db := c.Get("db").(*gorm.DB)

	if err := h.service.ResetPassword(db, req.Token, req.NewPassword); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Reset failed"))
	}

	return c.JSON(http.StatusOK, response.SuccessWithMessage("Password reset successfully", nil))
}
