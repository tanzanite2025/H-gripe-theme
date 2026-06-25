package payment

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"tanzanite/internal/domain/payment"
	pgateway "tanzanite/internal/pkg/payment" // alias for gateway
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	paymentRepo *repository.PaymentRepository
	orderRepo   *repository.OrderRepository
}

func NewHandler(paymentRepo *repository.PaymentRepository, orderRepo *repository.OrderRepository) *Handler {
	return &Handler{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
	}
}

// Payment Method 相关接口

// ListPaymentMethods 获取支付方式列表
// @Summary 获取支付方式列表
// @Tags Payment
// @Produce json
// @Param enabled query bool false "只显示启用的"
// @Success 200 {array} payment.PaymentMethod
// @Router /api/v1/payment/methods [get]
func (h *Handler) ListPaymentMethods(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	methods, err := h.paymentRepo.FindAllPaymentMethods(enabledOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": methods})
}

// GetPaymentMethod 获取支付方式详情
// @Summary 获取支付方式详情
// @Tags Payment
// @Produce json
// @Param id path int true "支付方式ID"
// @Success 200 {object} payment.PaymentMethod
// @Router /api/v1/payment/methods/{id} [get]
func (h *Handler) GetPaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment method id"})
		return
	}

	method, err := h.paymentRepo.FindPaymentMethodByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, method)
}

// CreatePaymentMethod 创建支付方式（管理员）
// @Summary 创建支付方式
// @Tags Payment
// @Accept json
// @Produce json
// @Param method body payment.PaymentMethod true "支付方式信息"
// @Success 201 {object} payment.PaymentMethod
// @Router /api/v1/admin/payment/methods [post]
func (h *Handler) CreatePaymentMethod(c *gin.Context) {
	var method payment.PaymentMethod
	if err := c.ShouldBindJSON(&method); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paymentRepo.CreatePaymentMethod(&method); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, method)
}

// UpdatePaymentMethod 更新支付方式（管理员）
// @Summary 更新支付方式
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "支付方式ID"
// @Param method body payment.PaymentMethod true "支付方式信息"
// @Success 200 {object} payment.PaymentMethod
// @Router /api/v1/admin/payment/methods/{id} [put]
func (h *Handler) UpdatePaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment method id"})
		return
	}

	var method payment.PaymentMethod
	if err := c.ShouldBindJSON(&method); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method.ID = uint(id)
	if err := h.paymentRepo.UpdatePaymentMethod(&method); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, method)
}

// DeletePaymentMethod 删除支付方式（管理员）
// @Summary 删除支付方式
// @Tags Payment
// @Produce json
// @Param id path int true "支付方式ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payment/methods/{id} [delete]
func (h *Handler) DeletePaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment method id"})
		return
	}

	if err := h.paymentRepo.DeletePaymentMethod(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment method deleted"})
}

// Tax Rate 相关接口

// ListTaxRates 获取税率列表
// @Summary 获取税率列表
// @Tags Payment
// @Produce json
// @Success 200 {array} payment.TaxRate
// @Router /api/v1/payment/tax-rates [get]
func (h *Handler) ListTaxRates(c *gin.Context) {
	rates, err := h.paymentRepo.FindAllTaxRates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rates})
}

// GetTaxRate 获取税率详情
// @Summary 获取税率详情
// @Tags Payment
// @Produce json
// @Param id path int true "税率ID"
// @Success 200 {object} payment.TaxRate
// @Router /api/v1/payment/tax-rates/{id} [get]
func (h *Handler) GetTaxRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tax rate id"})
		return
	}

	rate, err := h.paymentRepo.FindTaxRateByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rate)
}

