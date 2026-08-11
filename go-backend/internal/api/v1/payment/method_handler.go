package payment

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListPaymentMethods(c *gin.Context) {
	methods, err := h.paymentService.ListPaymentMethods(true)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	items, err := h.paymentMethodsToAvailabilityResponse(c, methods)
	if err != nil {
		apierror.RespondError(c, 503, "payment_methods_unavailable", "Payment methods are temporarily unavailable")
		return
	}

	response.Success(c, gin.H{"data": items})
}

func (h *Handler) GetPaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}

	method, err := h.paymentService.GetPaymentMethod(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Payment method")
		return
	}
	if !method.Enabled {
		apierror.RespondNotFound(c, "Payment method")
		return
	}

	context, err := h.resolvePaymentMethodAvailabilityContext(c)
	if err != nil {
		apierror.RespondError(c, 503, "payment_methods_unavailable", "Payment methods are temporarily unavailable")
		return
	}

	item, err := h.paymentMethodToAvailabilityResponse(c, *method, context)
	if err != nil {
		apierror.RespondError(c, 503, "payment_methods_unavailable", "Payment methods are temporarily unavailable")
		return
	}

	response.Success(c, item)
}
