# Backend Deployment Notes

These notes describe deployment preparation for the Go backend. They are not a production-readiness certification.

## Required Services

- PostgreSQL
- Redis
- Object storage if upload/storage features are enabled
- Payment provider credentials if payment flows are enabled

## Configuration

Start from the examples:

```powershell
cd go-backend
Copy-Item .env.example .env -ErrorAction SilentlyContinue
Copy-Item config/config.example.yaml config/config.yaml -ErrorAction SilentlyContinue
```

Before deploying, set real values for:

- Database host, username, password, and database name
- Redis host and credentials
- JWT and cookie secrets
- CORS origins
- Upload/storage provider settings
- Payment provider settings
- Log level and server mode

Do not commit real secrets.

## Admin Account Recovery

Production admin passwords are not recoverable because `users.password` stores bcrypt hashes only. Use the audited `adminctl` command to create or reset a backoffice account with a newly generated password.

See `docs/ADMIN_ACCOUNT_RECOVERY.md` for the runbook.

## Build

```powershell
cd go-backend
go test ./...
go build ./cmd/server
go build ./cmd/adminctl
```

<a id="schwalbe-migrations-357-370-preflight"></a>
<a id="schwalbe-migrations-357-361-preflight"></a>
<a id="schwalbe-migrations-357-360-preflight"></a>
<a id="schwalbe-migrations-357-359-preflight"></a>
<a id="schwalbe-migrations-357-358-preflight"></a>

## Schwalbe Catalog Migrations 357-370 Preflight

Migration 357 may remove only an empty legacy catalog; migration 358 recreates the standalone candidate catalog for the all-spectrum selector and template autofill; migration 359 seeds the official Schwalbe sitemap snapshot; migration 360 adds the database uniqueness boundary for Article No. on sales products; migration 361 creates the route-bound, backend-editable FAQ page for the Schwalbe selector and seeds the English/Chinese bead-label explanations. Migration 362 retires the old Schwalbe tire-width/inner-rim-width guidance table, migration 369 adds reviewed Hookless facts, and migration 370 removes any stale table left by the superseded 362 migration. The selector has no broad tire-width compatibility endpoint. Inspect each environment before applying 357. First check whether the table exists:

```sql
SELECT to_regclass('public.schwalbe_tire_specifications') AS legacy_table;
```

If it exists, check its row count:

```sql
SELECT count(*) FROM public.schwalbe_tire_specifications;
```

Migration 357 stops if the table contains any rows; it does not clear or convert them. If the count is nonzero, stop the deployment and preserve/inspect those catalog rows before deciding how to proceed. Do not map candidate rows to sales Products merely to satisfy the migration. After 357 succeeds, migration 358 recreates the candidate table without review or approval fields. Migration 359 applies the checked-in official snapshot as an upsert seed and must run after 358. Migration 360 fails closed if existing sales Products already contain duplicate Article No. values; resolve those duplicates before retrying and do not delete catalog or product data to bypass the check. Allow the migration runner enough statement and lock timeout for the seed transaction.

In release mode keep `DB_AUTO_MIGRATE=false`, run the migration command before starting or rolling the API, then deploy the API after the migration succeeds. From `go-backend`, run `go run ./cmd/server migrate`; it applies all pending migrations rather than only one numbered migration. After the migration and API rollout, the API startup sync and the Admin FAQ “同步前台路由” action reconcile the current Nuxt route manifest. Purge storefront HTML cache once after the rollout if old cached pages are still serving.

After the migration completes, verify:

```sql
SELECT version, dirty FROM schema_migrations;
SELECT count(*) FROM public.schwalbe_tire_specifications;
SELECT indexname FROM pg_indexes WHERE tablename = 'product_spec_values' AND indexname = 'uq_schwalbe_product_article_no';
SELECT page_id, route_path, locale, status FROM faq_pages WHERE page_id = 'guides-schwalbe-tire-selector';
SELECT locale, count(*) FROM faqs WHERE page_id = 'guides-schwalbe-tire-selector' AND status = 'published' GROUP BY locale;
SELECT to_regclass('public.schwalbe_tire_hookless_compatibility') AS hookless_compatibility_table;
SELECT to_regclass('public.schwalbe_tire_rim_width_combination_rules') AS retired_combination_table;
```

The selector has no broad Schwalbe combination endpoint. The calculator and
selector cards use the same versioned Go engineering model in
`internal/domain/tirerim`. The tire/rim API exposes only the `/solve` endpoint;
the DT Swiss comparison table is static front-end reference data and does not
participate in engineering results.

## Tire and Rim Width Guide FAQ and Route Migrations 407-409

