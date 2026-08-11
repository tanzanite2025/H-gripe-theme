package admin

import (
	"errors"
	"strconv"
	"strings"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	registrationService *service.RegistrationService
}

func NewRegistrationHandler(registrationService *service.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{
		registrationService: registrationService,
	}
}

func (h *RegistrationHandler) ListAllRegistrations(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	registrations, total, err := h.registrationService.GetAllRegistrations(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, registrations, params.Page, params.PageSize, total)
}

func (h *RegistrationHandler) UpdateRegistrationStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	if err := h.registrationService.UpdateRegistrationStatus(uint(id), req.Status); err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Registration status updated", nil)
}

func (h *RegistrationHandler) GetRegistrationStats(c *gin.Context) {
	stats, err := h.registrationService.GetRegistrationStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, stats)
}

func (h *RegistrationHandler) GetExpiringWarranties(c *gin.Context) {
	days := pagination.ParseLimit(c)
	if days > 365 {
		days = 30
	}

	registrations, err := h.registrationService.GetExpiringWarranties(days)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": registrations})
}

func (h *RegistrationHandler) ListAllWarrantyClaims(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	claims, total, err := h.registrationService.GetAllWarrantyClaims(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, claims, params.Page, params.PageSize, total)
}

func (h *RegistrationHandler) GetWarrantyClaim(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	claim, err := h.registrationService.GetWarrantyClaim(uint(id), 0, true)
	if err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.Success(c, claim)
}

func (h *RegistrationHandler) UpdateWarrantyClaimStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	processedBy := uint(0)
	if userID, exists := c.Get("user_id"); exists {
		processedBy = userID.(uint)
	}

	if err := h.registrationService.UpdateWarrantyClaimStatus(uint(id), req.Status, processedBy); err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Warranty claim status updated", nil)
}

func (h *RegistrationHandler) UpdateWarrantyClaimResolution(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	var req struct {
		Resolution string `json:"resolution"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	processedBy := uint(0)
	if userID, exists := c.Get("user_id"); exists {
		processedBy = userID.(uint)
	}

	if err := h.registrationService.UpdateWarrantyClaimResolution(uint(id), req.Resolution, processedBy); err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Warranty claim resolution updated", nil)
}

func (h *RegistrationHandler) ListWarrantyClaimOrderItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	items, err := h.registrationService.ListWarrantyClaimOrderItems(uint(id))
	if err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.Success(c, gin.H{"items": items})
}

func (h *RegistrationHandler) BindWarrantyClaimOrderItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	var req struct {
		OrderItemID *uint `json:"order_item_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	if err := h.registrationService.BindWarrantyClaimOrderItem(uint(id), req.OrderItemID); err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Warranty claim order item updated", nil)
}

func (h *RegistrationHandler) ListWarrantyServiceRecords(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	records, err := h.registrationService.ListWarrantyServiceRecords(uint(id))
	if err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.Success(c, gin.H{"records": records})
}

func (h *RegistrationHandler) CreateWarrantyServiceRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	var req struct {
		ServiceType string  `json:"service_type"`
		Status      string  `json:"status"`
		Summary     string  `json:"summary" binding:"required"`
		CostAmount  float64 `json:"cost_amount"`
		Currency    string  `json:"currency"`
		PerformedAt string  `json:"performed_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	performedAt, err := parseAdminWarrantyPerformedAt(req.PerformedAt)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid performed_at")
		return
	}

	createdBy := uint(0)
	if userID, exists := c.Get("user_id"); exists {
		createdBy = userID.(uint)
	}

	record, err := h.registrationService.CreateWarrantyServiceRecord(uint(id), service.WarrantyServiceRecordInput{
		ServiceType: req.ServiceType,
		Status:      req.Status,
		Summary:     req.Summary,
		CostAmount:  req.CostAmount,
		Currency:    req.Currency,
		PerformedAt: performedAt,
	}, createdBy)
	if err != nil {
		respondAdminRegistrationError(c, err)
		return
	}

	response.Created(c, record)
}

func parseAdminWarrantyPerformedAt(value string) (*time.Time, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, normalized); err == nil {
		return &parsed, nil
	}

	parsed, err := time.Parse("2006-01-02", normalized)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func respondAdminRegistrationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWarrantyEmailMismatch):
		apierror.RespondForbidden(c)
	case errors.Is(err, service.ErrWarrantyOrderItemMismatch):
		apierror.RespondBadRequest(c, "Order item does not match warranty claim")
	case errors.Is(err, service.ErrWarrantyOrderItemUnavailable):
		apierror.RespondBadRequest(c, "Order item binding is unavailable")
	case service.IsRecordNotFound(err):
		apierror.RespondNotFound(c, "Resource")
	case err.Error() == "unauthorized":
		apierror.RespondForbidden(c)
	case err.Error() == "serial number already registered":
		apierror.RespondConflict(c, "Serial number already registered")
	case err.Error() == "product not found":
		apierror.RespondNotFound(c, "Product")
	default:
		apierror.RespondBadRequest(c, err.Error())
	}
}
