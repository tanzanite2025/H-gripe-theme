# Customer-Service Auto Reply System

Last audited: 2026-07-29

> **Strict locale decision (2026-07-29):** This document is the base system
> design. The stricter multilingual rules in
> `AUTO_REPLY_I18N_ARCHITECTURE.md` are authoritative wherever the two
> documents differ. In particular, `*` and empty/unknown locales are legacy
> audit data only and must never match a localized request.

This document defines the durable design for Tanzanite customer-service
automatic replies. It is the implementation boundary for the admin panel,
Go backend, Nuxt storefront, database migrations, and future recommendation
orchestration.

## 1. Current Audit

The repository already contains a partial automatic-reply implementation:

- `ticket.AutoReplyRule` maps to `ticket_auto_replies`.
- Rules support `welcome` and `keyword` triggers.
- Keyword matching supports `exact` and `contains`.
- Public Chat exposes:
  - `GET /api/v1/customer-service/auto-reply/welcome`
  - `POST /api/v1/customer-service/auto-reply/match`
- These endpoints are legacy compatibility routes. The current Nuxt customer
  UI creates/reuses a conversation and reads persisted history instead of
  calling rule matching directly.
- Durable Public Chat facts are stored in `tickets` with
  `category = customer_service` and `ticket_messages`.

The implementation is incomplete and has these known gaps:

1. There is no admin CRUD API or admin UI for automatic-reply rules.
2. Rule content is a single `reply_message` string.
3. Automatic replies are forced to `message_type = text`; images, generic
   links, product cards, and order cards cannot be configured.
4. The automatic-reply table is only present in GORM `AutoMigrate`; there is
   no versioned SQL migration for an existing database.
5. Automatic replies are written with `user_id = 0`, while the current
   `ticket_messages.user_id` column is required and references `users.id`.
6. `agent_id` is stored on a rule but is not applied by the active-rule query.
7. Rules have no locale, cooldown, per-conversation policy, or audit state.
8. The storefront currently owns the trigger timing. A non-Nuxt client can
   send a message without invoking automatic replies.

The retired `chat_messages` and `chat_sessions` tables must stay retired.
They were a second message source and are not part of this design.

## 2. Goals

The automatic-reply system must:

- Be configurable from the authenticated admin panel.
- Use the backend as the only rule and delivery authority.
- Persist every delivered reply in the existing `ticket_messages` table.
- Support localized text and structured replies for:
  - text;
  - image;
  - generic link;
  - product reference;
  - order reference.
- Allow safe links and managed media assets without allowing arbitrary
  executable content.
- Match the selected public agent when a rule is agent-scoped.
- Preserve legacy global rules for audit and migration review, but do not use
  them as a runtime fallback for localized requests.
- Enforce cooldown and duplicate protection in the backend.
- Remain compatible with the existing customer-service SSE and HTTP history
  contract.
- Make an automatic reply distinguishable in admin and storefront displays.

## 3. Non-Goals

- Do not reintroduce `chat_messages` or `chat_sessions`.
- Do not put staff tools in the Nuxt storefront.
- Do not store product or order cards as plain text only.
- Do not make the browser the source of truth for rule matching or cooldown.
- Do not add machine-learning recommendation logic to the first release.
- Do not allow arbitrary HTML or JavaScript in reply content.

## 4. Ownership Boundaries

### Go backend

Owns:

- rule validation and persistence;
- locale and agent selection;
- keyword matching;
- cooldown and idempotency;
- structured message validation;
- automatic-reply message creation;
- audit metadata and delivery result;
- public and admin API contracts.

### Admin frontend

Owns:

- rule list, filters, create, edit, enable/disable, and delete actions;
- structured content editor;
- image selection/upload through the existing media boundary;
- link preview and validation feedback;
- rule preview by locale and message type.

### Nuxt storefront

Owns:

- rendering the persisted message contract;
- displaying automatic replies with the normal chat timeline;
- subscribing to SSE or refreshing HTTP history.

It must not decide whether a reply matches, whether cooldown has elapsed, or
whether an automatic reply should be persisted.

## 5. Durable Data Model

The existing `ticket_auto_replies` table will be versioned through SQL
migrations and extended instead of replaced.

### Rule fields

| Field | Meaning |
| --- | --- |
| `id` | Stable rule id |
| `type` | `welcome` or `keyword` |
| `trigger_keyword` | Keyword for `keyword` rules; empty for `welcome` |
| `match_type` | `exact` or `contains` |
| `locale` | Canonical storefront locale; legacy `*` rows are inactive and never match |
| `agent_id` | Public agent user id as text; empty means all agents |
| `message_type` | `text`, `image`, `link`, `product`, `order` |
| `reply_message` | Human-readable fallback text |
| `metadata` | Validated JSON payload for structured content |
| `attachments` | Validated JSON array of managed asset URLs |
| `is_active` | Whether the rule can match |
| `priority` | Higher values match first |
| `cooldown_seconds` | Minimum interval for the same rule and conversation |
| `created_at` / `updated_at` | Audit timestamps |

The current `reply_message` value remains the fallback display text and is
kept for backward compatibility. Structured values must not be hidden inside
that field.

### Message contract

Automatic replies use the existing `ticket_messages` row and add:

```json
{
  "_source": "auto_reply",
  "_rule_id": 123,
  "_dedupe_key": "autoreply:rule:123",
  "_trigger": "keyword",
  "_locale": "en"
}
```

The message content is represented by:

