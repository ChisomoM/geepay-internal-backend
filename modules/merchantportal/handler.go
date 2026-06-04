package merchantportal

import (
	"backend/pkg/response"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines merchant portal HTTP handlers.
type Handler interface {
	Login(c echo.Context) error
	ListTickets(c echo.Context) error
	GetTicket(c echo.Context) error
	CreateTicket(c echo.Context) error
	UpdateTicketStatus(c echo.Context) error
}

type handler struct {
	svc Service
}

// NewHandler creates a new merchant portal handler.
func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

// Login handles POST /merchant/auth/login
func (h *handler) Login(c echo.Context) error {
	var req MerchantLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	resp, err := h.svc.Login(db, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.Success(resp))
}

// ListTickets handles GET /merchant/v1/tickets
func (h *handler) ListTickets(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	merchantID, _ := c.Get("merchant_id").(string)
	tickets, err := h.svc.ListTickets(db, merchantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list tickets"))
	}
	return c.JSON(http.StatusOK, response.Success(tickets))
}

// GetTicket handles GET /merchant/v1/tickets/:id
func (h *handler) GetTicket(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	merchantID, _ := c.Get("merchant_id").(string)
	ticket, err := h.svc.GetTicket(db, merchantID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(ticket))
}

// CreateTicket handles POST /merchant/v1/tickets
func (h *handler) CreateTicket(c echo.Context) error {
	var req CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	merchantID, _ := c.Get("merchant_id").(string)
	ticket, err := h.svc.CreateTicket(db, merchantID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("ticket submitted", ticket))
}

// UpdateTicketStatus handles PATCH /merchant/v1/tickets/:id/status
func (h *handler) UpdateTicketStatus(c echo.Context) error {
	id := c.Param("id")
	var req UpdateTicketStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	if req.Status == "" {
		return c.JSON(http.StatusBadRequest, response.Error("status is required"))
	}
	db := c.Get("db").(*gorm.DB)
	merchantID, _ := c.Get("merchant_id").(string)
	ticket, err := h.svc.UpdateTicketStatus(db, merchantID, id, req.Status)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("status updated", ticket))
}
