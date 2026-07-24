# Product warranty current source

Last updated: 2026-07-24

This note replaces the old WordPress-plugin warranty design draft. The current implementation is Nuxt storefront + Go backend registration APIs.

## Current user-facing flow

- Nuxt component: `nuxt-i18n/app/components/WarrantyCheckPanel.vue`
- Shared composable: `nuxt-i18n/app/composables/useWarrantyCheck.ts`
- The panel requires the user to be logged in.
- The user enters a product/serial code.
- Nuxt calls the authenticated API path:
  - `/registrations/warranty/:code`
- Successful results show product code, product type/name, ship date, warranty period, warranty end date, status, remaining time, and service records if present.
- Missing records show the current not-found state and contact-support action.

## Current backend flow

Relevant Go files:

- `go-backend/internal/api/v1/registration/serial_number.go`
  - `GetWarrantyStatus()`
  - `VerifySerialNumber()`
  - `warrantyStatusResponse()`
- `go-backend/internal/api/v1/registration/warranty.go`
  - order verification and warranty claim submission
  - upload validation for warranty claim images/videos
- `go-backend/internal/api/v1/router.go`
  - registration and warranty route registration
- `go-backend/internal/api/admin/registration_handler.go`
  - admin listing/status endpoints for product registrations and warranty claims
- `go-backend/web/admin/src/api/registrations.js`
  - admin frontend API wrapper for registration stats, lists, expiring warranties, and claim status updates
- `go-backend/web/admin/src/views/Warranty.vue`
  - admin UI surface for registrations, warranty claims, expiring warranties, and source-boundary notes
- `go-backend/migrations/030_warranty_claim_registration_nullable.up.sql`
  - makes `warranty_claims.registration_id` nullable so order-verified claims do not use `0` as a fake registration fallback

Current warranty status response is built from `ProductRegistration` via `registrationSvc.VerifySerialNumber(code)`.

Current admin backend entry points are Go admin routes, not WordPress:

- `GET /admin/registrations`
- `PUT /admin/registrations/:id/status`
- `GET /admin/registrations/expiring`
- `GET /admin/registrations/stats`
- `GET /admin/registrations/warranty-claims`
- `PUT /admin/registrations/warranty-claims/:id/status`

## What is closed

- Nuxt has a single reusable warranty-check composable.
- The page renders logged-out, loading, error, valid, expired, and no-record states.
- Warranty status now comes from Go backend registration data, not from the old WordPress plugin draft.
- UI text is wired through i18n keys under the `warranty` namespace.
- Admin backend ownership is confirmed: registration and claim management belongs to Go admin registration routes.
- User-facing claim upload UX now mirrors backend validation for allowed image/video formats, image count, per-file size, total image size, and video size.
- Admin frontend now consumes the Go registration routes through a dedicated `保修管理` page.
- Admin UI covers registration list/status changes, warranty claim list/status changes, 30-day expiring warranties, and uploaded image/video evidence links.
- Order-based warranty claims no longer rely on a misleading `registration_id=0`; null is the only valid representation for "not linked yet".

## Still not fully closed

1. Claim-to-registration linkage and service records
   - Current `warrantyStatusResponse()` returns an empty `records` array.
   - Authenticated JSON claims can be tied to a `registration_id`, but the public order-number claim flow currently creates a claim from order number + email and does not bind a concrete `RegistrationID`.
   - Do not populate `records` from order-based claims until each claim can be linked to one registration/serial/product fact source.
   - If repair/extension/replacement records become real, define one backend source and update this response.

2. Product type naming
   - Current response maps product type from linked product SKU/name.
   - If multilingual product type names are required, decide whether they come from backend locale-aware data or Nuxt translation keys.

3. Claim review UX
   - Backend validates warranty claim image/video uploads and Nuxt pre-validates the public submit form before upload.
   - Admin can now view uploaded image/video links and change statuses.
   - Rich processing notes, replacement/repair actions, and a dedicated claim detail drawer still need a backend-owned service-record model.

4. Testing
   - Need end-to-end tests for valid serial, expired serial, missing serial, unauthorized access, and upload validation.
   - Need admin e2e coverage for registration status updates, claim status updates, and nullable order-based claim rows.

## Maintenance rule

Any warranty change must keep a single fact source:

- serial/registration/warranty data: Go backend registration domain;
- user-facing status UI: Nuxt `WarrantyCheckPanel` + `useWarrantyCheck`;
- admin entry point: whichever current Go admin module owns registration data.

Do not use the archived WordPress-plugin draft as current architecture.
