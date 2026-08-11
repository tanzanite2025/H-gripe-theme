package admin

import (
	"errors"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) ListStripeDisputes(c *gin.Context) {
	params := pagination.ParsePagination(c)
	disputes, total, err := h.paymentService.ListStripeDisputes(
		strings.TrimSpace(c.Query("status")),
		params.Page,
		params.PageSize,
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Paged(c, disputes, params.Page, params.PageSize, total)
}

func (h *PaymentHandler) GetStripeDispute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid dispute id")
		return
	}
	dispute, err := h.paymentService.GetStripeDispute(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Stripe dispute")
		return
	}
	response.Success(c, dispute)
}

func (h *PaymentHandler) ListPayPalDisputes(c *gin.Context) {
	params := pagination.ParsePagination(c)
	disputes, total, err := h.paymentService.ListPayPalDisputes(
		strings.TrimSpace(c.Query("status")),
		params.Page,
		params.PageSize,
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Paged(c, disputes, params.Page, params.PageSize, total)
}

func (h *PaymentHandler) GetPayPalDispute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid PayPal dispute id")
		return
	}
	dispute, err := h.paymentService.GetPayPalDispute(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrPayPalDisputeNotFound) {
			apierror.RespondNotFound(c, "PayPal dispute")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, dispute)
}