Migration 407 seeds editable English and Chinese FAQ content. Migration 408
renames the page and its FAQ associations to
`guides-tireguides-tire-rim-width-compatibility` and moves the existing sitemap
catalog rows to `/guides/tireguides/tire-rim-width-compatibility`. It updates
existing records without replacing administrator questions, answers, images,
ordering or publication state. Run the pending migrations before starting the
API with the updated Nuxt route manifest.

Migration 409 adds three English and three Chinese long-tail answers to the
same editable page: 28 mm on a 25 mm hookless rim, the current 28C recommendation,
and how inner width can affect measured tire width. Existing answers, including
the 105% calculation explanation, remain unchanged. PageFaq generates its
FAQPage JSON-LD from these same visible, published questions.

This guide has not launched. Its old `/guides/tireguides/choose` URL is removed
and returns 404; no redirect or alias is registered. Use the new route and page
identity in Admin. After migration, check that each of the English and Chinese
pages has its twelve published questions and that the sitemap contains only the
new guide URLs. The calculation engine and `/solve` endpoint are unchanged by
the URL rename.

Google stopped showing FAQ rich results on May 7, 2026. Retaining FAQPage does
not promise expanded search snippets or a larger search result card. See the
Google Search Central documentation updates and section 28 of the tire/rim
fitment specification for checked sources and the separately deferred 105%
formula review.

## Freehub and Cassette Compatibility Backend FAQ Migration 418

Migration 418 adds nine English and nine Simplified Chinese FAQs for `/tools/freehub-cassette-fitment-engine`. Questions, rich-text answers, headings, publication status and ordering remain editable in Admin FAQ. It seeds only empty locales, respects existing custom page IDs, and preserves drafts, soft deletions and administrator edits. Other locales remain available for editorial configuration without automatically publishing English answers under another language. Rollback removes only untouched seeded answers and keeps page metadata.

From `go-backend`, run `go run ./cmd/server migrate` during deployment; this runs all pending migrations. Synchronizing frontend routes alone creates page metadata and does not seed these answers. After migration, verify the public `/api/v1/content/faq-pages/by-route` response with `route_path=/tools/freehub-cassette-fitment-engine` and `locale=en` or `locale=zh_cn`.

From `nuxt-i18n`, run `npm run test:freehub-cassette-fitment-backend-faq-seo-ssr`. Set `FREEHUB_CASSETTE_FITMENT_BASE_URL` and `FREEHUB_CASSETTE_FITMENT_BACKEND_ORIGIN` to the storefront and API origins when they differ from the local verification ports 10240 and 10241. This check compares complete answers in the initial HTML with the published backend response, checks the H2 FAQ heading, and verifies a single localized FAQPage JSON-LD with the same questions and answers. It also verifies that cassette selection query parameters retain the tool page canonical URL.

## Runtime Checks

Expose these internal checks to your load balancer or platform health probes:

- `/health`
- `/ready`
- `/liveness`

Metrics are exposed at:

- `/metrics`

## Docker Compose

For local or staging-like environments, use the root compose file:

```powershell
docker compose up -d
```

The root compose file is mainly a development convenience. Review environment variables, volumes, secrets, ports, and persistence before using it outside local development.

## Kubernetes

Kubernetes manifests live under `k8s/`. Treat them as deployment templates that need environment-specific review before use.

## Media upload identity and guest attachment upgrade (migration 419)

Run migrations 419, 420, 421, 422 and 423 before deploying the combined updated API. From `go-backend`, the
normal `go run ./cmd/server migrate` command applies **all pending migrations**;
review that pending set before using it for an environment upgrade.

This migration adds uploader actor type and a hidden visitor session digest
without changing existing uploader IDs, URLs, storage keys or stored objects.
Historical rows remain `legacy_account`; new administrator, member and visitor
uploads carry explicit identities. Guest uploads require an existing, authorized
customer-service conversation and its verified visitor cookie.

`MEDIA_UPLOAD_VISITOR_SESSION_STORAGE_QUOTA_BYTES` defaults to 20971520
(20 MiB of registered originals per visitor session). The existing account
quota remains 20 GiB. New administrator and customer quotas are separate;
legacy rows with the same numeric ID conservatively count toward both.
Retained soft-deleted originals count toward usage. Derivatives and pending
uploads require the later capacity-accounting upgrade.

Update every upload-serving instance before relying on the PostgreSQL quota
lock protocol. Confirm guest upload followed by image-message send, member
and administrator uploads, cross-conversation rejection, and quota responses.
The migration and quota concurrency tests use temporary tables or independent
schemas and do not upgrade the application database.

The down migration refuses to remove identity columns once typed uploads exist.
Keep the additive schema when rolling back application code to preserve ownership.
Conversation-authorized original/derivative reads are described below.
The binding/session upgrade is described in step 3 below and attachment cleanup
in step 4. Media library business partitions and conversation-specific backoffice
uploads are described in step 5. Production private upload activation remains separate.

