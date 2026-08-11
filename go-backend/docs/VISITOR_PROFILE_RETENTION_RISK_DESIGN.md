# Visitor Profile Retention and Risk Design

Last updated: 2026-07-28

This document defines how Commerce Platform should retain visitor profile data, behavior events, and risk telemetry without letting low-value traffic fill the database. It is intentionally written as a long-term implementation guide so the work can pause and resume without re-deciding the same boundaries.

## Goal

Visitor data has two different jobs:

- Help commerce workflows: recommendations, cart context, customer-service context, and returning anonymous visitors.
- Preserve enough security evidence to detect scraping, abuse, bot traffic, suspicious IP ranges, and repeated attack patterns.

Those jobs should not share one growing table. Marketing-quality visitor profiles must stay clean. Security telemetry must stay compact and queryable. Raw event data must expire.

## Current State

The current implementation already has useful pieces:

- `visitor_profiles` is an upsert-style profile table, not a raw request log.
- `VisitorProfileService.Touch()` binds facts such as `user_id`, public chat visitor hash, cart session id, email, locale, coarse region, IP hash, and user-agent hash.
- IP-derived facts must use the application-resolved client IP from the configured trusted proxy chain. Edge country headers are accepted only after the trusted edge metadata middleware has verified the immediate proxy peer. Raw `X-Forwarded-For`, `X-Real-IP`, `CF-Connecting-IP`, or country headers from an untrusted client must not be stored as visitor evidence.
- `recommendation_events` is append-only behavior telemetry for recommendation analytics.
- Admin visitor profile pages list and filter `visitor_profiles`.

The main concern is that profile creation can happen too early:

- `nuxt-i18n/app/composables/useCart.ts` loads the cart from `/cart/summary` when `useCart()` initializes.
- `go-backend/internal/api/v1/cart/handler.go` creates a `session_id` cookie when missing.
- The same cart summary path calls `touchVisitorProfile()`.

If the storefront calls cart summary on most page loads just to show a cart count, then a visitor who only opened one page can become a `visitor_profiles` row. That is not a meaningful customer profile. It is also not the right place to store security evidence.

## Core Principle

Do not ask one table to be all three of these:

- Online customer profile
- Raw behavior log
- Security/risk evidence

Use three layers instead.

## Data Layers

### 1. Online Visitor Profile

Table: `visitor_profiles`

Purpose:

- Store only visitors with real business value or clear identity linkage.
- Serve customer service, cart recovery, recommendation identity stitching, and admin lookup.

Should create or update profile when:

- User logs in.
- Visitor provides email.
- Visitor sends a public chat message.
- Visitor adds to cart.
- Visitor syncs a non-empty cart.
- Visitor starts checkout.
- Visitor adds wishlist item.
- Visitor completes a high-intent tool interaction, such as spoke calculator usage with meaningful parameters.
- Existing profile is already known and receives fresh facts.

Should not create profile only because:

- Page loaded.
- Empty cart summary was requested.
- A frontend component initialized.
- A cookie was missing and got created.
- A single low-value page view happened.

Recommended future fields:

```sql
ALTER TABLE visitor_profiles
  ADD COLUMN IF NOT EXISTS profile_quality_score INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_meaningful_seen_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS first_meaningful_seen_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS profile_status VARCHAR(24) NOT NULL DEFAULT 'active',
  ADD COLUMN IF NOT EXISTS retention_until TIMESTAMPTZ NULL;
```

Suggested `profile_status` values:

- `candidate`: has a weak identifier but no meaningful action yet.
- `active`: has meaningful customer or commerce signal.
- `archived`: kept for history, not shown by default.
- `suppressed`: excluded from marketing/recommendation use because it looks abusive or invalid.

Important: `candidate` should be short-lived and hidden from the normal visitor profile list by default.

### 2. Behavior Event Log

Table: `recommendation_events`

Purpose:

- Feed recommendation analytics and profile scoring.
- Record useful short-term behavior facts.

This table is append-only and can grow quickly, so it must have retention rules.

