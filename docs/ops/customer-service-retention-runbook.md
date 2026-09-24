# Customer-Service Retention Runbook

This runbook covers the administrator-only conversation retention commands and
the optional bounded purge worker described in
`docs/design/customer-service-conversation-lifecycle-and-inbox-architecture.md`.
It assumes the shared SQL Outbox dispatcher is available in the deployment.

## Guardrails

- Routine support users never hard-delete conversations. Use the admin
  retention workbench or the explicit admin API only.
- Start with one or two ticket IDs and verify the eligibility response before
  processing a larger batch. Each request is capped at 100 IDs.
- A soft-delete writes `tickets.deleted_at` and creates a 30-day
  tombstone/purge-deferral window. This is not an application restore command;
  keep the ticket out of purge while an investigation or separately approved
  database recovery is pending. Purge is rejected until that window and all
  dependency checks pass.
- Purge removes ticket-owned rows and keeps audit rows. It does not cascade into
  orders, visitor profiles, or unrelated media references.
- Keep a recent database backup and confirm object-storage recovery before any
  production purge. Backups are a disaster-recovery safety net; they are not a
  substitute for online retention cleanup and are not rewritten by the purge.
  A hard purge is not undoable through the application UI.

## Preflight

1. Confirm the operator has the `admin` role and the `system:manage`
   permission. Verify the API and Outbox dispatcher are healthy.
2. Confirm the retention worker remains disabled unless a separately approved
   change enables it. The deployment variable is only the initial default;
   the live value is managed in the客服 retention workbench:

   ```text
   WORKER_CUSTOMER_SERVICE_RETENTION_ENABLED=false
   ```

3. Check the Outbox backlog by event type. Retention cleanup events must not be
   accumulating in `pending`, `failed`, or `dead_letter` status before a purge
   batch begins.
4. If (and only if) a separate customer-service search projection is deployed,
   configure both variables and verify the endpoint with a staging ticket
   before production use:

   ```text
   CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL=https://search.example.internal/conversations/{ticket_id}
   CUSTOMER_SERVICE_SEARCH_INDEX_TOKEN=<secret>
   ```

   The adapter sends an authenticated idempotent `DELETE`; a collection URL
   without `{ticket_id}` receives the ticket ID as its final path segment. If no
   external projection exists, leave both variables empty and skip this step.
5. Record the ticket IDs, policy approval reference, operator, and reason in
   the maintenance ticket.

## Dry-run and execution

Use the eligibility endpoint first. It does not mutate data:

```text
GET /api/admin/customer-service/conversations/retention/eligibility?conversation_ids=101,102
```

For a staging drill, use the authenticated admin session/token already used by
the workbench. The same request sequence can be run with `curl`:

```bash
BASE_URL=http://127.0.0.1:8080
AUTH='Authorization: Bearer <admin-token>'
TICKET_ID=101

curl -fsS -H "$AUTH" \
  "$BASE_URL/api/admin/customer-service/conversations/retention/eligibility?conversation_ids=$TICKET_ID"

curl -fsS -X POST -H "$AUTH" -H 'Content-Type: application/json' \
  "$BASE_URL/api/admin/customer-service/conversations/retention/soft-delete" \
  -d "{\"conversation_ids\":[$TICKET_ID],\"reason\":\"staging retention drill\"}"

# Run only after the fixture's recovery window is simulated as elapsed.
curl -fsS -X POST -H "$AUTH" -H 'Content-Type: application/json' \
  "$BASE_URL/api/admin/customer-service/conversations/retention/purge" \
  -d "{\"conversation_ids\":[$TICKET_ID],\"reason\":\"staging retention drill purge\"}"
```

Do not use these commands against production without the approval and backup
steps above. The API requires the admin role and rejects a purge while the
recovery window or any dependency hold is still active.

Expected actions are:

- `soft_delete`: closed/resolved, old enough, and free of dependency holds;
- `soft_deleted`: inside the 30-day recovery window; the workbench exposes the
  state but does not provide a retention restore action;
- `purge`: recovery window elapsed and dependency checks still clear;
- `ineligible`: missing, active, too recent, or held by an external record.

For a soft-delete or purge, send only the IDs with the matching action and a
non-empty reason. Re-run eligibility after each small batch and retain the API
response with the maintenance ticket. The command pre-validates the entire ID
list before its first write; mixed eligible/ineligible input is rejected as one
operation.

## Monitoring

Watch these Prometheus metrics during and after a batch:

- `customer_service_soft_deleted_conversations_total`
- `customer_service_purged_conversations_total`
- `customer_service_retention_eligibility_total{result="blocked|error"}`
- `customer_service_retention_attachment_reference_skips_total{reason="malformed_payload|shared_reference"}`
- `customer_service_retention_search_index_deletes_total{result="deleted|already_deleted|error"}` (only when an external search projection is configured)

For Outbox cleanup, monitor counts for
`customer_service.retention_cleanup_requested` by `pending`, `failed`,
`dead_letter`, and `processed`. A failed cleanup event should be retried after
fixing the relevant attachment, CDN, or optional search dependency; do not
manually delete the Outbox row.

The dispatcher claims both `pending` and retryable `failed` events. A handler
error returns the event to `failed` with a later `available_at`; after the
configured maximum attempts it becomes `dead_letter`. Verify the staging event
with a read-only SQL query (adapt the table name to the deployed migration):

```sql
SELECT id, event_key, status, attempts, available_at, last_error
FROM outbox_events
WHERE event_type = 'customer_service.retention_cleanup_requested'
ORDER BY id DESC
LIMIT 5;
```

The expected sequence for a forced external dependency failure is
`pending -> processing -> failed -> processed` after the dependency is fixed
and the next dispatcher pass runs. A `dead_letter` event requires on-call
review and must not be deleted manually.

For the single-service deployment, add these expressions to the existing
private Prometheus/Alertmanager rule set:

```promql
commerce_platform_customer_service_retention_cleanup_outbox_events{status="dead_letter"} > 0
commerce_platform_customer_service_retention_cleanup_outbox_events{status="failed"} > 0
increase(commerce_platform_customer_service_retention_search_index_deletes_total{result="error"}[10m]) > 0
```

Treat dead-letter as critical and failed cleanup as warning after 10 minutes.
Treat search-index errors as a warning after 1 minute only when an external
search projection is configured. Keep `/metrics` private and route the critical
dead-letter alert to the normal on-call receiver.

## Stop and recover

- If eligibility or dependency checks look wrong, stop the batch and leave the
  worker disabled. Soft-deleted rows remain retained during the 30-day window;
  investigate before issuing purge because the application does not provide a
  retention restore command.
- If the Outbox dispatcher is unhealthy, stop new purge commands. Existing
  ticket-owned deletion transactions may have committed with pending cleanup
  events; restore the dispatcher before considering the operation complete.
- If attachment cleanup, CDN invalidation, or an optional search cleanup
  repeatedly fails, leave the event in retryable status, restore the relevant
  dependency, and replay through the normal Outbox worker. Escalate
  `dead_letter` events with their event key and ticket ID.
- For an accidental physical purge, recover ticket-owned data from the database
  backup and restore application-owned objects from object-storage versioning or
  backup. Reconciliation must be performed as a separate approved maintenance
  operation; the application does not provide an undo command for hard purge.

## Rollout drill

Before enabling the worker in production, run the same flow in staging with a
fixture conversation containing a shared attachment, a malformed attachment
payload, and an active dependency hold. Verify that the shared object remains,
the malformed-payload metric increments, the held ticket is not purged, and the
Outbox cleanup event reaches `processed` after the dispatcher retries. If an
external search projection is configured, also force one temporary search
failure and verify the retry sequence before production approval.
