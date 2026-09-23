package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) BackfillHistoricalRefundFXSnapshot(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	id, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}
	var req struct {
		BaseCurrency  string     `json:"base_currency"`
		OrderCurrency string     `json:"order_currency"`
		RateDecimal   string     `json:"rate_decimal"`
		Source        string     `json:"source"`
		CapturedAt    time.Time  `json:"captured_at"`
		RateFetchedAt *time.Time `json:"rate_fetched_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{StartedAt: startedAt, Action: paymentAuditActionUpdate, Resource: paymentAuditResourceFXSnapshot, ResourceID: uint(id), Status: paymentAuditStatusFailed, ErrorMessage: err.Error()})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.paymentService == nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{StartedAt: startedAt, Action: paymentAuditActionUpdate, Resource: paymentAuditResourceFXSnapshot, ResourceID: uint(id), Status: paymentAuditStatusFailed, ErrorMessage: service.ErrHistoricalRefundFXSnapshotAdminUnavailable.Error()})
		apierror.RespondError(c, http.StatusServiceUnavailable, "historical_fx_snapshot_unavailable", "historical FX snapshot service is unavailable")
		return
	}
	result, err := h.paymentService.BackfillHistoricalRefundFXSnapshot(service.BackfillHistoricalRefundFXSnapshotInput{
		OrderID: uint(id), BaseCurrency: req.BaseCurrency, OrderCurrency: req.OrderCurrency,
		RateDecimal: req.RateDecimal, Source: req.Source, CapturedAt: req.CapturedAt, RateFetchedAt: req.RateFetchedAt,
	})
	if err != nil {
		status := paymentAuditStatusFailed
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt: startedAt, Action: paymentAuditActionUpdate, Resource: paymentAuditResourceFXSnapshot, ResourceID: uint(id), Status: status, ErrorMessage: err.Error(),
			Changes: map[string]interface{}{"base_currency": strings.ToUpper(strings.TrimSpace(req.BaseCurrency)), "order_currency": strings.ToUpper(strings.TrimSpace(req.OrderCurrency)), "source": strings.TrimSpace(req.Source)},
		})
		switch {
		case errors.Is(err, service.ErrHistoricalRefundFXSnapshotAlreadyPresent):
			apierror.RespondConflict(c, err.Error())
		case errors.Is(err, service.ErrHistoricalRefundFXSnapshotAdminUnavailable):
			apierror.RespondError(c, http.StatusServiceUnavailable, "historical_fx_snapshot_unavailable", "historical FX snapshot service is unavailable")
		case errors.Is(err, service.ErrHistoricalRefundFXSnapshotNotApplicable), strings.Contains(strings.ToLower(err.Error()), "does not match"):
			apierror.RespondBadRequest(c, err.Error())
		default:
			if strings.Contains(strings.ToLower(err.Error()), "record not found") {
				apierror.RespondNotFound(c, "Order")
			} else {
				apierror.RespondBadRequest(c, err.Error())
			}
		}
		return
	}
	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{StartedAt: startedAt, Action: paymentAuditActionUpdate, Resource: paymentAuditResourceFXSnapshot, ResourceID: uint(id), Status: paymentAuditStatusSuccess, Changes: map[string]interface{}{"base_currency": result.Snapshot.BaseCurrency, "order_currency": result.Snapshot.OrderCurrency, "source": result.Snapshot.Source, "captured_at": result.Snapshot.CapturedAt}})
	response.Success(c, result)
}
