# 客服会话生命周期与收件箱归档架构

Last reviewed: 2026-09-21  
Status: accepted design; committed lifecycle and realtime scope complete; optional context-update event remains explicitly out of scope

This is the single source of truth for customer-service conversation lifecycle,
per-staff inbox visibility, close/reopen/archive behavior, and deletion/retention
boundaries. It supplements the public-chat boundary document and the realtime
delivery document; those documents must link here instead of redefining these
semantics.

## 1. Problem and decisions

The conversation list is an operational inbox, not a permanent database dump.
If every historical conversation remains in the default list, the inbox becomes
unusable as volume grows. The solution is to separate two facts that are
currently mixed together:

1. **Global conversation lifecycle** — what happened to the customer request.
2. **Recipient inbox visibility** — whether a particular staff member wants the
   conversation in their working queue right now.

The following decisions are binding:

- Closing a conversation is a lifecycle mutation. It does not delete messages.
- Archiving is a per-recipient inbox mutation. It does not change the lifecycle
  or hide the conversation from another assigned staff member.
- The normal UI never hard-deletes a conversation. “Delete from my list” means
  archive.
- Customer messages, staff replies, transfer, close, reopen, archive, and
  restore are authorized server-side and persisted transactionally.
- A new customer message reopens a resolved/closed conversation and removes the
  current assignee's archive marker in the same transaction.
- Permanent deletion is an administrator-only retention operation after legal,
  dispute, after-sales, audit, and attachment checks. It is not part of the
  inbox interaction.

## 2. Current baseline and documentation correction

The current implementation has:

- `tickets.category = customer_service` as the conversation fact source;
- lifecycle values `open`, `in_progress`, `resolved`, and `closed`;
- `tickets.status_version` for optimistic concurrency;
- `customer_service_inbox_states` for per-user read cursor, unread count, and
  assignment version;
- GORM soft-delete fields on tickets and inbox-state rows;
- list filters for status, identity, assignment, search, and unread state.

Phase 1 foundation implemented on 2026-09-17:

- `customer_service_inbox_states.archived_at` and its recipient/list indexes;
- `view=inbox|closed|archived|all`, with `inbox` as the API/admin default;
- per-staff archive and restore HTTP commands;
- durable `conversation.inbox_state.changed` Outbox/realtime invalidation;
- customer-message and transfer paths clear the current/new assignee's archive
  marker;
- an admin row action for archive/restore and separate queue selection.

Phase 2 lifecycle commands implemented on 2026-09-17:

- optimistic `PATCH /conversations/:id/status` with required
  `expected_status_version` and HTTP `409` conflict handling;
- atomic close-and-archive and reopen-and-restore commands;
- durable `conversation.status.changed` events for explicit commands, agent
  replies, same-owner transfers, and customer-message reopening;
- customer-message reopening clears lifecycle timestamps and the assigned
  recipient's archive marker in the message transaction;
- row actions for close/reopen, eight-second server-backed undo, and atomic
  batch archive for up to 100 selected conversations.

Retention is now implemented as a separate administrator maintenance surface,
not as a routine inbox action:

- the eligibility query, administrator-only soft-delete command, and
  administrator-only purge command are implemented;
- the soft-delete command writes `tickets.deleted_at` and starts the 30-day
  tombstone/recovery window;
- the bounded purge scheduler re-checks dependencies before removing
  ticket-owned records and is opt-in (disabled by default);
- the administrator retention workbench provides eligibility preview,
  soft-delete confirmation, recovery-window visibility, and purge confirmation;
- ordinary support users still cannot hard-delete conversations; the backend
  `AdminOnly` middleware remains the final authorization boundary.

Status-change realtime events and lifecycle commands carry only a stable
`reason_code` taxonomy. Omitted or unknown codes are rejected; lifecycle
commands use explicit codes such as `operator_reopen`, `operator_resolve`, and
`operator_close_and_archive`. Free-text `reason` remains reserved for
retention audit requests and is not part of status events.