## Customer-service attachment authorized reads (step 2, read before write)

This change prepares conversation-authorized file routes for registered
customer-service image/video attachments and image derivatives. It introduces
no additional database migration by itself. The combined application now
requires migrations 419, 420, 421, 422 and 423, including when private uploads are disabled.

Public reads verify the signed visitor cookie or authenticated customer against
the parent conversation. Backoffice reads require ticket-view permission and the
current conversation assignment, with the existing manager/administrator scope.
Public reads exclude internal messages. Each request rechecks the parent,
message membership, active asset and derivative ownership; private attachments
cannot use the generic media download or copyright export route.

The new private namespaces are:

- Originals: `customer-service/attachments/originals/...`
- Derivatives: `customer-service/attachments/derivatives/{assetID}/{preset}/v{version}/...`

Local private objects use the separate private root (or the existing sibling
`{publicRoot}-private` default), outside the public uploads directory. S3/OSS
require a separate private bucket and must not share the public base URL.
Configure private buckets to deny anonymous reads and keep their objects outside
public CDN/static routes. OSS reader uploads now explicitly set the private
object ACL. S3 keeps ACL headers absent for bucket-owner-enforced compatibility
and relies on the private bucket policy. Private derivatives require the explicit
private reader-upload capability; there is no public upload fallback.

S3 `v1.27.11` was incompatible with the existing AWS core `v1.42.0`
(`ComputePayloadHash` middleware failure before sending a request). The S3
module is now `v1.104.0`, which depends on the same core and Smithy versions
already used by this repository. Retain the updated `go.mod` and `go.sum`
together when building. Storage HTTP tests verify configured endpoints, separate
buckets, upload cache/ACL metadata, reads, copies, listing, deletion and presigning.
Private namespace-root listing also selects the private bucket.

Message responses retain the existing `attachments` and `attachment_url`
fields. Private originals use the new authorized route and add
`attachment_media` access descriptions; historical public URLs keep their
existing display behavior (canonical storefront URLs for customers and stored
public references for backoffice). File and message responses use
`Cache-Control: private, no-store` and vary by cookie. Authorized files stream
through the application without a public/presigned redirect. Local files support
HTTP ranges; remote S3/OSS reads currently stream the full object. Verify remote
video range requirements before activating private video uploads.

Deploy all readers first and verify visitor/member isolation, internal-note
exclusion, assignment transfer, private original/thumbnail/card/large reads,
public-route denial and cache headers. The current production upload handlers
still create public objects with the default rollout configuration. Typed
attachment binding, ownership checks and unsent-upload preview authorization
are implemented in step 3 below; automatic cleanup is implemented in step 4. Once
private writes begin, rollback must retain
these authorized readers or disable the affected writes.

Existing files have not been moved, their ACLs have not been migrated, and the
business database has not been upgraded by these tests. Migrating historical
public attachments requires a separate controlled process covering originals,
derivatives, stored references and CDN caches; changing `visibility` alone is
insufficient.

## Customer-service attachment bindings and pending uploads (step 3, migration 420)

Apply migration 420 before deploying this code, even with
`MEDIA_UPLOAD_CUSTOMER_SERVICE_PRIVATE_ATTACHMENTS_ENABLED=false`. The new
message format column is used by normal message reads and writes. The migration
adds `media_upload_sessions` and `ticket_message_media_attachments` without
rewriting historical URLs or moving objects. Foreign keys restrict deletion of
managed messages, assets, receipts and their parent conversations. The down
migration refuses to remove the schema while managed records remain; retain the
schema and authorized readers when rolling back the application.

The private upload flag is **false by default** in configuration and examples.
Keep it disabled during the staged rollout. Step 4 below adds the shared cleanup
lifecycle. The step 3 `attachment retention hold` remains the fallback when the
media storage and durable cleanup dependencies are unavailable. A 24-hour receipt
expiry prevents preview and binding; physical deletion follows the expiry
transaction and Outbox processing rather than the receipt timestamp alone.

With the flag enabled in an isolated validation environment, the upload flow
persists the verified actor, parent conversation and exact private object key
before writing bytes. Asset registration and the `uploading` to `ready`
transition share the existing quota transaction. A storage failure leaves the
key available for recovery, including when best-effort compensation fails.
Sending locks the parent conversation, rechecks current ownership/assignment,
locks assets in ascending ID order and consumes each owned, unexpired receipt
once. Message, binding, `bound` state and realtime Outbox commit atomically.
The PostgreSQL tests verify receipt contention and actual asset-row lock waits
for both bind-before-delete and delete-before-bind ordering.