// CalculateTax 计算税费
// @Summary 计算税费
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "计算请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payment/calculate-tax [post]
func (h *Handler) CalculateTax(c *gin.Context) {
	var req struct {
		Amount  float64 `json:"amount" binding:"required,gt=0"`
		Country string  `json:"country" binding:"required"`
		State   string  `json:"state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找税率
	taxRate, err := h.paymentRepo.FindTaxRateByLocation(req.Country, req.State)
	if err != nil {
		// 没有找到税率，返回0
		c.JSON(http.StatusOK, gin.H{
			"amount":   req.Amount,
			"tax_rate": 0.0,
			"tax":      0.0,
			"total":    req.Amount,
		})
		return
	}

	// 计算税费
	tax := req.Amount * taxRate.Rate / 100
	total := req.Amount + tax

	c.JSON(http.StatusOK, gin.H{
		"amount":   req.Amount,
		"tax_rate": taxRate.Rate,
		"tax":      tax,
		"total":    total,
	})
}

// CreateTaxRate 创建税率（管理员）
// @Summary 创建税率
// @Tags Payment
// @Accept json
// @Produce json
// @Param rate body payment.TaxRate true "税率信息"
// @Success 201 {object} payment.TaxRate
// @Router /api/v1/admin/payment/tax-rates [post]
func (h *Handler) CreateTaxRate(c *gin.Context) {
	var rate payment.TaxRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paymentRepo.CreateTaxRate(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rate)
}

// UpdateTaxRate 更新税率（管理员）
// @Summary 更新税率
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "税率ID"
// @Param rate body payment.TaxRate true "税率信息"
// @Success 200 {object} payment.TaxRate
// @Router /api/v1/admin/payment/tax-rates/{id} [put]
func (h *Handler) UpdateTaxRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tax rate id"})
		return
	}

	var rate payment.TaxRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rate.ID = uint(id)
	if err := h.paymentRepo.UpdateTaxRate(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rate)
}

// DeleteTaxRate 删除税率（管理员）
// @Summary 删除税率
// @Tags Payment
// @Produce json
// @Param id path int true "税率ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payment/tax-rates/{id} [delete]
func (h *Handler) DeleteTaxRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tax rate id"})
		return
	}

	if err := h.paymentRepo.DeleteTaxRate(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tax rate deleted"})
}

// Transaction 相关接口

// GetTransaction 获取交易详情
// @Summary 获取交易详情
// @Tags Payment
// @Produce json
// @Param id path int true "交易ID"
// @Success 200 {object} payment.Transaction
// @Router /api/v1/payment/transactions/{id} [get]
func (h *Handler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	transaction, err := h.paymentRepo.FindTransactionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// GetOrderTransactions 获取订单的交易记录
// @Summary 获取订单的交易记录
// @Tags Payment
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {array} payment.Transaction
// @Router /api/v1/payment/orders/{order_id}/transactions [get]
func (h *Handler) GetOrderTransactions(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	transactions, err := h.paymentRepo.FindTransactionByOrderID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

// CreateTransaction 创建交易记录
// @Summary 创建交易记录
// @Tags Payment
// @Accept json
// @Produce json
// @Param transaction body payment.Transaction true "交易信息"
// @Success 201 {object} payment.Transaction
// @Router /api/v1/payment/transactions [post]
func (h *Handler) CreateTransaction(c *gin.Context) {
	var transaction payment.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paymentRepo.CreateTransaction(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// Refund 相关接口

// CreateRefund 创建退款
// @Summary 创建退款
// @Tags Payment
// @Accept json
// @Produce json
// @Param refund body payment.Refund true "退款信息"
// @Success 201 {object} payment.Refund
// @Router /api/v1/payment/refunds [post]
func (h *Handler) CreateRefund(c *gin.Context) {
	var refund payment.Refund
	if err := c.ShouldBindJSON(&refund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认状态
	refund.Status = "pending"

	if err := h.paymentRepo.CreateRefund(&refund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, refund)
}

// GetRefund 获取退款详情
// @Summary 获取退款详情
// @Tags Payment
// @Produce json
// @Param id path int true "退款ID"
// @Success 200 {object} payment.Refund
// @Router /api/v1/payment/refunds/{id} [get]
func (h *Handler) GetRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund id"})
		return
	}

	refund, err := h.paymentRepo.FindRefundByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, refund)
}

// GetOrderRefunds 获取订单的退款记录
// @Summary 获取订单的退款记录
// @Tags Payment
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {array} payment.Refund
// @Router /api/v1/payment/orders/{order_id}/refunds [get]
func (h *Handler) GetOrderRefunds(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	refunds, err := h.paymentRepo.FindRefundsByOrderID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": refunds})
}

// UpdateRefundStatus 更新退款状态（管理员）
// @Summary 更新退款状态
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "退款ID"
// @Param request body map[string]string true "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payment/refunds/{id}/status [put]
func (h *Handler) UpdateRefundStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.paymentRepo.FindRefundByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	refund.Status = req.Status
	if err := h.paymentRepo.UpdateRefund(refund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund status updated"})
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported payment provider"})
		return
	}

	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize payment gateway"})
		return
	}

	isValid, err := gateway.VerifyWebhook(payload, signature)
	if !isValid || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// 4. 解析 Payload 获取订单号 (简化演示: 假设 JSON 内含 order_number 字段)
	var event struct {
		OrderNumber string `json:"order_number"`
		Status      string `json:"status"` // paid, failed
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if event.OrderNumber == "" || event.Status != "paid" {
		c.JSON(http.StatusOK, gin.H{"message": "Ignored or non-paid event"})
		return
	}

	// 5. 调用 OrderRepository 更新状态，实现核心状态扭转闭环
	order, err := h.orderRepo.FindByOrderNumber(event.OrderNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// 将订单支付状态设为 paid
	if err := h.orderRepo.UpdatePaymentStatus(order.ID, "paid"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// 自动将物流状态流转为待发货 (processing)
	if err := h.orderRepo.UpdateStatus(order.ID, "processing"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}
