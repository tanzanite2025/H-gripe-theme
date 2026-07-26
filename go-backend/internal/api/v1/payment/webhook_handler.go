package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"tanzanite/internal/pkg/apierror"
	pgateway "tanzanite/internal/pkg/payment" // alias for gateway
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
)

// ============ Webhook 相关接口 ============

// HandleWebhook 处理外部支付服务的回调通知
// @Summary 处理外部支付回调
// @Tags Payment
// @Accept json
// @Produce json
// @Param provider path string true "支付渠道 (如: stripe, alipay)"
// @Router /api/v1/payment/webhook/{provider} [post]
func (h *Handler) HandleWebhook(c *gin.Context) {
	provider := c.Param("provider")

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		apierror.RespondBadRequest(c, "Failed to read request body")
		return
	}

	switch provider {
	case "stripe":
		h.handleStripeWebhook(c, payload)
	case "paypal":
		apierror.RespondError(c, http.StatusNotImplemented, "payment_webhook_not_ready", "PayPal webhook is disabled until official signature verification is implemented")
	case "alipay":
		apierror.RespondError(c, http.StatusNotImplemented, "payment_webhook_not_ready", "Alipay webhook is disabled until official signature verification is implemented")
	case "wechat":
		apierror.RespondError(c, http.StatusNotImplemented, "payment_webhook_not_ready", "WeChat Pay webhook is disabled until API v3 signature verification and resource decryption are implemented")
	default:
		apierror.RespondBadRequest(c, "Unsupported payment provider")
	}
}

func (h *Handler) handleStripeWebhook(c *gin.Context, payload []byte) {
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		apierror.RespondUnauthorized(c)
		return
	}

	config, err := h.loadGatewayConfig(pgateway.GatewayStripe)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if config.WebhookSecret == "" {
		apierror.RespondInternalError(c, fmt.Errorf("payment webhook is not configured"))
		return
	}

	event, err := pgateway.ParseStripeWebhookEvent(payload, signature, config.WebhookSecret)
	if err != nil {
		apierror.RespondUnauthorized(c)
		return
	}

	switch string(event.Type) {
	case "payment_intent.succeeded":
		h.handleStripePaymentIntentSucceeded(c, event, payload)
	case "payment_intent.payment_failed":
		h.handleStripePaymentIntentFailed(c, event)
	default:
		response.SuccessWithMessage(c, "Ignored unsupported Stripe event", gin.H{
			"event_id":   event.ID,
			"event_type": string(event.Type),
		})
	}
}

func (h *Handler) handleStripePaymentIntentFailed(c *gin.Context, event stripe.Event) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	orderNumber := stripeOrderNumber(intent.Metadata)
	if orderNumber != "" {
		_ = h.paymentService.RecordGatewayPaymentFailure(
			c.Request.Context(),
			string(pgateway.GatewayStripe),
			orderNumber,
		)
	}

	response.SuccessWithMessage(c, "Stripe payment failure recorded", gin.H{
		"event_id":          event.ID,
		"event_type":        string(event.Type),
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	})
}

func (h *Handler) handleStripePaymentIntentSucceeded(c *gin.Context, event stripe.Event, payload []byte) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	orderNumber := stripeOrderNumber(intent.Metadata)
	if orderNumber == "" {
		response.SuccessWithMessage(c, "Ignored Stripe payment_intent.succeeded without order metadata", gin.H{
			"event_id":          event.ID,
			"payment_intent_id": intent.ID,
		})
		return
	}

	amount := intent.AmountReceived
	if amount <= 0 {
		amount = intent.Amount
	}
	if amount <= 0 {
		response.SuccessWithMessage(c, "Ignored Stripe payment_intent.succeeded without amount", gin.H{
			"event_id":          event.ID,
			"payment_intent_id": intent.ID,
		})
		return
	}

	if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
		Provider:        string(pgateway.GatewayStripe),
		OrderNumber:     orderNumber,
		TransactionID:   intent.ID,
		PaymentMethod:   "stripe",
		Amount:          float64(amount) / 100,
		Currency:        string(intent.Currency),
		GatewayResponse: string(payload),
	}); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			apierror.RespondNotFound(c, "Order")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Stripe webhook processed successfully", gin.H{
		"event_id":          event.ID,
		"event_type":        string(event.Type),
		"payment_intent_id": intent.ID,
		"order_number":      orderNumber,
	})
}

func stripeOrderNumber(metadata map[string]string) string {
	for _, key := range []string{"order_number", "order_id", "order"} {
		if value := metadata[key]; value != "" {
			return value
		}
	}
	return ""
}

func (h *Handler) loadGatewayConfig(gatewayType pgateway.GatewayType) (*pgateway.Config, error) {
	if h.settingsService != nil {
		st, err := h.settingsService.GetSetting(pgateway.SecureGatewaySettingKey(gatewayType), "global")
		if err == nil {
			if !pgateway.PaymentConfigMasterKeyConfigured() {
				return nil, fmt.Errorf("%s is required to read encrypted payment config", pgateway.PaymentConfigMasterKeyEnv)
			}
			secureConfig, err := pgateway.DecryptSecureGatewayConfig(st.Value, pgateway.PaymentConfigMasterKey())
			if err != nil {
				return nil, err
			}
			if secureConfig.Provider != gatewayType {
				return nil, fmt.Errorf("payment config provider mismatch")
			}
			return pgateway.GatewayConfigFromSecureConfig(secureConfig), nil
		}
	}
	return pgateway.LoadConfigFromEnv(gatewayType), nil
}
