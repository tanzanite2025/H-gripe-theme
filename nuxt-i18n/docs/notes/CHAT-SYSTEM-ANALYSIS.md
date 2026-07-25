# Public Chat / customer-service architecture

Last audited: 2026-07-25

This document is the current source for Tanzanite's Public Chat boundary. Update it whenever chat routes, ownership rules, message payloads, or frontend/admin component responsibilities change.

## Hard boundary

The storefront must not contain a staff chat console.

- Nuxt storefront owns only the customer-facing chat experience: agent/profile selection, conversation creation, customer messages, customer message history, product/order sharing, and local UI cache.
- Go backend owns the durable source of truth: customer-service conversations and messages are stored through the ticket/customer-service tables and services.
- Admin frontend owns staff operations: conversation inbox, staff replies, read state, assignment/transfer, and future realtime controls.
- Public Chat Agent Profiles only decide which staff profiles are exposed to the storefront. They are configuration data, not a chat workspace.

The previous Nuxt `agentMode` approach was removed because it allowed a logged-in staff identity to turn the storefront chat modal into a staff console. That mixed customer UI and admin operations in the same bundle and created a long-term security and maintenance risk.

## Current route ownership

### Storefront customer routes

These routes are consumed by Nuxt customer UI:

| Capability | Route |
| --- | --- |
| List public chat agents | `GET /api/v1/customer-service/agents` |
| Create/reuse customer conversation | `POST /api/v1/customer-service/conversations` |
| Check existing customer conversation | `GET /api/v1/customer-service/has-conversation` |
| Send customer message | `POST /api/v1/customer-service/messages` |
| Publish customer typing state | `POST /api/v1/customer-service/typing` |
| Read customer-owned messages | `GET /api/v1/customer-service/messages/:conversation_id` |
| Subscribe to customer-owned realtime events | `GET /api/v1/customer-service/events` |
| Welcome auto-reply | `GET /api/v1/customer-service/auto-reply/welcome` |
| Keyword auto-reply | `POST /api/v1/customer-service/auto-reply/match` |
| Customer chat orders | `GET /api/v1/customer-service/orders` |
| Customer chat product search | `GET /api/v1/customer-service/products` |

Customer access is validated in Go using authenticated `user_id` or a signed visitor session cookie hash. A conversation id alone must never authorize access.

### Admin staff routes

These routes are consumed by the admin UI:

| Capability | Route |
| --- | --- |
| List staff-visible conversations | `GET /api/admin/customer-service/conversations` |
| Subscribe to staff-visible realtime events | `GET /api/admin/customer-service/events` |
| Read one conversation's customer context | `GET /api/admin/customer-service/conversations/:id/context` |
| Read one conversation's messages | `GET /api/admin/customer-service/conversations/:id/messages` |
| Send staff reply | `POST /api/admin/customer-service/conversations/:id/messages` |
| Publish staff typing state | `POST /api/admin/customer-service/conversations/:id/typing` |
| Mark messages read | `POST /api/admin/customer-service/conversations/:id/messages/mark-read` |
| Transfer conversation | `PATCH /api/admin/customer-service/conversations/:id/transfer` |
| List visitor profiles | `GET /api/admin/customer-service/visitor-profiles` |
| Visitor profile stats | `GET /api/admin/customer-service/visitor-profiles/stats` |

Admin and manager roles can see all customer-service conversations. Support users only see conversations assigned to their own backend user id.

Conversation listing filters are applied in the Go API/repository layer, not by filtering only the current frontend page. Supported query parameters:

- `search`: conversation id, ticket number, logged-in customer email/name, captured visitor email/cart session, visitor hash, or message content.
- `status`: `pending`/`open`, `active`/`in_progress`, `closed`/`resolved`.
- `identity`: `account`/`member`/`user` or `anonymous`/`visitor`/`guest`.
- `assigned_to`: backend user id of the assigned customer-service agent. Admin/manager users can use this filter; support users are always forced to their own backend user id.
- `unread`: `true`/`1`/`yes` for conversations with unread customer messages.

