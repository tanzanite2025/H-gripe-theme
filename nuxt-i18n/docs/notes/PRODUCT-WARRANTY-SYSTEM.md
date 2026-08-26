# Product Warranty Current Source

Last updated: 2026-08-23

## Operating Model

The warranty source of truth is the Go backend and is based on shipped orders. There is no product registration, barcode, serial-number, or product-code lookup workflow.

- `orders` remains the source of shipment status, customer identity, tracking data, and purchased-item snapshots.
- `shipment_records` is an optional after-sales attachment keyed by `order_id`. It stores operator notes, images, optional product identifiers, and warranty-window overrides only when someone explicitly adds evidence.
- `warranty_claims` stores customer requests keyed by order number and can optionally bind to `order_item_id`.
- `warranty_service_records` stores repair, inspection, replacement, refund, or shipping service history for a claim.

## Customer Flow

- `nuxt-i18n/app/components/WarrantyCheckPanel.vue` renders the query and result states.
- `nuxt-i18n/app/composables/useWarrantyCheck.ts` calls the authenticated endpoint `/warranty/orders/:order_number`.
- The customer enters an order number, sees shipped items, ship date, warranty period, expiry date, remaining time, and service records.
- A missing shipped order shows the not-found state and contact-support action.

## Claim Flow

- `POST /warranty/verify-order` verifies the order number and email and sends a one-time email challenge.
- `GET /warranty/verify/:token` validates the challenge token.
- `POST /warranty/claim` submits the order-number claim and attachments.
- Admin claim processing lives under `/api/admin/warranty/claims/...`.

The customer-facing key is always the order number. Optional product identifiers, if staff record them, are evidence only and are never used for lookup or registration.

## Admin Flow

The admin page is `go-backend/web/admin/src/views/Warranty.vue` with only these tabs:

- `已发货`: reads shipped orders from `orders` and edits optional `shipment_records` evidence.
- `保修申请`: processes order-number claims and service records.
- `数据边界`: documents ownership and source-of-truth rules.

The warranty page does not create shipment rows during shipping, does not participate in shipping idempotency, and does not maintain a parallel order state.

## Removal Rule

Migrations `212` and `213` permanently remove the retired product-registration table, warranty links, and deployed FAQ wording. Old migration files and archived documents may mention the historical design, but active code, routes, types, UI, and documentation must not depend on it or provide compatibility aliases.
