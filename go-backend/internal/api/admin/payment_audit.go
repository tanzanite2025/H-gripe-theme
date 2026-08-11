package admin

import (
	"sort"
	"strings"
	"time"

	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
)

const (
	paymentAuditActionUpdate = adminAuditActionUpdate
	paymentAuditActionDelete = adminAuditActionDelete
	paymentAuditActionProbe  = adminAuditActionProbe

	paymentAuditResourceGatewayConfig   = "payment_gateway_config"
	paymentAuditResourceGatewayCallback = "payment_gateway_callback"

	paymentAuditStatusSuccess = adminAuditStatusSuccess
	paymentAuditStatusFailed  = adminAuditStatusFailed
)

type paymentAuditRecorder = adminAuditRecorder

type paymentAdminAuditEvent = adminAuditEvent

func (h *PaymentHandler) ConfigureAuditService(recorder paymentAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func paymentAuditStartedAt() time.Time {
	return adminAuditStartedAt()
}

func (h *PaymentHandler) recordPaymentAdminAudit(c *gin.Context, event paymentAdminAuditEvent) {
	if h == nil {
		return
	}
	recordPaymentAdminAudit(h.auditService, c, event)
}

func recordPaymentAdminAudit(recorder paymentAuditRecorder, c *gin.Context, event paymentAdminAuditEvent) {
	recordAdminAudit(recorder, c, event)
}

func (h *PaymentHandler) recordPaymentGatewayConfigAudit(
	c *gin.Context,
	startedAt time.Time,
	provider pgateway.GatewayType,
	action string,
	status string,
	failureStage string,
	errorMessage string,
	details map[string]interface{},
) {
	if details == nil {
		details = map[string]interface{}{}
	}
	details["provider"] = string(provider)
	if failureStage != "" {
		details["failure_stage"] = failureStage
	}

	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:    startedAt,
		Action:       action,
		Resource:     paymentAuditResourceGatewayConfig,
		Status:       status,
		ErrorMessage: errorMessage,
		Changes:      details,
	})
}

func (h *PaymentHandler) recordPaymentGatewayCallbackAudit(
	c *gin.Context,
	startedAt time.Time,
	result paymentCallbackCheckResult,
) {
	status := paymentAuditStatusFailed
	errorMessage := result.Error
	if result.Reachable && result.ExpectedSignatureFailure {
		status = paymentAuditStatusSuccess
	} else if result.Reachable && errorMessage == "" {
		errorMessage = "callback probe did not receive an expected signature failure"
	}

	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:    startedAt,
		Action:       paymentAuditActionProbe,
		Resource:     paymentAuditResourceGatewayCallback,
		Status:       status,
		ErrorMessage: errorMessage,
		Changes: map[string]interface{}{
			"provider":                   string(result.Provider),
			"callback_url":               result.CallbackURL,
			"method":                     result.Method,
			"status_code":                result.StatusCode,
			"reachable":                  result.Reachable,
			"transport_reachable":        result.TransportReachable,
			"route_reachable":            result.RouteReachable,
			"expected_signature_failure": result.ExpectedSignatureFailure,
		},
	})
}

func paymentGatewayConfigRequestAuditDetails(provider pgateway.GatewayType, environment string, credentials map[string]string) map[string]interface{} {
	normalizedEnvironment := pgateway.NormalizeGatewayEnvironment(environment)
	details := map[string]interface{}{
		"provider":         string(provider),
		"environment":      normalizedEnvironment,
		"production":       normalizedEnvironment == gatewayEnvironmentProduction,
		"submitted_fields": paymentGatewayCredentialFieldNames(credentials),
	}
	return details
}

func paymentGatewayConfigStoredAuditDetails(config pgateway.SecureGatewayConfig) map[string]interface{} {
	return map[string]interface{}{
		"provider":          string(config.Provider),
		"environment":       config.Environment,
		"production":        config.Environment == gatewayEnvironmentProduction,
		"configured_fields": pgateway.SecureGatewayConfiguredFields(config),
	}
}

func paymentGatewayDeleteAuditDetails(provider pgateway.GatewayType, confirmationMatched bool) map[string]interface{} {
	return map[string]interface{}{
		"provider":             string(provider),
		"confirmation_matched": confirmationMatched,
	}
}

func paymentGatewayCredentialFieldNames(credentials map[string]string) []string {
	if len(credentials) == 0 {
		return []string{}
	}

	fields := make([]string, 0, len(credentials))
	seen := map[string]struct{}{}
	for key, value := range credentials {
		field := strings.TrimSpace(key)
		if field == "" || strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}
