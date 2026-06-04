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
	companyID, _ := c.Get("companyID").(string)
	perms, err := h.service.ListPermissions(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch permissions"))
	}
	return c.JSON(http.StatusOK, response.Success(perms))
}

// ListRoles handles GET /api/v1/rbac/roles
func (h *handler) ListRoles(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	roles, err := h.service.ListRoles(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch roles"))
	}
	return c.JSON(http.StatusOK, response.Success(roles))
}

// CreateRole handles POST /api/v1/rbac/roles
func (h *handler) CreateRole(c echo.Context) error {
	var req CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}

	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	role, err := h.service.CreateRole(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Failed to create role"))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("Role created", role))
}
