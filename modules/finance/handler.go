package finance

import (
	"net/http"
	"strconv"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Handler defines HTTP handlers for finance module.
type Handler interface {
	// Budgets
	ListBudgets(c echo.Context) error
	GetBudget(c echo.Context) error
	CreateBudget(c echo.Context) error
	UpdateBudget(c echo.Context) error
	DeleteBudget(c echo.Context) error
	RenewBudget(c echo.Context) error
	ExportBudgetsCSV(c echo.Context) error

	// Licenses
	ListLicenses(c echo.Context) error
	GetLicense(c echo.Context) error
	CreateLicense(c echo.Context) error
	UpdateLicense(c echo.Context) error
	DeleteLicense(c echo.Context) error
	RenewLicense(c echo.Context) error
	ExportLicensesCSV(c echo.Context) error

	// Statutories
	ListStatutories(c echo.Context) error
	GetStatutory(c echo.Context) error
	CreateStatutory(c echo.Context) error
	UpdateStatutory(c echo.Context) error
	DeleteStatutory(c echo.Context) error
	MarkStatutoryPaid(c echo.Context) error

	// Salary advances
	ListSalaryAdvances(c echo.Context) error
	GetSalaryAdvance(c echo.Context) error
	CreateSalaryAdvance(c echo.Context) error
	UpdateSalaryAdvance(c echo.Context) error
	DeleteSalaryAdvance(c echo.Context) error
	ApproveSalaryAdvance(c echo.Context) error
	DeductSalaryAdvance(c echo.Context) error
	ExportSalaryAdvanceCSV(c echo.Context) error
}

type handler struct {
	service Service
}

func NewHandler(s Service) Handler {
	return &handler{service: s}
}

// --- Budgets ---
func (h *handler) ListBudgets(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.service.ListBudgets(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list budgets"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetBudget(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.GetBudget(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("budget not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateBudget(c echo.Context) error {
	var req CreateBudgetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.CreateBudget(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("budget created", item))
}

func (h *handler) UpdateBudget(c echo.Context) error {
	id := c.Param("id")
	var req UpdateBudgetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.UpdateBudget(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("budget updated", item))
}

func (h *handler) DeleteBudget(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.service.DeleteBudget(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) RenewBudget(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.RenewBudget(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("failed to renew"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("renewed", item))
}

func (h *handler) ExportBudgetsCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.service.ExportBudgetsCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=budgets.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}

// --- Licenses ---
func (h *handler) ListLicenses(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.service.ListLicenses(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list licenses"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetLicense(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.GetLicense(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("license not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateLicense(c echo.Context) error {
	var req CreateLicenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.CreateLicense(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("license created", item))
}

func (h *handler) UpdateLicense(c echo.Context) error {
	id := c.Param("id")
	var req UpdateLicenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.UpdateLicense(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("license updated", item))
}

func (h *handler) DeleteLicense(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.service.DeleteLicense(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) RenewLicense(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.RenewLicense(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("failed to renew"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("renewed", item))
}

func (h *handler) ExportLicensesCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.service.ExportLicensesCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=licenses.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}

// --- Statutories ---
func (h *handler) ListStatutories(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.service.ListStatutories(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list statutories"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetStatutory(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.GetStatutory(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateStatutory(c echo.Context) error {
	var req CreateStatutoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.CreateStatutory(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("created", item))
}

func (h *handler) UpdateStatutory(c echo.Context) error {
	id := c.Param("id")
	var req UpdateStatutoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.UpdateStatutory(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("updated", item))
}

func (h *handler) DeleteStatutory(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.service.DeleteStatutory(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) MarkStatutoryPaid(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.MarkStatutoryPaid(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("failed to mark paid"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("marked paid", item))
}

// --- Salary Advances ---
func (h *handler) ListSalaryAdvances(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.service.ListSalaryAdvances(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetSalaryAdvance(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.GetSalaryAdvance(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateSalaryAdvance(c echo.Context) error {
	var req CreateSalaryAdvanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.CreateSalaryAdvance(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("created", item))
}

func (h *handler) UpdateSalaryAdvance(c echo.Context) error {
	id := c.Param("id")
	var req UpdateSalaryAdvanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.UpdateSalaryAdvance(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("updated", item))
}

func (h *handler) DeleteSalaryAdvance(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.service.DeleteSalaryAdvance(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) ApproveSalaryAdvance(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.ApproveSalaryAdvance(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("failed to approve"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("approved", item))
}

func (h *handler) DeductSalaryAdvance(c echo.Context) error {
	id := c.Param("id")
	monthsStr := c.QueryParam("months")
	months, _ := strconv.Atoi(monthsStr)
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.service.DeductSalaryAdvance(db, companyID, id, months)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("failed to set deduction"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deduction set", item))
}

func (h *handler) ExportSalaryAdvanceCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.service.ExportSalaryAdvancesCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=salary_advances.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}
