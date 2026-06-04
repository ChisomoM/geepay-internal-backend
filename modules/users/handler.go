package users

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines user HTTP handlers.
type Handler interface {
	Create(c echo.Context) error
	Get(c echo.Context) error
	List(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

func (h *handler) Create(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}

	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	user, err := h.service.Create(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("User created", user))
}

func (h *handler) Get(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	user, err := h.service.Get(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("user not found"))
	}
	return c.JSON(http.StatusOK, response.Success(user))
}

func (h *handler) List(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	users, err := h.service.List(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to fetch users"))
	}
	return c.JSON(http.StatusOK, response.Success(users))
}

func (h *handler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	user, err := h.service.Update(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("User updated", user))
}

func (h *handler) Delete(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	if err := h.service.Delete(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete user"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("User deleted", nil))
}