### Terms that must not be conflated

`visitor_profiles.profile_status = archived` is visitor-profile retention
state. It is independent from a staff member archiving a conversation in
`customer_service_inbox_states.archived_at`. Updating one must not implicitly
update the other.

## 3. Global lifecycle state machine

| State | Meaning | Appears in default inbox? |
| --- | --- | --- |
| `open` | New or reopened request waiting for staff action | Yes, when not archived |
| `in_progress` | Staff is actively handling the request | Yes, when not archived |
| `resolved` | Staff recorded a resolution; customer may still follow up | No; closed view |
| `closed` | Operationally closed; retained for history and audit | No; closed view |

Allowed transitions:

```text
open ---------> in_progress ---------> resolved ---------> closed
  |                 |                     |                  |
  +-----------------+---------------------+------------------+--> closed

resolved/closed -- customer message --> open
resolved/closed -- explicit reopen --> open
open/in_progress/resolved -- explicit close --> closed
```

Rules:

- Sending a staff reply changes `open` or `closed` to `in_progress`; it does
  not silently delete or archive the conversation.
- “Resolve” and “Close” remain separate server states even if the first UI
  release exposes one combined **关闭并归档** action.
- Reopening is idempotent. Repeating a reopen on `open` creates no additional
  status event.
- Every changed status increments `tickets.status_version` and checks the
  caller's expected version. A stale command returns HTTP `409` and the client
  reconciles from HTTP.
- Status is global. One staff member cannot make a conversation `closed` only
  for themselves; personal visibility belongs to the inbox state below.

## 4. Per-recipient inbox state

Extend `customer_service_inbox_states` with:

```sql
archived_at TIMESTAMPTZ NULL
```

The row is keyed by `(ticket_id, recipient_user_id)`. Its responsibilities are:

- `last_read_message_id`, `unread_count`, `last_read_at`;
- `assignment_version`;
- `archived_at` for that recipient's personal queue.

The following invariants apply:

1. `archived_at IS NULL` means visible in that recipient's active queue.
2. `archived_at IS NOT NULL` means hidden from that recipient's default queue,
   but still searchable and recoverable.
3. Archiving does not change assignment, status, messages, or another user's
   inbox state.
4. A transfer creates/resets the new assignee's state with `archived_at = NULL`.
   The previous recipient's historical archive marker is not copied to the new
   assignee.
5. A customer message clears `archived_at` for the current assigned recipient,
   increments/reconciles unread state, and re-enters that recipient's active
   queue.
6. Admin/manager users may archive a visible conversation for their own inbox
   state. Their all-conversation access does not turn a personal archive into a
   global delete.
7. If a recipient has no state row yet, it is treated as active for listing;
   archive/restore creates the row transactionally.

## 5. List and filter contract

The list API must receive an explicit `view` parameter. The legacy absence of a
view is normalized to `inbox`, not `all`.

| `view` | Definition |
| --- | --- |
| `inbox` | Current recipient's `archived_at IS NULL` and lifecycle `open`/`in_progress` |
| `closed` | Lifecycle `resolved`/`closed`; archived state does not remove history |
| `archived` | Current recipient's `archived_at IS NOT NULL`, any lifecycle state |
| `all` | Explicit operator view of all authorized lifecycle and inbox states |

`status`, `identity`, `assigned_to`, `unread`, and `search` remain additional
filters. When `view` and `status` overlap, the intersection is returned. The
server, not the current frontend page, applies all filters and pagination.

Default ordering is:

1. unread customer messages first;
2. `tickets.updated_at DESC`;
3. `tickets.id DESC` as a stable tie breaker.

The response should include per-row `status`, `status_version`,
`inbox_archived`, `archived_at`, `unread_count`, and capability flags such as
`can_archive`, `can_restore`, `can_close`, and `can_reopen`.

