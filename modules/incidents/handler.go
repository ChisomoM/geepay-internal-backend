package incidents

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	// Incidents
	ListIncidents(c echo.Context) error
	GetIncident(c echo.Context) error
	CreateIncident(c echo.Context) error
	UpdateIncident(c echo.Context) error
	ChangeIncidentStatus(c echo.Context) error
	AssignIncident(c echo.Context) error
	NotifyIncident(c echo.Context) error

	// Tickets
	ListTickets(c echo.Context) error
	GetTicket(c echo.Context) error
	CreateTicket(c echo.Context) error
	UpdateTicket(c echo.Context) error
	TrashTicket(c echo.Context) error
	RestoreTicket(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

// Incidents
func (h *handler) ListIncidents(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListIncidents(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list incidents"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetIncident(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.GetIncident(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(it))
}

func (h *handler) CreateIncident(c echo.Context) error {
	var req CreateIncidentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.CreateIncident(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("created", it))
}

func (h *handler) UpdateIncident(c echo.Context) error {
	id := c.Param("id")
	var req UpdateIncidentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.UpdateIncident(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("updated", it))
}

func (h *handler) ChangeIncidentStatus(c echo.Context) error {
	id := c.Param("id")
	status := c.FormValue("status")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.ChangeIncidentStatus(db, companyID, id, status)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("status changed", it))
}

func (h *handler) AssignIncident(c echo.Context) error {
	id := c.Param("id")
	assignee := c.FormValue("assignee")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.AssignIncident(db, companyID, id, assignee)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("assigned", it))
}

func (h *handler) NotifyIncident(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.NotifyIncident(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("notified", it))
}

// Tickets
func (h *handler) ListTickets(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListTickets(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list tickets"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetTicket(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.GetTicket(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(it))
}

func (h *handler) CreateTicket(c echo.Context) error {
	var req CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.CreateTicket(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("created", it))
}

func (h *handler) UpdateTicket(c echo.Context) error {
	id := c.Param("id")
	var req UpdateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	it, err := h.svc.UpdateTicket(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("updated", it))
}

func (h *handler) TrashTicket(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.TrashTicket(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to trash"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("trashed", nil))
}

func (h *handler) RestoreTicket(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.RestoreTicket(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to restore"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("restored", nil))
}
