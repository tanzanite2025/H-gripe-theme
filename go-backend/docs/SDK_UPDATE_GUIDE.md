# SDK更新指南

## 概述

本文档描述了支付网关SDK的更新需求和步骤。由于第三方SDK的API变更，部分支付功能暂时禁用，需要更新后重新启用。

## 状态概览

| SDK | 状态 | 当前版本 | 需要版本 | 优先级 |
|-----|------|----------|----------|--------|
| Stripe | ✅ 正常 | v76.25.0 | - | - |
| Alipay | ⚠️ 部分功能 | v3.2.29 | - | 中 |
| PayPal | ❌ 需要更新 | v4.17.0 | 最新 | 高 |
| WeChat | ❌ 需要更新 | v0.2.21 | 最新 | 高 |

## 1. PayPal SDK更新

### 当前问题

PayPal SDK v4的API发生了重大变更：

1. `PurchaseUnitRequest` 类型不匹配
2. `ApplicationContext` 字段移除
3. `CreateOrder` 方法签名变更

### 受影响的功能

- ❌ 创建支付订单
- ❌ 捕获支付
- ❌ 退款
- ❌ 查询订单
- ❌ Webhook验证

### 更新步骤

#### 第1步：查看官方文档

访问官方文档了解最新API：
- https://github.com/plutov/paypal
- https://developer.paypal.com/docs/api/overview/

#### 第2步：更新依赖

```bash
cd go-backend
go get -u github.com/plutov/paypal/v4
go mod tidy
```

#### 第3步：更新代码

需要更新的文件：
- `internal/pkg/payment/paypal.go`

**CreatePayment方法**:

```go
// 旧版本（已禁用）
func (g *paypalGatewayImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    // TODO: 根据最新SDK文档更新实现
    return nil, fmt.Errorf("PayPal integration requires SDK API update")
}
```

**需要实现的功能**:

1. ✅ 创建订单
2. ✅ 捕获支付
3. ✅ 退款处理
4. ✅ 查询订单状态
5. ✅ Webhook验证

#### 第4步：参考实现

```go
// 参考 - 需要根据实际SDK更新
func (g *paypalGatewayImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    if err := ValidatePaymentRequest(req); err != nil {
        return nil, fmt.Errorf("invalid payment request: %w", err)
    }

    // 查阅最新SDK文档，更新这里的实现
    // 1. 构建订单请求结构
    // 2. 调用CreateOrder API
    // 3. 提取approval URL
    // 4. 返回PaymentResponse

    return &PaymentResponse{
        ID:          orderID,
        Status:      status,
        Amount:      req.Amount,
        Currency:    req.Currency,
        RedirectURL: approvalURL,
        CreatedAt:   time.Now(),
    }, nil
}
```

#### 第5步：测试

```bash
# 运行单元测试
go test ./internal/pkg/payment/... -v

# 运行集成测试
go test ./internal/pkg/payment/... -tags=integration -v
```

#### 第6步：文档更新

更新以下文档：
- `internal/pkg/payment/README.md`
- `docs/PAYMENT_GATEWAY_IMPLEMENTATION.md`

---

## 2. 微信支付SDK更新

### 当前问题

微信支付SDK v3的API变更：

1. `notify.CertificateVisitor` 类型不存在
2. Webhook验证API变更
3. 证书处理方式变更

### 受影响的功能

- ✅ Native扫码支付（正常）
- ✅ 订单查询（正常）
- ✅ 退款（正常）
- ❌ Webhook验证（需要更新）
- ❌ JSAPI支付（部分功能）

### 更新步骤

#### 第1步：查看官方文档

访问官方文档：
- https://github.com/wechatpay-apiv3/wechatpay-go
- https://pay.weixin.qq.com/wiki/doc/apiv3/index.shtml

#### 第2步：更新依赖

```bash
cd go-backend
go get -u github.com/wechatpay-apiv3/wechatpay-go
go mod tidy
```

#### 第3步：更新代码

需要更新的文件：
- `internal/pkg/payment/wechat.go`

**VerifyWebhook方法**:

```go
// 当前实现（简化版本）
func (g *wechatGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
    // 基本验证逻辑
    return verifyHMACSHA256(payload, signature, g.config.APIKey), nil
}

// TODO: 实现完整的SDK验证
// 参考官方文档: https://pay.weixin.qq.com/wiki/doc/apiv3/wechatpay/wechatpay4_1.shtml
```