Messages marked `media_asset_bindings` use ordered asset relations as their
only attachment authority; their legacy attachment JSON remains empty. URL
arrays and attachment descriptions are generated for existing clients.
Messages marked `legacy_urls` continue to use their original JSON. Public
assets can be reused by ID across messages without copying objects or creating
additional assets. Private asset IDs cannot replace an owned upload receipt.

The send endpoints accept `media_attachments` containing either `media_asset_id`
or `upload_session_id` per entry. Existing clients may still submit the exact
preview URL returned by upload; the server resolves it to a receipt before
binding. Mixing explicit bindings with legacy URL input is rejected.

Pending previews require both current conversation access and the original
uploader identity. Even a manager cannot read another uploader's unsent draft.
After binding, previews stop working and normal message access governs reads.
Private reads validate the receipt's parent, registered asset, exact key and
uploader identity; they never expose the private storage origin or key in upload
responses. New routes include:

- Public preview: `GET /api/v1/customer-service/conversations/:conversation_id/attachment-uploads/:upload_session_id/file`
- Backoffice upload: `POST /api/admin/customer-service/conversations/:id/attachment-uploads` (ticket-edit permission and upload rate limit)
- Backoffice preview: `GET /api/admin/customer-service/conversations/:id/attachment-uploads/:upload_session_id/file` (ticket-view permission)

Private writes currently accept images only. Backoffice media-picker UI
integration, media-center business partitions, historical private migration
and full physical capacity accounting are not included in this milestone.
Upload expiry cleanup is covered by step 4 below.
Relevant HTTP/service/storage/configuration regressions and
isolated PostgreSQL migration/concurrency tests cover the staged implementation;
the tests do not migrate the business database or activate production uploads.

## Customer-service upload expiry and managed attachment retirement (step 4, migration 421)

Apply migration 421 before deploying this combined code, even with private
uploads disabled: normal media queries now include the derivative write-lease
columns. Together with 419, 420, 422 and 423, it is an upgrade prerequisite. The migration
adds only the paired lease token/expiry columns and their constraint; it does not
rewrite existing URLs, references or objects. Its down migration refuses to
discard outstanding lease records. Retain the additive schema during an
application rollback.

The existing `CustomerServiceRetentionScheduler` now expires unbound uploads on
each poll (normally once per minute), independently of the operator's
conversation-retention enabled setting. `uploading`, `ready` and `failed`
receipts must have expired and must be older than the five-minute original-write
window. Bound receipts follow conversation retention instead. Each expiry pass
has a 30-second context and processes at most the configured worker batch limit.
No additional scheduler, task table or object inventory is introduced.

Original uploads have a deadline of creation time plus five minutes, shortened
by receipt/request expiry. Private derivative generation, including background
rebuilds, obtains a persistent asset-row lease for at most five minutes; object
IO runs outside the SQL transaction with that deadline. Retirement waits for
active writes. Derivative metadata creation/replacement locks and rechecks the
active source asset, preventing late metadata writes after retirement. All
writers must run this protocol before managed private writes are enabled.

Expiry locks the asset before its receipt, matching message binding. It rechecks
the unbound state, ownership and deadlines inside the transaction. Conversation
purge locks the parent, then assets in ascending ID order, then receipts; it
removes message bindings and receipts before deleting messages. Registered
private customer-service assets are soft-deleted only after a fresh reference
check, with the existing `storage.object_cleanup_requested` Outbox event in the
same transaction. Any event-persistence failure rolls back the retirement.
Registered public assets remain owned by the media library even after their last
chat use. Private assets still referenced by another message or business remain
intact. Deployments missing the required cleanup dependencies retain the hold.

The shared Outbox handler deletes private originals, registered derivatives and
unregistered derivatives under that asset's exact private namespace, listing at
most 1000 keys per page. It validates derivative ownership before deleting files
and never scans the whole bucket. Failed or abandoned uploads without an asset
record queue their preallocated original key in the same transaction that
removes the receipt; replay refuses deletion if that key has been claimed again.
Historical originals missing `storage_key` are resolved through the configured
provider's full URL-to-key mapping; an unresolved URL fails without a filename
fallback. Neighboring objects with the same filename remain intact.

Failed storage deletion leaves a soft-deleted asset and retryable cleanup event.
Registered original quota remains charged until all physical cleanup succeeds
and the asset/derivative metadata is hard-deleted. This is still registered
original accounting; derivative bytes and in-flight reservations require the
later capacity upgrade. Legacy URL projection cleanup preserves registered
public assets and delegates registered private assets to the shared lifecycle.

