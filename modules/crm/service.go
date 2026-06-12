package crm

import (
	"encoding/json"
	"errors"
	"time"

	"backend/global"
	"backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service defines the business logic for the CRM module.
type Service interface {
	// Tickets
	ListTickets(db *gorm.DB, companyID string, filters TicketFilters) ([]models.Ticket, error)
	GetTicket(db *gorm.DB, companyID, id string) (*models.Ticket, error)
	CreateTicket(db *gorm.DB, companyID, actorID, actorType string, req CreateTicketRequest) (*models.Ticket, error)
	UpdateTicket(db *gorm.DB, companyID, id string, req UpdateTicketRequest) (*models.Ticket, error)
	ChangeStatus(db *gorm.DB, companyID, id, actorID, actorType string, req ChangeStatusRequest) (*models.Ticket, error)
	AssignTicket(db *gorm.DB, companyID, id, actorID string, req AssignTicketRequest) (*models.Ticket, error)
	EscalateTicket(db *gorm.DB, companyID, id, actorID string) (*models.Ticket, error)
	TrashTicket(db *gorm.DB, companyID, id string) error
	ListBreachedTickets(db *gorm.DB, companyID string) ([]models.Ticket, error)
	AddComment(db *gorm.DB, companyID, ticketID, actorID, actorType string, req AddCommentRequest) (*models.TicketEvent, error)
	ListEvents(db *gorm.DB, companyID, ticketID string, includeInternal bool) ([]models.TicketEvent, error)
	SweepBreachedTickets(db *gorm.DB)

	// Categories
	ListCategories(db *gorm.DB, companyID string) ([]models.TicketCategory, error)
	GetCategory(db *gorm.DB, companyID, id string) (*models.TicketCategory, error)
	CreateCategory(db *gorm.DB, companyID string, req CreateCategoryRequest) (*models.TicketCategory, error)
	UpdateCategory(db *gorm.DB, companyID, id string, req UpdateCategoryRequest) (*models.TicketCategory, error)
	DeleteCategory(db *gorm.DB, companyID, id string) error

	// Routing rules
	ListRoutingRules(db *gorm.DB, companyID string) ([]models.CRMRoutingRule, error)
	CreateRoutingRule(db *gorm.DB, companyID string, req CreateRoutingRuleRequest) (*models.CRMRoutingRule, error)
	UpdateRoutingRule(db *gorm.DB, companyID, id string, req UpdateRoutingRuleRequest) (*models.CRMRoutingRule, error)
	DeleteRoutingRule(db *gorm.DB, companyID, id string) error

	// SLA policies
	ListSLAPolicies(db *gorm.DB, companyID string) ([]models.SLAPolicy, error)
	CreateSLAPolicy(db *gorm.DB, companyID string, req CreateSLAPolicyRequest) (*models.SLAPolicy, error)
	UpdateSLAPolicy(db *gorm.DB, companyID, id string, req UpdateSLAPolicyRequest) (*models.SLAPolicy, error)
	DeleteSLAPolicy(db *gorm.DB, companyID, id string) error

	// Reports
	GetSummaryStats(db *gorm.DB, companyID string) (*SummaryStats, error)
	GetStatsByCategory(db *gorm.DB, companyID string) ([]CategoryStat, error)
	GetStatsByAgent(db *gorm.DB, companyID string) ([]AgentStat, error)

	// ApplyAutoRoute applies routing rules and SLA deadline to a ticket in-place.
	// Called by the merchant portal after creating a ticket directly in the DB.
	ApplyAutoRoute(db *gorm.DB, ticket *models.Ticket)
}

type service struct {
	app *global.App
}

func NewService(app *global.App) Service {
	return &service{app: app}
}

// ── Ticket operations ─────────────────────────────────────────────────────────

func (s *service) ListTickets(db *gorm.DB, companyID string, filters TicketFilters) ([]models.Ticket, error) {
	var tickets []models.Ticket
	q := db.Preload("Category").Preload("Merchant").Where("is_trashed = false")

	if filters.Kind != "" {
		q = q.Where("kind = ?", filters.Kind)
	}
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.Priority != "" {
		q = q.Where("priority = ?", filters.Priority)
	}
	if filters.CategoryID != "" {
		q = q.Where("category_id = ?", filters.CategoryID)
	}
	if filters.AssigneeID != "" {
		q = q.Where("assigned_to_user_id = ?", filters.AssigneeID)
	}
	if filters.TeamID != "" {
		q = q.Where("assigned_to_team_id = ?", filters.TeamID)
	}
	if filters.Breached {
		q = q.Where("sla_deadline < ? AND status NOT IN ?", time.Now(), []string{"resolved", "closed"})
	}

	// Priority order: critical > high > medium > low
	q = q.Order("CASE priority WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, created_at DESC")

	if err := q.Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (s *service) GetTicket(db *gorm.DB, companyID, id string) (*models.Ticket, error) {
	var ticket models.Ticket
	err := db.Preload("Category").Preload("Merchant").Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("id = ? AND is_trashed = false", id).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (s *service) CreateTicket(db *gorm.DB, companyID, actorID, actorType string, req CreateTicketRequest) (*models.Ticket, error) {
	if req.Kind == "" || req.Subject == "" {
		return nil, errors.New("kind and subject are required")
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	source := req.Source
	if source == "" {
		source = "admin"
	}

	ticket := models.Ticket{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Kind:             req.Kind,
		Subject:          req.Subject,
		Description:      req.Description,
		Priority:         priority,
		Status:           "open",
		Source:           source,
	}

	if req.CategoryID != "" {
		if uid, err := uuid.Parse(req.CategoryID); err == nil {
			ticket.CategoryID = &uid
		}
	}
	if req.MerchantID != "" {
		if uid, err := uuid.Parse(req.MerchantID); err == nil {
			ticket.MerchantID = &uid
		}
	}
	if actorID != "" && actorType == "user" {
		if uid, err := uuid.Parse(actorID); err == nil {
			ticket.CreatedByUserID = &uid
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		ticket.TicketNumber = s.nextTicketNumber(tx, companyID)
		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}
		s.autoRoute(tx, &ticket)
		if ticket.AssignedToTeamID != nil || ticket.AssignedToUserID != nil {
			tx.Save(&ticket)
		}
		s.writeEvent(tx, &ticket, "creation", actorID, actorType, "Ticket created", false, nil)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetTicket(db, companyID, ticket.ID.String())
}

func (s *service) UpdateTicket(db *gorm.DB, companyID, id string, req UpdateTicketRequest) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := db.Where("id = ? AND is_trashed = false", id).First(&ticket).Error; err != nil {
		return nil, err
	}

	if req.Subject != nil {
		ticket.Subject = *req.Subject
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			ticket.CategoryID = nil
		} else if uid, err := uuid.Parse(*req.CategoryID); err == nil {
			ticket.CategoryID = &uid
		}
	}

	if err := db.Save(&ticket).Error; err != nil {
		return nil, err
	}
	return s.GetTicket(db, companyID, id)
}

func (s *service) ChangeStatus(db *gorm.DB, companyID, id, actorID, actorType string, req ChangeStatusRequest) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := db.Where("id = ? AND is_trashed = false", id).First(&ticket).Error; err != nil {
		return nil, err
	}

	oldStatus := ticket.Status
	ticket.Status = req.Status

	now := time.Now()
	if req.Status == "resolved" && ticket.ResolvedAt == nil {
		ticket.ResolvedAt = &now
	}

	if err := db.Save(&ticket).Error; err != nil {
		return nil, err
	}

	meta := map[string]any{"old_status": oldStatus, "new_status": req.Status}
	body := req.Comment
	if body == "" {
		body = "Status changed from " + oldStatus + " to " + req.Status
	}
	s.writeEvent(db, &ticket, "status_change", actorID, actorType, body, req.Comment != "", meta)

	if req.Status == "resolved" && ticket.MerchantID != nil {
		s.notifyMerchantResolved(db, &ticket)
	}

	return s.GetTicket(db, companyID, id)
}

func (s *service) AssignTicket(db *gorm.DB, companyID, id, actorID string, req AssignTicketRequest) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := db.Where("id = ? AND is_trashed = false", id).First(&ticket).Error; err != nil {
		return nil, err
	}

	meta := map[string]any{}

	if req.AssignedToUserID != nil {
		if *req.AssignedToUserID == "" {
			ticket.AssignedToUserID = nil
		} else if uid, err := uuid.Parse(*req.AssignedToUserID); err == nil {
			meta["assigned_to_user_id"] = uid.String()
			ticket.AssignedToUserID = &uid
		}
	}
	if req.AssignedToTeamID != nil {
		if *req.AssignedToTeamID == "" {
			ticket.AssignedToTeamID = nil
		} else if uid, err := uuid.Parse(*req.AssignedToTeamID); err == nil {
			meta["assigned_to_team_id"] = uid.String()
			ticket.AssignedToTeamID = &uid
		}
	}

	if err := db.Save(&ticket).Error; err != nil {
		return nil, err
	}

	note := req.Note
	if note == "" {
		note = "Ticket assigned"
	}
	s.writeEvent(db, &ticket, "assignment", actorID, "user", note, true, meta)
	s.notifyAssignedUser(db, &ticket)

	return s.GetTicket(db, companyID, id)
}

func (s *service) EscalateTicket(db *gorm.DB, companyID, id, actorID string) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := db.Where("id = ? AND is_trashed = false", id).First(&ticket).Error; err != nil {
		return nil, err
	}

	oldPriority := ticket.Priority
	switch ticket.Priority {
	case "low":
		ticket.Priority = "medium"
	case "medium":
		ticket.Priority = "high"
	case "high", "critical":
		ticket.Priority = "critical"
	}

	if err := db.Save(&ticket).Error; err != nil {
		return nil, err
	}

	meta := map[string]any{"old_priority": oldPriority, "new_priority": ticket.Priority}
	s.writeEvent(db, &ticket, "escalation", actorID, "user",
		"Ticket escalated from "+oldPriority+" to "+ticket.Priority, true, meta)

	return s.GetTicket(db, companyID, id)
}

