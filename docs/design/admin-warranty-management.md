# Admin Warranty Management

Last updated: 2026-07-25

## Boundary

Warranty management is an admin surface over the Go backend facts:

- `product_registrations`
- `warranty_claims`
- `warranty_service_records`

The Nuxt storefront may submit or display warranty data, but it must not own warranty-processing state. Status changes, processing notes, service records, and future order-line bindings belong in the Go backend and the admin warranty modules.

## Frontend file responsibilities

| File | Responsibility |
| --- | --- |
| `go-backend/web/admin/src/views/Warranty.vue` | Page composition only: header, stats, tabs, and wiring events from the warranty composable to tab components. |
| `go-backend/web/admin/src/composables/warranty/useWarrantyAdmin.js` | Admin data flow: fetching stats/registrations/claims/expiring records, pagination, active tab, selected claim detail, and save actions. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyRegistrationsTab.vue` | Registration list UI and registration status controls. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyClaimsTab.vue` | Claim list UI and event forwarding to the selected-claim detail panel. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyClaimDetailPanel.vue` | Selected claim detail UI: claim facts, order-line binding, processing note editor, and service-record form. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyExpiringTab.vue` | Expiring warranty list UI. |
| `go-backend/web/admin/src/components/admin/warranty/WarrantyBoundaryTab.vue` | Operational boundary notes shown in the admin UI. |
| `go-backend/web/admin/src/lib/warrantyPresentation.js` | Display-only helpers: labels, tones, date formatting, product/user labels, and claim media parsing. |
| `go-backend/web/admin/src/api/registrations.js` | HTTP contract for admin registration and warranty endpoints. |

Do not put table markup, status label maps, detail editors, or API orchestration back into `Warranty.vue`.

## Backend file responsibilities

| File | Responsibility |
| --- | --- |
| `go-backend/internal/domain/registration/product_registration.go` | GORM fact for product registrations. |
| `go-backend/internal/domain/registration/warranty_claim.go` | GORM fact for warranty claims, including selected order-item relation. |
| `go-backend/internal/domain/registration/warranty_service_record.go` | GORM fact for warranty service history. |
| `go-backend/internal/repository/registration_repository.go` | Persistence queries and narrow updates for registration and warranty claim facts. |
| `go-backend/internal/repository/order_repository.go` | Order and order-item lookups used when an admin binds a claim to a concrete purchased item. |
| `go-backend/internal/service/registration_service.go` | Product registration business rules. |
| `go-backend/internal/service/warranty_claim_service.go` | Warranty claim business rules, including order-verified claims, admin processing updates, order-item binding, and service-record validation. |
| `go-backend/internal/api/admin/registration_handler.go` | Admin HTTP handlers for warranty management. |
| `go-backend/internal/api/admin/router.go` | Admin route registration and permission boundaries. |

## Current next-stage scope

Completed in this pass:

- The admin warranty page was split into page/composable/tab/presentation layers.
- Admin claim detail lookup is available at `GET /api/admin/registrations/warranty-claims/:id`.
- Admin processing notes are saved to `warranty_claims.resolution` through `PUT /api/admin/registrations/warranty-claims/:id/resolution`.
- The UI now lets admins select a claim and save a processing note from the claim detail panel.
- Registration status `claimed` is accepted by the backend to match the existing admin UI.
- Order-line binding is stored as `warranty_claims.order_item_id`, with admin routes for listing valid order items and binding/unbinding the selected item.
- Service history is stored in `warranty_service_records`, with admin routes for listing records and adding constrained service records.
- The selected-claim detail panel was split out of the claim list tab so future rich-text/media work does not bloat the table component.

Still separate future work:

- Rich text and media in processing notes need a constrained editor contract before HTML or images are accepted. Plain text is the current safe fact source.
- Product registration still needs a clearer automatic link to order item/product variant/serial-number source if the storefront later collects serial numbers during checkout or delivery.
- Service record attachments are not implemented yet. If added, store them as their own constrained media facts, not inside `warranty_claims.resolution`.
- Public order-number/email warranty claim flows now require a matching signed, expiring, single-use email challenge before claim creation. Production still needs `SMTP_*`, `STOREFRONT_BASE_URL`, and the storefront token flow; the release gate is tracked in `go-backend/docs/SECURITY_FOLLOW_UPS.md`.