Validation covers Outbox persistence rollback, storage-failure replay, partial
originals, orphan derivatives, shared private/public references, foreign
derivative metadata rejection, lease ownership, canceled writes and bounded
local/S3/OSS listing. Isolated PostgreSQL schemas verify both actual lock orders
for message binding versus upload expiry, prior binding/delete and quota races,
and migrations 419/420/421 with preservation and guarded rollback. These tests
do not migrate the business schema or enable production private uploads.

Keep `MEDIA_UPLOAD_CUSTOMER_SERVICE_PRIVATE_ATTACHMENTS_ENABLED=false` in this
milestone. Step 5 integrates media-library business partitions, trusted upload
origins and the backoffice conversation-specific upload UI. Historical public
attachments remain a separate access-migration project; the cleanup changes do
not convert their visibility or storage location. Remote private video Range
support and full physical capacity accounting remain subsequent work.

## Media library business partitions and trusted upload origins (step 5, migration 422)

Apply migration 422 before deploying the combined backend, even with private
uploads disabled: ordinary media queries now include the classification columns.
The combined upgrade requires 419/420/421/422/423. Migration 422 is additive and
repeatable; it preserves URLs, storage keys, uploader identity and references.
Existing rows default to `unclassified / legacy_unknown / historical_unknown`.
It does not infer ownership from current references or relocate stored files.
The down migration refuses to discard recorded ownership, origins or purposes;
retain the additive schema when rolling back application code.

New ownership is assigned by fixed authorized server endpoints. Clients cannot
set a trusted origin or partition through multipart fields. Operational uploads
use these routes, in addition to the existing media-view/create permissions:

| Endpoint under `/api/admin/media/uploads/` | Additional permission | Ownership / original directory |
| --- | --- | --- |
| `catalog` | product create or edit | catalog: `catalog/product-images`, `catalog/product-videos`, `catalog/product-detail-media`, `catalog/product-variant-swatches` |
| `content` | content create or edit | content: `content/page-media` |
| `website-profile` | settings edit | content: `content/page-media` |
| `shared` | none | shared: `shared/reusable-brand-and-instruction-media` |
| `customer-service-reusable` | ticket edit | customer service, reusable reply: `customer-service/reusable-reply-media` |

Workbench feed keeps its existing `workbench-feed` directory and now records its
trusted origin. The legacy generic media upload route remains compatible and
records a general administrator origin in the unclassified partition; an
`image_purpose=product_image` field alone does not establish catalog ownership.
This milestone covers the connected product media/detail/swatch, website-profile
and media-library upload flows. Remaining specialized business uploaders require
their own permission and lifecycle review before adding trusted classifications.

The library defaults to operational assets (catalog, content and shared). Counts
and pagination use the same server-side access predicates. Conversation
attachment metadata requires manager/administrator scope and ticket-view
permission. Such rows omit URLs, access URLs, object keys, SHA256 and derivatives,
and set `requires_conversation_access=true`. The library prohibits their edit,
delete, generic download and evidence-export paths; authorized file viewing
continues through the owning conversation. This metadata restriction also
recognizes existing private namespaces, including cloud URLs without a stored key.
Operational image-dimension preflight also excludes recognized conversation
attachments from both results and summary counts, and rejects direct reconcile
requests, so that entry point cannot bypass the library delivery restrictions.

Backoffice local/pasted attachments upload through the current conversation's
authorized endpoint. With the private flag false, existing public image/video
behavior remains compatible in `customer-service/public-conversation-attachments`.
With the flag true, images use the prior pending receipt and private file flow;
private videos return 415 pending remote Range support. The flag remains false.
Public compatibility files retain public delivery; hiding their URLs from the
library does not revoke previously shared addresses. Registered public objects
also retain the existing conservative cleanup behavior and are not automatically
retired after the last chat reference ends. A scoped retention/manual cleanup
protocol for these conversation-owned public compatibility files is still needed.

All ordinary media pickers request `selection_context=reusable`: active public
assets only, excluding recognized conversation attachments and customer
submissions. Unknown historical public assets remain selectable for compatibility
until verified backfill and authoritative business-save validation are in place.
Sending an existing catalog image to a chat adds a reference and does not change
its ownership, key or physical object. Product references distinguish active use,
soft-deleted parent retention and missing parent records; both retained and broken
references continue to protect the file from deletion.

Validation covers trusted routes and permission failures, forged client origins,
list/count scope and redaction, public upload delivery, disabled-rollout video
compatibility, retained/missing product references, and partition/conversation/picker
request races and disposal. Real PostgreSQL temporary tables or isolated schemas
verify migrations 419-422, quota/binding/delete/expiry races, and list/count filtering
of Windows keys and legacy cloud URLs. Frontend typecheck and build checks pass.
Browser checks use isolated mocked API responses and do not query or migrate the
business database. The missing storefront Logo/Favicon test fixture discovered in
step 5 is resolved in step 6 using a backend-owned embedded WebP fixture.

