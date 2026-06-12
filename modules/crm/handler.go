package crm

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines all HTTP handlers for the CRM module.
type Handler interface {
	// Tickets
	ListTickets(c echo.Context) error
	CreateTicket(c echo.Context) error
	GetTicket(c echo.Context) error
	UpdateTicket(c echo.Context) error
	TrashTicket(c echo.Context) error
	ChangeStatus(c echo.Context) error
	AssignTicket(c echo.Context) error
	EscalateTicket(c echo.Context) error
	ListBreachedTickets(c echo.Context) error
	ListEvents(c echo.Context) error
	AddComment(c echo.Context) error

	// Categories
	ListCategories(c echo.Context) error
	GetCategory(c echo.Context) error
	CreateCategory(c echo.Context) error
	UpdateCategory(c echo.Context) error
	DeleteCategory(c echo.Context) error

	// Routing rules
	ListRoutingRules(c echo.Context) error
	CreateRoutingRule(c echo.Context) error
	UpdateRoutingRule(c echo.Context) error
	DeleteRoutingRule(c echo.Context) error

	// SLA policies
	ListSLAPolicies(c echo.Context) error
	CreateSLAPolicy(c echo.Context) error
	UpdateSLAPolicy(c echo.Context) error
	DeleteSLAPolicy(c echo.Context) error

	// Reports
	GetSummaryStats(c echo.Context) error
	GetStatsByCategory(c echo.Context) error
	GetStatsByAgent(c echo.Context) error
}

type handler struct {
	service Service
}

func NewHandler(s Service) Handler {
	return &handler{service: s}
}

// ── Ticket handlers ───────────────────────────────────────────────────────────

func (h *handler) ListTickets(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	filters := TicketFilters{
		Kind:       c.QueryParam("kind"),
		Status:     c.QueryParam("status"),
		Priority:   c.QueryParam("priority"),
		CategoryID: c.QueryParam("category_id"),
		AssigneeID: c.QueryParam("assignee_id"),
		TeamID:     c.QueryParam("team_id"),
		Breached:   c.QueryParam("breached") == "true",
	}

	tickets, err := h.service.ListTickets(db, companyID, filters)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to list tickets"))
	}
	return c.JSON(http.StatusOK, response.Success(tickets))
}

func (h *handler) CreateTicket(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	actorID, _ := c.Get("user_id").(string)

	var req CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}
	req.Source = "admin"

	ticket, err := h.service.CreateTicket(db, companyID, actorID, "user", req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("Ticket created", ticket))
}

func (h *handler) GetTicket(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	id := c.Param("id")

	ticket, err := h.service.GetTicket(db, companyID, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Ticket not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to get ticket"))
	}
	return c.JSON(http.StatusOK, response.Success(ticket))
}

func (h *handler) UpdateTicket(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	id := c.Param("id")

	var req UpdateTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	ticket, err := h.service.UpdateTicket(db, companyID, id, req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Ticket not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to update ticket"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Ticket updated", ticket))
}

func (h *handler) TrashTicket(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	id := c.Param("id")

	if err := h.service.TrashTicket(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to trash ticket"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Ticket moved to trash", nil))
}

func (h *handler) ChangeStatus(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	actorID, _ := c.Get("user_id").(string)
	id := c.Param("id")

	var req ChangeStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	ticket, err := h.service.ChangeStatus(db, companyID, id, actorID, "user", req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Ticket not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to update status"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Status updated", ticket))
}

func (h *handler) AssignTicket(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	actorID, _ := c.Get("user_id").(string)
	id := c.Param("id")

	var req AssignTicketRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	ticket, err := h.service.AssignTicket(db, companyID, id, actorID, req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Ticket not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to assign ticket"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Ticket assigned", ticket))
}

func (h *handler) EscalateTicket(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	actorID, _ := c.Get("user_id").(string)
	id := c.Param("id")

	ticket, err := h.service.EscalateTicket(db, companyID, id, actorID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Ticket not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to escalate ticket"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Ticket escalated", ticket))
}

