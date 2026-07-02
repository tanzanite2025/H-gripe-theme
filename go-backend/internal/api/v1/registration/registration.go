package registration

import (
	"errors"
	"strconv"
	domainregistration "tanzanite/internal/domain/registration"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateRegistration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var reg domainregistration.ProductRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	reg.UserID = userID.(uint)
	reg.Status = "active"

	if err := h.registrationSvc.CreateRegistration(&reg); err != nil {
		respondRegistrationServiceError(c, err)
		return
	}

	response.Created(c, reg)
}

func (h *Handler) GetRegistration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	reg, err := h.registrationSvc.GetRegistration(uint(id), userID.(uint), false)
	if err != nil {
		respondRegistrationServiceError(c, err)
		return
	}

	response.Success(c, reg)
}

func (h *Handler) ListUserRegistrations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	params := pagination.ParsePagination(c)
	registrations, total, err := h.registrationSvc.GetUserRegistrations(userID.(uint), params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, registrations, params.Page, params.PageSize, total)
}

func (h *Handler) ListAllRegistrations(c *gin.Context) {
	params := pagination.ParsePagination(c)
	status := c.Query("status")

	registrations, total, err := h.registrationSvc.GetAllRegistrations(params.Page, params.PageSize, status)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, registrations, params.Page, params.PageSize, total)
}

func (h *Handler) UpdateRegistration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid registration ID")
		return
	}

	var reg domainregistration.ProductRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	reg.ID = uint(id)
	reg.UserID = userID.(uint)

	if err := h.registrationSvc.UpdateRegistration(&reg, userID.(uint), false); err != nil {
		respondRegistrationServiceError(c, err)
		return
	}

	response.Success(c, reg)
}

func (h *Handler) UpdateRegistrationStatus(c *gin.Context) {
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

	if err := h.registrationSvc.UpdateRegistrationStatus(uint(id), req.Status); err != nil {
		respondRegistrationServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Registration status updated", nil)
}

func (h *Handler) GetRegistrationStats(c *gin.Context) {
	stats, err := h.registrationSvc.GetRegistrationStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, stats)
}

func respondRegistrationServiceError(c *gin.Context, err error) {
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