func (s *service) TrashTicket(db *gorm.DB, companyID, id string) error {
	return db.Model(&models.Ticket{}).Where("id = ? AND company_id = ?", id, companyID).
		Update("is_trashed", true).Error
}

func (s *service) ListBreachedTickets(db *gorm.DB, companyID string) ([]models.Ticket, error) {
	var tickets []models.Ticket
	err := db.Preload("Category").Preload("Merchant").
		Where("sla_deadline < ? AND status NOT IN ? AND is_trashed = false",
			time.Now(), []string{"resolved", "closed"}).
		Order("sla_deadline ASC").
		Find(&tickets).Error
	return tickets, err
}

func (s *service) AddComment(db *gorm.DB, companyID, ticketID, actorID, actorType string, req AddCommentRequest) (*models.TicketEvent, error) {
	var ticket models.Ticket
	if err := db.Where("id = ? AND is_trashed = false", ticketID).First(&ticket).Error; err != nil {
		return nil, err
	}

	// Record first response time when staff replies for the first time
	if actorType == "user" && ticket.FirstResponseAt == nil {
		now := time.Now()
		ticket.FirstResponseAt = &now
		db.Save(&ticket)
	}

	actorUID, _ := uuid.Parse(actorID)
	event := models.TicketEvent{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		TicketID:         ticket.ID,
		EventType:        "comment",
		ActorID:          actorUID,
		ActorType:        actorType,
		Body:             req.Body,
		IsInternal:       req.IsInternal,
	}

	if err := db.Create(&event).Error; err != nil {
		return nil, err
	}

	// Notify merchant when staff posts a non-internal comment
	if actorType == "user" && !req.IsInternal && ticket.MerchantID != nil {
		s.notifyMerchantComment(db, &ticket)
	}

	return &event, nil
}

