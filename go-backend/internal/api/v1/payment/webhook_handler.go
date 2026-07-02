package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"tanzanite/internal/pkg/apierror"
	pgateway "tanzanite/internal/pkg/payment" // alias for gateway
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
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
	var gatewayType pgateway.GatewayType

	// 1. 读取原始 Payload
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		apierror.RespondBadRequest(c, "Failed to read request body")
		return
	}

	// 2. 提取网关签名 Header (例如 Stripe-Signature)
	var signature string
	switch provider {
	case "stripe":
		gatewayType = pgateway.GatewayStripe
		signature = c.GetHeader("Stripe-Signature")
	case "paypal":
		gatewayType = pgateway.GatewayPayPal
		signature = c.GetHeader("Paypal-Transmission-Sig")
	case "alipay":
		gatewayType = pgateway.GatewayAlipay
		signature = c.GetHeader("Alipay-Signature")
	default:
		apierror.RespondBadRequest(c, "Unsupported payment provider")
		return
	}

	if signature == "" {
		apierror.RespondUnauthorized(c)
		return
	}

	config := pgateway.LoadConfigFromEnv(gatewayType)
	if config.WebhookSecret == "" {
		apierror.RespondInternalError(c, fmt.Errorf("payment webhook is not configured"))
		return
	}

	gateway, err := pgateway.NewPaymentGateway(config)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	isValid, err := gateway.VerifyWebhook(payload, signature)
	if !isValid || err != nil {
		apierror.RespondUnauthorized(c)
		return
	}

	// 4. 解析 Payload 获取订单号与网关交易信息
	var event struct {
		OrderNumber   string  `json:"order_number"`
		Status        string  `json:"status"` // paid, completed, succeeded, failed
		TransactionID string  `json:"transaction_id"`
		RefundID      string  `json:"refund_id"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		PaymentMethod string  `json:"payment_method"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	switch {
	case isPaidWebhookStatus(event.Status):
		if event.OrderNumber == "" {
			response.SuccessWithMessage(c, "Ignored paid event without order_number", nil)
			return
		}
		if err := h.paymentService.RecordVerifiedGatewayPayment(service.VerifiedGatewayPaymentInput{
			Provider:        string(gatewayType),
			OrderNumber:     event.OrderNumber,
			TransactionID:   event.TransactionID,
			PaymentMethod:   event.PaymentMethod,
			Amount:          event.Amount,
			Currency:        event.Currency,
			GatewayResponse: string(payload),
		}); err != nil {
			if errors.Is(err, service.ErrOrderNotFound) {
				apierror.RespondNotFound(c, "Order")
				return
			}
			apierror.RespondBadRequest(c, err.Error())
			return
		}
	case isRefundWebhookStatus(event.Status):
		if err := h.paymentService.RecordVerifiedGatewayRefund(service.VerifiedGatewayRefundInput{
			Provider:        string(gatewayType),
			OrderNumber:     event.OrderNumber,
			TransactionID:   event.TransactionID,
			RefundID:        event.RefundID,
			Amount:          event.Amount,
			Currency:        event.Currency,
			GatewayResponse: string(payload),
		}); err != nil {
			if errors.Is(err, service.ErrOrderNotFound) {
				apierror.RespondNotFound(c, "Order")
				return
			}
			apierror.RespondBadRequest(c, err.Error())
			return
		}
	default:
		response.SuccessWithMessage(c, "Ignored unsupported payment event", nil)
		return
	}

	response.SuccessWithMessage(c, "Webhook processed successfully", nil)
}

func isPaidWebhookStatus(status string) bool {
	switch strings.ToLower(status) {
	case "paid", "completed", "succeeded":
		return true
	default:
		return false
	}
}

func isRefundWebhookStatus(status string) bool {
	switch strings.ToLower(status) {
	case "refunded", "refund_completed", "refund_succeeded", "refund.succeeded":
		return true
	default:
		return false
	}
}
