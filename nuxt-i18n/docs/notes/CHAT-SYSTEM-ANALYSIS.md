# Chat system current source

Last audited: 2026-07-24

This document describes the current Tanzanite storefront chat architecture. It replaces the old 2025 analysis that treated the chat window as one large component and described WordPress as the active backend.

## Current boundary

The live chat path is:

- Nuxt storefront for visitor and agent-facing UI.
- Go backend `/api/v1/customer-service/*` for agents, conversations, messages, auto-replies, and WebSocket handshake.
- Optional authenticated user identity or a visitor session hash for conversation ownership.
- Browser `localStorage` for the short-lived UI/message cache; the Go ticket store remains the durable conversation source.

There is no active WordPress chat API in the current Nuxt flow.

## Nuxt ownership

The responsibilities are intentionally split:

- `app/components/WhatsAppChatModal.vue`
  - Teleport shell, mode switching, drawer mounting, and event wiring only.
  - Do not add chat state or API calls here.
- `app/composables/chat/useWhatsAppState.ts`
  - Chat state orchestration for visitor mode and agent mode.
  - Conversation creation, message sending, auto-reply calls, product/order sharing, agent status, transfer, and local persistence coordination.
- `app/components/whatsapp/ChatWelcomePanel.vue`
  - Visitor welcome screen and customer-service selection.
- `app/components/whatsapp/AgentChatPanel.vue`
  - Agent conversation list, selected conversation, status control, message list, and reply input.
- `app/components/whatsapp/UserChatBody.vue`
  - Visitor tabs and message body.
- `app/components/whatsapp/ChatTransferModal.vue`
  - Agent transfer form.
- `app/composables/chat/useChatAgentDirectory.ts`
  - Public agent directory request, 30-minute browser cache, current-user filtering, and DEV-only fallback directory.
- `app/composables/chat/useChatStorage.ts`
  - Per-conversation/per-agent browser state with a five-day message expiry.
- `app/composables/useAuth.ts`
  - `/auth/profile` session lookup and `is_agent` / `agent_id` identity fields.

## Visitor flow

1. `useWhatsAppState` loads the public agent directory.
2. `useChatAgentDirectory` removes an agent whose `user_id` matches the current authenticated user.
3. The visitor selects an agent and enters the chat.
4. `POST /customer-service/conversations` creates or reuses a conversation using the authenticated user or visitor session owner.
5. `POST /customer-service/messages` persists visitor messages.
6. `GET /customer-service/messages/:conversation_id` reads durable messages when needed.
7. Welcome and keyword auto-reply endpoints may append system/agent messages.
8. The UI also stores the current room locally for fast reopen and five-day expiry cleanup.

The visitor flow must never trust a conversation ID by itself; the Go service checks the owner before returning or writing messages.

## Agent flow

1. `/auth/profile` returns `is_agent` and `agent_id` when the authenticated user is linked to customer service.
2. `agentMode` is derived from `useAuth().isAgent`.
3. Agent requests use `/customer-service/agent/*`.
4. The router requires authentication and one of `admin`, `manager`, or `support`.
5. Agents can list conversations, open messages, send staff replies, mark messages read, transfer a conversation, and update status.

Current agent endpoints:

| Capability | Endpoint |
| --- | --- |
| List agents for visitor selection | `GET /customer-service/agents` |
| List agent conversations | `GET /customer-service/agent/conversations` |
| Read conversation messages | `GET /customer-service/agent/conversations/:id/messages` |
| Send staff message | `POST /customer-service/agent/messages` |
| Mark messages read | `POST /customer-service/agent/messages/read` |
| Transfer conversation | `POST /customer-service/agent/conversations/:id/transfer` |
| Read/update status | `GET/POST /customer-service/agent/status` |

## Go ownership

- `internal/service/ticket_customer_service.go`
  - Conversation owner normalization, visitor/user access checks, conversation creation, public message persistence, and agent directory lookup.
- `internal/api/v1/ticket/customer_service_agent_handler.go`
  - Agent conversation/message/status/transfer HTTP handlers.
- `internal/api/v1/ticket/hub.go`
  - WebSocket upgrade, origin validation, connection limits, ping/pong, and connection lifecycle.
- `internal/api/v1/router.go`
  - Public and agent route registration and middleware boundary.

The Go service is the durable source of truth for conversation ownership and messages. Frontend localStorage is a cache, not an authoritative message store.

## Realtime status

The backend exposes `GET /customer-service/ws` and has a guarded WebSocket hub. The latest audit confirms the route currently upgrades authorized visitor/user sessions, enforces origin checks and connection limits, and keeps the connection alive with ping/pong.

That hub is not yet a full chat transport:

- it does not define a conversation-scoped event payload contract;
- it does not subscribe clients to a specific customer-service conversation;
- it does not broadcast persisted visitor or agent messages after `POST /customer-service/messages` or `POST /customer-service/agent/messages`;
- it does not send read-status, transfer, status-change, or typing events.

The current Nuxt code does not create a browser `WebSocket` or `EventSource` connection; message writes and reads still rely on HTTP calls and local cache hydration.

Therefore the system should currently be described as:

- durable HTTP chat: active;
- WebSocket route: backend keepalive foundation only;
- end-to-end realtime push: not complete.

Do not wire the Nuxt client directly to `/customer-service/ws` until the backend first defines and implements conversation-scoped events, authorization rules per conversation, message persistence before broadcast, and replay/read fallback through HTTP. After that, the Nuxt client must add reconnect/backoff handling and keep HTTP as the fallback.

## Config confirmation status

The first configuration-confirmation shell already exists in `CHAT_CONFIG_CONFIRM_DESIGN.md` inside the product search drawer. It intentionally uses placeholder product data.

Phase 2 remains open and must wait for the product/SKU data contract:

1. Define the canonical selected variant and option payload.
2. Render real SKU-driven fields in the confirmation view.
3. Send a `config_confirm` message through the existing customer-service message API.
4. Render the same structured card in visitor and agent message views.
5. Add an Orders entry point only after order configuration history is available.

Do not invent generic fields such as `size` or `metal` in the chat component. The payload must come from the product/SKU fact source.

## Current risk boundaries

- Keep visitor and agent modes in separate presentation components.
- Keep API URLs and response conversion in composables/data modules, not in `WhatsAppChatModal.vue`.
- Keep conversation ownership checks in Go; client filtering is only a display safeguard.
- Keep DEV fallback agents behind `import.meta.dev`; production must use the backend directory.
- Treat localStorage messages as recoverable cache data and tolerate expiry or parse failure.
- Do not call the backend WebSocket “realtime” until the Nuxt client consumes it end to end.

## Next implementation order

1. Finalize the product/SKU configuration contract.
2. Implement `config_confirm` end to end.
3. Add end-to-end tests for visitor ownership, agent authorization, transfer, and structured configuration messages.
4. Define and implement the customer-service WebSocket event contract and backend broadcasts.
5. Implement the Nuxt WebSocket client with reconnect/backoff and HTTP fallback.

Whenever any chat route, ownership rule, message type, or component responsibility changes, update this document and `CHAT_CONFIG_CONFIRM_DESIGN.md` together.
