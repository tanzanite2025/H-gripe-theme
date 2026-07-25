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

`customer_service` records are intentionally not part of the ordinary ticket inbox/list/stat/detail/message path. This is important because anonymous Public Chat conversations still need a persisted `tickets.user_id` for the legacy schema, while the real customer identity is `customer_user_id` or `visitor_session_hash`. Treating those records as ordinary tickets would leak or misclassify customer conversations.

The older generic `/api/v1/chat/messages` path has been retired. The Go route, handler, service, repository, and domain models were removed, and migration `033_drop_legacy_chat_tables.up.sql` drops `chat_messages` / `chat_sessions` so they cannot become a second message source. Do not reintroduce a generic chat message table for Public Chat; use the dedicated customer-service ticket path.

## Customer context boundary

The admin customer-service workspace now has a read-only customer context resolver. Its rule is simple: display existing facts, do not invent missing identity data.

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

- GeoIP quality. The backend stores coarse country/region headers when present, but there is no dedicated GeoIP provider, consent/audit UI, or enrichment job yet.

The current `visitor_profiles` source binds signed Public Chat visitor cookie, cart session id, optional captured email, locale, coarse region, and request hashes without guessing. Any future expansion must keep raw IP out of the table unless a privacy/legal decision explicitly changes that boundary.

## Realtime status

HTTP chat remains the durable source of truth. SSE is now the first realtime acceleration layer for customer-service events:

- backend event hub publishes only after HTTP persistence succeeds;
- admin inbox subscribes to `/api/admin/customer-service/events?scope=inbox`;
- customer storefront can subscribe to `/api/v1/customer-service/events?conversation_id={conversation_id}`;
- support users receive only events for conversations they are authorized to see;
- clients keep HTTP read/write and use SSE only as a refresh/invalidation signal.

The older WebSocket route `GET /api/v1/customer-service/ws` is still only a guarded ping/pong foundation and should not be wired to frontend chat until a bidirectional use case exists.

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
  - Payload: the same normalized message object returned by HTTP message endpoints.
- `conversation.messages.read`
  - Payload: `reader_kind`, `read_by_user_id`, and `read_at`.
- `conversation.assigned`
  - Payload: `assigned_to`, `assigned_to_name`, `assigned_by_user_id`.
- `conversation.status.changed`
  - Payload: `status` and `display_status`.
- `conversation.context.updated`
  - Payload: minimal invalidation only, not a full customer context snapshot. Clients should refetch `/context`.
- `heartbeat`
  - Payload: server timestamp.

### Implementation rules

- Persist first, broadcast second. A WebSocket event must only be emitted after the message/status/assignment is durable in Go.
- The event payload must be derived from the same response builders used by HTTP handlers.
- Missing realtime must never block chat. Nuxt/admin clients keep HTTP send/read and use reconnect with polling fallback.
- Do not broadcast raw visitor IP, raw user-agent, or hidden profile hashes. Only send display-safe ids and values already exposed by HTTP.
- Do not let clients send arbitrary chat events before the server has a validation path. Client-to-server WebSocket messages should initially be limited to `subscribe`, `pong`, and future `typing` once scoped authorization exists.

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
5. Keep WebSocket reserved for a future bidirectional use case such as scoped typing indicators.

## Next implementation order

1. Revisit product/SKU structured configuration messages after the product/SKU contract is final.
2. Add scoped typing indicators only after deciding whether WebSocket is worth the extra bidirectional complexity.
3. Add GeoIP/consent/audit enrichment only after privacy/legal boundaries are explicitly confirmed.