`/api/v1/customer-service/agent/*` has been removed. Any new staff UI must use the admin route namespace.

## Current file ownership

### Nuxt storefront

- `app/components/WhatsAppChatModal.vue`
  - Customer chat shell only: welcome panel, selected public agent header, user chat body, product drawer, wishlist/history/auth drawers.
  - Must not import staff conversation-list or transfer components.
- `app/composables/chat/useWhatsAppState.ts`
  - Customer-side chat modal orchestration only: selected public agent, active tab, drawer visibility, product search entry, order/product sharing, login/member subflows, and UI cache loading/saving.
  - Must not own customer-service HTTP persistence or SSE internals directly.
- `app/composables/chat/useCustomerServiceChatSync.ts`
  - Customer-service HTTP/SSE sync boundary for the storefront.
  - Owns `conversation_id` persistence from backend responses, customer message send, server message normalization, optimistic-message replacement/failure marking, welcome/keyword auto-reply refresh, and EventSource lifecycle.
  - Keeps HTTP `/customer-service/messages/:conversation_id` as the message source of truth; SSE remains a refresh/invalidation signal.
- `app/components/whatsapp/ChatWelcomePanel.vue`
  - Public agent/profile selection for customers.
- `app/components/whatsapp/UserChatBody.vue`
  - Customer tabs and customer message body.
- `app/components/whatsapp/ChatTab.vue`
  - Customer message list/input and optional visitor email capture.
- `app/composables/chat/useChatAgentDirectory.ts`
  - Public agent directory request and DEV-only fallback.
- `app/composables/chat/useChatStorage.ts`
  - Per-conversation/per-agent browser cache with expiry.

### Admin frontend

- `go-backend/web/admin/src/views/CustomerServiceChats.vue`
  - Staff inbox for Public Chat conversations.
  - Reads/writes only through `/api/admin/customer-service/*`.
  - Displays the read-only customer context panel beside the staff conversation.
  - The conversation list must expose display-safe customer context at a glance: visitor/member identity, coarse region, and member points-tier icon when the customer is a logged-in account.
- `go-backend/web/admin/src/views/VisitorProfiles.vue`
  - Read-only operations page for `visitor_profiles`.
  - Supports search and capture-state filters for account/anonymous identity, email, cart session, Public Chat visitor binding, country, locale, and last-seen window.
  - Must not edit visitor cookies, cart sessions, or guessed identity fields.
- `go-backend/web/admin/src/router/index.js`
  - Registers `/customer-service` and `/visitor-profiles`.
- `go-backend/web/admin/src/layouts/MainLayout.vue`
  - Adds the sidebar entry.

### Go backend

- `internal/service/ticket_customer_service.go`
  - Customer owner normalization, visitor/user access checks, conversation creation, public message persistence, and staff conversation-scope checks.
- `internal/service/customer_service_context_service.go`
  - Read-only resolver for the customer context shown to staff: account, contact, cart, wishlist, recent orders, browsing history, and capture signals.
- `internal/service/visitor_profile_service.go`
  - Unified visitor profile touch/merge logic for Public Chat visitor hash, cart session id, optional email, locale, coarse region, and privacy-preserving request hashes.
  - Read-only admin list/stat snapshots; output masks the Public Chat visitor hash and only exposes capture status for IP/User-Agent fingerprints.
- `internal/domain/visitor/model.go` and `internal/repository/visitor_profile_repository.go`
  - `visitor_profiles` fact source.
- `internal/api/admin/visitor_profile_handler.go`
  - Read-only admin API for visitor profile search and capture coverage stats.
- `internal/pkg/visitorcookie/customer_service.go`
  - Shared signed Public Chat visitor cookie signing, validation, and hash generation.
- `internal/api/v1/ticket/customer_service_public_handler.go`
  - Customer-side HTTP API.
- `internal/api/v1/cart/handler.go`
  - Touches visitor profiles when cart session data is created/read/updated.
- `internal/api/admin/customer_service_chat_handler.go`
  - Admin staff inbox HTTP API.
- `internal/repository/ticket_repository.go`
  - Customer-service conversation queries, including assigned-agent, status, unread, identity, and search filtering.
  - Ordinary ticket list/stat queries explicitly exclude `category = customer_service`.