func (s *service) ListEvents(db *gorm.DB, companyID, ticketID string, includeInternal bool) ([]models.TicketEvent, error) {
	var events []models.TicketEvent
	q := db.Where("ticket_id = ?", ticketID)
	if !includeInternal {
		q = q.Where("is_internal = false")
	}
	err := q.Order("created_at ASC").Find(&events).Error
	return events, err
}

func (s *service) SweepBreachedTickets(db *gorm.DB) {
	var tickets []models.Ticket
	db.Where("sla_deadline < ? AND status NOT IN ? AND breach_notified_at IS NULL AND is_trashed = false",
		time.Now(), []string{"resolved", "closed"}).Find(&tickets)

	for _, t := range tickets {
		s.notifyAssignedUser(db, &t)
		now := time.Now()
		db.Model(&t).Update("breach_notified_at", &now)
	}
}

// ── Category operations ───────────────────────────────────────────────────────

func (s *service) ListCategories(db *gorm.DB, companyID string) ([]models.TicketCategory, error) {
	var categories []models.TicketCategory
	err := db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (s *service) GetCategory(db *gorm.DB, companyID, id string) (*models.TicketCategory, error) {
	var category models.TicketCategory
	err := db.Where("id = ?", id).First(&category).Error
	return &category, err
}

func (s *service) CreateCategory(db *gorm.DB, companyID string, req CreateCategoryRequest) (*models.TicketCategory, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	cat := models.TicketCategory{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Name:             req.Name,
		IsActive:         true,
	}
	if req.DepartmentID != "" {
		if uid, err := uuid.Parse(req.DepartmentID); err == nil {
			cat.DepartmentID = &uid
		}
	}
	if err := db.Create(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *service) UpdateCategory(db *gorm.DB, companyID, id string, req UpdateCategoryRequest) (*models.TicketCategory, error) {
	var cat models.TicketCategory
	if err := db.Where("id = ?", id).First(&cat).Error; err != nil {
		return nil, err
	}
	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.DepartmentID != nil {
		if *req.DepartmentID == "" {
			cat.DepartmentID = nil
		} else if uid, err := uuid.Parse(*req.DepartmentID); err == nil {
			cat.DepartmentID = &uid
		}
	}
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}
	if err := db.Save(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *service) DeleteCategory(db *gorm.DB, companyID, id string) error {
	return db.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.TicketCategory{}).Error
}

// ── Routing rule operations ───────────────────────────────────────────────────

func (s *service) ListRoutingRules(db *gorm.DB, companyID string) ([]models.CRMRoutingRule, error) {
	var rules []models.CRMRoutingRule
	err := db.Order("created_at DESC").Find(&rules).Error
	return rules, err
}

func (s *service) CreateRoutingRule(db *gorm.DB, companyID string, req CreateRoutingRuleRequest) (*models.CRMRoutingRule, error) {
	rule := models.CRMRoutingRule{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Priority:         req.Priority,
		IsActive:         true,
	}
	if req.CategoryID != nil {
		if uid, err := uuid.Parse(*req.CategoryID); err == nil {
			rule.CategoryID = &uid
		}
	}
	if req.DepartmentID != nil {
		if uid, err := uuid.Parse(*req.DepartmentID); err == nil {
			rule.DepartmentID = &uid
		}
	}
	if req.AssignedUserID != nil {
		if uid, err := uuid.Parse(*req.AssignedUserID); err == nil {
			rule.AssignedUserID = &uid
		}
	}
	if err := db.Create(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *service) UpdateRoutingRule(db *gorm.DB, companyID, id string, req UpdateRoutingRuleRequest) (*models.CRMRoutingRule, error) {
	var rule models.CRMRoutingRule
	if err := db.Where("id = ?", id).First(&rule).Error; err != nil {
		return nil, err
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			rule.CategoryID = nil
		} else if uid, err := uuid.Parse(*req.CategoryID); err == nil {
			rule.CategoryID = &uid
		}
	}
	if req.DepartmentID != nil {
		if *req.DepartmentID == "" {
			rule.DepartmentID = nil
		} else if uid, err := uuid.Parse(*req.DepartmentID); err == nil {
			rule.DepartmentID = &uid
		}
	}
	if req.AssignedUserID != nil {
		if *req.AssignedUserID == "" {
			rule.AssignedUserID = nil
		} else if uid, err := uuid.Parse(*req.AssignedUserID); err == nil {
			rule.AssignedUserID = &uid
		}
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if err := db.Save(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *service) DeleteRoutingRule(db *gorm.DB, companyID, id string) error {
	return db.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.CRMRoutingRule{}).Error
}

// ── SLA policy operations ─────────────────────────────────────────────────────

func (s *service) ListSLAPolicies(db *gorm.DB, companyID string) ([]models.SLAPolicy, error) {
	var policies []models.SLAPolicy
	err := db.Order("CASE priority WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END").
		Find(&policies).Error
	return policies, err
}

func (s *service) CreateSLAPolicy(db *gorm.DB, companyID string, req CreateSLAPolicyRequest) (*models.SLAPolicy, error) {
	policy := models.SLAPolicy{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: companyID},
		Priority:         req.Priority,
		ResponseHours:    req.ResponseHours,
		ResolutionHours:  req.ResolutionHours,
		IsActive:         true,
	}
	if err := db.Create(&policy).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

func (s *service) UpdateSLAPolicy(db *gorm.DB, companyID, id string, req UpdateSLAPolicyRequest) (*models.SLAPolicy, error) {
	var policy models.SLAPolicy
	if err := db.Where("id = ?", id).First(&policy).Error; err != nil {
		return nil, err
	}
	if req.ResponseHours != nil {
		policy.ResponseHours = *req.ResponseHours
	}
	if req.ResolutionHours != nil {
		policy.ResolutionHours = *req.ResolutionHours
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}
	if err := db.Save(&policy).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

func (s *service) DeleteSLAPolicy(db *gorm.DB, companyID, id string) error {
	return db.Where("id = ? AND company_id = ?", id, companyID).Delete(&models.SLAPolicy{}).Error
}

// ── Reports ───────────────────────────────────────────────────────────────────

func (s *service) GetSummaryStats(db *gorm.DB, companyID string) (*SummaryStats, error) {
	var stats SummaryStats

	db.Model(&models.Ticket{}).Where("status NOT IN ? AND is_trashed = false", []string{"resolved", "closed"}).
		Count(&stats.TotalOpen)

	db.Model(&models.Ticket{}).
		Where("sla_deadline < ? AND status NOT IN ? AND is_trashed = false",
			time.Now(), []string{"resolved", "closed"}).
		Count(&stats.TotalBreached)

	db.Model(&models.Ticket{}).Where("status = ? AND is_trashed = false", "resolved").
		Count(&stats.TotalResolved)

	db.Model(&models.Ticket{}).
		Where("status = ? AND resolved_at IS NOT NULL AND is_trashed = false", "resolved").
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/3600), 0)").
		Scan(&stats.AvgResolutionH)

	return &stats, nil
}

func (s *service) GetStatsByCategory(db *gorm.DB, companyID string) ([]CategoryStat, error) {
	var stats []CategoryStat
	db.Model(&models.Ticket{}).
		Select("category_id, ticket_categories.name as category_name, COUNT(*) as total, "+
			"SUM(CASE WHEN tickets.status NOT IN ('resolved','closed') THEN 1 ELSE 0 END) as open, "+
			"SUM(CASE WHEN tickets.status = 'resolved' THEN 1 ELSE 0 END) as resolved").
		Joins("LEFT JOIN ticket_categories ON ticket_categories.id = tickets.category_id").
		Where("tickets.is_trashed = false").
		Group("tickets.category_id, ticket_categories.name").
		Scan(&stats)
	return stats, nil
}

func (s *service) GetStatsByAgent(db *gorm.DB, companyID string) ([]AgentStat, error) {
	var stats []AgentStat
	db.Model(&models.Ticket{}).
		Select("assigned_to_user_id as user_id, COUNT(*) as total, "+
			"SUM(CASE WHEN status NOT IN ('resolved','closed') THEN 1 ELSE 0 END) as open, "+
			"SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0 END) as resolved").
		Where("assigned_to_user_id IS NOT NULL AND is_trashed = false").
		Group("assigned_to_user_id").
		Scan(&stats)
	return stats, nil
}

// ApplyAutoRoute is the public wrapper around autoRoute for use by other modules.
func (s *service) ApplyAutoRoute(db *gorm.DB, ticket *models.Ticket) {
	s.autoRoute(db, ticket)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// nextTicketNumber returns the next sequential ticket number for the company.
// Must be called inside a DB transaction to avoid race conditions.
func (s *service) nextTicketNumber(tx *gorm.DB, companyID string) int64 {
	var max int64
	tx.Model(&models.Ticket{}).Where("company_id = ?", companyID).
		Select("COALESCE(MAX(ticket_number), 0)").Scan(&max)
	return max + 1
}

// autoRoute applies the best matching routing rule to the ticket and sets the SLA deadline.
func (s *service) autoRoute(db *gorm.DB, ticket *models.Ticket) {
	var rule models.CRMRoutingRule

	// Try exact match: category + priority
	err := db.Where("is_active = true AND (category_id = ? OR category_id IS NULL) AND (priority = ? OR priority = 'any')",
		ticket.CategoryID, ticket.Priority).
		Order("category_id DESC NULLS LAST, CASE priority WHEN 'any' THEN 1 ELSE 0 END").
		First(&rule).Error

	if err == nil {
		if rule.DepartmentID != nil {
			ticket.AssignedToTeamID = rule.DepartmentID
		}
		if rule.AssignedUserID != nil {
			ticket.AssignedToUserID = rule.AssignedUserID
		}
	} else if ticket.CategoryID != nil {
		// Fall back to category default department
		var cat models.TicketCategory
		if db.Where("id = ?", ticket.CategoryID).First(&cat).Error == nil && cat.DepartmentID != nil {
			ticket.AssignedToTeamID = cat.DepartmentID
		}
	}

	// Set SLA deadline from policy
	deadline := s.computeSLADeadline(db, ticket.CompanyID, ticket.Priority)
	ticket.SLADeadline = deadline
}

// computeSLADeadline looks up the active SLA policy for the priority and returns now + resolution_hours.
func (s *service) computeSLADeadline(db *gorm.DB, companyID, priority string) *time.Time {
	var policy models.SLAPolicy
	err := db.Where("priority = ? AND is_active = true", priority).First(&policy).Error
	if err != nil || policy.ResolutionHours == 0 {
		return nil
	}
	deadline := time.Now().Add(time.Duration(policy.ResolutionHours) * time.Hour)
	return &deadline
}

// writeEvent creates a TicketEvent record for any state change or comment.
func (s *service) writeEvent(db *gorm.DB, ticket *models.Ticket, eventType, actorID, actorType, body string, isInternal bool, metadata map[string]any) {
	actorUID, _ := uuid.Parse(actorID)
	event := models.TicketEvent{
		CompanyBaseModel: models.CompanyBaseModel{CompanyID: ticket.CompanyID},
		TicketID:         ticket.ID,
		EventType:        eventType,
		ActorID:          actorUID,
		ActorType:        actorType,
		Body:             body,
		IsInternal:       isInternal,
	}
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			event.Metadata = string(b)
		}
	}
	if err := db.Create(&event).Error; err != nil {
		s.app.Logger.Warnf("CRM: failed to write event for ticket %s: %v", ticket.ID, err)
	}
}