**需要更新的功能**:

1. ✅ Native支付（已实现）
2. ⚠️ JSAPI支付（需要测试）
3. ⚠️ APP支付（需要测试）
4. ❌ Webhook通知验证（需要更新）
5. ⚠️ 证书处理（需要验证）

#### 第4步：证书配置

微信支付需要证书文件：

```yaml
# config.yaml
payment:
  wechat:
    app_id: "wx1234567890"
    mch_id: "1234567890"
    api_key: "your-api-key"
    cert_path: "/path/to/apiclient_cert.pem"
    key_path: "/path/to/apiclient_key.pem"
    serial_no: "your-cert-serial-number"
```

#### 第5步：参考实现

```go
// 参考 - 完整的Webhook验证
func (g *wechatGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
    // 1. 解析HTTP头获取签名信息
    // 2. 验证时间戳
    // 3. 构建签名串
    // 4. 使用平台证书验证签名
    // 5. 解密回调内容
    
    // 查阅最新SDK文档实现
    return true, nil
}
```

#### 第6步：测试

```bash
# 单元测试
go test ./internal/pkg/payment/... -run TestWechat -v

# 使用微信支付沙箱环境测试
```

---

## 3. 支付宝SDK更新

### 当前状态

支付宝SDK v3基本可用，但有小问题：

1. ✅ 网页支付 - 正常
2. ✅ APP支付 - 正常
3. ✅ 退款 - 正常
4. ⚠️ Webhook验证 - API变更需要更新

### 更新步骤

#### 第1步：更新Webhook验证

```go
// 当前实现（临时禁用）
func VerifyAlipayNotification(client *alipay.Client, values map[string]string) (bool, error) {
    return false, fmt.Errorf("alipay notification verification requires SDK API update")
}

// 需要实现
func VerifyAlipayNotification(client *alipay.Client, values map[string]string) (bool, error) {
    // 查阅SDK v3文档，找到正确的验签方法
    // 参考: https://opendocs.alipay.com/open/270/105902
    
    // 步骤：
    // 1. 提取sign字段
    // 2. 构建待验签字符串
    // 3. 使用支付宝公钥验签
    
    return ok, nil
}
```

#### 第2步：测试

```bash
go test ./internal/pkg/payment/... -run TestAlipay -v
```

---

## 4. 通用更新流程

### 开发环境测试

```bash
# 1. 更新依赖
go get -u github.com/plutov/paypal/v4
go get -u github.com/wechatpay-apiv3/wechatpay-go
go get -u github.com/smartwalle/alipay/v3
go mod tidy

# 2. 运行测试
go test ./internal/pkg/payment/... -v

# 3. 运行集成测试（需要真实凭证）
export PAYPAL_CLIENT_ID="xxx"
export PAYPAL_SECRET="xxx"
go test ./internal/pkg/payment/... -tags=integration -v

# 4. 编译检查
go build ./...
```

### 沙箱环境测试

所有支付网关都提供沙箱环境：

1. **PayPal Sandbox**
   - https://developer.paypal.com/developer/accounts/

2. **微信支付沙箱**
   - https://pay.weixin.qq.com/wiki/doc/api/jsapi.php?chapter=23_1

3. **支付宝沙箱**
   - https://openhome.alipay.com/platform/appDaily.htm

### 配置管理

使用不同环境的配置：

```yaml
# config.sandbox.yaml
payment:
  paypal:
    client_id: ${PAYPAL_SANDBOX_CLIENT_ID}
    secret: ${PAYPAL_SANDBOX_SECRET}
    mode: sandbox
  
  wechat:
    app_id: ${WECHAT_SANDBOX_APP_ID}
    mch_id: ${WECHAT_SANDBOX_MCH_ID}
    
  alipay:
    app_id: ${ALIPAY_SANDBOX_APP_ID}
```

---

## 5. 回归测试

### 测试清单

更新SDK后需要测试的功能：

#### PayPal
- [ ] 创建订单
- [ ] 获取approval URL
- [ ] 捕获支付
- [ ] 退款
- [ ] 查询订单
- [ ] Webhook接收
- [ ] Webhook验证