Audited per-asset classification is implemented in step 6 below. Next implement
verified historical backfill and
authoritative scene-specific save/binding policies and any derived usage index.
No editable usage index, redundant object inventory or per-reference file copies
are introduced here. Do not enable production private uploads during this milestone.

## Audited manual media library ownership review (step 6, migration 423)

Apply 423 before deploying this combined backend, including when private uploads
remain disabled. The combined code requires 419/420/421/422/423. Migration 423 adds
`classification_revision` and permits `manual_reviewed`; it does not infer history,
move files or change upload origins. Existing classifications begin at revision 0.
Repeated up preserves reviewed rows. Down refuses to discard any manual review or
nonzero revision, including soft-deleted assets; roll back application code while
retaining the additive schema instead of resetting review data to bypass this guard.

Administrators and managers with `media:configure` can open the per-asset review and
PATCH ownership through the dedicated classification endpoint. The target also
requires `product:edit`, `content:edit` or `ticket:edit` for catalog, content or
customer-service reusable materials. Editors, support and viewers cannot classify.
Unknown JSON fields, including forged upload origins or reviewer identities, are
rejected. Allowed targets are catalog, content, shared, customer-service reusable
materials and unclassified. Conversation attachments, private/inactive files and
known customer/visitor submissions require their existing business lifecycle and
cannot be republished through this operation.

The reviewer must provide a 3–1000 character reason and the reviewed revision.
The asset lock, revision comparison, limited ownership update and existing
`audit_logs` insertion share one database transaction. An audit failure rolls back
the update; a stale revision returns 409 and requires another human review. Repeating
the same already-reviewed ownership at the current revision is a no-op. The before
and after snapshots retain original upload provenance. URL, object key, file bytes,
visibility, uploader, copyright evidence and business references are preserved.
Ordinary metadata edits now take the same asset lock and update only editable
columns; the former whole-row Save method has been removed to prevent stale writes
from overwriting ownership or other lifecycle fields.

Review hints count only recognized business references and do not expose customer
message contents. A single recognized catalog/content category on an unclassified
asset may suggest a target unless a parent record is missing. Human confirmation
is always required; these hints neither prove original upload source nor claim a
complete historical reference inventory. This milestone supplies per-asset review,
not automatic historical backfill or bulk classification.

Logo/Favicon lifecycle and upload validation tests now share one backend-owned
512×512 lossless WebP embedded by a dedicated test-fixture package. They no longer
depend on storefront assets, encoders or the test working directory.
Validation covers atomic audit failure rollback, immutable provenance/files and
references, role and payload restrictions, stale revisions and request races.
Real PostgreSQL temporary tables or isolated schemas verify migration preservation,
constraints and guarded rollback, plus simultaneous reviews and metadata edits
waiting for the same asset lock. Browser checks use isolated API fixtures, including
successful partition refresh, original-source labels and version-conflict recovery.
The full backend `go test ./... -count=1`, admin typecheck/build/font checks and
22 media/conversation frontend behavior tests pass. PostgreSQL-specific checks
were also run separately with test connection settings; the default full backend
run skips integration tests that require those settings. These checks use isolated
tables/schemas and do not migrate business data.

Product-media save/binding validation is implemented in step 7 below. Verified
historical backfill, remaining scene policies, derived usage projections and
public conversation attachment lifecycle remain subsequent work.
This milestone does not deploy, migrate the business database or enable private
uploads in production.

## Product-media identity validation and retirement serialization (step 7)

This step uses the existing 419–423 media schema; it adds no migration or automatic
historical backfill. `media_assets` remains the asset authority and `product_media`
remains the actual product placement authority. No duplicate binding inventory or
per-placement file copies are introduced. Existing file objects and provenance
are retained.

Create, replacement and translation copy now discover and lock all registered
assets involved in the primary URL, thumbnail and poster, deduplicating and locking
IDs in ascending order. Asset identity, public reuse eligibility and current image
derivatives are reread within the product transaction. Transaction repository clones
retain the configured storage resolver, covering both the application TxManager
and service fallback paths. Generic product creation with media associations also
uses this check. Retirement uses the same asset lock and rechecks real references.
A committed binding blocks retirement; a committed retirement rejects the binding
and rolls back the entire product mutation. Removing placements does not delete
assets or enqueue storage cleanup.