func (h *handler) ListBreachedTickets(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	tickets, err := h.service.ListBreachedTickets(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to list breached tickets"))
	}
	return c.JSON(http.StatusOK, response.Success(tickets))
}

func (h *handler) ListEvents(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	id := c.Param("id")
	// Internal staff always sees all events; query param allows client to opt-out
	includeInternal := c.QueryParam("internal") != "false"

	events, err := h.service.ListEvents(db, companyID, id, includeInternal)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to list events"))
	}
	return c.JSON(http.StatusOK, response.Success(events))
}

func (h *handler) AddComment(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	actorID, _ := c.Get("user_id").(string)
	id := c.Param("id")

	var req AddCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	event, err := h.service.AddComment(db, companyID, id, actorID, "user", req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Ticket not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to add comment"))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("Comment added", event))
}

// ── Category handlers ─────────────────────────────────────────────────────────

func (h *handler) ListCategories(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	cats, err := h.service.ListCategories(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to list categories"))
	}
	return c.JSON(http.StatusOK, response.Success(cats))
}

func (h *handler) GetCategory(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	cat, err := h.service.GetCategory(db, companyID, c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Category not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to get category"))
	}
	return c.JSON(http.StatusOK, response.Success(cat))
}

func (h *handler) CreateCategory(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	cat, err := h.service.CreateCategory(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("Category created", cat))
}

func (h *handler) UpdateCategory(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	var req UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	cat, err := h.service.UpdateCategory(db, companyID, c.Param("id"), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Category not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to update category"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Category updated", cat))
}

func (h *handler) DeleteCategory(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	if err := h.service.DeleteCategory(db, companyID, c.Param("id")); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to delete category"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Category deleted", nil))
}

// ── Routing rule handlers ─────────────────────────────────────────────────────

func (h *handler) ListRoutingRules(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	rules, err := h.service.ListRoutingRules(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to list routing rules"))
	}
	return c.JSON(http.StatusOK, response.Success(rules))
}

func (h *handler) CreateRoutingRule(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	var req CreateRoutingRuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	rule, err := h.service.CreateRoutingRule(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("Routing rule created", rule))
}

func (h *handler) UpdateRoutingRule(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	var req UpdateRoutingRuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	rule, err := h.service.UpdateRoutingRule(db, companyID, c.Param("id"), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("Routing rule not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to update routing rule"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Routing rule updated", rule))
}

func (h *handler) DeleteRoutingRule(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	if err := h.service.DeleteRoutingRule(db, companyID, c.Param("id")); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to delete routing rule"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("Routing rule deleted", nil))
}

// ── SLA policy handlers ───────────────────────────────────────────────────────

func (h *handler) ListSLAPolicies(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	policies, err := h.service.ListSLAPolicies(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to list SLA policies"))
	}
	return c.JSON(http.StatusOK, response.Success(policies))
}

func (h *handler) CreateSLAPolicy(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	var req CreateSLAPolicyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	policy, err := h.service.CreateSLAPolicy(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("SLA policy created", policy))
}

func (h *handler) UpdateSLAPolicy(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	var req UpdateSLAPolicyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
	}

	policy, err := h.service.UpdateSLAPolicy(db, companyID, c.Param("id"), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.Error("SLA policy not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to update SLA policy"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("SLA policy updated", policy))
}

func (h *handler) DeleteSLAPolicy(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	if err := h.service.DeleteSLAPolicy(db, companyID, c.Param("id")); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to delete SLA policy"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("SLA policy deleted", nil))
}

// ── Report handlers ───────────────────────────────────────────────────────────

func (h *handler) GetSummaryStats(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	stats, err := h.service.GetSummaryStats(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to get summary stats"))
	}
	return c.JSON(http.StatusOK, response.Success(stats))
}

func (h *handler) GetStatsByCategory(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	stats, err := h.service.GetStatsByCategory(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to get stats by category"))
	}
	return c.JSON(http.StatusOK, response.Success(stats))
}

func (h *handler) GetStatsByAgent(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)

	stats, err := h.service.GetStatsByAgent(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("Failed to get stats by agent"))
	}
	return c.JSON(http.StatusOK, response.Success(stats))
}