- `internal/service/ticket_service.go` and `internal/service/ticket_message_service.go`
  - Ordinary ticket APIs reject `category = customer_service`; Public Chat conversations must use the dedicated customer-service service methods.

## Data source

The active Public Chat fact source is:

- `tickets` with `category = customer_service`
- `ticket_messages`
- public chat staff profiles from `agent_profiles`
- visitor context from `visitor_profiles`

`ticket_messages` now stores structured chat payload facts directly:

- `content`: human-readable fallback summary for every message.
- `message_type`: `text`, `product`, `order`, `image`, or `config_confirm`.
- `metadata`: JSON string payload for structured message cards.

For `config_confirm`, Nuxt builds `metadata` from the product/SKU fact source only:

- `metadata.product`: product id, selected variant id, title, slug, SKU, URL, thumbnail, price label, and numeric price.
- `metadata.selections`: selected variant title, SKU, variant option rows from `variants.option_values`, stock, `weight_grams`, price label, and numeric price.
- The shared frontend builder is `app/composables/chat/useProductConfigConfirmPayload.ts`.
- Admin renders this payload read-only. It must not let staff edit or fork product configuration facts inside the chat workspace.

For `order`, Nuxt builds `metadata` from the customer's order fact source:

- `metadata`: order id/number, status, payment status, shipping status, total, currency, URL, item count, and order item summaries.
- The shared frontend builder is `app/composables/chat/useOrderChatPayload.ts`.
- Orders Tab uses an explicit confirm button. Do not make the whole order card a hidden send action.

Do not persist product/order/config chat payloads only in frontend local storage or transient API response data. Nuxt storefront, Admin staff UI, HTTP history reads, and SSE refreshes must all render from the same persisted `ticket_messages` row.

`customer_service` records are intentionally not part of the ordinary ticket inbox/list/stat/detail/message path. This is important because anonymous Public Chat conversations still need a persisted `tickets.user_id` for the legacy schema, while the real customer identity is `customer_user_id` or `visitor_session_hash`. Treating those records as ordinary tickets would leak or misclassify customer conversations.

The older generic `/api/v1/chat/messages` path has been retired. The Go route, handler, service, repository, and domain models were removed, and migration `033_drop_legacy_chat_tables.up.sql` drops `chat_messages` / `chat_sessions` so they cannot become a second message source. Do not reintroduce a generic chat message table for Public Chat; use the dedicated customer-service ticket path.

## Customer context boundary

The admin customer-service workspace now has a read-only customer context resolver. Its rule is simple: display existing facts, do not invent missing identity data.

### Admin conversation list display contract

The staff conversation list should help operations understand who is chatting without turning the inbox into an unsafe identity-profiling surface.

Each row should show:

- Identity state:
  - `Visitor` / `Guest` when the conversation belongs only to the signed Public Chat visitor cookie.
  - `Member` / `Logged-in customer` when the conversation is linked to a real customer account.
- Member tier:
  - For logged-in customers, show the customer's points / membership tier icon beside the identity label.
  - For anonymous visitors, do not show a fake tier, guessed tier, or default member icon.
- Coarse region:
  - Show only a broad operational region such as `中国台湾`, `中国香港`, `中国大陆`, `United States`, `Japan`, or `Unknown`.
  - The row can use region as customer-service context and operations analysis input, but must not expose raw IP, raw User-Agent, or hidden hashes.
- Contact/capture hint:
  - Show captured email only when the customer provided it or when it comes from an authenticated account.
  - Do not infer email or identity from cart/session fingerprints.

The list API should return these fields as display-ready values so the admin frontend does not duplicate identity, tier, or region mapping rules.

### Coarse region analytics boundary

The first GeoIP/region enhancement target is intentionally narrow: "today's chatting customers by broad region" for operations analysis.

Allowed:

- Count customer-service conversations/messages by coarse region.
- Connect coarse region to product-interest signals already present in the chat context, cart, wishlist, or shared product/config cards.
- Display aggregated analysis such as "customers from 中国台湾 asked more about wheelsets today."

