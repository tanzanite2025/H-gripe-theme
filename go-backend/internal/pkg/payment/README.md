# 支付网关实现指南

## 概述

本项目支持多种支付网关，包括：
- Stripe (国际信用卡支付)
- PayPal (国际支付)
- 支付宝 (中国大陆)
- 微信支付 (中国大陆)

## 安装依赖

```bash
# Stripe
go get github.com/stripe/stripe-go/v76

# PayPal
go get github.com/plutov/paypal/v4

# 支付宝
go get github.com/smartwalle/alipay/v3

# 微信支付
go get github.com/wechatpay-apiv3/wechatpay-go
```

## 环境变量配置

### Stripe

```env
STRIPE_API_KEY=sk_test_xxxxx
STRIPE_SECRET_KEY=sk_test_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
STRIPE_ENVIRONMENT=sandbox  # 或 production
```

### PayPal

```env
PAYPAL_API_KEY=xxxxx
PAYPAL_SECRET_KEY=xxxxx
PAYPAL_WEBHOOK_SECRET=xxxxx
PAYPAL_ENVIRONMENT=sandbox  # 或 production
```

### 支付宝

```env
ALIPAY_API_KEY=xxxxx        # AppID
ALIPAY_SECRET_KEY=xxxxx     # 应用私钥
ALIPAY_WEBHOOK_SECRET=xxxxx # 支付宝公钥
ALIPAY_ENVIRONMENT=sandbox  # 或 production
```

### 微信支付

```env
WECHAT_API_KEY=xxxxx        # AppID
WECHAT_SECRET_KEY=xxxxx     # 商户密钥
WECHAT_WEBHOOK_SECRET=xxxxx # API证书序列号
WECHAT_ENVIRONMENT=sandbox  # 或 production
```

## 使用示例

### 创建支付

```go
package main

import (
    "context"
    "fmt"
    "github.com/yourusername/tanzanite/go-backend/internal/pkg/payment"
)

func main() {
    // 从环境变量加载配置
    config := payment.LoadConfigFromEnv(payment.GatewayStripe)
    
    // 创建支付网关
    gateway, err := payment.NewPaymentGateway(config)
    if err != nil {
        panic(err)
    }
    
    // 创建支付请求
    req := &payment.PaymentRequest{
        Amount:      99.99,
        Currency:    "USD",
        OrderID:     "ORD-20240101-123456",
        Description: "Tanzanite Ring",
        Customer: &payment.Customer{
            Email: "customer@example.com",
            Name:  "John Doe",
            Phone: "+1234567890",
        },
        ReturnURL: "https://example.com/payment/success",
        CancelURL: "https://example.com/payment/cancel",
        Metadata: map[string]string{
            "order_id": "ORD-20240101-123456",
            "product": "tanzanite-ring",
        },
    }
    
    // 验证请求
    if err := payment.ValidatePaymentRequest(req); err != nil {
        panic(err)
    }
    
    // 创建支付
    ctx := context.Background()
    resp, err := gateway.CreatePayment(ctx, req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Payment created: %s\n", resp.ID)
    fmt.Printf("Payment URL: %s\n", resp.PaymentURL)
    fmt.Printf("Status: %s\n", resp.Status)
}
```

### 捕获支付

```go
func capturePayment(gateway payment.PaymentGateway, paymentID string) {
    ctx := context.Background()
    resp, err := gateway.CapturePayment(ctx, paymentID)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Payment captured: %s\n", resp.ID)
    fmt.Printf("Status: %s\n", resp.Status)
}
```

### 退款

```go
func refundPayment(gateway payment.PaymentGateway, paymentID string, amount float64) {
    ctx := context.Background()
    
    // 验证退款金额（假设原始金额为99.99）
    if err := payment.ValidateRefundAmount(amount, 99.99); err != nil {
        panic(err)
    }
    
    resp, err := gateway.RefundPayment(ctx, paymentID, amount)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Refund created: %s\n", resp.ID)
    fmt.Printf("Status: %s\n", resp.Status)
    fmt.Printf("Amount: %.2f\n", resp.Amount)
}
```

