# Payment Package

This package contains payment gateway adapters and shared payment request/response types.

## Implementations

- `stripe.go` - Stripe adapter
- `paypal.go` - PayPal adapter
- `alipay.go` - Alipay adapter
- `wechat.go` - WeChat Pay adapter
- `gateway.go` - shared gateway interfaces and validation

## Boundary Rules

- Order payment state should be changed by verified provider callbacks or controlled service flows.
- Webhook handlers must verify provider signatures before mutating state.
- Payment operations should be idempotent by provider transaction ID and order ID.
- Do not let frontend requests directly mark orders as paid, refunded, or failed.
- Keep provider-specific details inside adapters; services should depend on shared package types.

## Configuration

Use environment-specific credentials and never commit secrets.

```env
STRIPE_SECRET_KEY=...
STRIPE_WEBHOOK_SECRET=...
STRIPE_ENVIRONMENT=sandbox

PAYPAL_API_KEY=...
PAYPAL_SECRET_KEY=...
PAYPAL_WEBHOOK_SECRET=...
PAYPAL_ENVIRONMENT=sandbox

ALIPAY_API_KEY=...
ALIPAY_SECRET_KEY=...
ALIPAY_WEBHOOK_SECRET=...
ALIPAY_ENVIRONMENT=sandbox

WECHAT_API_KEY=...
WECHAT_SECRET_KEY=...
WECHAT_WEBHOOK_SECRET=...
WECHAT_ENVIRONMENT=sandbox
```

## Tests

```powershell
cd go-backend
go test ./internal/pkg/payment
```
