package admin

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type WarrantyHandler struct {
	warrantyService *service.WarrantyService
}

func NewWarrantyHandler(warrantyService *service.WarrantyService) *WarrantyHandler {
	return &WarrantyHandler{
		warrantyService: warrantyService,
	}
}

func (h *WarrantyHandler) ListAllWarrantyClaims(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	claims, total, err := h.warrantyService.GetAllWarrantyClaims(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, claims, params.Page, params.PageSize, total)
}

func (h *WarrantyHandler) GetWarrantyClaim(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	claim, err := h.warrantyService.GetWarrantyClaim(uint(id), 0, true)
	if err != nil {
		respondAdminWarrantyError(c, err)
		return
	}

	response.Success(c, claim)
}

func (h *WarrantyHandler) UpdateWarrantyClaimStatus(c *gin.Context) {
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

	if err := h.warrantyService.UpdateWarrantyClaimStatus(uint(id), req.Status, processedBy); err != nil {
		respondAdminWarrantyError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Warranty claim status updated", nil)
}

func (h *WarrantyHandler) UpdateWarrantyClaimResolution(c *gin.Context) {
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

	if err := h.warrantyService.UpdateWarrantyClaimResolution(uint(id), req.Resolution, processedBy); err != nil {
		respondAdminWarrantyError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Warranty claim resolution updated", nil)
}

func (h *WarrantyHandler) ListWarrantyClaimOrderItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	items, err := h.warrantyService.ListWarrantyClaimOrderItems(uint(id))
	if err != nil {
		respondAdminWarrantyError(c, err)
		return
	}

	response.Success(c, gin.H{"items": items})
}

func (h *WarrantyHandler) BindWarrantyClaimOrderItem(c *gin.Context) {
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

	if err := h.warrantyService.BindWarrantyClaimOrderItem(uint(id), req.OrderItemID); err != nil {
		respondAdminWarrantyError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Warranty claim order item updated", nil)
}

func (h *WarrantyHandler) ListWarrantyServiceRecords(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid claim ID")
		return
	}

	records, err := h.warrantyService.ListWarrantyServiceRecords(uint(id))
	if err != nil {
		respondAdminWarrantyError(c, err)
		return
	}

	response.Success(c, gin.H{"records": records})
}

func (h *WarrantyHandler) CreateWarrantyServiceRecord(c *gin.Context) {
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

	record, err := h.warrantyService.CreateWarrantyServiceRecord(uint(id), service.WarrantyServiceRecordInput{
		ServiceType: req.ServiceType,
		Status:      req.Status,
		Summary:     req.Summary,
		CostAmount:  req.CostAmount,
		Currency:    req.Currency,
		PerformedAt: performedAt,
	}, createdBy)
	if err != nil {
		respondAdminWarrantyError(c, err)
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

func respondAdminWarrantyError(c *gin.Context, err error) {
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
	default:
		apierror.RespondBadRequest(c, err.Error())
	}
}