Not allowed in this phase:

- Store or display raw IP addresses.
- Store precise geolocation, latitude/longitude, street-level location, or postal code.
- Build long-term hidden behavioral profiles from IP/User-Agent fingerprints.
- Let region data drive discriminatory or irreversible customer decisions.

Preferred source order:

1. CDN/request region headers, for example country/region headers supplied by Cloudflare or the deployment edge.
2. Existing `visitor_profiles` coarse fields when the visitor was already touched by cart/chat.
3. `Unknown` when the request does not provide reliable region context.

If a future phase adds a GeoIP provider, it must still output only coarse location into the customer-service/admin contract unless the privacy/legal boundary is explicitly changed.

### Currently resolved

- Logged-in customer account:
  - user id, email, username/display name, role, locale, status, registration time;
  - cart by `cart.user_id`;
  - wishlist by `wishlist.user_id`;
  - recent orders by `orders.user_id`;
  - browsing history by `browsing_history.user_id`.
- Anonymous Public Chat visitor:
  - conversation id, ticket number, assigned agent, status, creation/update time;
  - masked `visitor_session_hash`;
  - visitor profile id if one exists;
  - cart by `visitor_profiles.cart_session_id` when the chat visitor and cart session have been linked by backend cookies;
  - optional email from the customer chat email field or authenticated account;
  - locale and coarse region only when captured from request headers/CDN headers.
- Operations inspection:
  - `/visitor-profiles` in admin lists the `visitor_profiles` fact source without exposing raw IP or raw User-Agent.
  - Search covers visitor profile id, user id, email, Public Chat visitor hash, cart session, locale, country, region, and city.
  - Filters cover member/anonymous identity, email captured/missing, cart session linked/missing, Public Chat visitor linked/missing, country, locale, and last seen window.

### Not resolved yet

- GeoIP quality. The backend stores coarse country/region headers when present, but there is no dedicated GeoIP provider, consent/audit UI, or enrichment job yet. The next safe step is not precision; it is a clear admin display contract for broad region, source, and unknown-state handling.
- Admin conversation-list identity/tier display. The customer context resolver can expose logged-in customer facts, but the list UI still needs a compact `Visitor`/`Member` badge and member points-tier icon for logged-in customers.

The current `visitor_profiles` source binds signed Public Chat visitor cookie, cart session id, optional captured email, locale, coarse region, and request hashes without guessing. Any future expansion must keep raw IP out of the table unless a privacy/legal decision explicitly changes that boundary.

## Realtime status

HTTP chat remains the durable source of truth. SSE is now the first realtime acceleration layer for customer-service events:

- backend event hub publishes only after HTTP persistence succeeds;
- admin inbox subscribes to `/api/admin/customer-service/events?scope=inbox`;
- customer storefront can subscribe to `/api/v1/customer-service/events?conversation_id={conversation_id}`;
- support users receive only events for conversations they are authorized to see;
- clients keep HTTP read/write and use SSE as the refresh/invalidation signal;
- typing indicators use scoped HTTP `POST` endpoints to publish transient state, then SSE broadcasts `conversation.typing` to the opposite side.

The older WebSocket route `GET /api/v1/customer-service/ws` is still only a guarded ping/pong foundation. It has no conversation scope, no admin namespace, and no persisted chat contract, so it must not be wired to frontend/admin chat as a message or typing source.

## Realtime event contract target

Realtime should be an acceleration layer over the durable HTTP/ticket source, not a second source of truth.

### Subscription routes

- Customer storefront: `GET /api/v1/customer-service/events?conversation_id={conversation_id}`
  - Requires the same owner proof as HTTP: authenticated `user_id` or signed Public Chat visitor cookie hash.
  - The requested `conversation_id` must belong to that customer owner.
- Admin staff: `GET /api/admin/customer-service/events?scope=inbox|conversation&conversation_id={ticket_id}`
  - Requires backoffice access and `ticket:view`.
  - Admin/manager can subscribe to inbox and any customer-service conversation.
  - Support users can subscribe only to conversations assigned to their backend user id.

