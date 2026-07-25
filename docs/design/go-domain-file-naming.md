# Go Domain File Naming

Last updated: 2026-07-25

## Rule

Do not add generic `model.go` files under `go-backend/internal/domain`.

Domain files must be named after the business fact, aggregate, or contract they contain, for example:

- `product.go`
- `product_variant.go`
- `shipping_template.go`
- `tracking.go`
- `warranty_claim.go`
- `gift_card.go`
- `member_level.go`
- `ticket_message.go`

This keeps search results useful. A developer looking for warranty service history should land on `warranty_service_record.go`, not inspect one of many unrelated `model.go` files.

## Current State

The previous `model.go` files under `go-backend/internal/domain` have been removed or split into business-specific files.

Large domains now use multiple files by responsibility:

- `product`: product, media, variants, specs, product types, attributes, cart, response contracts.
- `shipping`: templates, carrier services, tracking, zones, packaging rules, template bindings.
- `registration`: product registration, warranty claim, warranty service record.
- `coupon`: coupons and gift cards are separate facts.
- `loyalty`: transactions, check-ins, referrals, member levels, user balances.
- `ticket`: ticket, ticket message, auto-reply rule.

## Adding New Domain Types

When adding a new domain type:

1. Name the file after the fact source, not after the layer.
2. Keep GORM model structs close to their table methods and lifecycle hooks.
3. Split response/request DTOs into `*_contract.go` or `*_response.go` when they grow beyond the model's direct display helpers.
4. Split embedded operational facts into their own file instead of appending to a nearby model.

Examples:

- A future warranty attachment table should be `warranty_attachment.go`.
- A future shipping rate quote snapshot should be `shipping_quote_snapshot.go`.
- A future customer chat participant table should be `chat_participant.go`, not `model.go`.