## 6. Staff interaction model

Each row has a `···` action menu. The first release exposes:

- **关闭并归档** — transition to `closed` and archive the current recipient;
- **仅归档** — remove from this recipient's active queue without changing
  lifecycle;
- **恢复到收件箱** — clear `archived_at`;
- **重新打开** — transition `resolved/closed` to `open` and clear the current
  recipient's archive marker;
- **标记已读/未读** — operate on the existing per-recipient cursor.

Batch actions:

- batch archive only for selected rows in the current recipient's scope;
- batch close is deliberately deferred until an explicit confirmation flow and
  an audit reason are available.

Every archive/close/reopen action shows a short undo toast. Undo calls the
opposite idempotent command with the server version; it is not a client-only
array mutation. If a new customer message arrives before undo, the server wins
and the UI reconciles.

There is no routine **删除** button. Product copy should explain that archive
removes clutter while preserving the conversation for history, support, and
audit.

## 7. HTTP command contract

The lifecycle routes and the separate retention maintenance routes are:

```text
GET   /api/admin/customer-service/conversations
      ?view=inbox|closed|archived|all&status=...&identity=...&unread=...

PATCH /api/admin/customer-service/conversations/:id/status
      {"status":"resolved|closed|open|in_progress",
       "expected_status_version": 3,
       "reason_code":"operator_resolve|operator_close_and_archive|operator_reopen|manual_status_change"}

POST  /api/admin/customer-service/conversations/:id/archive
POST  /api/admin/customer-service/conversations/:id/restore

POST  /api/admin/customer-service/conversations/bulk-archive
      {"conversation_ids":[1,2,3]}

GET   /api/admin/customer-service/conversations/retention/eligibility
      ?conversation_ids=1,2

POST  /api/admin/customer-service/conversations/retention/soft-delete
      {"conversation_ids":[1,2],"reason":"retention policy approved"}

POST  /api/admin/customer-service/conversations/retention/purge
      {"conversation_ids":[1,2],"reason":"retention policy approved"}
```

Archive/restore commands must accept an optional expected `assignment_version`
when the caller has a state row. All commands return the refreshed conversation
projection. A stale lifecycle or assignment version returns `409`; `404` does
not reveal conversations outside the caller's authorization scope.

The retention eligibility query accepts either `conversation_ids` or the
ticket-owned maintenance alias `ticket_ids` as a comma-separated query value.
Mutation bodies accept either array under the same names, require a non-empty
`reason` (maximum 500 characters), deduplicate IDs, and cap each request at 100
conversations. The response contains an `eligibility` array with the ticket
ID, eligibility/action, lifecycle status, last activity, and soft-delete/purge
timestamps when applicable. All three retention routes are Admin-only; purge
also requires the record to have passed the dependency checks and the full
recovery window.

## 8. Realtime and Outbox contract

Durable mutations are committed before local fan-out. The existing realtime
document remains authoritative for transport, while this document defines the
business meaning of the events.

Required backoffice events:

- `conversation.created` — emitted transactionally when a customer-service
  conversation is first created; deterministic key is
  `customer_service.conversation.created:<ticket_id>`;
- `conversation.message.created` — canonical payload remains `{message_id}`;
- `conversation.status.changed` — `previous_status`, `status`, and
  `status_version`;
- `conversation.assigned` — assignment and new assignment version;
- `conversation.inbox_state.changed` — `recipient_user_id`, `archived`,
  `archived_at`, and `assignment_version`;
- `conversation.messages.read` — existing read-cursor event.

When a customer message reopens a closed/resolved conversation, the same
database transaction may create message, status, and inbox-state events. Each
event has a deterministic id derived from the mutation id/version. The browser
deduplicates by `event_id`, then refetches the authorized list and selected
conversation. Event payloads never contain raw visitor hashes, IP addresses,
message bodies, or full customer context.

