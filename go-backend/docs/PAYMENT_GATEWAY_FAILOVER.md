# Payment Gateway Failover

## Scope

The payment API now records provider health in Redis and opens a short circuit
when a provider has a high transient failure rate. The feature covers Stripe,
PayPal, Alipay, and WeChat Pay payment API calls.

The circuit only counts network failures, timeouts, and provider/server errors
such as HTTP 500, 502, 503, and 504. Card declines, insufficient funds, 3DS
decisions, invalid requests, currency errors, and order mismatches are business
outcomes and do not open the circuit.

## Default policy

- Health window: 60 seconds
- Failure-rate threshold: strictly greater than 15%
- Minimum sample count: 20 provider API responses
- Open-circuit duration: 30 seconds

The values are configurable under `payment_gateway_circuit_breaker` or with
the matching `PAYMENT_GATEWAY_CIRCUIT_BREAKER_*` environment variables.

## API behavior

When a circuit is open, `/payment/methods` marks that provider unavailable with
`gateway_circuit_open`. A payment creation failure caused by a transient
provider incident returns `payment_gateway_degraded` with:

- `gateway`
- `circuit_open`
- `retry_after_seconds`
- `failure_rate`
- `sample_count`
- `payment_attempt_may_still_be_processing`
- `fallback_payment_methods`

The Nuxt checkout displays the available fallback methods and requires the
customer to choose one. The system does not silently convert a Stripe order
into a PayPal order because each order stores its selected payment provider
and each provider handler validates that relationship.

## Important retry boundary

Checkout currently creates the local unpaid order before it creates the
provider payment session. A transient provider failure therefore leaves that
local order unpaid. The checkout UI recommends an alternative method, but the
customer must explicitly choose it and submit again. Automatic cross-provider
switching is intentionally deferred until payment-attempt reconciliation can
prove that the first provider did not accept the request.

## File responsibilities

- `internal/service/payment_gateway_circuit_breaker_service.go`: Redis-backed
  health window and circuit state.
- `internal/pkg/payment/gateway_error_classifier.go`: transient provider error
  classification.
- `internal/api/v1/payment/gateway_failover_handler.go`: API response,
  fallback recommendation, and handler integration helpers.
- `internal/api/v1/payment/*_order_handler.go`: provider-specific calls only;
  each call reports success or transient failure to the common helper.
- `nuxt-i18n/app/components/CheckoutModal.vue`: customer-facing fallback
  recommendation and explicit payment-method selection.
