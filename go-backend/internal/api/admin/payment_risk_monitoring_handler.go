package admin

import (
	"errors"
	"strings"
	"time"

	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

var paymentRiskMonitoringProviders = []string{
	string(paymentdomain.PaymentRiskProviderStripe),
	string(paymentdomain.PaymentRiskProviderPayPal),
}

type PaymentRiskMonitoringHandler struct {
	monitoring   *service.PaymentRiskMonitoringService
	auditService paymentAuditRecorder
}

func NewPaymentRiskMonitoringHandler(monitoring *service.PaymentRiskMonitoringService) *PaymentRiskMonitoringHandler {
	return &PaymentRiskMonitoringHandler{monitoring: monitoring}
}

func (h *PaymentRiskMonitoringHandler) GetSummary(c *gin.Context) {
	if h == nil || h.monitoring == nil {
		apierror.RespondInternalError(c, errors.New("payment risk monitoring service is not configured"))
		return
	}
	providers, err := parsePaymentRiskMonitoringProviders(c.Query("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	reports, err := h.currentReports(providers)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{
		"enabled": h.monitoring.Enabled(),
		"reports": reports,
	})
}

func (h *PaymentRiskMonitoringHandler) RecomputeSummary(c *gin.Context) {
	if h == nil {
		apierror.RespondInternalError(c, errors.New("payment risk monitoring service is not configured"))
		return
	}
	startedAt := paymentAuditStartedAt()
	if h.monitoring == nil {
		err := errors.New("payment risk monitoring service is not configured")
		h.recordRiskMonitoringAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionRecompute,
			Resource:     paymentAuditResourceRiskMonitoring,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRiskRecomputeAuditDetails(nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	var req struct {
		Provider string `json:"provider"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			h.recordRiskMonitoringAudit(c, paymentAdminAuditEvent{
				StartedAt:    startedAt,
				Action:       paymentAuditActionRecompute,
				Resource:     paymentAuditResourceRiskMonitoring,
				Status:       paymentAuditStatusFailed,
				ErrorMessage: err.Error(),
				Changes:      paymentRiskRecomputeAuditDetails(nil, nil),
			})
			apierror.RespondBadRequest(c, err.Error())
			return
		}
	}
	providers, err := parsePaymentRiskMonitoringProviders(req.Provider)
	if err != nil {
		h.recordRiskMonitoringAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionRecompute,
			Resource:     paymentAuditResourceRiskMonitoring,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"raw_provider": strings.TrimSpace(req.Provider),
			},
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	reports := make(map[string]*service.PaymentRiskReport, len(providers))
	for _, provider := range providers {
		report, recomputeErr := h.monitoring.RecomputeProvider(c.Request.Context(), provider, time.Now().UTC())
		if recomputeErr != nil {
			details := paymentRiskRecomputeAuditDetails(providers, reports)
			details["failed_provider"] = provider
			h.recordRiskMonitoringAudit(c, paymentAdminAuditEvent{
				StartedAt:    startedAt,
				Action:       paymentAuditActionRecompute,
				Resource:     paymentAuditResourceRiskMonitoring,
				Status:       paymentAuditStatusFailed,
				ErrorMessage: recomputeErr.Error(),
				Changes:      details,
			})
			apierror.RespondInternalError(c, recomputeErr)
			return
		}
		reports[provider] = report
	}
	h.recordRiskMonitoringAudit(c, paymentAdminAuditEvent{
		StartedAt: startedAt,
		Action:    paymentAuditActionRecompute,
		Resource:  paymentAuditResourceRiskMonitoring,
		Status:    paymentAuditStatusSuccess,
		Changes:   paymentRiskRecomputeAuditDetails(providers, reports),
	})
	response.Success(c, gin.H{
		"enabled": h.monitoring.Enabled(),
		"reports": reports,
	})
}

func (h *PaymentRiskMonitoringHandler) currentReports(providers []string) (map[string]*service.PaymentRiskReport, error) {
	reports := make(map[string]*service.PaymentRiskReport, len(providers))
	for _, provider := range providers {
		report, err := h.monitoring.CurrentReport(provider)
		if err != nil {
			return nil, err
		}
		reports[provider] = report
	}
	return reports, nil
}

func parsePaymentRiskMonitoringProviders(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return append([]string(nil), paymentRiskMonitoringProviders...), nil
	}

	providers := make([]string, 0, len(paymentRiskMonitoringProviders))
	seen := make(map[string]struct{})
	for _, value := range strings.Split(raw, ",") {
		provider := strings.ToLower(strings.TrimSpace(value))
		if provider == "" {
			continue
		}
		if !isPaymentRiskMonitoringProvider(provider) {
			return nil, errors.New("unsupported payment risk provider")
		}
		if _, exists := seen[provider]; !exists {
			seen[provider] = struct{}{}
			providers = append(providers, provider)
		}
	}
	if len(providers) == 0 {
		return nil, errors.New("at least one payment risk provider is required")
	}
	return providers, nil
}

func isPaymentRiskMonitoringProvider(provider string) bool {
	for _, supported := range paymentRiskMonitoringProviders {
		if provider == supported {
			return true
		}
	}
	return false
}