// notifyAssignedUser sends an email to the ticket's assigned user if one is set.
func (s *service) notifyAssignedUser(db *gorm.DB, ticket *models.Ticket) {
	if s.app.Email == nil || ticket.AssignedToUserID == nil {
		return
	}
	var user models.User
	if err := db.First(&user, "id = ?", ticket.AssignedToUserID).Error; err != nil || !user.IsActive {
		return
	}
	subject := "[GeePay] Ticket assigned: " + ticket.Subject
	body := "You have been assigned ticket #" + formatTicketNumber(ticket.TicketNumber) + ": " + ticket.Subject
	if err := s.app.Email.Send(user.Email, subject, body); err != nil {
		s.app.Logger.Warnf("CRM: failed to email assignee for ticket %s: %v", ticket.ID, err)
	}
}

// notifyMerchantResolved sends a resolution email to the merchant.
func (s *service) notifyMerchantResolved(db *gorm.DB, ticket *models.Ticket) {
	if s.app.Email == nil || ticket.MerchantID == nil {
		return
	}
	var merchant models.Merchant
	if err := db.First(&merchant, "id = ?", ticket.MerchantID).Error; err != nil {
		return
	}
	subject := "[GeePay] Your ticket has been resolved: " + ticket.Subject
	body := "Your support ticket #" + formatTicketNumber(ticket.TicketNumber) + " has been resolved. Please contact us if you need further assistance."
	if err := s.app.Email.Send(merchant.Email, subject, body); err != nil {
		s.app.Logger.Warnf("CRM: failed to email merchant for resolved ticket %s: %v", ticket.ID, err)
	}
}

// notifyMerchantComment sends an email to the merchant when staff posts a public comment.
func (s *service) notifyMerchantComment(db *gorm.DB, ticket *models.Ticket) {
	if s.app.Email == nil || ticket.MerchantID == nil {
		return
	}
	var merchant models.Merchant
	if err := db.First(&merchant, "id = ?", ticket.MerchantID).Error; err != nil {
		return
	}
	subject := "[GeePay] Update on your ticket: " + ticket.Subject
	body := "There is a new update on your support ticket #" + formatTicketNumber(ticket.TicketNumber) + ". Please log in to view the response."
	if err := s.app.Email.Send(merchant.Email, subject, body); err != nil {
		s.app.Logger.Warnf("CRM: failed to email merchant for comment on ticket %s: %v", ticket.ID, err)
	}
}

func formatTicketNumber(n int64) string {
	return "TKT-" + padLeft(n, 4)
}

func padLeft(n int64, width int) string {
	s := ""
	tmp := n
	for tmp > 0 {
		s = string(rune('0'+tmp%10)) + s
		tmp /= 10
	}
	if n == 0 {
		s = "0"
	}
	for len(s) < width {
		s = "0" + s
	}
	return s
}