Archive/restore producers and the complete status producer are implemented.
Future lifecycle write paths must use the same transaction and deterministic
event rules rather than introducing direct status updates without Outbox facts.

`conversation.context.updated` is intentionally not a required event yet. The
admin context endpoint is read-only, while cart, wishlist, browsing, order,
and visitor-profile writes are owned by separate bounded contexts and do not
carry a customer-service conversation mutation boundary. Reads, refreshes, or
unrelated customer activity must not fabricate this event. It may be promoted
to a required event only when a concrete transactional context command defines
its actor, authorization, payload projection, deterministic identity, and
rollback tests.

## 9. Authorization and audit

| Operation | Support user | Admin/manager | Customer/storefront |
| --- | --- | --- | --- |
| Read assigned conversation | Yes | Yes | Own conversation only |
| Reply / mark read | Yes | Yes | Customer message only |
| Transfer | Existing permission and scope | Yes | No |
| Archive/restore own inbox state | Yes | Yes | No |
| Change global status | Existing ticket-edit permission | Yes | No |
| Permanent purge | No | Admin only | No |

Status, archive, restore, and bulk archive commands retain their existing audit
contract. Retention soft-delete and purge commands write an audit entry with
the actor (or `system` for the scheduler), ticket id, action/path, reason,
projection scope, and success/failure status. The current `audit_logs` schema
does not persist a `request_id`; correlation must be added as a separate
follow-up if operations require it. Audit records must not copy message bodies
or raw visitor fingerprints.

## 10. Retention and permanent deletion

Conversation history is a support and audit record, not disposable UI state.
The retention policy is owned by the customer-service conversation domain and is
configured from the customer-service retention workbench. It is not a generic
system setting:

The cleanup layers must stay distinct:

- **Archive** only changes a staff member's inbox visibility. It does not delete
  the conversation, messages, attachments, or audit evidence.
- **Soft-delete** marks an eligible ticket with `tickets.deleted_at` and starts
  the recovery/tombstone window. It removes the record from normal application
  reads but does not physically remove ticket-owned data yet.
- **Purge** is the administrator-only physical cleanup after the retention,
  dependency, and recovery-window checks pass. It removes ticket-owned database
  rows and application-owned, unshared attachment objects. It does not cascade
  into orders, visitor profiles, audit logs, or shared objects.
- **Backups** are a disaster-recovery copy, not an online restore mechanism for
  the inbox. An application purge does not rewrite existing database or object
  storage backups; backup retention and any legal-erasure treatment of expired
  backups are separate infrastructure policies.
- **External search-index cleanup** is optional. It applies only when a separate
  search projection for customer-service conversations is actually deployed. If
  no such projection exists, the endpoint remains empty and retention does not
  depend on it.

- deployment variables such as `WORKER_CUSTOMER_SERVICE_RETENTION_ENABLED`,
  `...INTERVAL_SECONDS`, `...BATCH_LIMIT`, `...MINIMUM_DAYS`, and
  `...RECOVERY_WINDOW_DAYS` provide initial defaults only;
  `...INTERVAL_SECONDS` and `...BATCH_LIMIT` bound each worker pass;
  `...MINIMUM_MONTHS` defaults to `24`, while a positive
  `...MINIMUM_DAYS` overrides the month value when an exact day policy is
  required; `...RECOVERY_WINDOW_DAYS` defaults to `30`;

- closed/resolved conversations remain queryable for at least 24 months after
  `closed_at` (or the latest customer message, whichever is later);
- legal hold, payment dispute, after-sales case, order evidence, or audit link
  suspends eligibility;
- eligible records are first soft-deleted through `tickets.deleted_at` and kept
  in a 30-day recovery/tombstone window. In this design, the window is a purge
  deferral and investigation period; it does not expose a retention-specific
  restore API. If an investigation requires recovery, the operator must stop
  the purge and use a separately approved database maintenance procedure. The
  normal inbox archive/restore commands remain unrelated to this tombstone;
