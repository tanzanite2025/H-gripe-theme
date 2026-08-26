# Timezone Policy

## Decision

Tanzanite does not use an external Time API for normal business logic.

The admin setting historically named `time_api_*` is now a compatibility wrapper for the built-in timezone policy:

- `time_api_enabled`: always `false`
- `time_api_provider`: `built-in`
- `time_api_endpoint`: empty
- `time_api_query_template`: empty
- `time_api_default_timezone`: default business timezone, currently `Asia/Shanghai`
- `time_api_refresh_minutes`: `0`
- `time_api_key_ref`: empty

## Rules

- Store and compare backend timestamps in UTC.
- Use the default business timezone only for display conventions, admin explanations, and features that need a business-local day boundary.
- For admin-local analytics windows, pass an explicit timezone offset or timezone from the caller.
- For visitor risk and profiling, keep collecting client timezone signals from request data such as `CF-Timezone` and `X-Timezone`.
- In customer service, display the visitor's local time from `visitor_profiles.timezone` when available; missing or invalid data must remain explicitly unknown.
- Reuse `visitor_profiles.timezone` for future customer-facing timing features such as notification windows or agent scheduling instead of adding another timezone setting or time API.
- Do not add a new third-party time provider unless a future integration requires server clock attestation or cross-system signature timestamp calibration.

## Current Code Paths

- Admin timezone setting UI: `web/admin/src/components/admin/settings/TimeApiSettingsCard.vue`
- Generic private settings persistence: `internal/api/admin/settings_handler.go`
- Customer-service analytics local window: `internal/service/customer_service_analytics_service.go`
- Visitor timezone capture: `internal/api/v1/cart/handler.go` and `internal/api/v1/ticket/customer_service_session.go`
- Customer-service context timezone mapping: `internal/service/customer_service_context_service.go`
- Customer-service local-time display: `web/admin/src/components/admin/customer-service/CustomerContextPanel.vue`
- Public chat timezone header: `../nuxt-i18n/app/composables/chat/useCustomerServiceChatSync.ts`