### 查询支付

```go
func getPayment(gateway payment.PaymentGateway, paymentID string) {
    ctx := context.Background()
    resp, err := gateway.GetPayment(ctx, paymentID)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Payment ID: %s\n", resp.ID)
    fmt.Printf("Status: %s\n", resp.Status)
    fmt.Printf("Amount: %.2f %s\n", resp.Amount, resp.Currency)
}
```

### Webhook验证

```go
func handleWebhook(gateway payment.PaymentGateway, payload []byte, signature string) {
    valid, err := gateway.VerifyWebhook(payload, signature)
    if err != nil {
        fmt.Printf("Webhook verification error: %v\n", err)
        return
    }
    
    if !valid {
        fmt.Println("Invalid webhook signature")
        return
    }
    
    fmt.Println("Webhook verified successfully")
    // 处理webhook事件
}
```

## 测试

### 使用Mock网关

```go
func testWithMock() {
    gateway := payment.NewMockPaymentGateway()
    
    req := &payment.PaymentRequest{
        Amount:   99.99,
        Currency: "USD",
        OrderID:  "TEST-001",
        Customer: &payment.Customer{
            Email: "test@example.com",
            Name:  "Test User",
        },
    }
    
    ctx := context.Background()
    resp, _ := gateway.CreatePayment(ctx, req)
    
    fmt.Printf("Mock payment created: %s\n", resp.ID)
}
```

## 支付状态说明

### Stripe
- `pending` - 待处理
- `succeeded` - 成功
- `failed` - 失败
- `canceled` - 已取消

### PayPal
- `created` - 已创建
- `approved` - 已批准
- `completed` - 已完成
- `failed` - 失败

### 支付宝
- `WAIT_BUYER_PAY` - 等待买家付款
- `TRADE_SUCCESS` - 交易成功
- `TRADE_CLOSED` - 交易关闭
- `TRADE_FINISHED` - 交易完成

### 微信支付
- `NOTPAY` - 未支付
- `SUCCESS` - 支付成功
- `REFUND` - 转入退款
- `CLOSED` - 已关闭
- `PAYERROR` - 支付失败

## Webhook URL配置

在各支付平台配置webhook URL：

- Stripe: `https://yourdomain.com/api/v1/webhooks/stripe`
- PayPal: `https://yourdomain.com/api/v1/webhooks/paypal`
- 支付宝: `https://yourdomain.com/api/v1/webhooks/alipay`
- 微信支付: `https://yourdomain.com/api/v1/webhooks/wechat`

## 注意事项

1. **生产环境配置**
   - 使用生产环境的API密钥
   - 设置 `ENVIRONMENT=production`
   - 配置正确的webhook secret

2. **安全建议**
   - 不要在代码中硬编码API密钥
   - 使用环境变量或密钥管理服务
   - webhook验证是必须的

3. **金额处理**
   - Stripe使用最小货币单位（美分）
   - 其他网关可能有不同的金额格式
   - 注意浮点数精度问题

4. **货币代码**
   - 使用ISO 4217标准货币代码
   - 例如：USD, EUR, CNY, GBP

5. **测试卡号**
   - Stripe测试: 4242 4242 4242 4242
   - PayPal: 使用sandbox账户
   - 支付宝/微信: 使用沙箱环境

## 故障排查

### 常见问题

1. **配置错误**
   ```
   Error: invalid config: API key is required
   ```
   解决：检查环境变量是否正确设置

2. **金额验证失败**
   ```
   Error: amount must be greater than 0
   ```
   解决：确保金额大于0且格式正确

3. **Webhook验证失败**
   ```
   Error: signature verification failed
   ```
   解决：检查webhook secret配置和签名计算方法

## 参考文档

- [Stripe API文档](https://stripe.com/docs/api)
- [PayPal API文档](https://developer.paypal.com/api/rest/)
- [支付宝开放平台](https://opendocs.alipay.com/open/270/105899)
- [微信支付开发文档](https://pay.weixin.qq.com/wiki/doc/apiv3/index.shtml)