- a later background purge removes `ticket_messages`,
  `customer_service_inbox_states`, and the soft-deleted ticket row after the
  dependency checks and recovery window are re-evaluated;
- before the ticket-owned rows are removed, message attachment references are
  inspected. Application-owned, unshared objects are deleted, shared objects
  are retained, and a configured CDN purger is invoked. No cascade is performed
  into orders, visitor profiles, audit logs, or unrelated objects;
- when a separate search projection is deployed, an optional HTTP adapter may
  also remove the conversation from that projection. It is configured with
  `CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL` and the optional
  `CUSTOMER_SERVICE_SEARCH_INDEX_TOKEN`; both remain empty when no external
  search index exists. This integration is not a prerequisite for ordinary
  archive, soft-delete, or database purge;
- when the shared SQL Outbox is configured, the purge transaction also records
  a durable `customer_service.retention_cleanup_requested` event containing the
  ticket id and captured attachment references. The post-commit handler retries
  attachment, CDN, and search-index cleanup independently of the API process;
- the monolithic service wires the shared SQL Outbox for the production path;
  attachment cleanup reads references only through
  `FindCustomerServiceConversationAttachmentReferenceScan`, and the retention
  contract has no legacy repository compatibility API;
- the bounded worker is opt-in and defaults to disabled. Its live policy is
  stored in `customer_service_retention_policies` and can be changed by an
  administrator in the客服 retention workbench without restarting the API;
- visitor profile retention follows `VISITOR_PROFILE_RETENTION_RISK_DESIGN.md`
  and is not shortened merely because a conversation was archived.

The inbox release implements archive/restore, closed views, explicit lifecycle
commands, and batch archive. Physical deletion is available only through the
administrator retention command or the explicitly enabled bounded worker.

## 11. Implementation sequence

### Phase 0 — documentation and canonical contract

- Keep this document as the lifecycle source of truth.
- Add links from project, storefront, public-chat, realtime, and workbench docs.
- Use only the canonical lifecycle values `open`, `in_progress`, `resolved`,
  and `closed` at the API boundary. This system is not yet in production, so
  legacy aliases (`pending`, `active`) are intentionally removed rather than
  preserved as a second status path.

### Phase 1 — inbox archive foundation

- Add migration `303_customer_service_inbox_archive` (or the next unused
  migration number at implementation time) with nullable `archived_at` and
  recipient/list indexes.
- Add domain/repository/service support for archive, restore, and view-aware
  listing.
- Default the admin client to `view=inbox` and add Inbox / Closed / Archived /
  All tabs.
- Add row menu and server-backed undo toast for archive/restore.
- Emit/reconcile `conversation.inbox_state.changed`.

### Phase 2 — explicit lifecycle commands

- Implemented: status endpoint with optimistic `status_version` checks.
- Implemented: close, resolve, reopen, stale-version, rollback, and idempotent
  transition tests.
- Implemented: customer-message reopening clears the current assignee's archive
  marker and emits status/inbox invalidations.
- Implemented: bounded atomic batch archive with server-backed undo.

### Phase 3 — retention operations

- Implemented: admin-only eligibility query, soft-delete command, dependency
  holds for active after-sales/payment disputes/order evidence, 30-day recovery
  window, and ticket-owned audited purge command.
- Implemented: retention mutation batches pre-validate every explicit ticket ID
  before the first soft-delete or purge write, so an ineligible ID rejects the
  command without a known partial batch.
- Implemented: scheduled bounded purge worker,客服对话域运行时配置的 minimum
  age/recovery window/batch settings, and Prometheus counters for archived,
  reopened, soft-deleted, purged, blocked, and failed retention operations.
- Implemented: cleanup of parseable message attachment references with shared
  object protection, optional CDN invalidation, and a provider-neutral HTTP
  search-index deletion adapter. Without a configured endpoint, database and
  configured object-storage cleanup still proceed.