Recommended retention:

- Low-intent events such as `page_view`: 14-30 days.
- Product view, search, filter, category click: 30-60 days.
- High-intent events such as add-to-cart, wishlist, checkout, purchase, calculator use: 90-180 days or until aggregated.
- Events tied to completed purchases may be summarized into order/user features and then deleted from raw logs.

Required follow-up:

- Add scheduled cleanup.
- Add daily/hourly aggregates before deleting raw events.
- Consider date partitioning if traffic grows.

### 3. Risk Aggregation Layer

Proposed table: `visitor_risk_daily_facts`

Purpose:

- Preserve security evidence without creating huge profile rows.
- Detect bot traffic, scraping, abuse bursts, and suspicious IP/UA combinations.
- Support future block/allow decisions.

This table should aggregate by day and hashed request identity, not insert one row per request.

Suggested schema:

```sql
CREATE TABLE IF NOT EXISTS visitor_risk_daily_facts (
    id BIGSERIAL PRIMARY KEY,
    day DATE NOT NULL,
    ip_hash VARCHAR(64) NOT NULL,
    user_agent_hash VARCHAR(64) NULL,
    country_code VARCHAR(8) NULL,
    first_seen_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 0,
    unique_path_count INTEGER NOT NULL DEFAULT 0,
    unique_anonymous_count INTEGER NOT NULL DEFAULT 0,
    unique_session_count INTEGER NOT NULL DEFAULT 0,
    invalid_request_count INTEGER NOT NULL DEFAULT 0,
    auth_failure_count INTEGER NOT NULL DEFAULT 0,
    checkout_failure_count INTEGER NOT NULL DEFAULT 0,
    bot_like_user_agent_count INTEGER NOT NULL DEFAULT 0,
    no_cookie_request_count INTEGER NOT NULL DEFAULT 0,
    meaningful_action_count INTEGER NOT NULL DEFAULT 0,
    risk_score INTEGER NOT NULL DEFAULT 0,
    risk_level VARCHAR(16) NOT NULL DEFAULT 'normal',
    sample_paths JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(day, ip_hash, user_agent_hash)
);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_daily_facts_day_score
    ON visitor_risk_daily_facts(day, risk_score DESC);

CREATE INDEX IF NOT EXISTS idx_visitor_risk_daily_facts_ip_day
    ON visitor_risk_daily_facts(ip_hash, day DESC);
```

Implemented manual-decision table:

```sql
CREATE TABLE IF NOT EXISTS visitor_risk_decisions (
    id BIGSERIAL PRIMARY KEY,
    scope VARCHAR(24) NOT NULL,
    value_hash VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL,
    reason VARCHAR(500) NOT NULL,
    expires_at TIMESTAMPTZ NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Current scopes:

- `ip_hash`
- `ip_ua_hash`

The service derives scope and hash from the selected risk fact. The admin
client cannot submit an arbitrary hash. A fact with a user-agent hash uses the
more precise `ip_ua_hash` scope; its value is a hash of the stored IP hash and
UA hash together. A fact without a user-agent hash uses `ip_hash`.

Current actions:

- `ignore`
- `watch`
- `temporary_block`
- `block_candidate`

Decisions are append-only historical records. The current effective decision
is the newest non-expired record for the exact `ip_ua_hash` scope, falling back
to the `ip_hash` scope when no exact decision exists. `temporary_block` must
have a future `expires_at`. Saving a decision does not directly change
middleware blocking or rate limiting; a future enforcement layer must consume
these records explicitly.

## Is Same-IP Repeated Access Abnormal?

It has some value, but only as a weak signal.

Same IP can represent many normal users:

- Company or school NAT.
- Mobile carrier NAT.
- Public Wi-Fi.
- VPN / proxy.
- Shared hosting or crawler infrastructure.

So this rule is wrong:

> Same IP visited many times, therefore abnormal.

The better rule is:

> Same IP can increase risk only when combined with behavior shape, identity spread, and intent quality.

Useful combinations:

- Same IP + same user-agent + many anonymous ids in a short window.
- Same IP + high request burst + no meaningful action.
- Same IP + many invalid paths, 404s, auth failures, or API validation failures.
- Same IP + repeated checkout/payment failures.
- Same IP + high product scraping pattern, many product pages, low dwell, no add-to-cart.
- Same IP + impossible language/country churn.
- Same IP + no cookie acceptance across many requests.

Weak or non-risk examples:

- Same IP + several users + normal dwell + cart or checkout actions.
- Same IP + high traffic during ad campaign + normal product mix.
- Same IP + returning logged-in user.
- Same IP + customer service interaction.

## Scoring Model

Use two independent scores:

- `profile_quality_score`: whether this visitor deserves an online profile.
- `risk_score`: whether this traffic deserves security attention.

### Profile Quality Score

Suggested weights:

| Signal | Score |
| --- | ---: |
| Plain page view | 0 |
| Product view | 1 |
| Product dwell over 30 seconds | 2 |
| Search submit | 2 |
| Filter apply | 2 |
| Category navigation | 1 |
| Spoke calculator meaningful use | 5 |
| Wishlist add | 5 |
| Add to cart | 8 |
| Non-empty cart sync | 8 |
| Begin checkout | 12 |
| Email captured | 12 |
| Public chat message | 12 |
| Login / account binding | 20 |
| Purchase | 30 |

Recommended thresholds:

- `0-2`: do not create profile, or keep only transient client identity.
- `3-7`: candidate profile, short TTL, hidden by default.
- `8+`: active visitor profile.
- `20+`: high-value profile for customer-service context and recommendation identity stitching.

### Risk Score

Suggested weights:

| Signal | Score |
| --- | ---: |
| More than 120 requests from same IP in 10 minutes | +10 |
| More than 20 anonymous ids from same IP + UA per day | +15 |
| More than 50 unique product pages with no dwell/add-to-cart | +12 |
| High 404/invalid path ratio | +10 |
| Repeated auth or checkout failure | +12 |
| Known bot-like UA | +8 |
| No cookies across repeated requests | +8 |
| Requesting admin/API paths from public client | +20 |
| Meaningful customer action present | -8 |
| Logged-in account with normal history | -12 |
| Public chat conversation present | -8 |
| Purchase present | -20 |

Recommended levels:

- `0-19`: normal
- `20-39`: watch
- `40-69`: suspicious
- `70+`: block candidate, require manual review or temporary rule

## Retention Policy

Recommended defaults:

| Data | Default TTL | Notes |
| --- | ---: | --- |
| Raw low-intent behavior events | 30 days | Page views, impressions, generic clicks |
| Raw standard-intent behavior events | 60 days | Product views, search, filters, category clicks |
| Raw high-intent behavior events | 180 days | Add-to-cart, checkout, calculator, wishlist, purchase |
| Aggregated recommendation features | 365 days | Product/category/user affinity scores |
| Candidate visitor profiles | 7-30 days | Hidden by default, soft delete if no meaningful action |
| Active anonymous profiles | 180 days after last meaningful action | Keep if cart/chat/email exists |
| Account-linked profiles | Follow account retention policy | Do not delete independently unless account lifecycle says so |
| Risk daily facts | 180-365 days | Compact enough to keep longer |
| Risk decisions / block evidence | 365+ days or until expiry | Depends on operational policy |
| Deleted/suppressed profile audit record | 180 days | Only hashed identifiers, no marketing use |

## Cleanup Jobs

Add backend scheduled jobs or CLI commands:

```text
visitor:promote-candidates
visitor:archive-stale-profiles
visitor:delete-expired-candidates
behavior:aggregate-daily
behavior:delete-expired-raw-events
risk:rollup-daily-facts
risk:expire-decisions
```

Recommended order:

1. Aggregate raw behavior events into daily/user/product features.
2. Promote profiles that crossed the quality threshold.
3. Archive or soft delete candidate profiles that did not cross the threshold.
4. Delete expired low-value raw events.
5. Keep risk facts separately.

## Admin UI Rules

Visitor profile management should not show all raw or candidate data by default.

Default view:

- Active profiles only.
- Sort by `last_meaningful_seen_at`, then `last_seen_at`.
- Show profile quality score.
- Show why profile exists: `email`, `cart`, `chat`, `checkout`, `account`, etc.

Add filters:

- Active / Candidate / Archived / Suppressed
- Quality score range
- Last meaningful action
- Has email
- Has cart
- Has public chat
- Has risk flag

Add separate risk view:

- Risk daily facts grouped by IP hash / UA hash.
- Risk level and score.
- Request count, unique anonymous/session count.
- Sample paths.
- Action buttons: ignore, watch, temporary block, permanent block candidate.

Do not mix risk-only rows into the normal visitor profile list.

## Implementation Plan

### Phase 1: Stop Creating Low-Value Profiles

Scope:

- Update cart summary behavior so empty cart summary does not create a visitor profile.
- Touch profile only if one of these is true:
  - authenticated user exists
  - existing public chat visitor hash exists
  - existing cart session has real cart items
  - action is add/update/sync non-empty/checkout
- Keep cart count behavior working without turning every passive visitor into a profile.

Suggested code direction:

- Split cart helper into:
  - `getUserIDAndSession(c)` for cart operations
  - `getUserIDAndOptionalSession(c)` or `readUserIDAndSession(c)` for summary-only reads
- Do not create `session_id` merely for empty summary reads unless frontend truly needs it.
- Move `touchVisitorProfile()` out of generic `getUserIDAndSession()` and call it from meaningful operations.

### Phase 2: Add Profile Quality

Scope:

- Add `profile_quality_score`, `profile_status`, `last_meaningful_seen_at`, and `retention_until`.
- Introduce service method names that encode intent:
  - `TouchPassiveSeen()`
  - `TouchMeaningfulAction()`
  - `BindIdentityFact()`
- Do not let passive events create an active profile.

### Phase 3: Add Risk Aggregation

Scope:

- Add `visitor_risk_daily_facts`.
- Update request middleware or selected handlers to increment aggregate counters.
- Hash IP and UA with a server-side salt or existing stable hash strategy.
- Keep sample paths short and capped.

Important:

- This layer should be cheap and bounded.
- Use upsert counters, not one row per request.
- Keep it independent from marketing profile logic.

### Phase 4: Add Retention Jobs

Scope:

- Add CLI commands or scheduled service jobs for cleanup.
- Make TTL configurable through environment variables or admin settings.
- Log counts deleted/archived per run.

Suggested config:

```env
VISITOR_CANDIDATE_TTL_DAYS=14
VISITOR_ACTIVE_ANON_TTL_DAYS=180
BEHAVIOR_EVENTS_LOW_INTENT_RETENTION_DAYS=30
BEHAVIOR_EVENTS_STANDARD_INTENT_RETENTION_DAYS=60
BEHAVIOR_EVENTS_HIGH_INTENT_RETENTION_DAYS=180
RISK_DAILY_FACT_TTL_DAYS=365
```

### Phase 5: Admin Visibility

Scope:

- Update visitor profile admin list defaults.
- Add profile quality/status filters.
- Add a separate risk view or tab.
- Show "why retained" so operators trust the data.

## Database Growth Guardrails

Hard rules:

- Never create one profile per request.
- Never create one risk row per request.
- Never keep raw low-intent events forever.
- Never use IP-only matching as a strong identity signal.
- Never let risk-only data enter recommendation or marketing personalization by default.

Operational metrics:

- New visitor profiles per day.
- Candidate to active conversion rate.
- Candidate deletion count per day.
- Raw behavior events ingested per day.
- Raw behavior events deleted per day.
- Risk facts per day.
- Top IP hashes by risk score.
- Profile table size and index size.

Alert examples:

- New visitor profiles/day increases by more than 3x week-over-week.
- Candidate profiles exceed 70% of total profiles.
- `recommendation_events` grows without daily cleanup.
- One IP hash produces more than N anonymous IDs/day.

## Recommended Immediate Fix

The first code fix should be narrow:

1. Stop `/cart/summary` from creating visitor profiles for empty/anonymous reads.
2. Keep visitor profile creation on add-to-cart, non-empty sync, checkout, login, email capture, and public chat.
3. Add a short backend note/test proving passive cart summary does not create `visitor_profiles`.

This gives the fastest reduction in useless rows while preserving future attack analysis through the planned risk aggregation layer.

## Phase 1 Implementation Status

Implemented on 2026-07-28:

- `/cart/summary` now performs a read-only cart lookup. Anonymous visitors with no existing cart get an empty summary without a new `session_id`, `carts` row, or `visitor_profiles` row.
- Cart write handlers no longer create missing carts for update, remove, or clear requests. Missing carts return an empty/idempotent result instead of inserting an empty cart.
- `POST /cart/add` validates that the requested product/variant/quantity is purchasable before creating a cart session or cart row.
- `POST /cart/sync` only creates cart state when the submitted payload contains at least one purchasable item. Empty payloads, invalid product ids, inactive variants, and out-of-stock items do not create cart or visitor profile rows.
- Visitor profiles are touched after successful meaningful cart actions: add-to-cart, successful item update, remove/clear only when the existing cart had real items, and non-empty cart sync.
- Tests cover passive summary reads, invalid add-to-cart, empty sync, invalid sync, successful add-to-cart, and successful non-empty sync.

## Phase 2 Implementation Status

Partially implemented on 2026-07-28:

- `visitor_profiles` now has profile quality and retention fields:
  - `profile_quality_score`
  - `profile_status`
  - `last_meaningful_action`
  - `first_meaningful_seen_at`
  - `last_meaningful_seen_at`
  - `retention_until`
- Existing live profiles are backfilled as `active` with conservative quality scores based on account/email/chat/cart linkage.
- `VisitorProfileService` now exposes intent-specific methods:
  - `TouchMeaningfulAction()` creates or promotes active profiles.
  - `BindIdentityFact()` records identity linkage as meaningful data.
  - `TouchPassiveSeen()` updates only an existing profile and does not create a row.
- Cart profile touches are explicitly marked as `cart_action`.
- Public Chat no longer touches visitor profiles from the owner helper itself. Passive checks such as "has conversation", message reads, typing, and SSE now use existing owner identity rather than creating a fresh visitor profile. A profile is touched after a customer message is successfully saved.
- Admin visitor profile APIs and panels expose status, quality score, meaningful action times, retention timestamp, and default to active profiles.
- Manual retention cleanup is available from the admin visitor profile page and API:
  - expired `candidate` profiles are soft-deleted
  - expired anonymous `active` profiles are moved to `archived`
  - account-linked profiles are not archived by anonymous TTL rules

Still pending:

- Admin risk view and manual risk decisions.

## Phase 3 Implementation Status

Implemented on 2026-07-28:

- Added `visitor_risk_daily_facts` as a compact daily aggregation table keyed by `day + ip_hash + user_agent_hash`.
- Risk facts store aggregate counters and capped sample paths only; raw IPs, full user agents, request query strings, and per-request rows are not stored.
- Public `/api/v1` requests now flow through `VisitorRiskTelemetry` when `VISITOR_RISK_ENABLED=true`.
- `VisitorRiskService` aggregates facts in memory first and flushes them in batches, so normal request handling does not synchronously write a risk row on every request.
- Risk score is intentionally conservative:
  - `no_cookie_request_count` is counted but does not directly raise risk score.
  - score increases for invalid responses, auth failures, checkout/payment/order failures, and bot-like user agents.
  - successful meaningful commerce actions slightly reduce risk.
- Risk telemetry is disabled by default and requires explicit production enablement:
  - `VISITOR_RISK_ENABLED`
  - `VISITOR_RISK_HASH_SALT`
  - `VISITOR_RISK_FLUSH_INTERVAL_SECONDS`
  - `VISITOR_RISK_MAX_PENDING_FACTS`
  - `VISITOR_RISK_SAMPLE_PATH_LIMIT`
  - `VISITOR_RISK_RETENTION_DAYS`

Still pending:

- More precise unique anonymous/session counts from frontend behavior identifiers.
- Admin-facing controls for retention windows.

## Phase 4 Implementation Status

Partially implemented on 2026-07-28:

- Visitor profile TTL cleanup can now run automatically when enabled:
  - `WORKER_VISITOR_PROFILE_CLEANUP_ENABLED`
  - `WORKER_VISITOR_PROFILE_CLEANUP_INTERVAL_SECONDS`
- The scheduled job calls the same service method as the admin manual cleanup endpoint, keeping cleanup behavior centralized.
- The scheduler is disabled by default for safe rollout.
- Visitor risk facts flush from memory on a separate scheduler when `VISITOR_RISK_ENABLED=true`.
- Visitor risk facts are deleted after `VISITOR_RISK_RETENTION_DAYS` days; default is 365 days.
- Expired temporary risk decisions are retained for the same window and then
  deleted; indefinite `ignore`, `watch`, and `block_candidate` decisions remain
  as operator evidence until a later explicit policy removes them.
- Raw behavior events can now be cleaned automatically when enabled:
  - `WORKER_BEHAVIOR_EVENT_CLEANUP_ENABLED`
  - `WORKER_BEHAVIOR_EVENT_CLEANUP_INTERVAL_SECONDS`
  - `BEHAVIOR_EVENTS_LOW_INTENT_RETENTION_DAYS`
  - `BEHAVIOR_EVENTS_STANDARD_INTENT_RETENTION_DAYS`
  - `BEHAVIOR_EVENTS_HIGH_INTENT_RETENTION_DAYS`
  - `BEHAVIOR_EVENTS_CLEANUP_BATCH_LIMIT`
- Behavior event retention is tiered:
  - low intent: `page_view`, `recommendation_impression`
  - standard intent: `product_view`, `product_dwell`, `search_submit`, `filter_apply`, `category_navigation_click`, `recommendation_click`, `quiz_completed`
  - high intent: `calculator_use`, `add_to_cart`, `wishlist_add`, `begin_checkout`, `purchase`

Still pending:

- Admin-facing controls for retention windows.

## Phase 5 Implementation Status

Partially implemented on 2026-07-28:

- The admin visitor profile page now separates:
  - `访客画像`: commerce/customer-service profile facts.
  - `风险事实`: compact daily risk aggregates.
- Risk facts have dedicated admin endpoints for:
  - paginated list with level/date/score filters
  - aggregate statistics
  - manual retention cleanup for facts and old expired decisions
- The admin risk table displays risk score/level, request volume, invalid/auth/checkout failures, bot-like UA count, identity spread, and capped sample paths.
- IP and user-agent hashes are only shown as short previews; raw values are never returned to the admin UI.

Implemented on 2026-07-28:

- Risk facts can be assigned append-only manual decisions from the admin panel.
- The decision API derives the effective IP/UA scope from the selected fact,
  validates action/reason/expiry, and records the current admin user.
- Creating a manual decision writes a global audit-log entry for both success
  and validation failures. The audit details include the fact id, requested
  action, derived scope, decision id, expiry presence/time, and reason
  presence/length only.
- The risk list includes the current non-expired decision without exposing the
  full hash values.
- `GET /api/admin/customer-service/visitor-risk-facts/:id/decision` exposes the
  current effective decision for later admin/detail views.
- `POST /api/admin/customer-service/visitor-risk-facts/:id/decision` is admin-only.
- The UI clearly separates “record a decision” from enforcement. No automatic
  block or rate-limit behavior is connected yet.
- Global audit logs do not store raw IPs, raw user agents, full hash values, or
  the manual decision reason text.

Still pending:

- More precise unique anonymous/session counts from frontend behavior identifiers.
- Admin-facing controls for retention windows.
- A separately reviewed enforcement consumer with dry-run and audit logging.
