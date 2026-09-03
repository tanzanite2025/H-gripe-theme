package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/currency"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/pkg/antifraud"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/cardtesting"
	"commerce-platform/internal/pkg/logger"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	paymentService        *service.PaymentService
	orderService          *service.OrderService
	settingsService       *service.AdminSettingsService
	threeDSPolicy         *service.PaymentThreeDSPolicyService
	riskMonitoring        *service.PaymentRiskMonitoringService
	protection            *service.PaymentProtectionService
	refundReview          *service.PaymentRefundRecommendationService
	storefrontContext     *service.StorefrontContextService
	antiBot               *antibot.Service
	antiFraud             *antifraud.Service
	cardBINLimiter        *cardtesting.Service
	gatewayCircuitBreaker *service.PaymentGatewayCircuitBreakerService
	gatewayFactory        func(*pgateway.Config) (pgateway.PaymentGateway, error)
	publicBaseURL         string
}

func NewHandler(
	paymentService *service.PaymentService,
	orderService *service.OrderService,
	settingsService *service.AdminSettingsService,
	threeDSPolicy *service.PaymentThreeDSPolicyService,
	riskMonitoring *service.PaymentRiskMonitoringService,
	protection *service.PaymentProtectionService,
	refundReview *service.PaymentRefundRecommendationService,
	antiBot *antibot.Service,
	antiFraud *antifraud.Service,
	storefrontContexts ...*service.StorefrontContextService,
) *Handler {
	handler := &Handler{
		paymentService:  paymentService,
		orderService:    orderService,
		settingsService: settingsService,
		threeDSPolicy:   threeDSPolicy,
		riskMonitoring:  riskMonitoring,
		protection:      protection,
		refundReview:    refundReview,
		antiBot:         antiBot,
		antiFraud:       antiFraud,
	}
	if len(storefrontContexts) > 0 {
		handler.storefrontContext = storefrontContexts[0]
	}
	return handler
}

func (h *Handler) createPaymentGatewayFromConfiguration(config *pgateway.Config) (pgateway.PaymentGateway, error) {
	if h != nil && h.gatewayFactory != nil {
		return h.gatewayFactory(config)
	}
	return pgateway.NewPaymentGateway(config)
}

func (h *Handler) ConfigurePublicBaseURL(baseURL string) {
	if h == nil {
		return
	}
	h.publicBaseURL = normalizePaymentBaseURL(baseURL)
}

func (h *Handler) ConfigureCardBINLimiter(limiter *cardtesting.Service) {
	if h == nil {
		return
	}
	h.cardBINLimiter = limiter
}

// GetStripeExpressCheckoutConfiguration returns only the publishable Stripe
// key required to render Apple Pay and Google Pay buttons. Secret gateway
// credentials never leave the backend.
func (h *Handler) GetStripeExpressCheckoutConfiguration(c *gin.Context) {
	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayStripe)
	if err != nil {
		apierror.RespondError(c, http.StatusServiceUnavailable, "stripe_gateway_config_unavailable", "Stripe configuration is temporarily unavailable")
		return
	}
	if config.PublishableKey == "" {
		config.PublishableKey = pgateway.LoadConfigFromEnv(pgateway.GatewayStripe).PublishableKey
	}
	if config.PublishableKey == "" {
		apierror.RespondError(c, http.StatusServiceUnavailable, "stripe_publishable_key_missing", "Stripe publishable key is not configured")
		return
	}
	if available, reason := h.checkPaymentGatewayConfigurationAvailability(pgateway.GatewayStripe); !available {
		apierror.RespondError(c, http.StatusServiceUnavailable, reason, "Stripe Express Checkout is temporarily unavailable")
		return
	}
	if available, reason := h.gatewayCircuitBreakerAvailability(c, pgateway.GatewayStripe); !available {
		apierror.RespondError(c, http.StatusServiceUnavailable, reason, "Stripe Express Checkout is temporarily unavailable")
		return
	}

	response.Success(c, gin.H{
		"publishable_key": config.PublishableKey,
	})
}

func (h *Handler) authorizeOrderPaymentRead(c *gin.Context, orderID uint) bool {
	if roleValue, exists := c.Get("role"); exists {
		if role, ok := roleValue.(string); ok && auth.NormalizeRole(role).HasPermission(auth.PermOrderView) {
			return true
		}
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return false
	}

	if _, err := h.orderService.GetOrder(orderID, userIDValue.(uint)); err != nil {
		apierror.RespondForbidden(c)
		return false
	}

	return true
}

