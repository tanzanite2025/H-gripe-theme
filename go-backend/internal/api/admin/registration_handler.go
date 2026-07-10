package admin

import (
	"errors"
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

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

func respondAdminRegistrationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWarrantyEmailMismatch):
		apierror.RespondForbidden(c)
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
