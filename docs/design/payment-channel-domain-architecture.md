# 收款渠道与支付风控域架构

Last updated: 2026-09-09

## Status

Active architecture note. The current code has already split the admin surface
into a collection side and a risk side, but naming still needs to stay aligned
while the remaining pages are moved.

## 1. Current State

The current admin routes show the intended split:

- The expanded sidebar presents top-level domains in a two-column grid.
- `/payment/methods` is the collection method library.
- `/payment/stripe/integration` is the Stripe collection entry.
- `/payment/stripe/installments` is the Stripe installments placeholder.
- `/payment/stripe/3ds` is the Stripe 3DS and risk strategy entry, backed by
  the existing risk component.
- `/payment/stripe/disputes` is the Stripe dispute handling entry, backed by
  the existing risk component.
- `/payment/wechat/integration` is the WeChat Pay collection entry.
- `/payment/wechat/installments` is the WeChat Pay installments placeholder.
- `/payment/alipay/integration` is the Alipay collection entry.
- `/payment/alipay/installments` is the Alipay installments placeholder.
- `/payment/paypal/integration` is the PayPal collection entry.
- `/payment/paypal/installments` is the PayPal installments placeholder.
- `/payment/paypal/disputes` is the PayPal dispute handling entry, backed by
  the existing risk component with PayPal selected by default.
- `/payment/paypal/invoice` is the PayPal seller profile page.
- `/payment-risk/*` is the separate top-level payment risk domain.

`PaymentCollectionMethods.vue` owns the shared collection-method library.
`PaymentIntegrations.vue` is the reusable provider integration page used by
Stripe and PayPal routes. `Settings.vue` remains the system settings shell for
site, email, SEO, social, and other shared admin settings, and is no longer a
payment page.

## 2. Boundary

Each collection provider is its own top-level domain and owns its onboarding and
collection configuration:

- provider credentials and callbacks;
- payout and invoice profile settings;
- channel-specific options such as installments;
- provider-specific enable/disable state.

`支付风控` is a separate top-level domain and owns operational risk work:

- risk overview;
- 3DS strategy;
- manual review;
- refund recommendations;
- manual protection.

Do not mix these two domains just because they both touch payment facts.
Collection is about how money is accepted. Risk is about how money is
controlled, reviewed, and disputed.

## 3. Payment Fact and Duplicate-Charge Boundary

The payment ledger distinguishes two cases that must not be collapsed into one
"duplicate webhook" path:

- The same provider transaction ID is delivered again. This is an idempotent
  webhook retry; it reuses the existing transaction row and does not create a
  second refund.
- A different provider transaction ID is verified after the order is already
  paid. This is a real duplicate charge, not merely a repeated notification.
  The system stores the second transaction with status `duplicate_paid`,
  preserves its amount, currency, provider response, and transaction ID, and
  creates one full-amount local `pending` refund linked to that transaction.

The second case does not create another order-paid event and does not change the
order back to unpaid. Webhook success is returned only after the transaction
and the duplicate-paid refund record have been persisted. That response means
the local record is durable; it does not mean the provider-side refund has
already completed.

The pending refund continues through the existing explicit admin refund
execution workflow. The gateway API call, provider refund ID, execution audit,
retry behavior, and final `completed` state are handled there. A local
pending refund is therefore an automatic accounting and remediation record, not
an automatic provider-side refund.

### Financial order retention boundary

An order, its payment transactions, refunds, disputes, and evidence package are one financial reconciliation and chargeback-defense chain. Once an order has a paid or refunded financial fact, the order record and its linked evidence must remain queryable for reconciliation and dispute response; removing it from an admin list must never mean physically deleting it.

- The canonical admin action is `POST /api/admin/orders/:id/hide-unpaid-terminal`. It is limited to `cancelled + unpaid` and `payment_expired + expired`. Paid, refunded, or state-mismatched orders are retained and the action is rejected.
- The current implementation uses GORM soft delete only for those never-paid terminal orders. Soft delete changes default query visibility; it is not a complete `Archived` order state, archive workflow, recovery contract, or permission model.
- `DELETE /api/admin/orders/:id` is retained only for legacy client compatibility and invokes the same guarded soft-hide behavior. New clients must use the explicit POST route and call the action hide, not delete.

## 4. Target Route Shape

```text
顶级域（左侧两列）
├── 收款方式
├── Stripe
│   ├── 接入配置
│   ├── 分期配置
│   ├── 3DS / 风控策略
│   └── 拒付处理
├── PayPal
│   ├── 接入配置
│   ├── 分期配置
│   ├── 拒付处理
│   └── 发票方资料
├── 微信支付
│   ├── 接入配置
│   └── 分期配置
├── 支付宝
│   ├── 接入配置
│   └── 分期配置
└── 支付风控
    ├── 风控总览
    ├── 3DS 策略
    ├── 人工复核
    ├── 退款建议
    └── 人工保护
```

The long-term rule is channel-first:

- Stripe should own Stripe integration, Stripe installments, and Stripe risk
  facts.
- PayPal should own PayPal integration, PayPal invoice details, and PayPal
  installments.
- WeChat Pay and Alipay should get the same treatment when they grow.

If a provider grows enough, its subtree can later expand into a deeper route
tree. A separate top-level installments domain should only exist if
installments become truly shared across providers.

## 5. Placeholder Strategy

Installments should not be dumped into a mixed settings bucket.

Recommended rollout:

1. Keep `收款方式` as its own top-level collection-method domain.
2. Keep Stripe, PayPal, 微信支付, and 支付宝 as independent top-level domains.
3. Keep provider-level placeholder pages for Stripe and PayPal installments.
4. Keep PayPal invoice details near PayPal-specific collection data.
5. Keep `PaymentRisk` as the separate top-level operational risk domain.
6. Only extract a separate installments domain if a shared policy layer
   actually appears later.

This keeps the future shape visible without pretending a shared abstraction
already exists.

## 6. Naming Rules

- Use business names in visible labels: `收款方式`, `支付风控`, `Stripe`,
  `PayPal`, `微信支付`, `支付宝`.
- Avoid a generic `Settings` label for provider-specific work.
- Treat `Settings.vue` as system-settings scaffolding only.
- Do not keep legacy payment redirects in the active admin surface.

## 7. Practical Implication

The current `payment-risk` grouping is directionally correct. `人工保护` belongs
there, not inside a provider collection domain.

The old `payment` grouping has been retired. The long-term admin model reads as:

- 收款方式;
- Stripe;
- PayPal;
- 微信支付;
- 支付宝;
- 支付风控;
- provider-owned subtrees beneath each channel.