An explicit asset ID must agree with the registered URL or a full object key from
configured public storage/site origins. Registered URL-only references acquire the
existing asset ID. Foreign hosts containing `/uploads/`, unconfigured ports and
matching basenames are not trusted aliases. Ambiguous registrations are rejected.
Configured private cloud buckets/origins are rejected even when the public bucket
contains the same key. Missing, retired, inactive, private, conversation-owned and
known customer/visitor submissions cannot become public product media. Public
catalog/content/shared/reusable-support assets can be reused across partitions;
ownership, original source and copyright remain unchanged. The reusable picker
also excludes known customer/visitor upload actors.

Independent poster/thumbnail derivatives are stored using their original image's
stable registered URL, so a derivative rebuild cannot break a different asset's
custom auxiliary image. Main-image display variants and its preferred generated
thumbnail use current preset versions read through the same transaction.

Unregistered external HTTP(S) and non-storage static paths remain unmanaged.
Unregistered managed storage paths cannot create new bindings or translations,
including after physical cleanup removes the registration. Existing unregistered
paths can remain only on the same placement ID and in the same field. An existing
unconfigured host/port alias with an asset ID may remain only with that exact
placement ID, asset ID and URL, and the locked asset must still be eligible. These
exceptions preserve old relationships without authorizing the alias/path for new
placements. A known conflicting identity, private origin or conversation namespace
cannot use the exception. An update that does not replace media leaves relations
untouched; a translation creates new placements and must revalidate.

Before rollout, inspect product media with mismatched IDs/URLs, duplicate complete
keys, ineligible assets, missing registrations and unconfigured historical aliases.
Resolve verified relationships using existing records or replace the material;
do not infer upload source or create duplicate assets to bypass rejection. The
admin's whole-product save submits the media collection, so existing ineligible
attachments may block that save and require business review before rollout.
Unknown paths/aliases have only the exact preservation exceptions above.

Validation includes the full backend `go test ./... -count=1`, complete reruns of
the affected packages after the final compatibility adjustment, HTTP 400 mapping,
both transaction paths, rollback, auxiliary images, current derivative versions,
foreign-origin/private-bucket isolation, legacy preservation, translation and
unbinding. Separate real PostgreSQL runs in isolated schemas observe actual row
lock waits for primary/thumbnail/poster bindings against retirement in both orders,
and verify URL-only resurrection is rejected after cleanup. PostgreSQL picker
queries are also run with test settings; the default full run skips integration
tests without those settings. These checks do not migrate business tables.

Product description HTML, swatches and category images were deferred at this
milestone and are covered by step 8 below. Page and other content save paths
still need their own transaction policies. Full historical reference
coverage, verified provenance backfill and public-conversation cleanup are separate
remaining stages. This step does not deploy, commit, migrate business data or enable
production private uploads.

## Product rich text, template/product swatches and category images (step 8)

This step adds no media schema migration, binding ledger, asset copies or storage
moves. The existing 419–423 media migrations remain prerequisites for the combined
code. Actual placements stay in product HTML, template option items, product
option values and category image columns; media assets remain the identity and
reuse authority. No historical provenance or business data is automatically changed.

Product create/update, template option materialization, template save, translation
copy and category image writes now validate registered identity and public reuse
eligibility inside their business transactions. Product gallery, changed HTML and
replacement swatches participate in a single deduplicated ascending asset lock
order. Parent rows use NO KEY UPDATE, compatible with foreign-key KEY SHARE;
asset locks use the same FOR UPDATE protocol as retirement. Omitted HTML fields
reload current values inside the transaction so stale ordinary edits cannot
restore detached references. Detaching any placement does not delete its asset.

Reference reports now parse img.src, video.src, video.poster and HTTP(S)/path
a.href attributes, including entities, registered derivatives and configured
storage aliases. Text and foreign-origin suffix matches are excluded. New managed
references use stable original URLs and preserve client fragments. Template and
product swatches are scanned separately; disabled options, missing parents and
retained deleted-product content remain protected and are labeled accordingly.
Parsing uses pages of 256 business rows to bound memory; it remains proportional
to stored content and does not introduce a second authoritative index.

Unchanged HTML preserves existing relationships. Changed HTML may retain an
unregistered historical storage URL only in the same product/field/element
attribute, without increasing its occurrence count. Swatch exceptions require
the same existing item, owning specification/product, asset ID where present and
URL. Materializing a template into a new product or copying a translation creates
new placements and must revalidate; these exceptions cannot be propagated.

Before rollout, review template swatches with missing registrations, mismatched
IDs/URLs or ineligible assets: these can block new product creation. Verify old
hosts/ports/cloud aliases using actual storage configuration; URL-only HTML has
no asset ID to prove an unconfigured alias. Inspect historical CSS/srcset/source
or custom HTML separately because the current sanitizer/parser does not support
those as media placements. Review old derivative URLs before derivative rebuilds;
this change protects the owner asset but does not rewrite all historical HTML or
change derivative cleanup policies. Resolve ineligible references before whole
admin forms resubmit them. Keep deleted-parent/orphan references until the owning
business has reviewed detachment. Do not guess source or duplicate assets to
bypass validation.