// CreateStripePaymentIntent creates a server-priced PaymentIntent for an
// existing unpaid order. The client receives only Stripe's publishable key and
// client secret; payment completion still comes from the verified webhook.
func (h *Handler) CreateStripePaymentIntent(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var req struct {
		OrderNumber  string `json:"order_number" binding:"required"`
		CaptchaToken string `json:"captcha_token"`
		CardBIN      string `json:"card_bin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	orderRecord, err := h.orderService.GetOrderByNumber(strings.TrimSpace(req.OrderNumber), userIDValue.(uint))
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return
	}
	if orderRecord.PaymentStatus == "paid" {
		apierror.RespondBadRequest(c, "Order is already paid")
		return
	}
	if orderRecord.Status == "cancelled" || orderRecord.Status == "refunded" || orderRecord.Status == "payment_expired" {
		apierror.RespondBadRequest(c, "Order is not payable")
		return
	}
	if !ensureOrderHasPayableAmount(c, orderRecord) {
		return
	}

	if !h.authorizePaymentStart(c, paymentStartProtectionInput{
		Provider: string(pgateway.GatewayStripe),
		Order:    orderRecord,
	}) {
		return
	}

	riskIdentity := h.stripePaymentRiskIdentity(c, userIDValue.(uint))
	if !h.authorizeStripePaymentRisk(c, riskIdentity, req.CaptchaToken, orderRecord) {
		return
	}

	cardBIN, err := pgateway.NormalizeCardBIN(req.CardBIN)
	if err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	if h.cardBINLimiter != nil && cardBIN != "" {
		decision, err := h.cardBINLimiter.Check(c.Request.Context(), cardBIN)
		if err != nil {
			apierror.RespondError(c, http.StatusServiceUnavailable, "payment_bin_risk_unavailable", "Payment card risk service is temporarily unavailable")
			return
		}
		if decision.Blocked {
			apierror.RespondErrorWithDetails(c, http.StatusTooManyRequests, "payment_bin_temporarily_blocked", "Payment could not be started", gin.H{
				"retry_after_seconds": int64(decision.RetryAfter / time.Second),
			})
			return
		}
	}

	config, err := h.loadPaymentGatewayConfiguration(pgateway.GatewayStripe)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if config.PublishableKey == "" {
		config.PublishableKey = pgateway.LoadConfigFromEnv(pgateway.GatewayStripe).PublishableKey
	}
	if config.ThreeDSecure == "" {
		config.ThreeDSecure = "automatic"
	}
	if config.PublishableKey == "" {
		apierror.RespondError(c, 503, "stripe_publishable_key_missing", "Stripe publishable key is not configured")
		return
	}

	orderCurrency, err := strictOrderCurrency(orderRecord)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	config.PaymentMethodTypes, err = h.resolveStripePaymentMethodTypes(orderRecord.ShippingAddress.Country, orderCurrency, orderRecord.TotalAmount, config.PaymentMethodTypes)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if !ensureGatewayCurrency(c, pgateway.GatewayStripe, orderCurrency) {
		return
	}
	if !h.allowPaymentGatewayAttemptOrRespondWithFallbackRecommendation(c, pgateway.GatewayStripe) {
		return
	}
	defer h.releasePaymentGatewayAttemptIfUnrecorded(c, pgateway.GatewayStripe)
	gateway, err := h.createPaymentGatewayFromConfiguration(config)
	if err != nil {
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayStripe,
			http.StatusInternalServerError,
			"stripe_gateway_initialization_failed",
			err,
		)
		return
	}
	threeDSDecision := h.decideStripeThreeDS(c, orderRecord, userIDValue.(uint), config.ThreeDSecure)
	metadata := map[string]string{
		"order_number": orderRecord.OrderNumber,
	}
	for key, value := range threeDSDecision.Metadata() {
		metadata[key] = value
	}

	attempt, ok := h.ensurePaymentAttempt(
		c,
		pgateway.GatewayStripe,
		"stripe",
		orderRecord,
		orderRecord.TotalAmount,
		orderCurrency,
	)
	if !ok {
		return
	}
	paymentResponse, err := gateway.CreatePayment(c.Request.Context(), &pgateway.PaymentRequest{
		Amount:         orderRecord.TotalAmount,
		Currency:       orderCurrency,
		OrderID:        orderRecord.OrderNumber,
		Description:    fmt.Sprintf("Order %s", orderRecord.OrderNumber),
		IdempotencyKey: attempt.ProviderRequestKey,
		ThreeDSecure:   threeDSDecision.Mode,
		CardBIN:        cardBIN,
		Customer: &pgateway.Customer{
			Name:  orderRecord.ShippingAddress.FirstName + " " + orderRecord.ShippingAddress.LastName,
			Email: orderRecord.ShippingAddress.Email,
			Phone: orderRecord.ShippingAddress.Phone,
		},
		Metadata: metadata,
	})
	if err != nil {
		_ = h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
			Provider:           string(pgateway.GatewayStripe),
			OrderNumber:        orderRecord.OrderNumber,
			TransactionID:      attempt.TransactionID,
			AttemptKey:         attempt.AttemptKey,
			ProviderRequestKey: attempt.ProviderRequestKey,
			PaymentMethod:      "stripe",
			Status:             "failed",
			Amount:             orderRecord.TotalAmount,
			Currency:           orderCurrency,
			ErrorMessage:       err.Error(),
		})
		h.respondToPaymentGatewayOperationFailure(
			c,
			pgateway.GatewayStripe,
			http.StatusBadGateway,
			"stripe_payment_intent_failed",
			err,
		)
		if h.antiFraud != nil {
			_ = h.antiFraud.RecordAttemptFailure(c.Request.Context(), riskIdentity)
		}
		return
	}
	h.recordSuccessfulPaymentGatewayAPIResponse(c, pgateway.GatewayStripe)
	if h.antiFraud != nil {
		if err := h.antiFraud.BindPaymentIntent(c.Request.Context(), paymentResponse.TransactionID, riskIdentity); err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
	}
	if h.cardBINLimiter != nil {
		if err := h.cardBINLimiter.BindPaymentIntent(c.Request.Context(), paymentResponse.TransactionID, cardBIN); err != nil {
			apierror.RespondError(c, http.StatusServiceUnavailable, "payment_bin_risk_unavailable", "Payment card risk service is temporarily unavailable")
			return
		}
	}
	if err := h.paymentService.RecordGatewayPaymentAttempt(service.GatewayPaymentAttemptInput{
		Provider:           string(pgateway.GatewayStripe),
		OrderNumber:        orderRecord.OrderNumber,
		TransactionID:      paymentResponse.TransactionID,
		AttemptKey:         attempt.AttemptKey,
		ProviderRequestKey: attempt.ProviderRequestKey,
		PaymentMethod:      "stripe",
		Status:             "pending",
		Amount:             paymentResponse.Amount,
		Currency:           paymentResponse.Currency,
	}); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if h.riskMonitoring != nil && h.riskMonitoring.Enabled() {
		orderID := orderRecord.ID
		if err := h.riskMonitoring.RecordCheckoutDecision(service.PaymentRiskCheckoutDecisionInput{
			Provider:           string(pgateway.GatewayStripe),
			OrderID:            &orderID,
			ProviderPaymentID:  paymentResponse.TransactionID,
			Mode:               threeDSDecision.Mode,
			Strategy:           threeDSDecision.Strategy,
			ExemptionCandidate: threeDSDecision.ExemptionCandidate,
			RiskLevel:          threeDSDecision.RiskLevel,
			RiskScore:          threeDSDecision.RiskScore,
			PortfolioRiskLevel: threeDSDecision.PortfolioRiskLevel,
			Reasons:            threeDSDecision.Reasons,
			Amount:             paymentResponse.Amount,
			Currency:           paymentResponse.Currency,
			OccurredAt:         time.Now().UTC(),
		}); err != nil {
			logger.Warn("record Stripe checkout risk decision failed",
				zap.String("payment_intent_id", paymentResponse.TransactionID),
				zap.Uint("order_id", orderRecord.ID),
				zap.Error(err),
			)
		}
	}

	response.Success(c, paymentResponse)
}

func (h *Handler) stripePaymentRiskIdentity(c *gin.Context, userID uint) antifraud.AttemptIdentity {
	return antifraud.AttemptIdentity{
		Provider:    string(pgateway.GatewayStripe),
		UserID:      fmt.Sprint(userID),
		SessionID:   stripeRequestSessionID(c),
		AnonymousID: stripeRequestAnonymousID(c),
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	}
}

func (h *Handler) authorizeStripePaymentRisk(c *gin.Context, identity antifraud.AttemptIdentity, captchaToken string, orderRecord *orderdomain.Order) bool {
	if h == nil || h.antiFraud == nil {
		return true
	}
	decision, err := h.antiFraud.EvaluateAttempt(c.Request.Context(), antifraud.AttemptInput{
		Identity: identity,
		Signals: antifraud.Signals{
			IPCountry:      stripeRequestCountry(c),
			BillingCountry: orderRecord.BillingAddress.Country,
			UserAgent:      c.Request.UserAgent(),
		},
	})
	if err != nil {
		apierror.RespondError(c, http.StatusServiceUnavailable, "payment_risk_unavailable", "Payment risk service is temporarily unavailable")
		return false
	}
	if decision.Delay > 0 && !waitPaymentRiskDelay(c.Request.Context(), decision.Delay) {
		apierror.RespondError(c, http.StatusRequestTimeout, "payment_risk_check_timeout", "Payment risk check timed out")
		return false
	}
	if decision.Action == antifraud.DecisionActionBlock {
		apierror.RespondError(c, http.StatusForbidden, "payment_blocked", "Payment could not be started")
		return false
	}
	if !decision.ChallengeRequired {
		return true
	}
	if h.antiBot == nil || !h.antiBot.Required() {
		apierror.RespondError(c, http.StatusServiceUnavailable, "payment_challenge_unavailable", "Payment verification challenge is not configured")
		return false
	}
	if strings.TrimSpace(captchaToken) == "" {
		respondPaymentChallengeRequired(c, decision)
		return false
	}
	if err := h.antiBot.VerifyChallenge(c.Request.Context(), captchaToken, c.ClientIP()); err != nil {
		if errors.Is(err, antibot.ErrChallengeRequired) || errors.Is(err, antibot.ErrChallengeInvalid) {
			respondPaymentChallengeRequired(c, decision)
			return false
		}
		apierror.RespondError(c, http.StatusServiceUnavailable, "payment_challenge_unavailable", "Payment verification challenge is temporarily unavailable")
		return false
	}
	return true
}

func waitPaymentRiskDelay(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func respondPaymentChallengeRequired(c *gin.Context, decision antifraud.Decision) {
	apierror.RespondErrorWithDetails(c, http.StatusForbidden, "payment_challenge_required", "Payment verification challenge required", gin.H{
		"challenge_required": true,
		"action":             "payment",
		"risk_action":        decision.Action,
		"reason":             decision.ChallengeReason,
		"failures":           decision.Failures,
		"reasons":            decision.Reasons,
	})
}

func (h *Handler) decideStripeThreeDS(
	c *gin.Context,
	orderRecord *orderdomain.Order,
	userID uint,
	baseMode string,
) service.PaymentThreeDSDecision {
	if h == nil || h.threeDSPolicy == nil || orderRecord == nil {
		return service.PaymentThreeDSDecision{
			Mode:      pgateway.NormalizeThreeDSecureMode(baseMode),
			Strategy:  "configured",
			RiskLevel: "normal",
			Reasons:   []string{"adaptive_3ds_unavailable"},
		}
	}

	return h.threeDSPolicy.Decide(c.Request.Context(), service.PaymentThreeDSDecisionInput{
		Provider:          string(pgateway.GatewayStripe),
		UserID:            userID,
		OrderID:           orderRecord.ID,
		Amount:            orderRecord.TotalAmount,
		Currency:          normalizedOrderCurrency(orderRecord),
		BaseMode:          baseMode,
		IPAddress:         c.ClientIP(),
		DeviceFingerprint: stripeRequestDeviceFingerprint(c),
		IPCountry:         stripeRequestCountry(c),
		UserAgent:         c.Request.UserAgent(),
		SessionID:         stripeRequestSessionID(c),
		BillingCountry:    orderRecord.BillingAddress.Country,
		ShippingCountry:   orderRecord.ShippingAddress.Country,
		PaymentMethod:     orderRecord.PaymentMethod,
	})
}

func normalizedOrderCurrency(orderRecord *orderdomain.Order) string {
	if orderRecord == nil {
		return ""
	}
	return currency.NormalizeCode(orderRecord.Currency)
}

func strictOrderCurrency(orderRecord *orderdomain.Order) (string, error) {
	value := normalizedOrderCurrency(orderRecord)
	if !currency.IsValidCode(value) || !currency.IsCatalogCode(value) {
		return "", fmt.Errorf("order currency is not configured")
	}
	return value, nil
}

func stripeRequestCountry(c *gin.Context) string {
	return paymentRequestCountry(c)
}

func stripeRequestSessionID(c *gin.Context) string {
	for _, cookieName := range []string{"session_id", visitorcookie.CustomerServiceVisitorCookie} {
		if value, err := c.Cookie(cookieName); err == nil && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stripeRequestDeviceFingerprint(c *gin.Context) string {
	return strings.ToLower(strings.TrimSpace(c.GetHeader("X-Device-Fingerprint")))
}

func stripeRequestAnonymousID(c *gin.Context) string {
	for _, header := range []string{"X-Platform-Anonymous-ID", "X-Anonymous-ID", "X-Visitor-ID"} {
		if value := strings.TrimSpace(c.GetHeader(header)); value != "" {
			return value
		}
	}
	if value, ok := visitorcookie.ExistingCustomerServiceVisitorHash(c, nil); ok {
		return value
	}
	return ""
}
