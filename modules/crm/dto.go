package crm

// --- Ticket DTOs ---

type CreateTicketRequest struct {
	Kind        string `json:"kind"`        // "support" | "incident"
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Priority    string `json:"priority"`    // "low" | "medium" | "high" | "critical"
	CategoryID  string `json:"category_id"`
	MerchantID  string `json:"merchant_id"` // optional: admin creating on behalf of a merchant
	Source      string `json:"-"`           // set by handler, not client input
}

type UpdateTicketRequest struct {
	Subject     *string `json:"subject"`
	Description *string `json:"description"`
	Priority    *string `json:"priority"`
	CategoryID  *string `json:"category_id"`
}

type ChangeStatusRequest struct {
	Status  string `json:"status"`
	Comment string `json:"comment"` // optional internal note written as an event
}

type AssignTicketRequest struct {
	AssignedToUserID *string `json:"assigned_to_user_id"`
	AssignedToTeamID *string `json:"assigned_to_team_id"`
	Note             string  `json:"note"`
}

type AddCommentRequest struct {
	Body       string `json:"body"`
	IsInternal bool   `json:"is_internal"`
}

type TicketFilters struct {
	Kind       string
	Status     string
	Priority   string
	CategoryID string
	AssigneeID string
	TeamID     string
	Breached   bool   // filter to only SLA-breached tickets
}

// --- Category DTOs ---

type CreateCategoryRequest struct {
	Name         string `json:"name"`
	DepartmentID string `json:"department_id"`
}

type UpdateCategoryRequest struct {
	Name         *string `json:"name"`
	DepartmentID *string `json:"department_id"`
	IsActive     *bool   `json:"is_active"`
}

// --- Routing Rule DTOs ---

type CreateRoutingRuleRequest struct {
	CategoryID     *string `json:"category_id"`
	Priority       string  `json:"priority"` // "low"|"medium"|"high"|"critical"|"any"
	DepartmentID   *string `json:"department_id"`
	AssignedUserID *string `json:"assigned_user_id"`
}

type UpdateRoutingRuleRequest struct {
	CategoryID     *string `json:"category_id"`
	Priority       *string `json:"priority"`
	DepartmentID   *string `json:"department_id"`
	AssignedUserID *string `json:"assigned_user_id"`
	IsActive       *bool   `json:"is_active"`
}

// --- SLA Policy DTOs ---

type CreateSLAPolicyRequest struct {
	Priority        string `json:"priority"`
	ResponseHours   int    `json:"response_hours"`
	ResolutionHours int    `json:"resolution_hours"`
}

type UpdateSLAPolicyRequest struct {
	ResponseHours   *int  `json:"response_hours"`
	ResolutionHours *int  `json:"resolution_hours"`
	IsActive        *bool `json:"is_active"`
}

// --- Report DTOs ---

type SummaryStats struct {
	TotalOpen      int64   `json:"total_open"`
	TotalBreached  int64   `json:"total_breached"`
	TotalResolved  int64   `json:"total_resolved"`
	AvgResolutionH float64 `json:"avg_resolution_hours"`
}

type CategoryStat struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	Total        int64  `json:"total"`
	Open         int64  `json:"open"`
	Resolved     int64  `json:"resolved"`
}

type AgentStat struct {
	UserID   string `json:"user_id"`
	Total    int64  `json:"total"`
	Open     int64  `json:"open"`
	Resolved int64  `json:"resolved"`
}
