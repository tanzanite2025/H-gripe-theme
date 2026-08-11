package payment

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type gatewayFallbackPaymentMethodResponse struct {
	Code     string `json:"code"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

func (h *Handler) ConfigurePaymentGatewayCircuitBreaker(
	circuitBreaker *service.PaymentGatewayCircuitBreakerService,
) {
	if h == nil {
		return
	}
	h.gatewayCircuitBreaker = circuitBreaker
}

func (h *Handler) allowPaymentGatewayAttemptOrRespondWithFallbackRecommendation(
	c *gin.Context,
	provider pgateway.GatewayType,
) bool {
	if h == nil || h.gatewayCircuitBreaker == nil {
		return true
	}

	decision, err := h.gatewayCircuitBreaker.IsGatewayPaymentAttemptAllowed(
		c.Request.Context(),
		string(provider),
	)
	if err != nil {
		// Circuit-breaker storage is deliberately fail-open. A Redis outage
		// must not turn into a site-wide payment outage.
		return true
	}
	if decision.Allowed {
		return true
	}

	h.respondPaymentGatewayDegradedWithFallbackRecommendation(c, provider, decision)
	return false
}

func (h *Handler) recordSuccessfulPaymentGatewayAPIResponse(
	ctx context.Context,
	provider pgateway.GatewayType,
) {
	if h == nil {
		return
	}
	if h.gatewayCircuitBreaker != nil {
		_ = h.gatewayCircuitBreaker.RecordSuccessfulGatewayAPIResponse(ctx, string(provider))
	}
	if h.antiFraud != nil {
		h.antiFraud.RecordProviderSuccess(string(provider))
	}
}

func (h *Handler) respondToPaymentGatewayOperationFailure(
	c *gin.Context,
	provider pgateway.GatewayType,
	status int,
	errorCode string,
	err error,
) {
	if !pgateway.IsTransientGatewayNetworkOrServerError(err) {
		apierror.RespondError(c, status, errorCode, err.Error())
		return
	}

	decision := service.PaymentGatewayCircuitBreakerDecision{Allowed: true}
	if h != nil && h.gatewayCircuitBreaker != nil {
		recordedDecision, recordErr := h.gatewayCircuitBreaker.RecordFailedGatewayAPIResponse(
			c.Request.Context(),
			string(provider),
		)
		if recordErr == nil {
			decision = recordedDecision
		}
	}
	if h != nil && h.antiFraud != nil {
		h.antiFraud.RecordProviderFailure(string(provider))
	}

	h.respondPaymentGatewayDegradedWithFallbackRecommendation(c, provider, decision)
}

func (h *Handler) respondPaymentGatewayDegradedWithFallbackRecommendation(
	c *gin.Context,
	provider pgateway.GatewayType,
	decision service.PaymentGatewayCircuitBreakerDecision,
) {
	retryAfterSeconds := durationToCeilingSeconds(decision.RetryAfter)
	if retryAfterSeconds > 0 {
		c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
	}

	status := http.StatusBadGateway
	if decision.CircuitOpen {
		status = http.StatusServiceUnavailable
	}

	apierror.RespondErrorWithDetails(
		c,
		status,
		"payment_gateway_degraded",
		"The selected payment provider is temporarily unavailable. Choose another payment method to continue.",
		gin.H{
			"gateway":             string(provider),
			"circuit_open":        decision.CircuitOpen,
			"retry_after_seconds": retryAfterSeconds,
			"failure_rate":        decision.FailureRate,
			"sample_count":        decision.SampleCount,
			"failure_count":       decision.FailureCount,
			"payment_attempt_may_still_be_processing": true,
			"fallback_payment_methods":                h.availableFallbackPaymentMethods(c, provider),
		},
	)
}

func (h *Handler) availableFallbackPaymentMethods(
	c *gin.Context,
	failedProvider pgateway.GatewayType,
) []gatewayFallbackPaymentMethodResponse {
	if h == nil || h.paymentService == nil {
		return nil
	}

	methods, err := h.paymentService.ListPaymentMethods(true)
	if err != nil {
		return nil
	}
	availabilityContext, err := h.resolvePaymentMethodAvailabilityContext(c)
	if err != nil {
		return nil
	}

	fallbackMethods := make([]gatewayFallbackPaymentMethodResponse, 0, len(methods))
	seen := make(map[string]struct{})
	for _, method := range methods {
		item, err := h.paymentMethodToAvailabilityResponse(c, method, availabilityContext)
		if err != nil || !item.Available {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.Provider), string(failedProvider)) {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(item.Code))
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(item.Provider))
		}
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		fallbackMethods = append(fallbackMethods, gatewayFallbackPaymentMethodResponse{
			Code:     item.Code,
			Provider: item.Provider,
			Name:     item.Name,
		})
	}
	return fallbackMethods
}

func (h *Handler) gatewayCircuitBreakerAvailability(
	c *gin.Context,
	provider pgateway.GatewayType,
) (bool, string) {
	if h == nil || h.gatewayCircuitBreaker == nil {
		return true, ""
	}
	decision, err := h.gatewayCircuitBreaker.IsGatewayPaymentAttemptAllowed(
		c.Request.Context(),
		string(provider),
	)
	if err != nil || decision.Allowed {
		return true, ""
	}
	return false, "gateway_circuit_open"
}

func durationToCeilingSeconds(value time.Duration) int64 {
	if value <= 0 {
		return 0
	}
	return int64((value + time.Second - 1) / time.Second)
}