- Implemented: ticket-owned purge deletion and its successful retention audit
  row commit in one SQL transaction, with rollback coverage when the audit
  insert fails.
- Implemented: durable post-commit retention cleanup Outbox events carrying
  attachment references that would otherwise disappear with ticket messages;
  the shared Outbox dispatcher invokes the retryable cleanup handler.
- Implemented: malformed attachment payload and shared-reference skip counters
  for retention cleanup operations.
- Implemented: aggregate retention cleanup Outbox gauges and Prometheus alerts
  for failed/dead-letter cleanup and search-index deletion errors.
- The worker only purges already soft-deleted conversations after the recovery
  window and re-runs dependency checks; ordinary staff actions remain archive
  or close, never hard delete. Physical purge is disabled by default and must
  be explicitly enabled by deployment configuration.

## 12. Acceptance matrix

| Scenario | Required result |
| --- | --- |
| New message arrives | Conversation is in assigned recipient's `inbox`, unread, not archived |
| Staff archives active conversation | It disappears from that staff member's inbox only; messages remain |
| Staff restores archived conversation | It returns to the inbox with original lifecycle and read cursor |
| Staff closes conversation | It leaves default inbox and appears in `closed`; no message deletion |
| Closed customer replies | Status becomes `open`, assigned recipient is unarchived, unread increments, realtime invalidation is delivered |
| Transfer archived conversation | New assignee receives an unarchived state; old assignee does not gain ownership |
| Two staff archive independently | Each user's inbox state is independent |
| Stale close/archive command | HTTP `409`; client refetches and does not overwrite newer state |
| Search archived history | Explicit `view=archived` or `view=all` finds it within authorization scope |
| Visitor profile is archived | Conversation inbox visibility is unchanged unless a separate inbox command is made |
| Routine delete click | UI offers archive/close; no hard delete occurs |
| Retention purge attempt | Rejected unless admin, eligible, unheld, audited, and past soft-delete window |

## 13. Current limitations and next steps

The admin workbench now provides an administrator-only retention page for
eligibility preview, soft-delete confirmation, recovery-window state, and purge
results. The recovery-window state is intentionally informational and blocks
purge; the workbench does not offer a retention restore command. The backend
remains the final authorization boundary, and ordinary support users still have
no hard-delete control. The provider-neutral HTTP search-index adapter is
implemented but no endpoint is configured by default.
Attachment cleanup currently
parses historical JSON-array payloads; malformed legacy payloads are skipped,
counted by the `customer_service_retention_attachment_reference_skips_total`
metric, and require an operator report before their referenced objects can be
considered for cleanup. Ticket-owned purge
deletion and the success audit row now commit in one SQL transaction; external
attachment/CDN/search cleanup is represented by a durable post-commit Outbox
event when configured and must stay idempotent. Failure audit records are still
written separately when a purge cannot complete. Production must enable the
shared Outbox dispatcher for the durable path to execute.

Recommended next steps, in order:

1. Run the bounded staging drill from the retention runbook: soft-delete a
   fixture, simulate the 30-day window, purge it, retry the Outbox event, and
   verify shared-attachment protection, malformed-payload metrics, and audit
   results. If an external search projection is deployed, include its idempotent
   delete and retry checks in the same drill; otherwise this step is skipped.
2. After the drill is accepted, enable the worker only through the monolith's
   deployment environment, add the existing retention PromQL rules to the
   private monitoring system, and record the backup/rollback approval before
   production use.
3. Only when a concrete external search projection is selected, document its
   delete contract and configure `CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL` and
   its token in staging before production use.

Focused service/repository/scheduler tests now cover the recovery window,
non-customer-service IDs, ticket/message/inbox cleanup, audit retention,
attachment reference deduplication, malformed attachment payload handling,
transient search-index delete retry, and an expired-tombstone worker pass.
