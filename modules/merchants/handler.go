package merchants

import (
	"net/http"

	"backend/pkg/response"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler interface {
	// Merchants
	ListMerchants(c echo.Context) error
	GetMerchant(c echo.Context) error
	CreateMerchant(c echo.Context) error
	UpdateMerchant(c echo.Context) error
	DeleteMerchant(c echo.Context) error
	ExportMerchantsCSV(c echo.Context) error
	EnableMerchantPortal(c echo.Context) error

	// Merchant Statements
	ListStatements(c echo.Context) error
	GetStatement(c echo.Context) error
	CreateStatement(c echo.Context) error
	UpdateStatement(c echo.Context) error
	DeleteStatement(c echo.Context) error
	ExportStatementsCSV(c echo.Context) error
}

type handler struct {
	svc Service
}

func NewHandler(s Service) Handler {
	return &handler{svc: s}
}

// --- Merchants ---

func (h *handler) ListMerchants(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListMerchants(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list merchants"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetMerchant(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetMerchant(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("merchant not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateMerchant(c echo.Context) error {
	var req CreateMerchantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateMerchant(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("merchant created", item))
}

func (h *handler) UpdateMerchant(c echo.Context) error {
	id := c.Param("id")
	var req UpdateMerchantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateMerchant(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("merchant updated", item))
}

func (h *handler) DeleteMerchant(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteMerchant(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) EnableMerchantPortal(c echo.Context) error {
	merchantID := c.Param("id")
	var req EnableMerchantPortalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	merchant, err := h.svc.EnableMerchantPortal(db, merchantID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("merchant portal updated", merchant))
}

func (h *handler) ExportMerchantsCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.svc.ExportMerchantsCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=merchants.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}

// --- Merchant Statements ---

func (h *handler) ListStatements(c echo.Context) error {
	merchantID := c.Param("merchantId")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	items, err := h.svc.ListStatements(db, companyID, merchantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to list statements"))
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *handler) GetStatement(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.GetStatement(db, companyID, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.Error("statement not found"))
	}
	return c.JSON(http.StatusOK, response.Success(item))
}

func (h *handler) CreateStatement(c echo.Context) error {
	merchantID := c.Param("merchantId")
	var req CreateStatementRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	req.MerchantID = merchantID
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.CreateStatement(db, companyID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusCreated, response.SuccessWithMessage("statement created", item))
}

func (h *handler) UpdateStatement(c echo.Context) error {
	id := c.Param("id")
	var req UpdateStatementRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Error("invalid request"))
	}
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	item, err := h.svc.UpdateStatement(db, companyID, id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Error(err.Error()))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("statement updated", item))
}

func (h *handler) DeleteStatement(c echo.Context) error {
	id := c.Param("id")
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	if err := h.svc.DeleteStatement(db, companyID, id); err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to delete"))
	}
	return c.JSON(http.StatusOK, response.SuccessWithMessage("deleted", nil))
}

func (h *handler) ExportStatementsCSV(c echo.Context) error {
	db := c.Get("db").(*gorm.DB)
	companyID, _ := c.Get("companyID").(string)
	data, err := h.svc.ExportStatementsCSV(db, companyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Error("failed to export csv"))
	}
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=merchant_statements.csv")
	return c.Blob(http.StatusOK, "text/csv", data)
}
