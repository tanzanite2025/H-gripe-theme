# Admin Warranty Management

Last updated: 2026-08-23

## Boundary

Warranty management is an admin surface over order-backed facts in the Go backend:

- `orders`: shipment status, customer data, tracking data, and purchased items.
- `shipment_records`: optional after-sales evidence attached to an already shipped order.
- `warranty_claims`: customer warranty requests identified by order number.
- `warranty_service_records`: operational service history for a claim.

There is no product registration, barcode, serial-number, or registration-status workflow. An order may have no `shipment_records` row until an operator adds a note, image, optional product identifier, or warranty override. This table is not part of the shipping write path and does not control shipping idempotency.

The Nuxt storefront may query warranty status by authenticated order number and submit an order-number/email claim. It must not own warranty-processing state.

## Frontend Responsibilities

| File | Responsibility |
| --- | --- |
| `go-backend/web/admin/src/views/Warranty.vue` | Page composition: header, stats, tabs, and event wiring. |
| `go-backend/web/admin/src/composables/warranty/useWarrantyAdmin.ts` | Fetching shipped orders and claims, pagination, selection, and save actions. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyShipmentsTab.vue` | Lists shipped orders and edits optional after-sales evidence. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyClaimsTab.vue` | Lists claims and forwards events to claim details. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyClaimDetailPanel.vue` | Claim facts, order-line binding, processing notes, and service records. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyBoundaryTab.vue` | Displays source-of-truth and ownership boundaries. |
| `go-backend/web/admin/src/lib/warrantyPresentation.ts` | Display-only labels, tones, date formatting, and order-line/media helpers. |
| `go-backend/web/admin/src/api/warranty.ts` | HTTP contract for `/api/admin/warranty/...`. |

The warranty page must not reintroduce registration tabs, expiring-registration lists, serial-number lookup, or a second shipment workflow.

## Backend Responsibilities

| File | Responsibility |
| --- | --- |
| `go-backend/internal/domain/shipping/shipment_record.go` | Optional order-backed after-sales evidence fact. |
| `go-backend/internal/domain/warranty/warranty_claim.go` | Warranty claim fact with order number and optional order-line relation. |
| `go-backend/internal/domain/warranty/warranty_service_record.go` | Service history fact owned by a claim. |
| `go-backend/internal/repository/shipment_record_repository.go` | Reads shipped orders and left-joins optional evidence; writes only explicit evidence updates. |
| `go-backend/internal/repository/warranty_repository.go` | Claim and service-record persistence. |
| `go-backend/internal/service/shipment_record_service.go` | Validates optional evidence and warranty window updates. |
| `go-backend/internal/service/warranty_claim_service.go` | Order verification, claim creation, order-line binding, and service-record validation. |
| `go-backend/internal/api/admin/warranty_handler.go` | Admin claim and service-record handlers. |
| `go-backend/internal/api/admin/shipment_record_handler.go` | Admin shipped-order evidence handlers. |
| `go-backend/internal/api/admin/router.go` | `/admin/warranty/...` route and permission boundary. |

## Public Flow

- `GET /api/v1/warranty/orders/:order_number` reads the authenticated user's shipped order and computes the warranty window from the order plus optional `shipment_records` evidence.
- `POST /api/v1/warranty/verify-order` verifies order number and email before sending the claim email challenge.
- `GET /api/v1/warranty/verify/:token` validates the signed, expiring challenge.
- `POST /api/v1/warranty/claim` creates a claim bound to the order number; a later admin action may bind an exact `order_item_id`.

The customer needs the order number to see remaining warranty time. Product codes and images are evidence for staff, not customer-facing warranty keys.

## Data Removal

Migrations `212` and `213` permanently remove the retired product-registration schema, its warranty-table links, and its deployed FAQ wording. Their down migrations are intentionally irreversible so a rollback cannot recreate a second warranty domain.