#### 微信支付
- [ ] Native支付二维码生成
- [ ] JSAPI支付
- [ ] APP支付
- [ ] 订单查询
- [ ] 退款
- [ ] Webhook通知接收
- [ ] Webhook签名验证

#### 支付宝
- [ ] 网页支付
- [ ] APP支付
- [ ] WAP支付
- [ ] 交易查询
- [ ] 退款
- [ ] 异步通知接收
- [ ] 异步通知验签

### 测试脚本

```bash
#!/bin/bash

echo "Running payment gateway tests..."

# PayPal tests
echo "Testing PayPal..."
go test ./internal/pkg/payment/... -run TestPayPal -v

# WeChat tests
echo "Testing WeChat Pay..."
go test ./internal/pkg/payment/... -run TestWechat -v

# Alipay tests
echo "Testing Alipay..."
go test ./internal/pkg/payment/... -run TestAlipay -v

echo "All tests completed!"
```

---

## 6. 文档更新

更新完成后需要更新的文档：

1. **API文档**
   - 更新支付API的请求/响应示例
   - 更新Webhook处理说明

2. **配置文档**
   - 更新支付网关配置示例
   - 添加新的配置项说明

3. **部署文档**
   - 更新环境变量列表
   - 更新Secret配置示例

4. **变更日志**
   - 记录SDK版本变更
   - 记录API变更

---

## 7. 优先级和时间线

### 高优先级（1-2周）

1. **PayPal SDK更新**
   - 估计工作量: 3-5天
   - 影响: 高（核心支付功能）
   
2. **微信支付Webhook验证**
   - 估计工作量: 2-3天
   - 影响: 中（安全相关）

### 中优先级（2-4周）

3. **支付宝Webhook验证**
   - 估计工作量: 1-2天
   - 影响: 中（安全相关）

4. **微信支付JSAPI/APP测试**
   - 估计工作量: 2-3天
   - 影响: 中（扩展功能）

### 低优先级（按需）

5. **性能优化**
6. **监控和日志增强**
7. **错误处理完善**

---

## 8. 风险评估

### 技术风险

1. **API兼容性**
   - 风险: SDK更新可能引入破坏性变更
   - 缓解: 在沙箱环境充分测试

2. **数据迁移**
   - 风险: 现有支付数据格式变化
   - 缓解: 保持数据库Schema稳定

3. **第三方依赖**
   - 风险: SDK依赖其他包的版本冲突
   - 缓解: 使用go mod确保依赖隔离

### 业务风险

1. **服务中断**
   - 风险: 更新期间支付功能不可用
   - 缓解: 使用feature flag渐进式发布

2. **资金安全**
   - 风险: Webhook验证失败导致资金损失
   - 缓解: 双重验证机制

---

## 9. 发布计划

### Phase 1: 准备阶段（Week 1）
- [ ] 研究SDK文档
- [ ] 搭建沙箱测试环境
- [ ] 准备测试数据

### Phase 2: 开发阶段（Week 2-3）
- [ ] 更新PayPal SDK
- [ ] 更新微信支付SDK
- [ ] 更新支付宝SDK
- [ ] 编写单元测试

### Phase 3: 测试阶段（Week 3-4）
- [ ] 沙箱环境测试
- [ ] 集成测试
- [ ] 性能测试
- [ ] 安全测试

### Phase 4: 发布阶段（Week 4）
- [ ] Code Review
- [ ] 文档更新
- [ ] 灰度发布
- [ ] 全量发布

---

## 10. 参考资源

### 官方文档

- **PayPal**: https://developer.paypal.com/docs/api/overview/
- **微信支付**: https://pay.weixin.qq.com/wiki/doc/apiv3/index.shtml
- **支付宝**: https://opendocs.alipay.com/

### SDK仓库

- **PayPal Go SDK**: https://github.com/plutov/paypal
- **微信支付Go SDK**: https://github.com/wechatpay-apiv3/wechatpay-go
- **支付宝Go SDK**: https://github.com/smartwalle/alipay

### 社区资源

- **Go Payment Libraries**: https://github.com/avelino/awesome-go#financial
- **支付集成最佳实践**: 参考各大电商平台的技术博客

---

## 联系方式

如有问题，请联系：

- **技术负责人**: tech-lead@tanzanite.example.com
- **支付团队**: payment-team@tanzanite.example.com
- **紧急支持**: oncall@tanzanite.example.com