### Event envelope

Every pushed event should use one envelope so Nuxt and admin do not drift:

```json
{
  "type": "conversation.message.created",
  "event_id": "uuid",
  "ticket_id": 123,
  "conversation_id": "public-conversation-id",
  "occurred_at": "2026-07-25T00:00:00Z",
  "actor": {
    "kind": "customer",
    "user_id": 10,
    "anonymous": false
  },
  "payload": {}
}
```

Planned event types:

- `conversation.message.created`
  - Payload: the same normalized message object returned by HTTP message endpoints, including `message_type` and parsed `metadata` for structured cards such as `config_confirm`.
- `conversation.messages.read`
  - Payload: `reader_kind`, `read_by_user_id`, and `read_at`.
- `conversation.assigned`
  - Payload: `assigned_to`, `assigned_to_name`, `assigned_by_user_id`.
- `conversation.status.changed`
  - Payload: `status` and `display_status`.
- `conversation.context.updated`
  - Payload: minimal invalidation only, not a full customer context snapshot. Clients should refetch `/context`.
- `conversation.typing`
  - Payload: `is_typing`, `display_name`, and `expires_at`.
  - This is transient UI state only. It is not stored in `ticket_messages` and must not trigger HTTP message refreshes.
- `heartbeat`
  - Payload: server timestamp.

### Implementation rules

- Persist first, broadcast second. Durable message/status/assignment realtime events must only be emitted after the Go write succeeds.
- Transient events such as `conversation.typing` are allowed to skip persistence, but must still be scoped by the same customer/admin authorization as HTTP message reads.
- The event payload must be derived from the same response builders used by HTTP handlers.
- Missing realtime must never block chat. Nuxt/admin clients keep HTTP send/read and use reconnect with polling fallback.
- Do not broadcast raw visitor IP, raw user-agent, or hidden profile hashes. Only send display-safe ids and values already exposed by HTTP.
- Do not let clients send arbitrary chat events before the server has a validation path. If WebSocket is expanded later, client-to-server messages must start with scoped `subscribe`/`pong` style control frames only and reuse the same authorization boundaries defined here.

### Implementation order

1. Backend SSE event hub and admin inbox SSE route are active.
2. Broadcasts are active for:
   - customer message send;
   - staff reply;
   - mark-read;
   - transfer assignment;
   - visitor email/customer context invalidation.
3. Admin inbox frontend is wired to SSE and refetches HTTP facts on events.
4. Nuxt customer chat is wired to SSE. It listens for customer-owned conversation events, then refetches `/api/v1/customer-service/messages/:conversation_id` and merges persisted messages into the local room cache.
5. Structured message payload persistence is active for `message_type` + `metadata`; `config_confirm` cards render in both Nuxt customer chat and Admin staff chat after HTTP reload/SSE refresh.
6. Scoped typing indicators are active:
   - Nuxt publishes customer typing through `POST /api/v1/customer-service/typing`;
   - Admin publishes staff typing through `POST /api/admin/customer-service/conversations/:id/typing`;
   - both sides receive `conversation.typing` through the existing SSE event hub and clear the indicator by expiry.
7. Keep WebSocket reserved for a future true bidirectional use case that cannot be handled cleanly by HTTP + SSE.

## Next implementation order

1. Revisit real product/SKU configurable fields after the product/SKU contract is final, then populate `config_confirm.metadata.selections`.
2. Add the admin conversation-list customer summary:
   - `Visitor`/`Member` display label;
   - member points-tier icon for logged-in customers;
   - coarse region display with `Unknown` fallback;
   - no raw IP/User-Agent/hash exposure.
3. Add today's coarse-region operations stats for Public Chat, then connect the aggregate region signal to product-interest facts already present in chat/cart/wishlist/product cards.
4. Add GeoIP provider, consent, audit, or enrichment jobs only after the broad-region contract proves useful and the privacy/legal boundary is explicitly confirmed.
5. If a future feature genuinely requires WebSocket, design a scoped admin + storefront protocol first; do not reuse the legacy `/ws` ping/pong route as-is.
