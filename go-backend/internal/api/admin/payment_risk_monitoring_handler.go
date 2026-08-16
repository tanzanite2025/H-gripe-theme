package admin

import (
	"errors"
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/config"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

var paymentRiskMonitoringProviders = []string{
	string(paymentdomain.PaymentRiskProviderStripe),
	string(paymentdomain.PaymentRiskProviderPayPal),
}

type PaymentRiskMonitoringHandler struct {
	monitoring     *service.PaymentRiskMonitoringService
	threeDS        *service.PaymentThreeDSPolicyService
	protection     *service.PaymentProtectionService
	circuitBreaker *service.PaymentGatewayCircuitBreakerService
	config         *config.Config
	runtimeReader  func() pgateway.RuntimeReadiness
	auditService   paymentAuditRecorder
}

func NewPaymentRiskMonitoringHandler(monitoring *service.PaymentRiskMonitoringService) *PaymentRiskMonitoringHandler {
	return &PaymentRiskMonitoringHandler{monitoring: monitoring}
}

func (h *PaymentRiskMonitoringHandler) ConfigureRiskConfiguration(
	cfg *config.Config,
	threeDS *service.PaymentThreeDSPolicyService,
	protection *service.PaymentProtectionService,
	circuitBreaker *service.PaymentGatewayCircuitBreakerService,
) {
	if h == nil {
		return
	}
	h.config = cfg
	h.threeDS = threeDS
	h.protection = protection
	h.circuitBreaker = circuitBreaker
}

func (h *PaymentRiskMonitoringHandler) ConfigureGatewayRuntimeReader(reader func() pgateway.RuntimeReadiness) {
	if h == nil {
		return
	}
	h.runtimeReader = reader
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
	response.Success(c, h.summaryPayload(c, reports))
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
	response.Success(c, h.summaryPayload(c, reports))
}

func (h *PaymentRiskMonitoringHandler) summaryPayload(
	c *gin.Context,
	reports map[string]*service.PaymentRiskReport,
) gin.H {
	payload := gin.H{
		"enabled": h.monitoring.Enabled(),
		"policy":  h.monitoring.PolicyView(),
		"reports": reports,
		"configuration": service.BuildPaymentRiskConfigurationView(
			h.config,
			h.monitoring,
			h.threeDS,
			h.protection,
			h.circuitBreaker,
		),
		"gateway_health": h.gatewayHealth(c),
	}
	if h.runtimeReader != nil {
		payload["gateway_runtime"] = h.runtimeReader()
	}
	return payload
}

func (h *PaymentRiskMonitoringHandler) gatewayHealth(c *gin.Context) map[string]service.PaymentGatewayHealthView {
	health := make(map[string]service.PaymentGatewayHealthView)
	if h == nil || h.circuitBreaker == nil {
		return health
	}
	for _, provider := range []string{"stripe", "paypal", "alipay", "wechat"} {
		health[provider] = h.circuitBreaker.CurrentHealth(c.Request.Context(), provider)
	}
	return health
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