Validation includes the full backend regression, affected-package checks after
final adjustments, admin type checks and isolated PostgreSQL schemas. Seven new
placement kinds are raced against retirement in both orders (14 cases), observing
actual pg_blocking_pids waits and rejecting binding after cleanup. A separate
PostgreSQL test swaps assets between two products' HTML/gallery fields to verify
consistent lock ordering; the prior six primary/thumbnail/poster races also pass.
Default backend runs skip these PostgreSQL tests without connection settings.

Page/article/FAQ/gallery and other content writes, complete historical coverage,
verified provenance backfill and bulk classification remain subsequent stages.
This step does not commit, deploy, migrate business tables or enable production
private conversation uploads.

## Media Upgrade Step 9: Registered FAQ Uploads, Answer Images and File Links

FAQ image uploads now use the authoritative MediaService instead of writing
directly to storage. The existing response shape is preserved. The authenticated
administrator owns the upload quota; the authorized route fixes the content
origin/partition and `faq_answer_image` purpose. Originals use the existing
`content/page-media` directory. Multipart actor, origin and partition fields do
not override this policy. No FAQ asset table, file copy or new migration is added.
The previous 419–423 media migrations remain prerequisites; the unrelated 424
FAQ content migration in this workspace requires its own review.

New dedicated answer images require a unique active public reusable asset whose
registered original is an 800x800 WebP. Client dimensions and an 800px derivative
of an incompatible original cannot satisfy this contract. Answer links are
collected after the FAQ-specific sanitizer, which removes embedded image/video
tags. Supported links and new images participate in one ascending asset-lock
plan within the FAQ save transaction, sharing retirement's FOR UPDATE protocol.
The generic HTML binding-plan builder is shared with products.

FAQ updates lock and reload the current record, cannot upsert a deleted row, and
reload omitted media fields to prevent a stale edit from restoring a detached
reference. Omitting `answer_image_url` preserves the image; an explicit empty
string detaches it. Existing unchanged images and unchanged answers remain
compatible. Changed answers may retain unregistered historical file links only
in the same FAQ/attribute and without increasing occurrences. New copies or
translations revalidate rather than inheriting this exception. Detachment does
not delete the asset. The offline FAQ importer also configures media identity
resolution and preserves image fields absent from its input format; use the same
storage environment and STOREFRONT_BASE_URL as the serving application.

Reference scanning now matches dedicated images and rendered file-link
attributes exactly, including registered derivatives and trusted aliases. Text,
removed tags, foreign hosts with the same suffix and longer filenames are not
usage. Scanning pages 256 FAQ rows at a time bounds memory without introducing
a second reference ledger. Deleted FAQs with retained content continue to
protect their assets and report `retained_deleted_parent`.

Before rollout, review unregistered, ambiguous or ineligible historical images
and file links, missing objects, and inaccurate MIME/dimension metadata. Preserve
existing placements while reviewing; new reuse or translation requires verified
registration or a valid replacement. Do not infer ownership from filenames or
URLs. Check historical domains/ports/CDN aliases against actual configuration
and review derivative links before rebuilding or deleting old derivatives.
Complete the upload/save protocol upgrade across FAQ writer instances before
resuming related editing, so old uploaders cannot produce unregistered files
for new writers. Retained deleted FAQs require explicit business review before
permanent detachment.

The final code passes `go test ./... -count=1` across the backend, including
the offline FAQ importer build. Validation covers real backend-owned WebP uploads, trusted identity, quota
rejection without files, single-asset reuse, transaction rollback, actual
original specifications, historical compatibility, omitted fields, explicit
detach, no deleted-row resurrection and paged exact scanning. Isolated
PostgreSQL tests race answer images/file links through create/update in both
binding/retirement orders (8 cases), observe pg_blocking_pids waits and reject
rebinding after physical cleanup. A ninth test swaps image/link asset order
between two FAQs. PostgreSQL tests skip without connection settings; targeted
product HTML regression covers extraction of the shared binding-plan helper.

Article/gallery and other public content save protocols remain subsequent
steps. This step does not commit, deploy, migrate business schemas, rewrite old
files, infer historical provenance or enable production private chat uploads.

## Pre-Launch Checklist

- `go test ./...` passes.
- Database migrations are reviewed and reversible where practical.
- Logs do not print secrets or sensitive customer data.
- Cookie, CSRF, CORS, and WebSocket origin settings are configured for the real domains.
- Upload endpoints enforce type, size, and storage limits.
- Payment callbacks are verified and idempotent.
- Backups and restore drills exist for PostgreSQL and object storage.
