package payment

import (
	"encoding/json"
	"io"
	"tanzanite/internal/pkg/apierror"
	pgateway "tanzanite/internal/pkg/payment" // alias for gateway
	"tanzanite/internal/pkg/response"

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
		signature = c.GetHeader("Stripe-Signature")
	case "paypal":
		signature = c.GetHeader("Paypal-Transmission-Sig")
	case "alipay":
		signature = c.GetHeader("Alipay-Signature")
	default:
		apierror.RespondBadRequest(c, "Unsupported payment provider")
		return
	}

	if signature == "" {
		apierror.RespondUnauthorized(c)
		return
	}

	// 3. 构建 Gateway 并验签
	// 注意：在实际生产中，配置应从 settingService 或环境变量拉取。这里为了安全演示，默认尝试读取对应环境变量。
	config := &pgateway.Config{
		Type:          pgateway.GatewayType(provider),
		WebhookSecret: "test_secret_for_dev", // 实际应为 os.Getenv(fmt.Sprintf("%s_WEBHOOK_SECRET", strings.ToUpper(provider)))
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

	// 4. 解析 Payload 获取订单号 (简化演示: 假设 JSON 内含 order_number 字段)
	var event struct {
		OrderNumber string `json:"order_number"`
		Status      string `json:"status"` // paid, failed
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		apierror.RespondBadRequest(c, "Invalid JSON payload")
		return
	}

	if event.OrderNumber == "" || event.Status != "paid" {
		response.SuccessWithMessage(c, "Ignored or non-paid event", nil)
		return
	}

	// 5. 调用 OrderRepository 更新状态，实现核心状态扭转闭环
	order, err := h.orderRepo.FindByOrderNumber(event.OrderNumber)
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return
	}

	// 将订单支付状态设为 paid
	if err := h.orderRepo.UpdatePaymentStatus(order.ID, "paid"); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	// 自动将物流状态流转为待发货 (processing)
	if err := h.orderRepo.UpdateStatus(order.ID, "processing"); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Webhook processed successfully", nil)
}
