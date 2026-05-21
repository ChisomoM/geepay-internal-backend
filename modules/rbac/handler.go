package rbac

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines RBAC HTTP handlers.
type Handler interface {
	ListPermissions(c echo.Context) error
	ListRoles(c echo.Context) error
	CreateRole(c echo.Context) error
}

type handler struct {
	service Service
}

// NewHandler creates a new RBAC handler.
func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// ListPermissions handles GET /api/v1/rbac/permissions
func (h *handler) ListPermissions(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	perms, err := h.service.ListPermissions(db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch permissions"))
	}
	return c.JSON(http.StatusOK, response.Success(perms))
}

// ListRoles handles GET /api/v1/rbac/roles
func (h *handler) ListRoles(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	roles, err := h.service.ListRoles(db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch roles"))
	}
	return c.JSON(http.StatusOK, response.Success(roles))
}

// CreateRole handles POST /api/v1/rbac/roles
func (h *handler) CreateRole(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	if err := h.service.CreateRole(db); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Failed to create role"))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("Role created", nil))
}