- `content`: safe human-readable fallback;
- `message_type`: normalized message type;
- `metadata`: structured JSON payload;
- `attachments`: JSON array of managed URLs.
- `source`: response-level marker derived from metadata; automatic replies use
  `auto_reply`.

For generic links, metadata is limited to:

```json
{
  "kind": "link",
  "url": "https://example.com/path",
  "title": "Safe display title",
  "description": "Optional short description",
  "thumbnail": "https://cdn.example.com/image.webp"
}
```

The backend must allow only `http` and `https` URLs, reject control
characters, and never render raw HTML from a rule.

Relative URLs are limited to application paths beginning with a single `/`;
protocol-relative URLs such as `//host/path` are rejected. Image rules require
at least one managed attachment, and product/order rules require structured
metadata.

Automatic reply cooldown is keyed by conversation plus stable rule identity,
not by the rendered locale metadata. This prevents the same welcome or keyword
rule from being resent merely because the browser sends a different
`Accept-Language` value. Rows written before `_dedupe_key` existed are still
recognized through `_rule_id` when present, and through content for very old
text-only rows without metadata.

## 6. Trigger and Delivery Flow

### Welcome

1. Public Chat creates or reuses a customer-service conversation through
   `POST /api/v1/customer-service/conversations`.
2. The backend selects the highest-priority active welcome rule for the
   canonical requested locale and selected agent/group scope. If the locale
   is missing or no rule exists for that locale, no localized automatic reply
   is sent.
3. The backend checks the rule cooldown for the conversation.
4. The backend creates one `ticket_messages` row if delivery is allowed.
5. The backend publishes the normal customer-service message-created event.

### Keyword

1. The backend persists the customer message.
2. In the same customer-service request flow, the backend evaluates active
   keyword rules.
3. Matching uses normalized text and priority order.
4. The backend applies cooldown and duplicate protection.
5. The backend persists the automatic reply and publishes its event.
6. The public client only refreshes from HTTP/SSE; it does not call a second
   matching endpoint.

The legacy public welcome and match endpoints remain temporarily available for
older clients during rollout, but the current Nuxt client must stop invoking
them directly. These endpoints must use the same idempotency checks so old
clients cannot create unbounded duplicates.

The cooldown read and message insert are one repository operation. On
PostgreSQL, the conversation row is locked during that operation so concurrent
legacy retries cannot both pass the cooldown check.

## 7. Admin API

The rule management API belongs under the existing authenticated customer
service namespace:

```text
GET    /api/admin/customer-service/auto-reply/rules
GET    /api/admin/customer-service/auto-reply/rules/:id
POST   /api/admin/customer-service/auto-reply/rules
PUT    /api/admin/customer-service/auto-reply/rules/:id
DELETE /api/admin/customer-service/auto-reply/rules/:id
```

Permissions:

- list/read: `ticket:view`;
- create/update/enable/disable: `ticket:edit`;
- delete: `ticket:delete`.

The API must validate:

- trigger type and match type;
- locale and agent scope;
- required keyword for keyword rules;
- message type;
- metadata schema;
- URL scheme and media URL format;
- non-empty fallback text;
- non-negative priority and cooldown.

## 8. Migration and Rollout

### Phase 0: Repair the existing foundation

- Add a versioned SQL migration for `ticket_auto_replies`.
- Add missing structured rule columns.
- Preserve existing rules with:
  - their current locale for audit;
  - legacy `*`, empty, and unknown locales marked inactive;
  - `message_type = 'text'`;
  - empty structured payloads;
  - welcome cooldown of 24 hours when absent.
- Ensure automatic replies use a real staff user id from the conversation or
  assigned agent so the existing foreign key remains valid.
- Add metadata that identifies automatic replies.

### Phase 1: Admin text rules

- Add backend CRUD and admin UI.
- Keep welcome and keyword behavior compatible with existing records.
- Move keyword triggering into the backend message flow.

### Phase 2: Structured replies

- Add image, link, product, and order editors.
- Reuse existing media storage and product/order payload builders.
- Extend the shared message normalizer/renderers only where needed.

### Phase 3: Operations

- Add rule preview, delivery counters, audit logs, and safe test-send.
- Add missing-locale diagnostics and per-rule monitoring.

Every phase must use the same `ticket_messages` source and must be
backward-compatible with HTTP history reads.

## 9. Acceptance Criteria

- Existing text welcome and keyword rules still work after migration.
- An existing database receives `ticket_auto_replies` through SQL migration,
  without relying on `AutoMigrate`.
- Automatic replies no longer insert an invalid `user_id = 0`.
- Current Nuxt sends one customer message and receives at most one matching
  automatic reply.
- A rule can be enabled, disabled, edited, and deleted from the admin panel.
- A locale-specific rule never falls back to a global or other-language rule
  when no translation exists.
- A link reply survives reload and renders as a safe link.
- An image reply survives reload and renders from a managed URL.
- Product/order replies remain structured and render in Admin and Nuxt.
- No legacy generic chat table or second message source is introduced.

## 10. Verification Checklist

Before shipping each phase:

1. Run the Go unit and repository tests.
2. Run the admin type/build checks.
3. Inspect migration ordering on an existing database.
4. Test anonymous and authenticated conversations.
5. Test legacy wildcard rows remain inactive and agent-scoped rules.
6. Test canonical locale exact match with no cross-language fallback.
7. Test repeated keyword messages inside and outside cooldown.
8. Reload both Admin and Nuxt history and verify identical message facts.
