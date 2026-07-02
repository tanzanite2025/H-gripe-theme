# Backend Maintainability Guide

This guide captures the backend boundaries we want to preserve as the project grows.

## Layering Rules

- Handlers parse HTTP input, call services, and format responses.
- Services own business rules and use-case transactions.
- Repositories own persistence queries and database-specific details.
- Domain packages define models and domain constants.
- Shared packages under `internal/pkg/` should stay small and dependency-light.

## Transaction Rules

- A transaction should cover a complete business invariant, not a single repository convenience call.
- Do not validate balance/stock/coupon limits outside the transaction that mutates them.
- Avoid async side effects for value-bearing deductions such as points, gift cards, coupons, or inventory.
- Prefer atomic SQL predicates or row locks for counters and balances.

## Handler Rules

- Do not inject raw `*gorm.DB` into handlers.
- Do not call repositories directly from handlers when business rules exist.
- Keep request DTOs close to handlers and business DTOs close to services.
- Return sanitized errors through the shared API error/response helpers.

## Repository Rules

- Repositories should not decide business policy.
- Keep repository methods small and composable.
- If a method requires a transaction, accept a transaction-bound DB/repository through an explicit pattern used by the service.

## Documentation Rules

- Current backend docs live in `go-backend/` and `go-backend/docs/`.
- Historical reports live in `../docs/archive/`.
- Do not keep completion reports in active docs.
- Avoid "production ready", "100% complete", and performance claims unless verified by current tests or measurements.

## Current Priority

The backend should keep moving toward one source of truth:

- Go backend is the API source of truth.
- Frontend should display server-calculated checkout/price results instead of recalculating business totals.
- WordPress compatibility should only exist as explicit migration tooling, not runtime product behavior.
