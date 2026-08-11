# BIN 级卡片测试防护

## 当前实现

Stripe 创建 PaymentIntent 接口支持可选的 `card_bin` 字段。该字段只能是发卡行 BIN 的前 6 位或前 8 位数字，不能传完整卡号、有效期或 CVV。

默认策略：

- 1 分钟内同一 BIN 累计 5 次真实 `payment_intent.payment_failed`
- 达到阈值后 Redis 封禁该 BIN 30 分钟
- 成功支付会清理该 BIN 的失败窗口
- Redis 只保存 BIN 哈希和 PaymentIntent 到哈希键的绑定，不保存原始 BIN

请求示例：

```json
{
  "order_number": "ORDER-001",
  "card_bin": "41111111"
}
```

## 数据来源边界

后端不接收完整 PAN 或 CVV。前端只有在已经通过合规支付组件取得 BIN 的情况下，才应提交 `card_bin`。如果支付组件不能安全提供 BIN，则省略该字段，此时不会启用 BIN 维度的拦截，但现有 IP、Session 和用户风控仍然生效。

PayPal 当前使用 hosted checkout，后端不接触卡号，因此本限流器暂不适用于 PayPal 卡支付。

## 配置

配置段为 `payment_bin_rate_limit`，也可以使用环境变量：

- `PAYMENT_BIN_RATE_LIMIT_ENABLED`
- `PAYMENT_BIN_RATE_LIMIT_WINDOW_SECONDS`
- `PAYMENT_BIN_RATE_LIMIT_FAILURE_THRESHOLD`
- `PAYMENT_BIN_RATE_LIMIT_BLOCK_DURATION_SECONDS`
