# Backend Deployment Notes

These notes describe deployment preparation for the Go backend. They are not a production-readiness certification.

Latest media-center checkpoint: follow-up 51 adds agent private-video upload preparation, receipts, preview, binding and expiry under the existing disabled-by-default private-upload switch. Follow-up 50 supplies authorized S3/OSS video byte ranges. Customer uploads remain image-only. Production activation and real-cloud/browser acceptance remain outstanding; earlier follow-up sections retain their historical scope.

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

## Media Upgrade Step 10: Article Content, Featured Images and Atomic Categories

This step adds no schema migration, asset copy, storage move or reference ledger.
The existing 419–423 media migrations remain prerequisites for the combined code.
Actual placements remain in `posts.content/featured_img`; explicit categories
remain in `post_categories`. Media assets and derivative registrations supply
identity and reuse eligibility. Existing authorized content uploads already use
the content partition/origin and `content/page-media` directory.

Article create/update and preflight content-link repair validate changed content
and featured images inside the repository transaction. Sanitized img.src,
video.src, video.poster and HTTP(S)/path a.href attributes join the changed
featured image in one deduplicated ascending asset lock plan. Active, public,
eligible registered assets are rechecked under the same FOR UPDATE locks used
by retirement. Featured images require image type, without FAQ's dedicated
800x800 WebP restriction. Reuse preserves the original asset's partition/source.

Updates lock the active post with NO KEY UPDATE and reload omitted media fields,
preventing stale snapshots from restoring detached files. Empty strings detach;
omitted fields preserve. Updates cannot upsert deleted rows. Post writes omit
implicit associations; only explicit category replacement changes relationships.
Category validation, post saving and relationship replacement share one
transaction, so failures roll back together without compensating soft deletion.
This does not implement the full concurrent category deletion/locale lifecycle.
SEO updates write only supplied SEO fields and cannot restore media or categories.

Registered originals, derivatives and configured aliases use canonical original
URLs, retaining client fragments. External public URLs and static paths remain
compatible; a foreign host with the same upload suffix cannot claim a managed
asset. Unchanged historical fields are preserved. Changed HTML may retain an
unregistered managed URL only in the same article/element attribute and up to
the original occurrence count. New articles, translations and other placements
must revalidate. Detachment does not delete assets. The blog-translations importer
only updates existing grouping/SEO metadata and does not write media fields.

Reference reports scan 256 posts per page and parse sanitized attributes exactly,
excluding text, removed tags, similar filenames and foreign-host suffixes.
Deleted articles with retained content protect assets as `retained_deleted_parent`.

Before rollout, review missing or ambiguous registrations, invalid upload
identities and ineligible references that new reuse/translation will reject.
Verify historical hosts, ports, cloud aliases and derivative registrations against
actual configuration. Review old derivative links before rebuild/cleanup, and
inspect CSS/srcset/custom media outside the supported sanitizer separately.
Resolve verified records or replace material without guessing provenance or
duplicating assets. Upgrade all article upload/save/preflight writer instances
to the same protocol before resuming edits. Permanently detaching retained deleted
articles requires business review; no automatic cleanup or data rewrite occurs.

Validation includes service/HTTP regression for canonical reuse, eligibility,
legacy compatibility, omitted fields, rollback, external URLs, explicit category
changes and SEO concurrency, plus the full backend `go test ./... -count=1`.
Real PostgreSQL tests in isolated schemas race featured images, inline images,
videos, posters and file links through create/update in both operation orders
(20 cases), observe pg_blocking_pids waits and reject rebinding after cleanup.
A separate test swaps content/cover assets to verify combined lock ordering.
PostgreSQL tests skip without connection settings; no business schema is migrated.

Gallery cover/image/thumbnail transactions, single/bulk entry points and its
direct database importer remain the next stage. Other public settings, historical
provenance backfill and bulk classification also remain. This step does not
commit, deploy or enable production private chat uploads.

## Media Upgrade Step 11: Gallery Covers, Originals, Thumbnails and Import

Existing gallery cover, image ID/URL/thumbnail, deleted_at and product-link fields
support this step. No migration, reference ledger, asset copy or file move is
added. The existing 419–423 media migrations remain prerequisites for the combined
code. Upload routing continues to determine origin/partition/storage directories;
using an asset in a gallery does not change its existing ownership or source.

Gallery creation/import combines its cover and all appended images' originals
and thumbnails in one deduplicated ascending asset FOR UPDATE lock plan. Batch
append combines the entire batch. Registered identities, eligible public reuse
and image type are rechecked under retirement's locks. Independent thumbnail
assets receive the same protection as original IDs. ID-only selections resolve
their canonical original URL inside the transaction. Public external/static URLs
remain supported by generic/import paths; admin image selection still requires
an asset ID. Foreign-origin path suffixes do not claim managed assets.

Mutations first lock active gallery parents with NO KEY UPDATE; multiple parents
are locked in ascending ID order before assets. Image updates constrain ownership,
cannot move rows or upsert deleted images, and reload omitted media fields.
Gallery metadata updates reload the omitted cover. Explicit association omission
prevents stale Images/ProductLinks/Product snapshots from being saved implicitly.
Only supplied product ID collections replace links, atomically with gallery
writes. Full product deletion/status concurrency remains a separate lifecycle.
Existing batch sorting remains a sequence of individually protected writes.

Gallery deletion locks its parent before deleting images/links. Single/bulk image
deletion, product-link replacement and ordering also use active parent locks.
Competing append/update/import either completes before deletion or sees the
deleted parent and rolls back, leaving no active child behind.

The offline importer uses the same repository protocol and storage identity
configuration as the application, including STOREFRONT_BASE_URL. Each imported
gallery and all newly appended images commit together. One invalid new media
placement or insert rolls back that gallery. Originals, derivatives and trusted
aliases deduplicate by canonical identity, including existing URL-only placements
with a verified registration. Reruns do not duplicate assets or restore soft-deleted
images. Existing image metadata and product links are preserved. Dry-run remains
a preview and does not prove media validation; review input in an isolated
environment before an actual import.

Reference scans page 256 rows and match original IDs, exact registered URLs and
trusted complete object keys. Deleted galleries report retained_deleted_parent;
deleted image records under active galleries report retained_deleted_placement;
orphan images report missing_parent. The admin detail explains these states, and
the existing retained-deleted-record review counter includes deleted images.
Soft deletion preserves stored references and continues to protect files.
Permanent detachment/purge requires a later lifecycle workflow; this step adds
no automatic hard deletion or cleanup. Replacing media fields never deletes assets.

Before rollout, review mismatched IDs/URLs, unregistered paths, ambiguous object
keys, invalid independent thumbnails and private/customer/conversation assets.
Verify historical hosts/ports/CDN aliases and derivative registrations against
actual configuration. New reuse/copy/import requires verified registration or
valid replacements; preserve old unchanged fields without guessing provenance.
Upgrade every gallery writer and importer to the same parent/asset lock protocol
before resuming related editing. Retained deleted/orphan records need business
review before permanent removal. Other public settings, historical provenance
backfill and bulk classification remain subsequent work.

The final code passes the complete backend `go test ./... -count=1` and admin
`npm run typecheck`, including font/locale policies and vue-tsc. Validation covers
service/admin HTTP contracts and the actual importer, including
rollback, canonical deduplication, reruns, omitted fields, stale associations,
eligibility, exact paging and retained references. Real PostgreSQL tests use
isolated schemas for 26 binding/retirement races across applicable write entry
points, one combined cover/original/thumbnail asset order test and eight parent
deletion races. They observe pg_blocking_pids waits and reject rebinding after
cleanup by ID or URL. Default tests skip these without connection settings.

This step does not commit, deploy, migrate business tables or enable production
private chat uploads.

## Media Upgrade Step 12: Website Profile Settings and Atomic Import

The existing settings key/locale/value, visibility, deleted_at and unique index
support this stage. No new migration, asset copy, file move or reference ledger
is added. The combined code still requires the existing 419–423 media migrations.
Media registrations own identity/eligibility; settings values own actual usage.
Cross-domain reuse does not change asset origin or library ownership.

Website profile avatar/factory images and statement HTML (including retained
legacy paragraphs) now validate at the settings repository. Single Set, BatchSet,
existing transaction writes, the profile API, generic admin settings and the
offline settings importer share the same protocol. Media keys are recognized
independently of mutable group metadata. Invalid media maps to HTTP 400 rather
than 500. The media usage dialog identifies these references as website profile
settings under content.

Transactions lock supplied key/locale pairs in stable order with NO KEY UPDATE.
Missing keys reserve a row using the existing unique index, then reload under
lock; concurrent first writes serialize and failures roll reservations back.
Ordinary updates only select and lock existing rows. All changed image and HTML
attributes form one deduplicated ascending media-asset FOR UPDATE plan, shared
with retirement. Identity/eligibility and field media type are checked under lock.
Registered originals/derivatives/trusted aliases resolve to canonical asset URLs;
private, conversation, customer/visitor, inactive, ambiguous, unregistered managed
or retired files cannot be newly attached. External public/static URLs remain
compatible; matching foreign-host path suffixes never establish ownership.

Unchanged historical values stay at their existing key/locale/group. Existing
HTML media occurrences can remain while surrounding text changes; new occurrences,
fields, languages, group changes and transitions to public require validation.
Plain legacy paragraphs keep their original storage and merge whitespace.
The dedicated profile form retains its existing full-replacement contract;
generic single/batch settings update only supplied keys. Set retains existing
visibility/description semantics; BatchSet explicitly writes both. WithTx keeps
the configured trusted resolver. Writes do not upsert soft-deleted settings,
implicitly restore other keys or delete detached assets.

Reference scanning pages 256 records and reads exact image values and sanitized
HTML attributes, including legacy paragraphs and hidden/deleted settings. Text,
removed tags, near filenames and foreign-host suffixes do not match. Deleted
settings report retained_deleted_parent and continue protecting files. Permanent
detachment/purge requires a separate reviewed lifecycle.

The importer uses application storage configuration and STOREFRONT_BASE_URL.
The entire input file commits atomically; any invalid media rolls back all of it.
The batch value import preserves the old Set contract for existing visibility
and descriptions; new rows retain the existing default-public behavior. Use
explicit admin BatchSet to change visibility. Dry-run only previews data and
does not prove media validation. Review input in isolation before use; the old
partial-success import behavior no longer applies.

Before rollout, inventory unregistered/ambiguous files, ineligible provenance,
visibility changes and deleted key/locale conflicts. Preserve unchanged legacy
placements while resolving verified registrations or choosing valid replacements.
Verify actual hosts/ports/CDN/provider aliases and derivative registrations.
Upgrade online and offline settings writers together before resuming edits.

Home visual tiles, site logos and favicons have dedicated namespaces and cleanup
owners, so this step does not register them a second time in media_assets.
Their complete reuse/cleanup concurrency, legacy site_logo fallback settings and
brand logos remain later work. Before opening cross-domain reuse, establish one
identity owner, complete consumer scans, shared writer/cleanup locks and failed
upload recovery. Homepage SEO currently stores only text title/description.

Final code passes the full backend `go test ./... -count=1`, admin
`npm run typecheck`, gofmt and `git diff --check`. Service, real admin handler and
import tests cover eligibility, canonical reuse, whole-batch rollback, historical
placements/plain paragraphs, single/batch/WithTx writers, external/static paths,
exact pagination, hidden/deleted references, reruns and existing import metadata.
Real PostgreSQL tests in isolated schemas pass 24 binding/retirement races across
the six image/HTML placements and create/update in both operation orders, one
combined image/HTML ascending asset lock test and two concurrent setting first
write/reverse-input-order tests (27 cases). They observe pg_blocking_pids waits,
reject URL-only rebinding after physical cleanup and verify no partial rows or
reservations remain after failure. Default tests skip PostgreSQL-specific cases
without connection settings; no business schema is migrated.

This step does not commit, deploy, migrate business tables or enable production
private chat uploads.

## Media Upgrade Step 13: Dedicated Homepage and Site Identity Object Lifecycles

This stage repairs dedicated owner lifecycles. site_logo_assets, site_favicon_assets
and active visual_showcase_items remain their owning records. Objects stay under
site-logo/, site-favicon/ and visual-showcase/<set>/<locale>/, using the existing
separate logo storage configuration where applicable. There is no second media
registration, reference ledger, lock table, cleanup queue, file copy or file move.
No new migration is needed; verify existing dedicated table migrations and the
375 active homepage slot unique index. Combined media prerequisites remain 419–423.

PostgreSQL transactions acquire a namespace advisory lock before owner row locks,
including empty singleton tables and empty slots. Homepage writes across all
sets/locales and physical cleanup share one namespace lock; logos and favicons
each use their own. READ COMMITTED reloads current ownership after waiting.
Request contexts reach lock waits. Dedicated cleanup holds the same lock from
the ownership/reference check through object deletion. Existing outbox retry and
idempotent delete behavior remain in use.

New homepage original and independent thumbnail bindings resolve complete keys
through the configured public provider, configured storefront origin or root
/uploads/ path. Original URLs must agree with explicit storage keys. Foreign
hosts, unconfigured ports, private origins, other namespaces and missing objects
are rejected. Existing ObjectOpener implementations check new objects under the
cleanup lock and close their streams. Missing read capability rejects new writes.
The repository rejects writes without its configured transactional validator.
Unchanged historical fields can stay at the same set/locale/slot; new placements
must validate. Single-slot persistence explicitly creates or updates the locked
slot instead of Save/upsert, ignoring caller-supplied foreign IDs on new slots.
A failed save never deletes a client-selected file that another slot may use.

The real local /uploads/ allow-list now includes dedicated homepage originals and
thumbnails. Configured home-hero slots keep the existing always-visible contract;
other sets obey publication flags and time windows. Unknown/detached objects and
colliding generic media registrations cannot bypass the dedicated namespace.
Cleanup reads active owners across all locales/sets, including drafts, URL-only
originals and thumbnails. Retained references prevent deletion; cleanup winning
the lock prevents later rebinding of a missing object. Replaced soft-deleted
homepage rows remain previous versions under the existing dedicated cleanup
contract, not a promise to restore already removed files.

Each committed homepage detachment now gets a distinct cleanup operation identity.
An earlier event may have completed while another slot still used the object;
the final detachment must create a fresh task instead of colliding with that
processed event key. Delivery retries of one task remain idempotent.

Before rollout, inventory duplicate active slots, URL/key conflicts, missing
objects, retired provider aliases, legacy external thumbnails and queued cleanup.
Stop old writers/cleanup workers before switching all related entry points and
workers together: old processes do not participate in the new namespace locks.
Confirm provider read permissions without copying files to bypass validation.

This does not complete cross-domain dedicated object reuse. Inventory historical
product/content/settings consumers and duplicate media registrations before
allowing their cleanup; retain affected objects pending verified identity,
complete scans, shared cross-domain locks and compatible public access. Current
logo/favicon access still authorizes only the current object. Legacy site_logo
fallback settings, historical favicon locale shadows and brand logos remain
follow-up work. Abandoned homepage uploads have no trusted receipt/expiry record,
so this stage does not automatically delete them. Previously processed retained
cleanup events are not reopened in bulk; historical orphan review remains separate.

Dedicated regression and actual admin-save/local-upload HTTP tests pass. Fourteen
real PostgreSQL cases in isolated schemas cover concurrent first logo/favicon
uploads and original/thumbnail binding through new slots, existing slots and group
replacement against cleanup in both orders. They observe pg_blocking_pids waits,
pause cleanup inside Delete after reference checking, preserve committed reuse,
reject missing-object rebinding and verify rollback. Default tests skip these
without connection settings. Final code passes the full backend
`go test ./... -count=1`, gofmt and `git diff --check`, using the task's isolated
GOCACHE. No frontend code changes in this stage.

This step does not commit, deploy, migrate business schemas, move existing files
or enable production private chat uploads.

## Media Upgrade Step 14: Atomic Site Identity Mirrors and Legacy Shared Logo References

Logo and favicon owner writes now commit their settings mirrors and cleanup
outbox events in the same transaction. A mirror failure rolls back the owner and
event; failed uploads compensate only their newly uploaded object. Both services
require SettingService wiring for mutations. Logo HTTP handlers no longer write
settings in a separate transaction. Cache invalidation happens after owner commit.

The current dedicated owner determines mirror URLs from its complete storage key
and configured storefront URL. All existing active locales plus the requested
locale are synchronized in sorted order, retaining existing group, visibility,
type and description. Deleting without a current owner still clears stale active
mirrors. A deleted key/locale is not revived: requesting that occupied unique key
fails and rolls back the whole operation. Review those rows before rollout.

Generic single/batch/transactional/import writes acquire favicon then logo
namespace locks before sorted setting row locks. Unchanged settings forms stay
compatible. Generic changes cannot replace a dedicated current logo, override
favicon ownership, or create a new empty locale shadow while an owner exists.
When no dedicated logo exists, site_logo can still select eligible shared media;
the shared original/derivative/trusted-key validation and asset retirement lock
apply. External/static fallback addresses keep their existing behavior.

Public site, group and single-setting responses consistently prefer current
owners. Without a current logo, an obsolete trusted site-logo/ URL is cleared;
shared media, static files and external logo fallbacks remain compatible. Favicon
public values continue to come from its dedicated current owner.

Reference scans now include site_logo/site_favicon settings, including hidden
and soft-deleted rows, registered derivatives and configured exact aliases.
Current dedicated logo/favicon owners and active homepage originals/thumbnails,
including drafts, also protect conflicting shared registrations. A matching full
key in separate logo storage is a conservative identity conflict to review, not
proof that both stores contain the same object. Deleted historical homepage rows
retain the existing dedicated cleanup contract. The admin reference dialog names
these consumers. No duplicate asset or persistent reference ledger is created.

New public placements cannot use site-logo/, site-favicon/ or visual-showcase/
objects through duplicate shared registrations. This covers combined content
bindings, product main/thumbnail/poster bindings and category image selection.
ID-only selection also resolves the registered URL through configured trusted
origins when historical StorageKey is missing; foreign path suffixes do not
establish a dedicated object identity. Category repository wiring shares the
same configured media resolver.
Unchanged historical placements retain their original addresses and existing
variant snapshots; copying to a new placement or locale must validate again.

Before rollout, inventory active and retained site identity settings, conflicting
full keys, separate logo storage identity, deleted setting unique keys, configured
provider/storefront aliases and pending dedicated cleanup. Do not merge or copy
objects based on a path match. Switch all writers and workers together; older
processes do not honor these locks. This stage needs no new migration. Existing
dedicated tables, settings key/locale unique index, homepage index 375 and combined
media migrations 419–423 remain prerequisites.

Historical site-logo/, site-favicon/ and visual-showcase/ objects now retain
bytes for exact references in products, rich text, galleries, FAQ answers,
website settings and legacy community/private attachments, including hidden,
orphaned and retained deleted records. Dedicated cleanup checks these consumers
inside its existing namespace transaction. A retained business consumer returns
a retryable error: the original Outbox event stays eligible until its last
reference is removed. Shared media retirement and replayed cleanup reject
dedicated original or derivative identities, including deleted derivative
metadata and historical URL-only records, before deleting any bytes.

Public access keeps the current dedicated owner/public showcase rules and also
recognizes visible published business consumers of historical objects. This
includes active product content and visible product media, enabled swatches on
visible variant/custom-option definitions belonging to the product's template,
published posts/FAQs/galleries, and public website-profile settings. Draft,
hidden, deleted, orphaned and private consumers retain bytes without granting
anonymous access. Site identity mirrors, specification templates and shared
registrations alone also cannot grant access. Reads use the dedicated owner's
provider/storefront aliases and database, preserving separate Logo storage.
Private attachment JSON is not inspected for public authorization. Last
detachment revokes origin access before the original cleanup event removes the
file; existing browser/CDN cache behavior is unchanged.

Business updates reload their locked authoritative rows and cannot restore a
detached dedicated URL from an old editor snapshot or copy it to a new placement.
Uncommitted detachment remains a committed reference for cleanup and public
reads; retry after commit removes the object. Deploy all validated business
writers and cleanup workers together so older processes cannot bypass these
checks. Brand identities, abandoned uploads and historical bulk cleanup remain
separate follow-up work. No business schema migration or file move is required.

Validation covers owner/mirror/outbox rollback, new-upload compensation, locale
metadata preservation, deleted-key conflicts, shared fallback validation and
retirement, exact retained-reference scans, duplicate registration boundaries,
real Logo HTTP failures and public settings responses. Eleven new PostgreSQL
cases use isolated schemas and observe pg_blocking_pids: eight owner-vs-generic
upload/delete races, two shared fallback-vs-retirement races and one reversed
logo/favicon batch. Existing dedicated first-write and homepage cleanup races are
also rerun. Final checks are full backend tests, admin typecheck, gofmt and diff
whitespace validation. Historical access adds visibility/last-detachment service
cases for all three namespaces, real /uploads HTTP replacement/revocation/original
Outbox replay cases, separate Logo-store byte verification, and six PostgreSQL
detachment/stale-restoration races. Run the database cases on an isolated test
PostgreSQL instance with DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD and DB_NAME
configured. The tests create and remove isolated schemas; they do not migrate
the business schema. Final verification passed `go test ./... -count=1` with the
isolated PostgreSQL connection configured, including the six historical-object
database races, plus admin `npm run typecheck`, gofmt and `git diff --check`.
Without connection settings, skipped database cases do not count as executed.

## Media Upgrade Step 15: Product Brand Logo References and Retirement Protection

Product brand logos continue to use `product_brands.logo_url`; they do not gain a
dedicated storage namespace or a second asset registration. No migration, asset
copy, storage move or reference ledger is added. Existing media migrations
419–423 remain prerequisites. Reusing a logo preserves its original library
partition, upload origin and attribution.

Brand create/update validates changed managed logo URLs in the repository
transaction. Original, registered derivative and configured trusted aliases
resolve to the eligible public image's canonical original URL. The validator
holds the same media-asset FOR UPDATE lock as retirement through the brand write.
Private, customer/visitor, conversation, inactive, retired, ambiguous,
unregistered managed files and non-image assets cannot be newly selected.
Dedicated site logo/favicon/showcase objects cannot be newly attached through a
duplicate registration. External HTTP(S) and static paths remain supported;
foreign hosts with the same upload suffix do not identify a managed file.

Updates acquire NO KEY UPDATE on the current brand before asset locks. An omitted
HTTP `logo_url` reloads the locked current value; an explicit empty string detaches
it. Unchanged historical placements retain their addresses. An old editor cannot
restore a detached unregistered or dedicated URL, and an update cannot recreate a
deleted brand. Transaction-manager repository clones preserve the configured
storage resolver. Logo writes and existing dependent outbox events roll back
together when an event write fails. Brand replacement/deletion does not directly
delete files; unreferenced media uses the existing explicit retirement workflow.

Reference reports scan 256 brands per page and match exact originals, derivatives
and trusted complete object keys. Disabled brands continue retaining bytes. The
admin reference dialog labels these consumers as product brand logos. Historical
dedicated files used by enabled brands remain publicly readable through their
owning provider; disabled brands only retain bytes. Last detachment revokes that
permission and lets the original dedicated cleanup event retry. A shared
registration alone still grants no dedicated-file access.

Validation covers invalid create/update rollback with and without the application
transaction manager, canonical reuse without source changes, exact paging and
foreign-host isolation, disabled-brand retention, omitted/empty fields, stale
restoration, dependent-event rollback and the real admin HTTP contract. Four real
PostgreSQL cases race brand create/update against media retirement in both orders,
observe `pg_blocking_pids` waits, reject binding after physical cleanup and verify
failed updates preserve the old brand. Run them with an isolated PostgreSQL
connection configured; skipped database cases do not count as executed. Final
verification passed `go test ./... -count=1` with the isolated PostgreSQL
connection configured, including all four brand-logo database races, plus admin
`npm run typecheck`, gofmt and `git diff --check`.

Deploy brand writers and media cleanup workers together. Review historical
unregistered or ambiguous logo URLs before selecting them in new placements;
unchanged values can stay while verified registrations or replacements are
prepared. Abandoned dedicated uploads and historical bulk cleanup remain separate
work. This step does not commit, deploy or migrate business data.

## Media Upgrade Step 16: Read-Only Historical Media Registration and Reference Inventory

`adminctl audit-historical-media` produces a review report before historical
classification, registration repair or cleanup. It opens a database connection
directly and runs one read-only, repeatable-read transaction. It does not start
the application, migrate schemas, initialize storage clients, create upload
directories, change business records or deliver cleanup events. A dedicated
read-only database account can be used. PostgreSQL and MySQL are supported; the
read-only snapshot and concurrent-write behavior are verified on PostgreSQL.

Run from `go-backend` using the intended database environment (`DB_DRIVER`,
`DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`; the existing
`DATABASE_*` aliases also apply). Storage identity must match the application:
`STORAGE_TYPE`, `STORAGE_BASE_URL`, public provider bucket/region/endpoint and
`STOREFRONT_BASE_URL`. `SITE_LOGO_STORAGE_*` overrides identify a separate logo
provider; otherwise logo identity inherits shared storage. Existing private
storage origin exclusions still apply. No provider credentials or network access
to storage are needed. Do not add historical external origins merely to make
findings disappear; verify provider ownership first.

For the current `start-dev.ps1` defaults, with the intended database environment
already configured:

```powershell
$env:STORAGE_TYPE = 'local'
$env:STORAGE_BASE_URL = 'http://localhost:10241'
$env:STOREFRONT_BASE_URL = 'http://localhost:10240'
go run ./cmd/adminctl audit-historical-media -format json -max-issues 10000
```

`-config path` supplies database configuration only and requires an explicit
`STORAGE_BASE_URL`. This command does not load dotenv files. Without `-config`,
the existing adminctl database environment defaults apply; do not assume an
unrelated YAML file or application's process environment was loaded. Always
check `database_target` and `identity_configuration` in the report. Configuration
URLs omit user information, query strings and fragments, and the database target
omits credentials. Current identity metadata records public base URLs and storage
types; it is not an ownership attestation for retired origins or providers.

The scan includes registered originals/derivatives and recognized product media,
product HTML, categories, brand logos, swatches, posts, FAQ answers, galleries,
dedicated site identity/showcase owners and website/site identity settings.
Retained soft-deleted rows are included and identified in issue details. Exact
registered URLs, complete object keys and configured trusted aliases determine
identity. Dedicated object references are reported separately and do not require
a shared-media registration. Same-name files on foreign hosts do not establish
managed ownership.

Review findings include missing/invalid or conflicting registration identities,
duplicate identities, orphan derivatives, unverified historical classification,
unregistered managed references, retired or otherwise ineligible assets, media
type mismatches, ID/URL disagreement, dedicated namespace conflicts and external
upload-shaped URLs requiring origin review. Several findings may describe one
reference; totals are not file counts. Historical references that cannot be newly
selected may remain valid retention evidence. No finding authorizes deletion,
ownership reassignment, backfill or automatic canonicalization.

Defaults are `-batch-size 256` (1..1000), `-max-records-per-source 100000`
(1..1000000), `-max-issues 1000` (1..100000), `-timeout 5m`, and `-format text`.
Use `-format json` for a machine-readable report. Missing tables/columns and
per-source row limits are visible in `sources`; limited sources include their
last record ID and remaining matching row count. `requested_sources_complete`
is false if any requested source is unavailable or truncated. A partial
registration index sets `registration_index_complete=false` and suppresses
absence-based findings. The issue detail cap sets `issues_truncated`; category
totals still include every finding in scanned records. A timeout/query failure
returns an error instead of a successful report. Zero findings do not establish
complete coverage when any source is unavailable or truncated.

Issue details contain table, record ID, field, related IDs, HTTP(S) origin and
SHA256 of the reference. They omit raw object names/URLs, rich text, customer
content and private attachment payloads. The command excludes private/community
payload inspection, unrecognized custom fields, physical storage existence,
bucket listings, unused-object detection, upload receipts and cleanup queue
reconciliation. Those exclusions apply even when requested sources are complete.
Keep the report outside version control and review it against the correct
environment before planning the next repair step.

Validation covers exact original/derivative/trusted alias matching, retained
records, issue limits, incomplete-index false positives, missing schema,
cancellation, privacy and output failures. SQLite query-only mode rejects any
audit write. A real PostgreSQL concurrency test checks enforced read-only and
repeatable-read settings, allows a concurrent business write and confirms the
audit continues seeing its original snapshot. Use an isolated PostgreSQL test
instance for the database tests; never point the test suite at the business
database. No frontend changes are required for this step.

Verification passed `go test ./... -count=1` with the isolated PostgreSQL
connection configured, the final inventory regression tests, gofmt and
`git diff --check`. A local development snapshot on 2026-10-15 scanned 12
registrations and 9 media references with all requested sources complete and no
issue detail truncation. Its 17 review findings comprise 12 registration origins
unverified by the current configuration, 3 unverified historical classifications,
1 unverified dedicated owner identity and 1 upload-shaped external setting URL.
The addresses include historical local ports 9200/9199 while this report used
the start-dev public identities on 10241/10240. Verify historical origin/provider
mapping before proposing repairs; these counts do not describe production or
establish physical object existence. The report is stored outside the repository.

## Media Upgrade Step 17: Local Historical Origin and Classification Evidence Review

A read-only local review on 2026-10-16 checked the objects behind the Step 16
findings. The old Compose backend advertises localhost:9200 and mounts
`go-backend/uploads` separately from `go-backend/site-logo-uploads`. README and
environment examples identify localhost:9199 as the old storefront. The current
API launch explicitly uses localhost:10241 and storefront localhost:10240.

The three originals, nine derivatives and one dedicated Logo file all exist in
their respective owning directories. Every historical URL resolves to its full
recorded key under the reviewed historical configuration, every recorded size
matches, and current API GET responses match the corresponding local bytes.
The three originals and dedicated Logo also match their registered SHA256 values.
Derivatives have no registered digest; their current file/size/HTTP agreement is
not proof of historical content. The old site_logo setting's full key matches
the dedicated owner. This explains the reviewed legacy addresses without adding
origins to the application's trusted configuration or rewriting rows.

Classification review remains evidence-based. Asset 4 has a recognized legacy
ticket-message attachment reference; review its business access rules before
assigning reusable customer-service ownership. The original and derivatives
currently return 200 anonymously, so this access question must be explicitly
resolved before any lifecycle change. Asset 5 is retained by a product-media
record whose product is deleted; the existing review service suggests catalog
ownership, but that is a current-use hint requiring manual review. Asset 6 has no
recognized references and has identical bytes to the dedicated Logo, while its
full key and storage owner differ. Preserve those distinct objects and ownership
boundaries; matching bytes do not permit merged registration or linked deletion.

All three assets remain unclassified with legacy_unknown origin, historical_unknown
method and revision 0. No media-related audit records were found in this local
database. Neither uploader legacy_account nor present uses establishes original
upload provenance. A later approved classification must use the existing audited,
revision-checked workflow and preserve the original upload origin. Empty recognized
reference results do not authorize deletion or establish absence in excluded
business sources.

The current-origin inventory still reports 17 findings. A separate historical-origin
comparison reports three classification findings and one dedicated business
reference. That comparison is not a runtime configuration change or a completed
repair. Before proposing address compatibility or rewrites, resolve the ticket
attachment access question and preserve the dedicated owner and retained deleted
product references.

The evidence run reused production reference scanners, the manual classification
review service and configuration-only identity parsers, inside an enforced
PostgreSQL read-only repeatable-read transaction. Read-only local file and HTTP
checks verified 13 files and four registered digests. Assertions also verified
separate shared/Logo keys and owners and the legacy setting binding. Reports and
the one-time evidence source are outside version control and omit raw URLs,
object names, business text and attachment payloads. No application code, business
data, runtime aliases, storage files or cleanup tasks changed in this step; the
previous full backend test result remains the code-validation baseline. Physical
checks are local observations, not an entire storage snapshot, production audit
or unused-object inventory.

### Step 18 — Historical unclassified ticket attachments require conversation access (2026-10-16)

The local evidence in step 17 exposed a real origin mismatch: media 4 is
registered at the old backend origin `http://localhost:9200`, while ticket 10 /
message 5 stores an attachment at the old storefront origin
`http://localhost:9199`. Treating only the current public flag or only exact
registered URLs as permission leaves the original and its three derivatives
anonymously readable.

The media access policy now computes a read projection for registrations whose
origin, library classification and uploader identity are still historical and
unknown. A ticket attachment reference to the registered original or derivative
requires conversation access when no recognized published public business
consumer exists. Retained and orphaned messages participate. Nested/escaped JSON
values and query/fragment variants are inspected individually; names, arbitrary
substrings, content hashes and unrelated object keys are not identities. A
complete root `/uploads/<registered-key>` at an unresolved historical origin
can withhold anonymous access as a privacy precaution, but cannot authorize a
file read. Reference-read failures also withhold access. Explicit public ownership
and recognized published public consumers retain their existing public access;
draft and deleted content do not grant it.

The original and derivative public routes share this policy, and authorization
failures use `Cache-Control: private, no-store`. New legacy attachment URL
bindings, transactional public-business bindings, generic media downloads and
copyright evidence downloads cannot bypass a protected historical attachment.
Library delivery fields are redacted using the existing transient
`requires_conversation_access` projection. No registration visibility, upload
origin, classification, revision, storage key, object bytes or cleanup event is
rewritten. The existing explicit public conversation-upload rollout is unchanged.

Authorized conversation downloads resolve a unique registered original by its
exact stored URL or configured trusted full key, including retained registrations
in ambiguity checks. Every download rechecks conversation ownership, message
membership, internal-message scope and agent assignment; it streams bytes rather
than redirecting to a public/presigned URL. History responses replace protected
original and thumbnail addresses, including repeated nested metadata values,
with these authorized routes. Existing private attachment behavior is preserved.

For an old storefront origin that differs from both the current origin and the
registered original URL, prepare the explicit, comma-separated setting:

```dotenv
MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS=http://localhost:9199
```

That value is supported by the local evidence only; it is not a production
default. Review the actual historical origins for each installation. The setting
defaults to empty; invalid origins fail application construction. Credentials,
paths, queries, fragments, non-HTTP(S) schemes and wildcard entries are rejected.
The allowlist applies only to persisted unclassified historical attachments read
through an authorized conversation. It does not expand general public-media
trust, permit new URL bindings or publish files. Without a reviewed mapping,
unresolved old-origin authorized downloads remain unavailable, while anonymous
access is still withheld. The current runtime environment was not modified.

A separate enforced PostgreSQL read-only/repeatable-read local snapshot verified
the new code against actual registrations: media 4 and all three derivatives are
denied anonymously, media 5 and 6 retain their prior access decisions, a foreign
visitor is rejected before storage opens, and generic download/new attachment
binding is denied. Enabling the reviewed `9199` origin only inside that evidence
process restored ticket 10 / message 5 authorized streaming: the original SHA256
matched its registration and all three derivative sizes matched. History used
the conversation route. The evidence process used configuration-only identity
resolution and read-only file opens, without constructing storage clients,
migrating, writing business data, moving files or restarting the current API.

Regression coverage includes anonymous original/derivative HTTP denial and
uncacheable errors; visitor/member cookie-authorized history, repeated metadata
and Range responses; foreign owners/origins, internal messages, reassigned agents
and ambiguous retained registrations; reviewed-origin configuration boundaries;
published/draft/deleted uses; rejection of new public bindings; and independent
same-content/different-key files. The full backend suite
(`go test ./... -count=1`) passed with a disposable, isolated PostgreSQL instance
rather than the local business database. The final origin-parser hardening also
passed the focused reviewed-origin regression tests. Detailed
sanitized evidence and its one-time Go overlay source are stored outside version
control with the step 17 local reports.

Release still requires applying the reviewed environment value with the new API
binary and checking the public and authorized endpoints after rollout. Existing
browser/CDN caches and objects already public in an S3/OSS bucket can bypass this
application route: origin denial does not purge previously cached bytes or change
bucket permissions. Those delivery boundaries require installation-specific
review before claiming the historical file is private everywhere. No runtime
rollout, cache purge or bucket-policy change was performed in this step.

### Step 19 — Local API rollout and live historical attachment authorization (2026-10-16)

The local API on `http://localhost:10241` now runs the step 18 implementation
with the reviewed historical attachment origin `http://localhost:9199`. The
switch completed at `2026-10-16T02:07:54Z` (10:07:54 Asia/Shanghai), replacing
process 35000 with process 23416; the new executable owned the expected listener
and returned health 200. This is a local development rollout only.

The prepared executable is
`output/dev/go-api-server-with-historical-ticket-attachment-conversation-access-2026-10-16.exe`,
SHA256 `a84750e76117ff60bfca2a75d94eeff3054d49e452d4c885dd57f693f776faa7`.
The previous executable remains available for rollback. The repository-root
`restart-local-api-with-reviewed-historical-ticket-attachment-access-and-preserved-runtime-environment.ps1`
validates the existing listener, workspace executable paths and local database
host before stopping the selected API process. It preserves the launcher's 27
explicit environment assignments, adds the reviewed origin and sets
`DB_AUTO_MIGRATE=false`. It uses a hidden launcher, waits for the expected new
executable and health response, and restores the prior executable if startup
fails. `-ValidateOnly` performs the preparation checks without a restart. The
sanitized switch result and runtime logs are under ignored `output/dev`.

Real HTTP requests before and after the switch verified these results:

- Media 4's original and all three derivatives changed from anonymous 200 to
  404 with `Cache-Control: private, no-store`; conditional/Range requests also
  returned 404 without a redirect.
- An existing active administrator with ticket permission downloaded the
  original through the conversation endpoint with its registered SHA256 and
  all three derivatives with bytes matching the local files. Authorized Range
  returned 206. Downloads were private/no-store, varied by Cookie and did not
  redirect. Staff history returned protected conversation attachment links.
- The generic media file endpoint returned 403 for that administrator. The
  public conversation endpoint returned private/no-store 404 both without an
  owner cookie and with an unrelated signed visitor cookie.
- The other nine reviewed objects, including the dedicated Logo, retained
  anonymous 200 responses matching their local files. All 13 reviewed file
  keys, sizes and SHA256 values remained identical. Complete-row digests for
  the selected media, derivatives, Logo, setting, ticket and message records
  also matched the baseline.

The evidence helper used an enforced read-only/repeatable-read database
snapshot, local file reads and HTTP GETs. Its staff token was generated through
the existing authentication service for an existing account, held only in
memory and never logged or saved; no account or role was created. The original
visitor cookie was unavailable, so live positive authorization was verified as
staff. Step 18's isolated visitor/member HTTP regression tests and read-only
owner service check remain the evidence for those positive paths. The startup
used no migrations; normal configured startup services and schedulers still
run, so the unchanged-row assertion applies to the selected registrations,
not to the entire business database.

The origin setting is currently applied to the running API process. No machine
environment or `start-dev.ps1` default was changed. A later full development
restart must retain the reviewed setting explicitly, for example from the
repository root:

```powershell
$env:MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS = 'http://localhost:9199'
$env:DB_AUTO_MIGRATE = 'false'
./start-dev.ps1
```

Without that setting the new code still withholds anonymous access, but cannot
resolve this reviewed old-origin attachment for authorized downloads. The
restart helper may also be used for a later prepared executable while a
supported local API launcher is running; supply its current process ID through
`-ExpectedExistingApiProcessId` rather than reusing the historical ID above.

The storefront at port 10240 and the historical listeners at 9199/9200 were not
running during verification. Its source routes `/api/**` through the configured
internal API; its optional image transformer is a separate delivery boundary.
No live storefront/proxy/image-cache result is claimed. Production rollout,
existing browser/CDN/image caches and any public bucket access still require
their own environment-specific verification; no cache purge or bucket-policy
change was performed. Sanitized before/after evidence and its one-time Go
overlay source are stored outside the repository with the prior local reports.
PowerShell parsing, the switch preparation and live checks passed; step 18's
full isolated backend suite remains the code-validation baseline.

### Step 20 — Persist local restart configuration and verify storefront delivery (2026-10-16)

The repository-root `start-dev.ps1` now loads the optional, Git-ignored
`local-development-reviewed-historical-customer-service-attachment-access.json`
before stopping listeners or starting services. This installation's local file
persists the reviewed `http://localhost:9199` attachment origin and
`database_auto_migrate: false`. The tracked `.example.json` file contains an
empty origin list and a false migration switch; another checkout must review
its own historical origins before adding them. No machine environment setting
is needed to retain this configuration across `npm run dev` restarts.

The local JSON supports only
`reviewed_historical_customer_service_attachment_origins` (an array of explicit
HTTP(S) origins) and `database_auto_migrate` (a JSON boolean). Nonempty process
environment values for the corresponding two settings take precedence. Unknown
properties, invalid value types, non-origin paths (including dot segments),
credentials, queries, fragments, backslashes and wildcard hosts fail validation.
The generated API launch assignments escape literal quotes. Missing local
configuration leaves the existing environment/configuration behavior in place.
Port inspection and stop-only commands remain available independently of local
media configuration. To inspect effective values without changing services:

```powershell
./start-dev.ps1 -ValidateLocalMediaAccessConfiguration
```

`-LocalMediaAccessConfigurationPath` accepts an alternate explicit JSON path.
Windows PowerShell 5.1 checks verified the saved local settings, the empty
example, environment precedence, rejection of ten invalid-origin cases and a
non-boolean migration value. Both PowerShell scripts parse successfully. The
existing API process 23416 was preserved; a complete development-stack restart
or account bootstrap was not needed to validate this startup change.

The Nuxt development storefront was separately started on port 10240 with
public/internal API origins pointing to 10241 and HTML caching disabled. Its
listener is process 39456. A Playwright browser opened the localized homepage
successfully, with visible main content and one primary heading. Actual HTTP
requests through the storefront's `/api/**` proxy then verified:

- Staff-authorized original and three derivative downloads match the reviewed
  bytes; Range returns 206. The proxy preserves `private, no-store`, Cookie
  variance and streaming without a redirect.
- Staff history retains the authorized conversation links. The generic media
  file route returns 403; missing owner and unrelated signed visitor requests
  return private/no-store 404.
- All four protected original/derivative `/uploads/` paths return 404 at the
  storefront. The IPX image-transform route returns 404 for all four objects,
  including repeated conditional/Range requests. A reviewed public object's
  transform returns an image with status 200, confirming that the transform
  service itself is functional.

The same enforced read-only/repeatable-read evidence helper also repeated the
direct API checks. Selected registration digests and all 13 file identities
still match the step 19 pre-switch baseline. The helper saves only sanitized
statuses, counts and digests outside the repository; credentials, raw object
keys and business messages are not in the reports.

This completes local startup persistence and the live storefront proxy/image
delivery checks. The original visitor's historical conversation UI and positive
download still need that actual session credential; staff authorization and
isolated visitor/member tests do not recreate the original visitor session.
The local cold/repeated transform checks do not establish the state of any old
production image/CDN/browser cache. Production rollout, delivery-cache/bucket
review and historical manual classification remain separate pending work.
No migrations, media classification changes, URL rewrites, file deletion or
cache purge were performed in this step. The previous isolated backend suite
remains the code baseline because this step changed startup configuration and
documentation, not the Go application.

### Step 21 — Prepare production origin wiring and preserve API cache decisions at Nginx (2026-10-16)

The next release configuration exposed two gaps beyond the local Nuxt proxy.
Production Compose did not pass the reviewed historical attachment-origin
setting to the API. Its Nginx `/uploads/` and `/uploads/site-logo/` locations
also appended `Cache-Control: public, max-age=31536000, immutable` with `always`,
including when the API returned a private/no-store denial. A real Nginx test
reproduced contradictory private/public headers on rejected originals and
derivatives, and duplicate public headers on successful responses.

Both upload locations now preserve the API's cache policy without adding their
own public TTL. They still proxy through the API and retain the existing
security headers. An isolated HTTP regression loads the complete repository
Nginx configuration into the release Dockerfile's Nginx base image, with a
separate synthetic API and a temporary Docker network. Its 18 cases cover both
upload prefixes, public original/derivative 200 and revalidated 304 responses,
and private original/derivative 404 responses with ordinary, Range and
conditional requests. All now return exactly the upstream cache policy, with
no redirect and with `nosniff` preserved. The test uses no business storage,
database or production service, and removes its own containers and network.

`compose.prod.yml` now passes
`MEDIA_REVIEWED_HISTORICAL_CUSTOMER_SERVICE_ATTACHMENT_ORIGINS` through the API
environment, defaulting to empty. The production environment example documents
that empty default and requires per-installation review; the local `9199`
alias is not a production default. Rendered Compose checks verified both the
empty value and an explicit synthetic reviewed origin, with automatic
migration still disabled. The static release-boundary verifier also requires
this API setting to be present, and its example-environment network checks
passed without starting services.

The backend CI and image-publishing validation jobs now pull the Nginx base
image named by `deployment/docker/web.Dockerfile` and enable the real HTTP
regression. Backend CI also triggers for deployment configuration and
`compose.prod.yml` changes. Image-publishing validation also runs the static
release-boundary verifier with the example environment and connectivity checks
disabled. The new test opts in through
`PRODUCTION_NGINX_UPLOAD_CACHE_BOUNDARY_TEST_IMAGE`; an ordinary test run
without that setting skips the Docker-specific test. To run it locally from
`go-backend`, with Docker available and the release base image present:

```powershell
$env:PRODUCTION_NGINX_UPLOAD_CACHE_BOUNDARY_TEST_IMAGE = 'nginx:1.30.4-alpine3.24'
go test ./internal/api/edge ./internal/api/middleware -count=1
```

That focused package run passed with the real Nginx test enabled. Compose
default/explicit-origin checks, static network-boundary checks, actionlint,
workflow YAML and shell syntax, gofmt and `git diff --check` also passed. GitHub workflow runs
and release-image publication were not triggered from this workspace.

The user confirmed there is no deployed production environment and asked to
finish local release preparation. No production environment file is present,
and no live deployment, CDN purge or bucket-policy change was attempted. The
original visitor session is unavailable or uncertain, so historical visitor UI
and positive-download verification remains an explicit acceptance limitation;
existing staff live checks and isolated visitor/member tests remain valid.

Historical manual review retains the established local decisions: media 4
stays historical/unclassified with conversation access enforced by the read
policy; media 5 remains retained by the deleted product and has a catalog
candidate requiring a business ownership decision; media 6 remains unknown and
independent of the same-content dedicated Logo. No classification, historical
URL, file or cleanup task was changed in this preparation step. Before a future
deployment, review that installation's actual origins and storage delivery,
publish the API/web images containing these changes, then verify the deployed
proxy and any existing caches with the installation's real session credentials.

## Follow-up media data-chain audit: full upgrade remains incomplete (2026-10-16)

Step 21 completes local release preparation, not the entire media-library plan.
A retrospective of the conversation, staged implementation notes and current
code found five unresolved classes of defects. Isolated SQLite/temp-file
reproductions demonstrate the following behavior:

- Workbench feed file_path placements are absent from shared-media reference
  scanning, despite feed uploads registering media assets. A referenced object
  can be physically deleted. Feed saves also lack the common media eligibility
  and retirement-lock protocol.
- Ticket-message and automatic-reply metadata thumbnails are absent from the
  reference scan. Automatic-reply creation also accepts an image URL after its
  managed asset has been physically deleted.
- Legacy shared JSON scanners use LIKE/substrings: escaped-slash attachment
  values can be missed, while a foreign host with the same path or a longer
  filename can incorrectly retain an unrelated asset.
- Shared reference scans omit soft-deleted reviews, warranty claims and
  suggestion feedback, allowing deletion despite their retained media fields.
- Cleanup can remove bytes still named by another active registration for the
  identical key in the same storage provider. URL uniqueness does not establish
  object identity uniqueness.

These are audit findings, not fixes or evidence of a business-data incident.
The twelve new isolated scenarios include one passing exact-attachment control
and eleven failed protection expectations, grouped into the five classes above.
They were loaded through a local Go overlay outside version control; the current
CI does not include them. Existing focused regressions passed in five backend
packages, and seventeen admin media behavior tests passed. The workbench feed
package has no test files. This audit did not connect to PostgreSQL, rerun its
integration races or repeat Step 21's Nginx tests.

The original remaining scope also includes abandoned/general draft-upload
receipts and recovery, full physical capacity accounting, derived usage queries,
historical identity/registration reconciliation and reviewed bulk classification,
customer-submission promotion, public conversation-attachment lifecycle and
remote private-video Range support. Private upload activation, production rollout
and original-visitor acceptance remain subject to their previously recorded
conditions. The existing read-only inventory covers its named sources only;
source completeness does not establish full business or physical-object coverage.

Resolve object-identity and reference-coverage defects, then complete the missing
writer/retirement protocols with regression and PostgreSQL race verification
before treating cleanup as complete across all domains. This audit changed no
business data, storage files, cleanup tasks, runtime configuration or deployment.
The detailed local report is named
media-center-upgrade-conversation-retrospective-data-chain-audit-and-remaining-implementation.zh-cn.md
under the ignored docs/design directory; reproduction source and logs are retained
with the existing external verification evidence.

## Step 22: retain existing duplicate object registrations during retirement and cleanup replay (2026-10-16)

Shared-media reference discovery now includes other assets' original and
derivative registrations that claim the same full object identity in the
configured storage. Active, inactive and soft-deleted registrations retain
their objects. The current owner is excluded; complete keys, registered URLs
and trusted configured aliases identify conflicts without basename or
foreign-host suffix matching. The bounded scan also protects registered claims
inside the exact private derivative namespace that retirement will enumerate,
including objects absent from the retiring asset's derivative metadata.

Both retirement and durable cleanup replay use this reference discovery. A
conflict before retirement leaves the asset live and creates no cleanup event.
A conflict during replay leaves its tombstone, original event and physical
objects available for retry. Explicit repair of synthetic conflicting metadata
lets the same event complete, and another replay remains idempotent. There is
no automatic duplicate-owner selection or business-data repair.

Admin views display these as a separate registration category, with original
and derivative conflict types. Registration conflicts do not imply a content
ownership suggestion. Manual-classification repository copies now preserve the
configured identity resolver, so their trusted storefront-alias results match
the deletion endpoint. The pre-existing private namespace corruption regression
uses an unregistered foreign-namespace object to keep exercising namespace
validation independently of the earlier registration check.

Twenty-four new SQLite/temp-file scenarios and one PostgreSQL retirement/replay
scenario passed, covering original/derivative claims, retained registrations,
URL-only local/S3/OSS identities, trusted storefront aliases, Windows keys,
pagination, delayed conflicts, private namespace cleanup and independent
same-name objects. The original duplicate-registration audit reproduction now
passes. The PostgreSQL check used a new disposable container and isolated test
schema; that container was removed. Full tests passed for internal/service,
internal/repository, internal/pkg/storage, internal/api/admin,
internal/api/v1/ticket, cmd/server, cmd/adminctl, cmd/import/galleries and
cmd/import/settings; internal/domain/media has no standalone tests. All
seventeen existing admin media behavior tests and admin typecheck passed.
This step did not run all backend ./... packages or the entire PostgreSQL race
suite. Evidence is retained with the prior external audit logs as
media-center-duplicate-object-registration-backend-regression-results-2026-10-16.log,
media-center-duplicate-object-registration-retirement-and-replay-postgresql-verification-2026-10-16.log
and media-center-audit-original-duplicate-registration-reproduction-after-retention-fix-2026-10-16.log.

This is protection for existing claims at each check, not completion of all
object ownership or cleanup paths. Object-level admission/locking against a
new same-key registration after the check remains unimplemented. Derivative
backfill still removes replaced URLs through its separate best-effort path;
failed-upload rollback paths also need their own ownership and durable-recovery
review. Historical storage-location identity remains an open design item.
The other four audit classes (workbench feed, structured ticket/auto-reply
metadata, legacy JSON precision and retained deleted business rows) remain
unresolved. No business database, file, classification, historical URL, cleanup
task or runtime configuration was changed; no commit, deployment or existing
service restart was performed. Next, address exact JSON matching and retained
deleted references before extending the missing writers and their lock protocol.

## Step 23: exact legacy JSON references and retained deleted business rows (2026-10-16)

Legacy shared-media scans now decode the designated JSON fields and compare
complete registered or configured-provider identities instead of LIKE or
substring matches. They read ID-ordered pages of 256 rows for showcase/review
images, warranty images and video_url, suggestion attachments, automatic-reply
attachments and ticket-message attachments. Escaped slashes/Unicode, nested
attachment values, query strings/fragments, trusted storefront aliases and
URL-only original/replaced-derivative registrations are covered. Unrelated
foreign hosts, longer filenames and prose containing an address do not count
as references. Configured private aliases retain their objects without changing
visibility or granting public reuse.

Soft-deleted reviews, warranty claims and suggestion feedback retain their
media while those fields remain. References expose retained_deleted_parent;
admin reference details, manual-review counters and the deletion dialog display
that state. Restoring a record leaves its media protected. A warranty row using
both images and video_url yields one reference listing both fields.

Reference inspection, retirement and Outbox replay share these scanners.
References introduced after retirement block the existing cleanup event and
preserve the tombstone/bytes. Explicit final detachment in synthetic tests lets
that same event complete and replay idempotently. Dedicated historical JSON
scanning also uses the shared decoder/matcher. Empty/null values are accepted;
malformed JSON returns an inspection/cleanup error with only table, field and
record ID, never the private payload. The existing admin deletion flow closes
the dialog when inspection fails rather than treating an error as zero uses.
Obsolete substring helpers and their SQL-shape-only tests were removed in favor
of behavioral retention and physical-cleanup regressions.

Ninety-six new SQLite/temp-file cases passed, plus all seven relevant original
audit scenarios (six previously failing expectations and the exact-attachment
control). A new disposable PostgreSQL container verified all six sources across
TEXT/JSON/JSONB fields, retained states, foreign-host isolation, restoration and
original-event replay; it was removed afterward. Final package tests passed for
internal/service, internal/repository, internal/pkg/storage, internal/api/admin,
internal/api/v1/ticket, cmd/server, cmd/adminctl, cmd/import/galleries and
cmd/import/settings. All seventeen existing admin media behavior tests and
admin typecheck passed. This step did not run every backend ./... package, the
entire PostgreSQL race suite or browser acceptance. Final results and individual
logs are recorded in
media-center-legacy-json-and-retained-business-rows-completed-local-validation-summary-2026-10-16.md
alongside the external audit evidence; the detailed ignored local audit report
has the Step 23 scope and log filenames.

The scan remains a per-asset authority read, not a complete derived usage index.
Pages bound read memory, not total scan work. A corrupt historical JSON record
requires review before cleanup can establish absence of references; no automatic
data repair was added. This step does not implement every business writer or
restoration lock protocol. Workbench file_path and message/auto-reply metadata
remain outside this change. Step 22's concurrent same-key registration and other
physical cleanup entry points also remain open. No business data, object,
classification, historical URL, cleanup task or runtime configuration was
changed, and no commit, deployment or existing-service restart was performed.
Next, integrate workbench saves/reference discovery with retirement locking,
then complete structured message and automatic-reply media dependencies.

## Step 24: workbench feed media references and transactional binding (2026-10-16)

Workbench feed file_path now participates in shared-media reference inspection,
retirement and original Outbox replay. Its scanner compares exact registered
originals, retained derivatives and configured-provider aliases in ID-ordered
pages of 256 rows. It includes draft/archived entries, soft-deleted parents and
placements, missing parents and legacy schemas with only the media table. These
states retain bytes; only a published, active parent and active placement grant
public access to a historical dedicated Logo/Favicon/homepage object. The same
published-consumer lookup now includes workbench when evaluating historical
unclassified ticket attachments. Read-only inventory includes file_path too;
its coverage flag remains limited to explicitly recognized sources.

Create/update now validate new image paths inside the business transaction,
using the common original/derivative identity resolver and one ascending asset
row-lock plan shared with retirement. Invalid, private, retired, inactive,
ambiguous, wrong-type and newly attached dedicated objects reject the entire
save with HTTP 400. Eligible aliases canonicalize to the registered original;
external addresses retain existing compatibility. The unchanged occurrences of
an exact historical path may remain within the same active entry. The legacy
input has no placement IDs, so this compatibility uses occurrence counts from
the locked parent's current media rows: reordering is allowed, but duplication
or restoring a detached unregistered/dedicated path requires new validation.
Retained deleted rows cannot authorize new placements.

Entry update, status change and deletion acquire the active parent row lock
before reading its children. Parent field updates omit associations and cannot
upsert a deleted entry. Status changes update only status/actor and, when
necessary, published_at, so they cannot restore stale images or product links.
Existing child replacement and entry deletion retain their soft-deleted rows;
this step does not silently purge history to make a file deletable. An original
cleanup event remains blocked while any such retained reference exists, and
can finish idempotently after explicit permanent detachment in synthetic tests.

New SQLite/temp-file and HTTP regressions passed for upload provenance,
canonical binding, invalid-save atomic rollback, historical occurrence limits,
pagination, retained states, dedicated public access and original-event retry.
The original workbench audit now finds its reference, blocks deletion and keeps
the physical object. A disposable PostgreSQL container passed eight concurrency
scenarios: create/update versus retirement in both orders, reversed multi-image
asset order, update versus status/delete and delete before update. The container
was removed. Full tests passed for internal/workbenchfeed, internal/repository,
internal/service, internal/api/admin/workbenchfeed, internal/api/admin,
internal/api/v1/ticket, cmd/server, cmd/adminctl, cmd/import/galleries and
cmd/import/settings; internal/api/v1/workbenchfeed compiled with no test files.
All seventeen existing admin media behavior tests and admin typecheck passed.
The complete backend ./... suite, complete PostgreSQL race matrix and browser
acceptance were not run. Results are indexed in
media-center-workbench-media-chain-completed-local-validation-summary-2026-10-16.md
in the external evidence directory.

Structured message/automatic-reply metadata and their writers remain the next
confirmed audit gap. Concurrent same-key registration admission, derivative
backfill/failed-upload physical cleanup, abandoned uploads, complete capacity
accounting and the broader original design remain outstanding. No business
database, classification, historical URL, file, existing cleanup event or runtime
configuration was changed. No commit, deployment or existing service restart
was performed; this step needs no new migration.

## Step 25: structured customer-service message and automatic-reply media chains (2026-10-16)

Message and automatic-reply metadata now use one explicit media-field contract.
Link/product/order thumbnails, FAQ answer images, configuration/wheel-selection
product thumbnails, and image/video URLs and posters participate in exact
original/derivative/provider-alias reference discovery. Navigation URLs, captions
and arbitrary nested strings do not become media dependencies. Empty historical
message types conservatively retain named image/video fields without guessing
the meaning of url. ID-ordered pages include disabled rules, messages whose
parents were deleted or lost, and minimal historical schemas. Attachments and
metadata in one business row merge into one reference with both fields listed.
Malformed recognized metadata prevents retirement/cleanup; diagnostics identify
only the table, field and record ID. Read-only inventory includes both metadata
sources and reports parse/identity issues without returning customer payloads.

Rule create/update and message insertion validate newly attached managed media
inside their business transaction. Public card/attachment media and private
upload-session asset IDs acquire one ascending asset lock plan shared with
retirement. Private IDs receive no public reuse permission: the existing actor,
conversation, expiry, object identity and one-use upload-session checks still
apply. Invalid, ambiguous, retired, private, wrong-type or newly attached
dedicated/unregistered managed objects reject saves without inserting messages
or realtime events. Eligible aliases canonicalize to registered originals;
unmanaged external URL compatibility remains. A rule can retain the exact
historical address only in the same message type and media field/attachment
index read from its locked current row. A detached address cannot be restored
under that exception. Sending a rule creates a new message placement and
revalidates its media instead of copying the rule's historical exception.

Before sending, the conversation transaction locks and checks the current rule
against the selected snapshot, including its stored update timestamp, active
state, message type, metadata, attachments and reply text. Concurrent deletion,
disablement or media detachment makes the old snapshot produce no reply/event.
Rule updates use Updates instead of Save/upsert; missing rules are not recreated.
Create/update return the reread database row so timestamp precision does not
incorrectly reject a fresh PostgreSQL rule. Request decoding, canonical media
rewriting, automatic-reply metadata decoration and authorized HTTP responses
preserve large JSON integers and unrelated payloads.

Metadata-only private or sensitive unclassified historical images keep a
positive authorized read path. Card fields receive conversation-protected file
URLs with membership indices following the original attachments/ID bindings;
they do not become extra displayed attachments. Every file request rechecks the
conversation owner/agent, current message membership, registration and original
or derivative storage scope. A damaged card does not break an independently
valid original attachment. Public-origin historical ticket use can withhold
anonymous access without changing registration visibility/classification. A
message or rule referencing historical dedicated Logo/Favicon/homepage objects
retains their bytes and the original cleanup event, but grants no anonymous
access. The original event can finish idempotently after the last detachment.
Metadata is not added to the old retention worker's bare-URL deletion list.

With private uploads disabled, the existing public conversation-upload path
continues to work for the verified uploader. Only an active public original
explicitly recorded as a conversation upload and matching the current actor can
use this compatibility path; its asset joins the same sorted lock plan and its
identity is rechecked after locking. Image/video metadata can repeat that exact
attached original. This does not permit a standalone card, automatic reply,
another actor or a private upload without its receipt. These legacy public
uploads lack per-conversation receipts, so this step does not invent one or
claim a storage/CDN privacy migration.

Validation passed for the focused domain/service/public/admin HTTP regressions,
the three original audit reproductions, and eleven related backend packages in
full: internal/domain/ticket, internal/repository, internal/service,
internal/api/admin, internal/api/v1/ticket, internal/workbenchfeed,
internal/api/admin/workbenchfeed, cmd/server, cmd/adminctl,
cmd/import/galleries and cmd/import/settings. An independently created disposable
PostgreSQL container passed fifteen competition scenarios: five write entry
points versus retirement in both orders, opposite metadata/attachment asset
orders, mixed private-session/card locking, and rule detachment/deletion/disable
before delivery. Fresh create/update delivery also passed with database timestamp
precision. The container was removed. All seventeen existing admin media
behavior tests and admin typecheck passed; disabled-rule retention is now
explained in reference details and deletion dialogs. Evidence is indexed in
media-center-structured-message-media-chain-completed-local-validation-summary-2026-10-16.md
in the external evidence directory.

This step adds no migration and does not complete the entire media-center design.
Concurrent same-key registration admission and other physical cleanup writers
remain the next safety work, followed by abandoned uploads, complete capacity
accounting, usage indexing, historical storage-location governance and production
implementation. The complete backend ./... suite, complete PostgreSQL race
matrix and browser acceptance were not run. No business database, classification,
historical URL, business object, existing cleanup event or runtime configuration
was changed. No commit, deployment or existing-service restart was performed.

## Step 26: same-object registration admission versus retired-asset physical cleanup (2026-10-16)

Immediate retired-media cleanup and original Outbox replay now share one
READ COMMITTED transaction. They acquire the retained asset row FOR UPDATE,
then acquire sorted transaction advisory locks for the complete original and
current/replaced derivative object identities. Private customer-service assets
also lock admission to their entire derivative namespace, including files absent
from SQL metadata. Reference revalidation, bounded namespace listing, provider
deletes and final registration hard deletion all execute while holding these
locks. Another worker reloads after obtaining the row lock and does not repeat
physical deletion once the asset is gone.

Original registration, quota registration, associated derivative insertion,
derivative creation and replacement acquire the same sorted object/namespace
locks. They then verify the exact object still exists at the configured origin.
Registration committed first remains visible to cleanup and retains the bytes;
cleanup completed first causes the waiting registration to reject absent bytes.
This also covers cleanup whose provider deletes succeeded but whose SQL hard
delete rolled back. Failed derivative registration batches preserve the previous
active metadata. A repository without configured origin verification rejects new
registrations. New shared original/derivative registrations cannot acquire
site-logo, site-favicon or visual-showcase objects; dedicated owners continue to
use their own lifecycle and namespace protocol. Historical registrations retain
the existing conservative reference/ownership handling.

Local verification checks regular files in the correct public/private root.
S3 and OSS use origin HEAD requests in the correct public/private bucket, without
downloading object bodies or consulting CDN existence. Open-only storage adapters
are closed immediately without reading. Provider verification errors do not
return private URLs or credentials. All existing media asset rows must be locked
before object admission locks; actual advisory IDs are deduplicated and sorted,
including private namespace IDs. PostgreSQL rejects an enclosing stale-snapshot
transaction instead of silently scanning under REPEATABLE READ. No new lock
table, reference ledger or migration is required. Historical storage-location
identity remains outside this step.

Validation passed for 33 focused leaf cases, including absent bytes, atomic
derivative batches, dedicated ownership rejection, aliases/private locations,
provider-error privacy and origin HEAD. A newly created isolated disposable
PostgreSQL container passed 53 leaf cases: 28 new object-admission cases plus
quota, receipt/binding/expiry and structured-message/rule concurrency regression
coverage. Twelve related backend packages passed in full, including storage,
domain, repository, service, admin/public ticket APIs, workbench APIs and command
entry points. The temporary container was removed. The complete backend ./...
suite, complete PostgreSQL matrix, browser acceptance and real remote production
storage acceptance were not run. Evidence is indexed in
media-center-object-registration-cleanup-admission-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. This step does not cover every physical
deletion writer: derivative generation/backfill compensation, old derivative
deletes, unregistered customer-service upload cleanup and other raw URL cleanup
still need the same admission/retention boundary. Those are next, followed by
abandoned uploads, complete capacity accounting, usage indexing, historical
storage-location governance and production implementation. No business database,
classification, historical URL, business object, existing cleanup event or
runtime configuration was changed. No commit, deployment or existing-service
restart was performed.

## Step 27: durable failed-upload compensation and retained derivative generations (2026-10-16)

Ordinary media upload failure compensation, partial derivative generation failure,
derivative metadata failure and derivative backfill now persist exact-object
cleanup events before attempting provider deletion. Prepared private originals
continue to use their existing pending-upload receipt and expiry lifecycle.
Compensation never falls back to unchecked deletion when event persistence fails:
it retains bytes and propagates the durability error, including when image
conversion would otherwise be best effort. Request cancellation does not cancel
the bounded attempt to persist compensation. Providers that fail without returning
an exact key, process crashes before recording a receipt, and database outages
still require the planned preallocated ordinary-upload receipt/reconciliation
work; this is not complete abandoned-upload recovery.

Replacing derivatives, recording each old generation's cleanup event, and updating
generated product variant/thumbnail snapshots now share one SQL transaction. A
snapshot or Outbox failure preserves the previous active generation and makes a
later missing-generation backfill retry possible. A committed worker can only see
the replacement snapshots after the same commit. Cleanup acquires the parent row,
the exact retained derivative row where applicable, and the same sorted object
admission lock used by new registrations, before rechecking references and deleting
bytes. It keeps the original event retryable on references or provider failures.

Exact-generation scans do not expand to unrelated originals or current generations.
They include retained business rows, product image_variants JSON, manual thumbnails,
rich text, structured/legacy message media, and every original/derivative registration,
including other registrations under the same parent. Only the event's own retired
derivative tombstone is excluded. Whole-asset retirement/cleanup also includes
retained product-media rows so it cannot bypass this boundary. Dedicated objects
and foreign private/parent derivative namespaces cannot be cleaned through shared
upload compensation. Retired derivative tombstones remain after physical cleanup
to reject later binding of the retired address; duplicate Delete is idempotent.

Existing-consumer reads use a separate resolver that can identify a retained
derivative under an active, uniquely registered parent. Anonymous old-generation
reads additionally require an exact published consumer and preserve the historical
ticket privacy checks. The last published detachment returns a non-cacheable 404
from the local uploads route, while retained draft/deleted consumers can still keep
the bytes for cleanup purposes. Authorized structured-message membership streams
the exact old private derivative after conversation and parent validation. New
public bindings still use the strict resolver and reject retired generations.
Remote CDN/bucket revocation is still an environment-specific deployment concern.

Validation: 49 focused leaf cases passed, including five real uploads-route HTTP
cases, transaction rollback/retry, partial generation, persistence/provider failure,
same-parent and paginated registration conflicts, exact private membership, and
whole-parent retention. A new isolated disposable PostgreSQL container passed 45
leaf cases: 17 new cases (four cleanup kinds versus two registration kinds in both
orders, plus atomic snapshot rollback/retry), and 28 object-admission regressions.
Twelve related backend packages with tests passed in full; domain/media had no test
files. The disposable container was verified by full ID and task label and removed.
Evidence is indexed in
media-center-derivative-generation-compensation-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. No migration was added. No business database,
business file, historical URL, existing task, runtime configuration, deployment or
existing service was modified. No commit was made. The entire backend ./... suite,
complete PostgreSQL matrix, real browser and real remote production storage were
not tested in this step.

Next physical-deletion boundaries to address are unregistered customer-service
original cleanup and remaining dedicated/avatar/community/evidence upload failure
paths. Ordinary draft/abandoned-upload receipts, complete physical capacity,
rebuildable usage indexing, historical storage-location identities, reviewed
historical reconciliation and production implementation remain unfinished.

## Step 28: unregistered customer-service originals, receipt admission and cleanup lock lifetime (2026-10-16)

Unregistered private customer-service original cleanup now acquires the same exact
object admission lock as original/derivative registration and pending-upload receipt
creation. It rechecks all registered owners, receipts and identified historical
business references inside that transaction before deleting. URL-only, Windows-key,
retained derivative and retained business rows are included. Exact retention can
resolve configured private origins without authorizing public reuse. Foreign hosts,
similar names and independent paths do not claim the object.

Receipt creation validates a canonical customer-service original key and shares the
object lock, but intentionally does not require existing bytes: the receipt must
commit before the first upload write. Actual media registration still verifies the
provider object after acquiring its lock. A receipt that commits first retains the
original task; cleanup that commits first permits a later preallocated receipt but
rejects a media registration whose bytes are gone. All receipt states retain their
exact object claim until the owning lifecycle atomically removes the receipt. Whole
asset retirement and cleanup also scan same-key receipts without a media_asset_id,
including historical Windows keys, rather than relying on asset-ID relations alone.
Existing ID-based receipt/message results are not duplicated.

A real PostgreSQL reproduction also found that request cancellation could make
database/sql release a transaction's locks while its provider Delete was still
running. Whole-asset cleanup, failed-upload/retired-generation exact cleanup and
unregistered customer-service original cleanup now share a transaction wrapper
that retains row/object locks until the synchronous provider operation returns.
Individual SQL statements, lock waits and provider calls still use the request
context; cancellation is propagated and leaves the durable event retryable. This
does not introduce a background delete or another ownership ledger.

Validation: 29 focused local leaf cases passed without failures/skips. A newly
created independent disposable PostgreSQL container passed 69 leaf cases without
failures/skips: 19 new cases (six claim/cleanup scenarios in both orders, three
cancellation/rollback cases, stale-snapshot rejection, and three in-flight provider
cancellation cases) plus 50 existing registration, generation, receipt/binding and
expiry regressions. Twelve related backend packages with tests passed in full;
domain/media had no test files. The canceled-provider failure before the fix and
the passing final evidence are indexed in
media-center-unregistered-customer-service-original-cleanup-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. The disposable container was checked by full
ID and task label and removed. No migration, business-data/file change, historical
URL change, existing-task change, runtime change, commit, deployment or existing
service restart was performed. The complete backend ./... suite, full PostgreSQL
matrix, real browser and real remote production storage were not tested here.

Next, review remaining physical deletion and failure compensation in dedicated
site identity/homepage, avatar, community and evidence lifecycles. Ordinary
draft/abandoned-upload receipts, complete physical capacity, rebuildable usage
indexing, historical storage-location identities/reconciliation and production
implementation remain unfinished; this step does not complete the media center.

### Media center follow-up step 29: dedicated upload compensation and cleanup lock lifetime (2026-10-16)

Logo and Favicon failed uploads now persist compensation in the existing object
cleanup Outbox before attempting physical deletion, including provider results
that contain both an object address and an error and owner/mirror transactions
that roll back. Homepage image uploads use the same compensation for partial
provider failures. The uploading provider must resolve a complete canonical key
inside its own namespace; invalid, foreign or unresolvable results grant no
deletion authority. Compensation persistence uses a separate bounded context so
request cancellation cannot silently discard it. Persistence failure retains the
bytes and returns an error. Every failed attempt has a new event identity, so a
previously completed same-key event cannot swallow a later required retry.

Compensation reuses each dedicated owner's existing cleanup handler, namespace
lock, current-owner/draft checks and retained business-reference scanning.
Historical references leave the same durable event retryable until detached.
Current owners retain bytes; normal owner detachment continues to create its own
cleanup task. No second ownership ledger, table or migration is introduced.

Logo/Favicon repositories now require configured validation before owner writes.
After acquiring the namespace/current-owner locks they verify the URL/key agree
with the owning provider and that the original object still exists. Context-bound
repository copies preserve the validator. Thus a cleanup that wins the lock
prevents a waiting upload from committing dangling ownership. Historical access
fixtures seed historical rows directly instead of bypassing production validation.

All three dedicated cleanup paths hold their namespace transaction locks until
synchronous provider deletion returns, including request cancellation. SQL, lock
waits and provider IO retain the caller context. Cleanup requires its own
transaction lifetime; externally owned transactions are rejected before IO.
PostgreSQL namespace locks also verify actual READ COMMITTED isolation, rejecting
an outer REPEATABLE READ snapshot before owner writes.

Validation: 43 new focused local leaf cases passed without failures/skips. A newly
created independent disposable PostgreSQL container passed 49 leaf cases without
failures/skips: 18 new cases and 31 existing dedicated lifecycle regressions.
The new cases cover three namespaces, owner/cleanup in both orders, owner rollback,
in-flight provider cancellation, stale-snapshot rejection and outer transaction
cleanup rejection. Twelve related backend packages with tests passed in full;
domain/media had no test files. Database integrations are skipped in that ordinary
package run and were exercised separately only within the stated PostgreSQL scope.
An initial test PID barrier was corrected to run before the expected first-owner
not-found result; its exploratory log is retained separately from passing evidence.
Formatting and changed-path whitespace checks passed. Evidence is indexed by
media-center-dedicated-public-upload-compensation-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. The disposable container was checked by full
ID/task label and removed, along with its temporary credential file.

No business database/file, historical URL, existing cleanup task, runtime config,
commit, deployment or existing service restart was changed. The complete backend
./... suite, full PostgreSQL matrix, browser and real remote production storage
were not tested here. Avatar, UGC and evidence cleanup/compensation remain to be
reviewed. Successful but abandoned homepage uploads and crashes before a trusted
upload result/durable event still need receipts and reconciliation. Capacity,
usage indexing, historical storage identities and production implementation also
remain unfinished; this step does not complete the media center.

### Media center follow-up step 30: customer-service avatar ownership and durable compensation (2026-10-16)

Customer-service avatar cleanup now holds the avatar namespace transaction lock
while checking exact provider object identity, all retained profile rows and the
existing business/original/derivative reference scanners, and until synchronous
provider deletion returns. Disabled and soft-deleted profiles and retained product
or rich-text placements keep the same cleanup event retryable. After the final
detachment, that event can retry successfully. Public avatar access separately
requires a non-deleted profile's exact object reference; retained business history
alone does not grant public access.

Avatar uploads now persist compensation in the existing avatar cleanup Outbox
before attempting owner-checked deletion. This includes provider results with
both an address and an error, canceled requests and failed owner transactions.
Compensation uses a bounded independent context. Event persistence failure retains
the bytes and propagates the durability error; an untrusted or foreign result is
never converted into deletion authority. Every failure or detachment has a fresh
event identity so an earlier completed same-key task cannot swallow new cleanup.

Avatar binding acquires the same namespace lock before the profile row and then
checks canonical provider identity and object existence. A cleanup that wins the
lock makes a waiting binding reject missing bytes. Repository context copies keep
the validator; unconfigured bindings are rejected. Ordinary profile creation
cannot insert an avatar, and ordinary contact/status updates explicitly exclude
the avatar and cannot upsert or revive a deleted profile. This prevents a stale
metadata save from restoring a detached or physically deleted avatar.

The avatar namespace is also included in shared-media admission and cleanup
ownership guards. New shared originals/derivatives and public business bindings
cannot claim it, and old shared cleanup events cannot delete its objects even
after the current avatar is detached. Existing historical consumers remain in
their original placements and participate in exact retention. Namespace locking
uses the existing READ COMMITTED and provider-cancellation lifetime protections;
cleanup in an externally owned transaction is rejected. No new table, migration
or ownership ledger was introduced.

Validation: 38 focused local leaf cases passed without failures/skips. A newly
created independent disposable PostgreSQL container passed 57 leaf cases without
failures/skips: eight new avatar cases and 49 existing dedicated lifecycle cases.
These cover binding/cleanup in both orders, binding rollback, provider cancellation,
canceled lock wait, stale metadata save, old transaction snapshots and external
cleanup transaction rejection. Twelve related backend packages with tests passed
in full; domain/media had no test files. The ordinary package run passed 3359 leaf
cases and skipped 280 database cases in the unconfigured environment; the stated
PostgreSQL scope was exercised separately. Three final local rollback/query-failure
cases were added after that package run and passed in the final 38-case focused
run. The initial PostgreSQL fixture incorrectly remigrated an existing associated
user table; its exploratory log is retained separately, and the final combined
57-case run passed after creating only the missing profile table.

Formatting and changed-path whitespace checks passed. Evidence is indexed in
media-center-customer-service-avatar-ownership-and-compensation-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. The disposable container was checked by full
ID/task label and removed, along with its exact temporary credential file. No
business database/file, historical URL, existing cleanup event, runtime config,
commit, deployment or existing service restart was changed. The complete backend
./... suite, complete PostgreSQL matrix, browser and real remote production
storage were not tested here.

Next are UGC publication/copy cleanup and evidence upload compensation. Successful
but abandoned ordinary/dedicated uploads, preallocated receipts, crashes before a
trusted upload result or durable event, database-outage reconciliation, complete
physical capacity, rebuildable usage indexing, historical storage-location
governance and production implementation remain unfinished. This step does not
complete the media center.

### Media center follow-up step 31: UGC publication, exact retention and durable compensation (2026-10-16)

UGC pending uploads, approved publication destinations, moderation and physical
cleanup now share one lifecycle namespace lock before owner rows. Approval reloads
the pending owner inside that transaction, copies under the lock, verifies the
approved objects, and commits the new image list/status with its pending-source
cleanup event. Rejection and expiration also commit their state change with the
existing cleanup event. A competing approval/rejection/expiry rechecks the current
status after waiting. Repository context/transaction copies keep the validators;
new image claims require canonical owning-provider identity and existing bytes.
Rejected/approved records cannot be restored to pending by a stale status write.

Upload failures and publication rollback now persist compensation in the existing
object-storage cleanup Outbox before attempting physical deletion. Both a provider
upload returning an address with an error and a copy that writes its known target
before returning an error are covered. Every attempted copy target is recorded
before IO. Compensation gets a fresh attempt identity and a bounded independent
context, so cancellation and an earlier completed same-key task cannot silently
discard the new retry. If event persistence fails, bytes remain and the durability
error is returned. Foreign, unresolvable and other-namespace results grant no
cleanup authority. Publication cannot copy from other private/dedicated lifecycles.

Scheduled immediate cleanup and the Outbox worker now use the same locked exact
reference check and deletion/pruning transaction. All identified UGC, historical
business, original and derivative claims are scanned in bounded pages with JSON
decoding and trusted provider/site identity. Only the originating rejected owner's
pending reference may be excluded. Other references or scan/provider errors keep
the original event retryable. Successful pending deletion prunes only that exact
reference under the same lock; failed pruning remains retryable and idempotent.
Both copy and delete hold the lifecycle lock until their synchronous provider
operation returns, including cancellation. Old outer snapshots and externally
owned provider-operation transactions are rejected.

Shared-media admission, public business new bindings, retirement and old generic
cleanup tasks also recognize showcase/pending and showcase/approved as dedicated
objects, including objects with no current owner. Public UGC lookup no longer uses
a 25-row substring query or lets an earlier rejected duplicate hide an approved
consumer. It scans decoded exact references across pages, permits approved UGC
consumers, and keeps pending objects private. Expiration candidate selection also
decodes JSON and preserves oldest-first batching, including escaped legacy keys.
Retention alone grants no public access. No new table, migration or ownership
ledger was introduced; existing historical rows and URLs are retained in place.

Validation: 56 new local focused leaf cases passed without failures/skips. A new
independent disposable PostgreSQL container passed 72 leaf cases without failures
or skips: 15 new UGC cases plus 57 existing avatar/dedicated lifecycle regressions.
The new database cases cover pending/approved binding versus cleanup in both
orders and rollback, actual approval versus rejection/expiration in both orders,
canceled in-flight copy/delete, stale isolation and external transaction rejection.
Twelve related backend packages with tests passed in full: 3418 leaf cases passed
and 295 database cases skipped in the ordinary unconfigured environment;
domain/media had no test files. The stated PostgreSQL scope ran separately.
Existing and new UGC local regressions also passed after final cleanup of obsolete
query code. Formatting and changed-path whitespace checks passed.

Evidence is indexed in
media-center-ugc-showcase-ownership-publication-and-compensation-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. The disposable container was verified by full
ID/task label and removed, along with its exact temporary credential file. No
business database/file, historical URL, original cleanup event, runtime config,
commit, deployment or existing-service restart was changed. The full backend ./...
suite, complete PostgreSQL matrix, browser and real remote production storage
were not tested in this step.

Next is order-evidence upload/reference compensation, preserving its immutable
evidence semantics. Successful abandoned uploads, preallocated receipts, crashes
before a trusted upload result or durable task, database-outage reconciliation,
complete capacity accounting, rebuildable usage indexing, historical storage
identity/governance, reviewed operational-material conversion and production
implementation remain unfinished. This step does not complete the media center.

### Media center follow-up step 32: immutable order-evidence ownership and durable upload compensation (2026-10-16)

Order-evidence uploads now use a fresh per-attempt directory below the exact
order/item namespace. A provider address returned with an upload error, or a
registration transaction failure, first persists the existing object-cleanup
Outbox event before attempting deletion. The owning provider must resolve the
complete canonical key within that attempt directory. Event-persistence failure
retains bytes and returns the durability error; foreign/unresolvable results and
legacy directories grant no physical cleanup authority. Compensation uses a
bounded context independent of request cancellation. The worker reuses the
original event after provider, reference-scan or retained-ownership failures.

New reference registration now checks the exact order AND evidence item, allowed
type/size/uploader, private provider identity, actual byte length, SHA256 and
content-derived MIME after acquiring the evidence lifecycle namespace lock.
Registration, detachment, item-state refresh, package locking/revision and failed
upload cleanup use that same lock. Physical Delete holds it until the synchronous
provider returns, including request cancellation. Stale outer isolation and an
externally owned physical-cleanup transaction lifetime are rejected.

Apply migration 425_retain_detached_order_evidence_attachment_ownership_history
before running this application version. It adds nullable detached_at and an index
to the existing attachment table; it does not change existing keys, rows or
snapshots. Removing an attachment now hides its placement while retaining that
immutable row as permanent object ownership. Normal lookup/preloads exclude
detached placements. Detaching the last required attachment from a complete item
also changes the item to draft and refreshes package readiness atomically; a
concurrent completion rechecks attachments inside the lock. Revisions copy active
attachments with their original keys, and their authenticated reads still work.
The down migration refuses to discard detached history or restore those placements.
Keep evidence writers/workers on the same application version; old binaries still
hard-delete references and do not participate in this ownership protocol.

Exact retention scans include all attachment rows, detached history, item data,
order snapshots, submission snapshots and export snapshots, using paged reads and
decoded JSON. Shared-media registration and generic original/derivative cleanup
cannot acquire the order-evidence namespace even if a legacy shared row collides.
Retention never grants public access. Once evidence was registered, even its last
detachment does NOT authorize physical deletion. Pre-upgrade hard-deleted history
cannot be reconstructed: those old objects remain retained, and automatic failed
upload cleanup is restricted to the new per-attempt directories.

Validation: 23 new local leaf cases and the existing order-evidence service/HTTP
cases passed (58 focused cases total). Eight related backend packages passed in
full: 3607 leaf cases passed and 313 database-dependent cases were skipped with
database connection variables cleared. An independent disposable PostgreSQL
container passed 81 leaf cases with no failures/skips: eight new evidence races,
one migration preservation/idempotence/guarded-rollback case, and 72 existing
UGC/avatar/dedicated lifecycle cases. Initial full-package failures were test
fixtures missing the newly required Outbox dependency or using unverified fake
registration; fixtures now configure durable cleanup and use actual uploads or
explicit historical rows. Final focused and full-package logs pass. Formatting
and whitespace checks pass. Evidence is indexed in
media-center-order-evidence-completed-local-validation-summary-2026-10-16.md
in the external evidence directory.

The temporary container and its anonymous volume were verified by full ID/task
label and removed; the exact temporary credential file was also removed. No
business database/file, historical URL, original cleanup task, runtime config,
commit, deployment or existing-service restart was changed. Migration 425 was
exercised only in isolation, not applied to a business database. The entire backend
./..., full PostgreSQL matrix, browser and remote production storage were not
tested. Successful abandoned uploads, preallocated receipts, crashes before a
trusted result/durable task, database-outage reconciliation, complete capacity,
rebuildable usage indexing, historical storage identity/governance, operational
material conversion and production implementation remain unfinished. This step
does not complete the media center.

## Media-center follow-up 33: business-editor upload receipts and abandoned-upload expiry (2026-10-16)

New uploads made while editing product media, descriptions and variant images,
website-profile images, FAQ answer images, workbench images and refund-policy
images now use an explicit draft workflow. Uploads made directly in the media
library retain their permanent semantics, including catalog/content uploads.
The authorized route selects the workflow and actor; multipart receipt, origin,
identity and expiry fields cannot enroll or promote a file.

Migration 426 creates business_draft_media_upload_receipts for this workflow.
Every attempt receives a UUID namespace under business-draft-media. The service
prepares image conversions and commits the complete original/derivative key list
before the first provider write. Local/S3/OSS writers accept only that exact
public namespace. Original and derivative bytes are verified by size and SHA256;
foreign/mismatched responses cannot authorize deletion of the returned address.
An unsupported conversion can retain the validated original with a one-key
receipt. Shared upload/registration/failed-upload cleanup cannot claim this
namespace without the prepared registration protocol.

Writing holds receipt/object locks until synchronous provider operations return,
including cancellation; each provider/SQL operation retains its request deadline.
The write context is bounded to five minutes. Registration separately acquires
quota -> receipt -> object locks, validates the actor/origin/key/digest/type/size,
and commits the asset, inline derivatives and ready receipt together. It never
acquires a quota lock while holding the upload's object locks. Business binding
locks assets in the existing sorted order and marks receipts bound in the owning
transaction, including the separate category-image path. Rollback restores the
ready state. Bound receipts stop automatic expiry; later manual deletion still
depends on authoritative reference scans and removes the receipt on retirement.

Unbound uploads become eligible after 24 hours. The existing scheduler checks
them independently of conversation-retention settings. Asset -> receipt locking
rechecks status and registration races. Registered originals use the existing
atomic retirement/Outbox path; unregistered attempts atomically exchange their
receipt for one exact-key cleanup event. Workers recheck all retained references
and keep locks until Delete returns. Referenced files and scan/provider failures
retain retryable work. A separate five-minute retry time prevents blocked rows
from monopolizing a bounded batch without changing their original expiry. Draft
derivative rebuilds wait until business binding; expiry cannot race additional
unreceipted rebuild writes. Provider operations reject externally controlled
transaction lifetimes; receipt locks also reject stale PostgreSQL isolation.

The audit additionally found that refund-policy JSON settings were outside the
previous binding/reference plan. Section images and media contact links now join
the same transaction and sorted asset locks. Active, private and soft-deleted
settings retain exact references; order_policy_disclosures.policy_json retains
immutable policy snapshots. New draft-media snapshot capture locks active assets
before inserting and leaves the snapshot bytes/hash unchanged. Snapshot-first
blocks retirement; retirement-first rejects a stale snapshot. This new write
guard covers the draft namespace; it does not retrofit every legacy snapshot or
dedicated-object writer into that protocol. Retained legacy snapshots are scanned
conservatively and do not gain public access.

The admin API exposes an unbound draft's deadline. Product/profile upload notices
and the media card explain the save window. Classification or metadata edits do
not turn an unbound editor upload into permanent library material.

Apply migration 426 before enabling the new routes. Drain old application and
worker versions before switching writers; mixed versions can bypass the draft
binding/rebuild protocol or fail to dispatch its new cleanup resource. The down
migration refuses any remaining receipts, including bound workflows. Do not
bypass that guard by discarding receipts. No business migration has been applied.

Validation: 30 new local leaf cases passed. Eight related backend packages passed
in full (3639 leaf cases; 333 database scenarios skipped with connection variables
cleared). Disposable PostgreSQL passed 115 cases with no failures/skips: 12 new
receipt/policy/snapshot races and migration checks, plus 103 existing binding
regressions. Admin behavior tests passed 22 cases across four files; typecheck
and production build passed. All 39 touched Go files passed formatting checks;
the delivery-file whitespace check also passed. Exploration logs are retained;
the completed-delivery logs contain the final passing results.

Final validation and disposable-resource removal are indexed in
media-center-business-draft-upload-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. This work changes local source/release
preparation only: no business database/file, historical URL, original cleanup
task, runtime config, commit, deployment or existing-service restart was changed.
Real remote storage, browser acceptance and production execution remain pending.
Unreceipted dedicated uploads, provider operations surviving process/database
connection loss, outage reconciliation, complete physical capacity, usage
indexing and historical storage governance remain outside this completed step.

## Media-center follow-up 34: write-ahead receipts for immediate-owner dedicated uploads (2026-10-16)

Site Logo, Favicon and customer-service avatar uploads now preallocate one unique
key under their existing dedicated namespace, commit an existing ownership-aware
cleanup Outbox event, and only then write provider bytes. No schema migration or
second ownership ledger is added. Local/S3/OSS expose a restricted preallocated
writer for these three namespaces; shared media, private attachments and showcase
drafts cannot use that capability. Unique keys use an immutable public cache
policy. There is no fallback to an adapter that allocates its own unknown key.

The cleanup receipt is delayed to the request's write deadline, bounded to five
minutes. A fresh independent READ COMMITTED transaction locks namespace -> receipt,
rechecks that the exact receipt is pending, unclaimed, unchanged and unexpired,
and holds both locks through synchronous provider IO and owner binding. A worker
claim, prior attempt, completed/missing/expired receipt or changed payload fences
the late writer before any bytes are sent. Worker claims do not hold Outbox row
locks while waiting for namespace/provider operations. The upload rejects an
externally controlled transaction before receipt persistence.

Only the precommitted key grants compensation authority. No URL returned by a
provider can expand it. Errors with no returned URL, mismatched/empty responses,
request cancellation, owner/mirror failure or receipt-exchange failure preserve
the original exact-key event. Existing workers recheck current ownership and
retained historical references before deleting. A blocked/failed cleanup retries
the same task. Request cancellation does not release namespace/receipt locks
until synchronous provider IO returns; individual SQL/provider calls retain
the request deadline.

Success atomically removes only the newly created temporary workflow receipt
while committing the dedicated owner, site-setting mirrors where applicable, and
cleanup of the previous owner. A failed receipt exchange rolls back all those
writes. Existing cleanup tasks are never consumed or rewritten by this exchange.
The owner's existing lifecycle takes over successful uploads immediately; these
three endpoints do not return an unbound editing draft.

Deploy application and worker code together after the existing required
migrations through 426. This step adds no migration. Drain old writers before
switching: they still have an unreceipted write window. Keep the existing Outbox
worker and namespace-aware handlers enabled. New failed-attempt receipts can
become eligible by five minutes; a shorter request deadline may make them ready
earlier. Never delete a pending receipt to suppress a retained-reference failure.

Validation: 41 new local leaf cases passed. Eight related package targets passed
in full (3652 leaf cases passed; 340 database cases skipped with connection
variables cleared; domain/media has no tests). Disposable PostgreSQL passed all
58 cases without failures/skips: 15 new receipt/claim/cancellation/atomicity cases
and 43 existing ownership regressions. All 17 touched Go files passed formatting;
all 19 delivery files passed whitespace checks. Existing tests now distinguish
legacy returned-URL compensation from the new prepared upload entry points.

The disposable container, anonymous volume and exact temporary credential file
were verified and removed. Evidence is indexed in
media-center-immediate-owner-upload-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. No business database, existing uploaded file,
historical URL, original task, runtime configuration, commit, deployment or
running service was changed. Full backend ./..., the entire PostgreSQL matrix,
browser flows and real S3/OSS/CDN/production acceptance were not run.

This completes only the three immediate-owner upload entry points. Home visual
showcase uploads are a separate upload-then-save workflow and still need a
write-ahead receipt, atomic binding and abandoned-draft expiry. UGC upload/copy
and order-evidence uploads also retain the previously documented pre-result
crash window. Processes or remote operations surviving database-connection loss,
outage reconciliation, legacy snapshot admission, physical capacity, usage
indexing, historical storage governance, conversion and production acceptance
remain unfinished. This step does not complete the media-center redesign.

## Media-center follow-up 35: showcase draft write-ahead receipts, atomic save and expiry (2026-10-16)

Visual-showcase image uploads now allocate a unique exact key under
`visual-showcase/<set>/<locale>/upload-drafts/<UUID>.<ext>` and independently
commit a temporary cleanup Outbox receipt before writing bytes. This reuses the
existing `home_visual_tile` cleanup handler and does not add a table, migration
or second ownership ledger. Local/S3/OSS provide a restricted writer for this
namespace, with immutable public cache headers; an unsupported adapter cannot
fall back to uploading an unknown key. A single upload writes one object and
returns the same URL for original and thumbnail. Independent thumbnails use a
separate upload receipt.

An independent READ COMMITTED transaction locks namespace -> receipt and holds
both locks until synchronous provider IO returns. The receipt must still be
pending, never claimed, unchanged and within its bounded five-minute write
deadline. Provider responses must resolve to the preallocated key; source size
and SHA256 are verified by reading back the object. Success changes the receipt
to ready and schedules cleanup 24 hours after completion. Upload responses
include `draft_expires_at`; admin instructions now explain the save window.
Empty/wrong responses, write errors, cancellation or a failed ready transition
leave the original uploading receipt for the existing worker. A prior claim,
completion, manual retry, missing/expired receipt or changed payload fences a
writer waiting for the namespace before it sends any bytes.

Both single-card save and group replacement consume ready receipts in the same
transaction as the showcase owner and old-owner cleanup events. Original and
independent-thumbnail keys are deduplicated and locked in sorted order. Every
receipt must be ready, unclaimed and before its original draft expiry; postponing
the task's retry time does not extend that expiry. Failure restores all consumed
receipts and the previous owner. Successful save removes only these temporary
receipts; existing tasks remain untouched. Saved images use the existing owner
lifecycle and do not expire after 24 hours.

A new-format key without a receipt needs an existing active or soft-deleted
showcase owner row as evidence of a previous successful save, in addition to the
existing new-binding object checks. Readable bytes alone do not authorize a new
binding. This is deliberately conservative: single-card updates overwrite their
owner row, so an overwritten image with no other showcase owner history requires
a new upload even if another business's historical reference retains its bytes.
Legacy keys keep the existing compatibility rules. Existing local public-upload
authorization still requires a saved public owner or published business consumer;
ready receipts do not grant access. Remote bucket/CDN access remains governed by
the deployment's existing policies and was not tested against real providers.

The existing Outbox worker handles abandoned drafts when their AvailableAt is
due; no new scheduler is needed. It rechecks active owners and retained business
references under the namespace lock. Historical product placements, descriptions
and post content preserve the object and retry the original event. After the
last detachment the same event can complete. Worker claim transactions do not
hold a receipt lock while waiting for namespace/provider operations. Request
cancellation does not release the upload's locks before the provider returns.

Deploy application and workers together after the existing migrations through
426; this step adds no migration. Drain old writers before switching, keep the
existing Outbox worker enabled, and do not discard receipts to suppress a
retained-reference failure. Previously abandoned uploads are not retroactively
assigned receipts or expiry times.

Validation: 35 new local leaf cases passed. Eight backend packages passed in full
(3751 leaf cases passed; 353 database cases skipped with connection variables
cleared). A final focused run also passed after strengthening whole-payload
comparison. Disposable PostgreSQL passed 14 new concurrency/rollback cases and
70 existing ownership cases, without failures or skips. Admin behavior tests
passed 22 cases across four files; typecheck and production build passed.
All 10 touched Go files passed formatting, and all 14 delivery files passed
whitespace validation. The disposable container, anonymous volume and exact
temporary credential file were verified and removed. Evidence is indexed in
media-center-visual-showcase-draft-upload-completed-local-validation-summary-2026-10-16.md
in the external evidence directory.

Only local source and release preparation were changed. No business database,
existing uploaded file, historical URL, original task, runtime config, commit,
deployment or running service was changed. Full backend ./..., the complete
PostgreSQL matrix, browser flows and real S3/OSS/CDN/production acceptance were
not run. UGC upload/review-copy and order-evidence pre-result crash windows,
provider operations surviving process/database-connection loss, outage
reconciliation, complete snapshot admission, physical capacity, usage indexing,
historical storage governance, conversion and production acceptance remain
unfinished. This step does not complete the media-center redesign.

## Media-center follow-up 36: UGC upload and review-copy write-ahead receipts (2026-10-16)

UGC multi-image uploads and review publication now allocate all exact destination
keys and independently commit one existing cleanup Outbox event before any object
write or copy. The receipt covers unattempted destinations as well as partial
writes, using the existing `ugc_showcase_failed_upload` handler. This step adds no
table, migration or ownership ledger; migrations through 426 remain prerequisites.
Pending uploads retain their private namespace and store keys rather than public
URLs. Existing completed-order eligibility remains required, and per-user pending
limits are rechecked under the owner lock before the first upload writes bytes.

An independent READ COMMITTED transaction locks namespace -> receipt -> owner and
holds the locks until synchronous provider operations return. The receipt must
remain pending, never claimed, within its original write deadline (at most five
minutes), and have an unchanged whole JSON payload. A prior claim, completion,
manual retry, expiry, missing row or payload change fences a waiting writer before
upload/copy IO. Externally owned owner/Outbox transactions are rejected. Worker
claims skip rows held by active writers; cancellation does not release the locks
before the provider returns. OSS copies now receive the request context in all
bucket-routing branches, and local copies check cancellation while reading.

Uploads verify the returned exact key and read back source storage to compare size
and SHA256 against the multipart file. Publication plans destinations without
object IO, then rechecks the locked submission's pending status and ordered image
references before copying. Source and destination size/SHA256 must match. Each
distinct source gets a separate UUID directory, so a legacy public source and a
pending source with the same suffix cannot overwrite each other. Duplicate
references to the same source copy once and preserve display order; existing
approved objects are reused and are excluded from new-destination receipts.

Success atomically creates the pending owner, or updates the approved owner and
enqueues pending-source cleanup, while deleting only this request's temporary
receipt. Saved objects return to the existing owner lifecycle; this is not the
24-hour showcase-draft workflow. Any write/copy, integrity, cancellation, owner or
receipt-exchange failure keeps the original cleanup event and rolls back owner and
source-cleanup changes. There is no synchronous compensation deletion or second
failure task. Failed receipts become eligible at their original bounded write
deadline; worker scheduling and retained references can defer actual cleanup.
Cleanup rechecks current owners and historical product/content references. A
retained reference preserves the object and retries the same event until the last
detachment. Existing historical cleanup events are not consumed or rewritten.

Deploy application and workers together, drain old writers before switching, keep
the existing cleanup worker enabled, and preserve retained-reference failures for
retry. Previously abandoned uploads are not retroactively assigned receipts. This
step does not change provider access policies or pending-image authorization.

Validation: 33 new local leaf cases passed, including Local/S3/OSS adapter contracts
with local test endpoints. Eight relevant backend packages passed in full (3765
leaf cases passed; 372 database cases skipped with connection variables cleared).
Disposable PostgreSQL passed 18 new receipt/concurrency/rollback cases and 44
existing UGC, immediate-owner and showcase-draft cases, without failures or skips.
All 13 touched Go files passed formatting, and all 15 delivery files passed
whitespace validation. The disposable container, anonymous volume and exact
temporary credential file were verified and removed. Final evidence is indexed in
media-center-ugc-precommitted-upload-and-publication-completed-local-validation-summary-2026-10-16.md
in the external evidence directory. Exploratory logs, including an earlier OSS
copy-source header-encoding assertion failure, are retained separately; the two
final package/PostgreSQL logs validate the completed implementation.

Only local code and release preparation changed. No business database, existing
uploaded file, historical URL, original task, runtime configuration, commit,
deployment or running service was changed. Full backend ./..., the entire
PostgreSQL matrix, browser flows and real S3/OSS/CDN/production acceptance were not
run. Order-evidence uploads still need write-ahead receipts. Remote operations
surviving process/database-connection loss, outage reconciliation, complete
snapshot admission, physical capacity, usage indexing, historical storage
governance, conversion and production acceptance remain unfinished. This step
does not complete the media-center redesign.

## Media-center follow-up 37: order-evidence upload write-ahead receipts (2026-10-17)

Order-evidence attachment uploads now allocate the complete private key under
`order-evidence/<order>/<item>/owned-upload-<UUID>/` and independently commit an
existing order_evidence_failed_upload cleanup Outbox receipt before sending any
bytes. Local/S3/OSS reuse the exact-key private writer with private storage and
no-store metadata. An unsupported adapter cannot fall back to unknown-key uploads.
No table or migration is added; existing migrations through 426 remain required.

An independent READ COMMITTED transaction acquires the order-evidence namespace
lock, then the receipt row lock. Receipt status, previous attempts, original
deadline (at most five minutes), event type and whole JSON payload are checked
before provider IO. A claimed, completed, manually retried, missing, expired or
changed receipt prevents a waiting writer from uploading. Externally owned owner
or Outbox transactions are rejected. Active writers retain both locks until the
synchronous provider operation returns, including after request cancellation;
worker claims skip their locked receipt rows.

Inside the same namespace transaction, the upload rechecks the order/item scope,
planned package and item type, and current package mutability before writing.
Package locking or supersession while an upload waits therefore prevents bytes
from being sent. Existing signed-POD PDF and image validation remains in place.
The returned identity must match the preallocated key exactly; registration reads
source storage and verifies actual size, SHA256 and MIME before saving the owner.
The attachment and consumption of this request's temporary receipt commit
atomically. A failed owner write or receipt exchange restores the receipt and
leaves no committed attachment. Existing cleanup events are not consumed.

Write errors with no response, empty/foreign/wrong-key responses, missing or
corrupt bytes, cancellation and transaction failures keep the original exact-key
receipt. No synchronous compensation deletion or second failure event is issued.
The original task becomes eligible at the bounded write deadline; scheduling and
retained references may delay actual cleanup. The existing worker preserves active
attachments, detached immutable ownership and snapshot references. A removable
snapshot reference can release the same task after its last detachment. Detaching
an attachment does not authorize physical deletion of its immutable evidence.
Successfully registered attachments retain their existing evidence lifecycle and
do not receive a draft expiry. Prior abandoned uploads are not backfilled.

Deploy application and workers together after draining old writers. Keep the
existing Outbox worker enabled and retain reference-blocked events for retry.
Private authorization, immutable history and revision behavior remain governed by
the existing evidence rules. Source changes and validation are local release
preparation; production deployment and real provider acceptance remain pending.

Validation: 24 new local leaf cases passed, including Local/S3/OSS contracts using
local test endpoints. Eight relevant backend packages passed in full: 3786 leaf
cases passed and 383 database cases skipped with connection variables cleared.
Disposable PostgreSQL passed 11 new receipt/concurrency/rollback cases and 55
existing order-evidence, UGC, immediate-owner and showcase-draft cases, without
failures or skips. Nine Go files passed formatting and eleven delivery files passed whitespace
checks. The disposable container, anonymous volume and exact temporary credential
file were verified and removed. Final evidence is indexed in:
media-center-order-evidence-precommitted-upload-completed-local-validation-summary-2026-10-17.md
in the external evidence directory. Exploratory compilation and fixture failures
are preserved separately from the final passing package and PostgreSQL logs.

No business database, existing uploaded file, historical URL, original cleanup
task, runtime configuration, commit, deployment or running service was changed.
Full backend ./..., the entire PostgreSQL matrix, browser flows and real
S3/OSS/CDN/production acceptance were not run. No frontend source changed.

This closes the identified order-evidence upload pre-result receipt gap. Remote
provider operations that outlive process/database-connection loss, outage
reconciliation, remaining upload/snapshot entry-point inventory, physical capacity
and in-flight reservations, usage indexing, historical storage governance,
attachment conversion, private-video Range and production acceptance remain
unfinished. This step does not complete the media-center redesign.

## Media-center follow-up 38: remaining upload and snapshot interruption audit (2026-10-17)

This step inventories remaining entry points and reproduces gaps; it does not
implement their fixes or establish a passing release gate. The latest isolated
Go-overlay probes failed all six local leaf cases and all five disposable
PostgreSQL leaf cases. Each failure reached its intended protection assertion;
there were no build failures, fixture failures or skipped cases in these two
final audit logs. Eleven failed cases represent three failure mechanisms, not
eleven observed business incidents.

1. Permanent-library originals and public derivative generation can persist bytes
   before the provider returns an empty identity plus an error. The original has
   no asset or cleanup receipt; the derivative has no registered derivative or
   cleanup receipt while the original remains valid. Compensation starts from a
   returned URL, so it cannot recover the unknown object key in either probe.
   Public rebuild/backfill shares the derivative generator. Private conversation
   derivatives already have a bounded write lease and an asset-scoped retirement
   namespace scan; they are not equivalent to the unprotected public case.
2. Four immutable snapshot repository entry points admit a fresh reference after
   exact-object cleanup: initial Create, submission Create, submission
   CreateOrGetLatest, and export CreateOrGetForPackage. Four PostgreSQL races also
   commit the snapshot between the retention scan and provider deletion despite
   the cleanup namespace lock. These use constructed JSON payloads to test the
   repository boundary, not normal production HTTP requests. Current export
   attachment projections omit storage_key; copied DataJSON and other payload
   fields require a separate caller-level inventory before changing admission.
3. A controlled provider accepts bytes, then the disposable PostgreSQL writer
   connection is terminated. After the request deadline, the original cleanup
   receipt completes while the object is absent. The accepted write then finishes,
   owner registration fails, and a second worker pass has no pending task for
   the remaining object. This combines actual temporary PostgreSQL connection
   loss with simulated late remote completion, not a real S3/OSS or process-crash
   test. Extending a deadline alone cannot prove this recovery gap closed.

Static inspection also identifies direct uploads for warranty claims, after-sales
attachments, shipment evidence, suggestion-feedback attachments and PayPal
commercial-invoice PDFs. The three private handlers rely on returned URLs and
best-effort synchronous deletion before/after owner failure. Suggestion feedback
has a live authenticated upload route, returns a canonical URL without media
registration or a durable upload receipt, and saves attachment JSON later. PayPal
uploads the PDF before submission-snapshot persistence and only checks URL shape.
Its storage is wired to the general provider. The local public allow-list rejects
unregistered keys, so feedback/PDF access and ownership need end-to-end checks;
this audit does not claim a real PayPal download or browser failure was reproduced.

Next implement a bounded permanent-library original upload receipt and atomic
asset exchange, then public derivative receipts and the remaining caller-specific
upload contracts. Snapshot reference admission must share cleanup ordering and
reject missing objects without rewriting immutable history. Late accepted writes
need durable reconciliation or provider-confirmed completion, not a finite delay
presented as universal recovery. Preserve dedicated ownership boundaries.

Only documentation and external audit evidence changed in this step. No new
migration was added; the previous migration ceiling remains 426. The prior step's
passing package/provider-contract matrix was not rerun and does not turn these
audit failures green. The temporary PostgreSQL container, anonymous volume and
exact credential file were identity-checked and removed. No business database,
uploaded business file, historical URL, original task, runtime configuration,
commit, deployment or existing service was changed. Full backend, browser and
real-provider acceptance remain pending. Evidence index:
media-center-remaining-upload-and-snapshot-interruption-audit-findings-and-release-gaps-2026-10-17.md
in the external evidence directory. The media-center redesign remains incomplete.

## Media-center follow-up 39: permanent original write-ahead receipts and atomic registration (2026-10-17)

Permanent-library originals now allocate the complete unique object key under
`<trusted feature prefix>/preallocated-permanent-originals/<UUID>.<ext>` and
independently commit an exact-key cleanup Outbox receipt before provider IO.
General uploads use the root preallocated directory. Local/S3/OSS implement a
restricted public writer that rejects dedicated-owner, business-draft and private
attachment namespaces. Unsupported adapters fail before writing. Existing paths
are not rewritten. This covers public compatibility attachments routed through
UploadAsset; prepared private attachments keep their existing receipt lifecycle.
No migration is added; the existing migration ceiling remains 426.

The asset repository commits the receipt using its own database, then starts an
independent READ COMMITTED transaction with quota -> exact object -> receipt row
lock order. It rechecks quota, pending/unattempted status, event type, complete
payload and the original bounded write deadline (at most five minutes) before IO.
Worker claims skip active receipt locks; a claim or completion before the writer
obtains those locks fences the writer. Request cancellation does not release the
transaction locks before the synchronous provider call returns. Holding quota
locks through IO serializes uploads for the same quota owner and consumes a
database connection for the duration; rollout capacity must account for this.

The returned canonical URL must match the planned key. Reading source storage
then checks actual byte count, SHA256 and MIME using the same image/video sniffing
rules as upload validation. Asset creation and deletion of only this request's
temporary receipt commit together. Receipt/asset persistence failure, lost or
wrong response, missing/corrupt source and cancellation leave no new committed
asset and retain the original receipt after it has been committed. No second
failure event or synchronous compensation deletion is required. Receipt creation
failure prevents provider IO. Successful assets receive no draft expiry, and
existing cleanup tasks survive unchanged.

The original failed-upload task becomes eligible after its write deadline.
Existing shared-object ownership and retained-reference checks still govern
physical cleanup. A soft-deleted historical reference retains the object and the
same task until its last unbind. This closes the permanent-original lost-result
gap from step 38; it does not repair prior abandoned objects automatically.

Validation: 35 new local leaf cases passed, including Local/S3/OSS contracts
against local test endpoints. Nine related package targets completed: eight
passed and domain/media has no test files; 3637 leaf cases passed and 376 database
cases skipped with database connection variables cleared. Disposable PostgreSQL
passed 40 leaf cases (four new receipt/cancellation/worker/quota cases and 36
existing ownership/quota cases), with no failures or skips. Final logs and scope
are indexed in the external evidence directory under
media-center-permanent-original-receipt-completed-local-validation-summary-2026-10-17.md.
Exploratory failures remain separate from these final passing logs.

Deploy the application and worker together after draining old writers. Preserve
the existing cleanup worker and reference-blocked retry tasks. This is local
release preparation only: no business database, uploaded business file,
historical URL, original business cleanup task, runtime configuration, commit,
deployment or existing service was changed. The isolated PostgreSQL container,
anonymous volume and exact temporary credential file were verified and removed.
Full backend ./..., the entire PostgreSQL matrix, browser flows and real
S3/OSS/CDN/production acceptance were not run.

Public derivative generation/rebuild/backfill receipts, warranty/after-sales/
shipment/feedback/PayPal upload contracts, snapshot reference admission, and
remote accepted writes surviving process or database-connection loss remain
unfinished. A receipt and a finite deadline alone do not solve late remote
completion. Physical capacity, usage indexing, historical storage governance,
attachment conversion, private-video Range and production acceptance also remain
open. The next implementation slice is public derivative write-ahead receipts.
This step does not complete the media-center redesign.

## Media-center follow-up 40: public derivative write-ahead receipts (2026-10-17)

Public derivative generation, rebuild and backfill now plan immutable exact keys
under `media-derivatives/<asset>/<preset>/v<version>/preallocated-generations/<UUID>.<jpg|png>`.
All planned keys receive independently committed existing unregistered-derivative
cleanup Outbox receipts before provider IO. Local/S3/OSS have a restricted public
exact-key writer with the existing immutable cache policy. It rejects original,
dedicated-owner, draft and private namespaces; public generation has no generic
reader-upload fallback. Private conversation derivatives keep their bounded write
lease and private storage/cache policy. No table or migration is added; the
existing migration ceiling remains 426.

The batch then takes an independent READ COMMITTED transaction, locking the
active public parent, sorted exact object keys and receipt rows. It rechecks the
parent's original identity and permanent draft binding where applicable, plus
every receipt's original bounded deadline (at most five minutes), whole payload,
event type and unattempted pending state. Worker claims skip active receipt rows.
A claimed/completed receipt or retired parent fences a waiting batch before IO.
Request cancellation keeps these locks until the synchronous provider operation
returns. Parent locks and a database connection span conversion and provider IO;
deployment capacity must account for that duration.

The source is decoded once, all key receipts are committed, and each derivative
is then encoded, written and verified in sequence. Encoded payloads are not kept
for the whole batch. Returned canonical identity, actual size, SHA256 and MIME
must match the planned content. All new derivative registrations, replacement
tombstones, old-object cleanup events, product snapshots and consumption of this
batch's temporary receipts share the final transaction. A metadata, snapshot or
receipt-exchange failure rolls them all back and preserves old generations and
the original committed receipts.

Lost/empty/wrong responses, missing/corrupt bytes, partial batches and cancellation
retain every planned receipt, including keys that were never written. There is
no synchronous public compensation deletion and no second failure event. Failure
to commit the receipt batch prevents all provider IO. Ordinary conversion/write
failure retains the existing best-effort original-upload behavior; durable
receipt or metadata-exchange failures are reported while the registered original
remains intact. Successful retry uses new unique keys, consumes only its own
receipts and preserves earlier cleanup tasks. Historical references continue to
block the exact original task until its last unbind. Existing URLs and historical
generations are not rewritten; old abandoned objects are not backfilled.

Validation: 34 new local leaf cases passed, including Local/S3/OSS contracts
against local test endpoints. Nine related package targets completed: eight
passed and domain/media has no test files; 3671 leaf cases passed and 382 database
cases skipped with database connection variables cleared. Disposable PostgreSQL
passed 46 leaf cases (six new public-derivative concurrency/rollback cases and 40
existing original-upload, ownership, replacement/snapshot and quota cases), with
no failures or skips. Final scope and logs are indexed in the external evidence
directory under
media-center-public-derivative-receipt-completed-local-validation-summary-2026-10-17.md.

Deploy application and workers together after draining old public derivative
writers, retaining the cleanup worker and reference-blocked retry tasks. This
is local release preparation only. No business database, uploaded business file,
historical URL, original business task, runtime configuration, commit, deployment
or existing service was changed. The isolated PostgreSQL container, anonymous
volume and exact temporary credential file were verified and removed. Full
backend ./..., the entire PostgreSQL matrix, browser flows and real S3/OSS/CDN/
production acceptance were not run.

The step-38 public derivative lost-result gap is now covered locally. Warranty,
after-sales, shipment, feedback and PayPal upload ownership contracts, immutable
snapshot reference admission, and remote accepted writes surviving process or
database-connection loss remain unfinished. Finite write deadlines do not solve
late remote completion. Capacity accounting, usage indexing, historical storage
governance, attachment conversion, private-video Range and production acceptance
also remain open. Next inventory and implement the remaining business attachment
upload contracts. The media-center redesign remains incomplete.

## Media-center follow-up 41: warranty, after-sales and shipment upload receipts (2026-10-17)

All three HTTP upload endpoints now collect validated files and use the shared
private business attachment workflow. A whole batch receives independently
committed exact-key cleanup Outbox coverage before any provider write. Unique
attempt namespaces are warranty/claims/owned-upload-<UUID>,
after-sales/pending/owned-upload-<UUID>, and
warranty/shipment/<order>/owned-upload-<UUID>. Local/S3/OSS use the existing private
exact-key writer and private/no-store policy. No new migration is required; the
migration ceiling remains 426.

The independent READ COMMITTED transaction locks the private business namespace
and receipt row, checks the full payload, pending/unattempted state and original
bounded deadline, then tentatively persists authorized business ownership before
provider IO. Warranty challenge consumption, after-sales cases/items/attachments
and applicable notifications, shipment image merging, and receipt consumption
commit with successful canonical URL/key and source size/SHA256/MIME validation.
A storage, owner or final receipt-exchange failure rolls them back while retaining
the independently committed original receipt, including unwritten planned keys.
There is no synchronous compensation deletion or second failure task. Successful
retries use new keys and consume only their own receipt. After-sales enrichment
queries use the same transaction, including operator and refund-review lookups.

Worker claims skip active receipt rows. Claimed, completed, previously attempted,
expired, missing or modified receipts fence waiting writers before IO. Cancellation
keeps the namespace until the synchronous provider operation returns. All three
workflows share this namespace; database connections and namespace locks span
provider writes and reads, so concurrent throughput and timeout sizing must account
for that serialization. Shipment read/merge/update now runs under the same lock,
including the first optional shipment record. Direct service creation/update
cannot attach another preallocated upload through copied or escaped URL input;
the same already-bound shipment URLs remain editable. Upload input URLs and
attachment metadata are generated from server-selected files, and after-sales
file uploads recheck customer ownership before sending bytes.

Exact cleanup/reference scanning now includes shipment_records.shipping_images
and after_sales_case_attachments.storage_url, with no parent join that could
hide retained attachment rows. Existing warranty scans include soft-deleted
claims and video URLs. The cleanup checks every planned key before deleting any
bytes; retained ownership leaves the same original task retryable until final
detachment. Existing objects and historical URLs are not backfilled or rewritten.

Validation: 43 new local leaf cases passed (including nine Local/S3/OSS contract
cases against local endpoints). All nine related package targets passed: 3673
leaf cases passed and 384 database cases skipped with connection variables
cleared. Disposable PostgreSQL passed 56 leaf cases, including ten new receipt,
worker fencing, cancellation and concurrent shipment append cases plus 46
existing ownership/original/derivative/quota cases. There were no failures or
skips in that PostgreSQL run. Its container, anonymous volume and exact temporary
credential file were verified removed. Final evidence is indexed under
media-center-private-order-attachment-receipt-completed-local-validation-summary-2026-10-17.md
in the external evidence directory. Earlier fixture/exploratory failures remain
separate from the final passing logs.

Deploy application and cleanup workers together after draining old writers. This
step is local release preparation. No business database, existing upload,
historical URL, original business task or runtime configuration was changed;
there was no commit, deployment or existing-service restart. Full backend ./...,
the entire PostgreSQL matrix, browser flows and real-provider acceptance remain
outside this validation.

The next slice is authorized private attachment reading and UI delivery: warranty
and shipment currently return raw private URLs without a complete authorized read
route; customer after-sales responses expose attachment metadata without a
customer download route. Existing administrator after-sales reads are separate.
This upload slice does not close those read chains. Feedback attachments, PayPal
PDFs, snapshot reference admission, process/database-disconnection late remote
writes, capacity accounting, usage indexing, historical storage governance,
attachment conversion, private-video Range and production acceptance remain open.
The media-center redesign is not complete.

## Media-center follow-up 42: authorized private order attachment reads and UI delivery (2026-10-17)

Warranty claim files now use an authenticated/guest-authorized record selector at
GET /api/v1/warranty/claims/:id/attachments/:attachment_kind/:attachment_index.
Guest file access requires X-Warranty-Claim-Token in the request header; query
credentials are not accepted for files. Account ownership and the existing
verified-email/guest-token rules remain authoritative. Administrator claim files
use the equivalent /api/admin/warranty/claims path with product:view permission.
Historical warranty/<filename> references are supported from private storage;
shipment namespaces and deleted claims cannot be read through a claim selector.

Shipment images use GET /api/v1/warranty/orders/:order_number/images/:attachment_index
for the current order owner and GET /api/admin/warranty/shipment-records/:id/images/:attachment_index
for product:view administrators. The administrator id is the order ID. New
owned-upload keys must match that order; historical warranty/shipment references
remain readable from private storage. Customer warranty-status responses expose
authorized image paths. Administrator editing keeps the original stored references,
including their original array indices; preview and Blob URLs are never saved.

Customer after-sales files use GET /api/v1/orders/:order_number/after-sales/attachments/:attachment_locator.
The opaque locator hides internal case/attachment IDs and grants no access by
itself. The service checks current order ownership, order soft deletion and the
case/attachment parent relationship on every request. Create and list responses
now share the customer contract: metadata plus authorized download_path, with no
raw storage URL or internal identifiers. Paths use the already authorized order
number instead of a potentially empty order-number projection. Administrator
after-sales files retain order:view permission and use the same private reader.

Reads preserve the exact provider/private-container identity. Foreign same-path
URLs, public-bucket same-key aliases, traversal, userinfo and unexpected relative
paths are rejected. Local legacy /uploads locators resolve only within the separate
private root; public copies are never used as a fallback. Existing objects and
historical URLs are not moved or rewritten. ObjectOpener is required: signer-only
custom providers fail closed instead of redirecting a private file to storage.
Responses use private/no-store, nosniff, no-referrer and sandbox CSP. Downloads are
the default; explicit inline preview is restricted to supported media types.

Both applications now fetch attachments with their existing authenticated client
and render/download Blob URLs. The storefront keeps claim credentials in headers
and parses a Blob only on a successful response. Switching records or storefront
identity and unmounting cancels requests and releases temporary URLs. Missing
filename extensions are inferred from the content type. Video preview fetches the
whole file into browser memory and remains subject to browser codec support;
private-video Range remains open.

Validation: all 10 related Go packages passed, with 3709 local leaf tests passed
and 384 database leaf tests skipped after clearing test-process connection
variables. The new/updated read contracts cover 37 leaves, including guest token
binding, current ownership, historical references, wrong parents, exact Local/S3/OSS
container mapping, response headers and MOV/MP4/WebM byte/MIME preservation.
Administrator and Nuxt type checks passed; two historical image-index Vitest cases
passed. Isolated Playwright verification used the actual shared components and
storefront request client against synthetic authenticated endpoints: four decoded
image previews, four browser downloads, header-only credentials, record switching,
visible 403 errors and URL release passed. This was component-level browser QA,
not full business-page, real-video or real-provider acceptance. Evidence is indexed
in media-center-private-order-attachment-authorized-reading-completed-local-validation-summary-2026-10-17.md.
Exploratory fixture/command-ordering failures remain separate from final evidence.
The isolated browser and port 14842 server have been stopped.

This remains local release preparation. There is no new migration (highest 426),
no business-database access or real-upload/history/task/runtime mutation, no new
database container, and no commit, deployment or existing-service restart.
Feedback uploads, PayPal PDFs, snapshot admission, late remote write reconciliation,
full physical capacity and usage indexing, historical storage governance,
attachment conversion, private-video Range and production acceptance remain open.
The media-center redesign is not complete.

## Media-center follow-up 43: feedback and PayPal invoice durable upload ownership (2026-10-17)

Feedback images and new PayPal commercial invoices now allocate exact keys under dedicated owner namespaces, commit an Outbox cleanup receipt before provider IO, verify stored size/SHA256/MIME, and exchange ready receipts atomically for the feedback row or selected immutable submission snapshot. Uploads use a maximum five-minute write window and a 24-hour binding window. A claimed, attempted, expired or completed receipt fences the writer. Canceled requests retain the lifecycle lock until synchronous provider IO returns.

Feedback uploads and creation enforce current membership and the original uploader. New submissions reject legacy/unreceipted URLs, duplicate attachments and cross-user reuse; a previously uploaded but unsubmitted legacy attachment must be uploaded again. Existing saved feedback URLs are not rewritten. PayPal retries retain the selected snapshot, while unused concurrent invoices keep their original cleanup tasks. Both namespaces are excluded from generic media ownership. Cleanup retains soft-deleted feedback, immutable snapshots and PayPal submission audits until the last exact reference is removed.

Local /uploads access now recognizes these owners and returns private/no-store and nosniff. Ready feedback images can be previewed; new invoices require a selected snapshot before public retrieval. S3/OSS use the configured public container with private/no-store metadata; actual container/CDN behavior still needs deployment validation. This preserves the existing public attachment contract and does not migrate historical files into private storage.

New Stripe/PayPal submission snapshots use JSON key/whitespace normalization for stable integrity after PostgreSQL JSONB reads, preserving large integer IDs. Verification also accepts existing byte-exact hashes. Historical JSONB snapshots whose original text cannot be reconstructed still need compatibility review; no stored hashes were rewritten. The four generic snapshot reference-admission entry points remain pending.

Validation: 3749 local test leaves passed across nine tested packages; the outbox domain compiled with no tests. 391 environment-related leaves were skipped with database connection variables cleared. The isolated PostgreSQL suite passed seven new cases, including writer fencing, cancellation lock lifetime and invoice snapshot integrity/atomic binding. The temporary labeled tmpfs container and its exact credential file were removed. Local/S3/OSS provider contracts use mock HTTP endpoints. The final evidence index is media-center-feedback-paypal-final-validation-and-delivery-manifest-2026-10-17.json in the external evidence directory.

No new migration (highest remains 426), business database access, live file/URL/task/configuration changes, commits, deployment or existing-service restart. Remaining work includes legacy snapshot compatibility and admission, disconnected late-write reconciliation, physical capacity accounting, usage index, storage-location/history governance, customer attachment conversion, private video Range, release organization and production acceptance.

## Media-center follow-up 44: JSONB snapshot integrity and transactional media admission (2026-10-17)

Initial snapshot Create, submission Create, submission CreateOrGetLatest and export CreateOrGetForPackage now validate new managed media references within their insertion transaction. Shared media must remain active, public, eligible for reuse, physically present at the configured provider, and permanently bound when it belongs to a business draft. Missing verification capability fails closed. Dedicated objects require an authenticated retained snapshot from the same order or a retained order-evidence attachment belonging to that order. A shared registration cannot supply dedicated ownership. Existing immutable rows remain available for retries without admitting the unused candidate's references.

PayPal's storage-aware adapter binds only newly selected invoice receipts before inserting the selected snapshot, within the same transaction. Generic snapshot entry points cannot bypass that exchange. Dispute ownership is checked for new submissions; an order export can retain its own already-bound historical invoice without requiring a second receipt. Candidate failures roll back receipt consumption. The existing receipt lifecycle namespace is acquired before PayPal selection, other dedicated namespaces and sorted shared asset rows. Separate selection locks serialize the absence of a canonical submission/export row. WithTx retains configured storage identity verification.

New initial, submission and export snapshots use sorted JSON keys and exact decimal normalization before hashing, preserving large integers and surviving JSONB key, whitespace and exponent changes. Shared domain payload contracts preserve the known schema-one Go serialization order. Legacy verification reconstructs that serialization, checks the original stored digest, and separately verifies equality of the complete JSON content. Unknown or ignored fields and content changes cannot authenticate through partial reconstruction. No historical row or digest is rewritten. If historical formatting or numeric notation has been irreversibly lost and none of the supported complete encodings matches its original digest, verification still rejects it; such rows require read-only inventory and evidence review, not automatic rehashing.

Retained JSON reference scanning now also parses supported image, video, poster and file-link attributes inside rich-text strings. Snapshot admission and retirement therefore recognize the same embedded references. Retention protects bytes without granting public access.

Validation: 3801 local leaf tests passed across nine tested packages; the outbox domain compiled with no tests. 401 environment-related leaves were skipped with test-process database connection variables cleared. The isolated PostgreSQL matrix passed 21 leaves without skips: four legacy payload kinds, exact decimal JSONB integrity, both orders of binding/physical cleanup for all four entries, retained invoice export and the previous receipt fencing/cancellation cases. The labeled tmpfs container and its exact credential file were verified and removed. Results and delivery hashes are indexed in media-center-immutable-snapshot-final-validation-and-delivery-manifest-2026-10-17.json in the external evidence directory. No frontend changes or real cloud/CDN/browser acceptance are included.

This remains local release preparation: no new migration (highest 426), business-database access, real upload/history/task/runtime changes, commit, deployment or existing-service restart. Disconnected late-write reconciliation, full physical capacity and usage indexing, storage-location/history governance, attachment conversion, private-video Range, release organization and production acceptance remain open. The media-center redesign is not complete.

## Media-center follow-up 45: durable reconciliation of late remote writes (2026-10-17)

The existing Outbox scheduler now continuously rechecks surviving original object-cleanup and customer-service avatar-cleanup receipts after successful, unknown or exhausted dispatch. A successful absence check schedules another check one hour later; references, provider failures and missing handlers defer for fifteen minutes. Neither repeated absence nor an attempt count retires this reconciliation. Startup resumes the persisted schedule. Normal dispatch failures do not suppress the independent reconciliation pass. Payment and notification events, manually ignored tasks and consumed/deleted upload receipts are excluded.

Reconciliation reuses the original event identity, payload and exact keys with the existing domain ownership and retained-reference checks. It preserves the original status, attempts, processed timestamp, last attempt, error and update timestamp. Separate last-reconciled/error fields record its result; reconcile_after and locked_at/locked_by schedule and lease each pass. A unique token fences every pass, including overlapping passes in one process. PostgreSQL uses SKIP LOCKED and its own clock for lease timestamps/expiry; heartbeat and completion require the current token. Worker failure leaves a recoverable lease. Successful ownership exchanges never become writable upload receipts again.

Whole-asset cleanup can now replay its authorized exact keys after the registration and derivative rows have been hard-deleted. It reacquires the shared object locks and checks new registrations and retained business references, while refusing dedicated or foreign private namespaces. For retired private customer-service assets, the original authorization also retains the already authorized derivative namespace; later unregistered generations are listed within that namespace and checked for surviving exact claims before deletion. Feedback and PayPal cleanup revalidate the original receipt and reconciliation token under their lifecycle lock. Missing exact keys and conflicting historical metadata remain visible reconciliation errors requiring review; there is no URL-based reconstruction of deletion authority.

Apply migration **427_add_original_storage_cleanup_continuous_late_write_reconciliation.up.sql** before deploying the new worker. It adds two reconciliation-result columns and a partial due-work index without rewriting events, history, keys or writer fences. Rollback removes only the new result metadata/index and retains original receipts and reconcile_after. Keep original cleanup authorizations, including processed rows, in retention/backup plans: deleting them would remove the durable late-write ledger. Finite automatic purging is unsafe until the provider can prove no writer can still complete. Reconciliation uses the configured provider; historical storage-location changes still require the outstanding location/governance work. Enable the existing Outbox worker continuously and monitor reconciliation logs plus storage_cleanup_reconciliation_error; the hourly interval is a cadence, not a completion deadline or a hard latency guarantee under backlog.

Validation: 3825 distinct local test leaves passed across nine tested packages, plus the outbox domain compiled with no tests; 406 environment-related leaves were skipped with database variables cleared. This combines the related-package run with final focused checks after four additional local tests and the missing-key guard, and includes 43 new local leaves. The isolated PostgreSQL matrix passed 15 distinct leaves with no skips, including five new cases: actual writer-connection termination before delayed provider completion for logo/favicon/avatar, disjoint concurrent claims with database-clock lease recovery, and migration idempotency/rollback preservation. The five new PostgreSQL cases were rechecked after the final lease-context change. The exact labeled tmpfs container and credential file were verified and removed. Final evidence is indexed by media-center-original-cleanup-late-write-final-validation-and-delivery-manifest-2026-10-17.json in the external evidence directory.

This is local release preparation. No business database, real uploads/history/tasks/runtime configuration, commit, deployment or existing-service restart was involved. Real cloud/CDN and production acceptance remain pending. Full physical capacity accounting, usage indexing, storage-location/history governance, customer attachment conversion, private-video Range and release organization remain open; the entire media-center redesign is not complete.

## Media-center follow-up 46: read-only current-storage capacity audit (2026-10-17)

The media library now offers an explicit capacity audit. POST /api/admin/media/storage-capacity-audits requires both media:view and media:configure and returns only no-store aggregates. It observes current public/private storage and the independent logo provider. Local stat/walk and fully paginated S3/OSS listings read metadata without downloading bodies or changing storage. Shared physical identities and local nested paths are deduplicated; equal keys in different containers remain distinct. No second asset inventory, editable capacity ledger or deletion path was introduced.

A read-only REPEATABLE READ transaction snapshots original/derivative registrations including tombstones, surviving non-ignored object/avatar/customer-service-retention cleanup receipts, and unbound conversation/business-draft uploads. It ends before provider IO and neither locks business ownership nor skips active writer receipts. Observed originals, derivatives, cleanup-receipt-only objects and unmatched registrations form disjoint totals. Cleanup/upload receipts and soft-deleted registrations are separate overlapping annotations. Observed uploads are not counted again as reservations; unobserved known plans and unknown plan sizes remain separate. Neither unmatched objects nor cleanup receipt annotations imply permission to delete.

Apply **428_add_pending_customer_service_upload_expected_capacity_bytes.up.sql** before this code. It adds a nullable nonnegative expected-byte column to customer-service upload sessions. Existing NULL values remain unknown; rollback removes only this accounting column. Newly precommitted permanent/dedicated/order-evidence/private-order/UGC-original upload receipts also retain optional expected bytes inside their existing payloads. These fields never expand authorized object keys or change writer fences. Existing feedback/PayPal/visual/business-draft metadata is reused. Unencoded derivatives and unmeasured UGC copy plans remain explicitly unknown. Historical registration size zero also stays unknown. Conflicting keys/URLs, retired origins and unverifiable public/private container identities stay unresolved rather than being attributed to a current same-name object.

Limits are 100,000 total source rows, 200,000 object observations and a 45-second context, with one audit per process shared across transactional MediaService copies. A source-row overflow returns no capacity conclusion. Provider errors, pagination loops, duplicate objects, concurrent size changes, cancellation and observation limits produce a partial report. Unobserved registrations in incomplete locations are not declared absent. Cancellation stops subsequent registration-location resolution. Byte values use exact decimal strings in the API and BigInt formatting in the UI. Closing, disposal or permission revocation cancels the request and clears reports; late responses cannot restore them.

Coverage is current object content bytes at configured identities, observed over a time interval. This is not an atomic database/provider snapshot, allocated disk space or a cloud billing total. Historical locations, unrecognized endpoint aliases of one container, prior versions and incomplete multipart uploads still require location/provider governance. Cleanup receipts without persistent location IDs remain tied to the configured provider. Existing per-actor/session upload quota admission remains based on its current original-registration rules; this read-only scan does not replace that hot path. Keep normal cleanup ownership checks and original receipts unchanged.

Validation: the related Go run plus the final focused cancellation check passed 4,082 distinct local leaves across ten tested packages, with 427 environment-dependent leaves skipped and the outbox domain compiled. This includes 26 new local leaves. A separate labeled tmpfs PostgreSQL container passed 108 leaves without skips, including two new read-only snapshot/concurrent-registration and migration preservation/constraint tests. Admin default Vitest passed 86 tests, including 11 new tests; type checking, font/locale checks and production build passed. The playwright skill verified the real component/request logic and production CSS against synthetic capacity responses: six checks cover disjoint usage/reservations, incomplete results, mobile bounds, actual cancellation, failed-retry clearing and permission revocation. This is component-level browser validation, not live business authentication or real cloud acceptance. Initial fixture/old-assertion failures and exploratory browser setup checks are retained separately from successful evidence.

The exact temporary PostgreSQL container, credential file, browser session, Vite process and listener were verified removed. Evidence and delivery hashes are indexed by media-center-physical-capacity-final-validation-and-delivery-manifest-2026-10-17.json in the external evidence directory. No business database, live uploads/history/tasks/runtime configuration, commit, deployment or existing-service restart was involved. The following step adds bounded usage summaries; historical location governance, attachment conversion, private-video Range, release organization and production acceptance remain open. The overall media-center redesign is not complete.

## Media-center follow-up 47: bounded rebuildable usage summary and permission-safe reference presentation (2026-10-19)

The media library offers an explicit usage-summary dialog via POST /api/admin/media/usage-summaries, protected by media:view and the existing conversation-metadata scope. Candidates use the current partition, search, origin, type, status and customer-service-purpose predicates. One read-only REPEATABLE READ transaction reads candidates, unfinished business-draft receipts and the existing authoritative reference scanners, including historical derivatives and retained business rows. No migration or persistent usage ledger is added; migration 428 remains the latest required migration.

The dialog filters recognized-empty, cross-business, retained/pending, anomalous and restricted-reference rows. Counts overlap and share the same snapshot. Unconsumed draft uploads, including expired receipts still awaiting lifecycle processing, are separately marked and excluded from the recognized-empty filter. Each row previews up to 20 visible references; flags and totals include the remaining recognized references. Resource identities use decimal strings. The report explicitly covers recognized business sources only and is never deletion authority.

Limits are 200 candidate assets, 20,000 returned scanner references and a 30-second context, with one active summary per process shared across service copies. Overflow, invalid historical JSON, database/commit failure or cancellation returns no partial conclusion. Narrow the media filters before retrying an oversized scope. This does not bound every scanner's source-row reads; complex sources can time out. The implementation still reuses individual asset scanners inside an explicit bounded report. It is not the planned global persistent index or an assertion of large-library query performance; ordinary media cards do not trigger these scans.

Known resource types require their business view permission. Global ticket metadata additionally requires the existing full customer-service scope; admin/manager policy is unchanged. Customer/order sources without a record-level policy at this boundary remain restricted. Restricted rows retain their protection without exposing source, count, IDs or labels. Registration collisions show a generic conflict without another owner's identity. The existing references endpoint and in-use deletion responses now use the same presentation boundary: their total counts visible presentation entries, with one restricted_reference placeholder for any hidden set. Consumers must not interpret that value as a complete reference count. Authoritative deletion still uses the full unredacted reference set.

Ownership review returns visible counts and has_restricted_references; restricted references suppress ownership suggestions derived from incomplete visibility. Its write/audit/revision contract is unchanged. Summary links open verified existing business lists only; they do not invent unsupported record-ID routing or links to deleted/missing records. Anomalies are reported for review; no orphan unlink/recovery/repair action is delivered in this step.

Closing, disposal, authorization or identity changes, filter changes and asset refresh cancel and clear the UI report, including late responses. Edits elsewhere require an explicit rebuild; the snapshot time is displayed. Invalid or contradictory responses do not become unused conclusions. Full-library indexing/performance, direct authorized record positioning and reviewed anomaly repair remain outstanding, as do historical locations, attachment conversion, private-video Range, release organization and real-environment acceptance. No batch deletion is enabled.

Validation and delivery hashes are indexed by media-center-usage-summary-final-validation-and-delivery-manifest-2026-10-19.json and media-center-usage-summary-completed-local-validation-summary-2026-10-19.md in the external evidence directory. Browser checks use the actual component/composable and production CSS with synthetic endpoints and identities. This remains local release preparation; no business database, live uploads/history/tasks/runtime configuration, commit, deployment or existing-service restart is involved.

## Media-center follow-up 48: association diagnosis and reviewed orphan-gallery reattachment (2026-10-19)

GET /api/admin/media/assets/:id/association-diagnosis uses media:view, existing asset metadata authorization and the same business-reference visibility policy as the usage summary. One read-only REPEATABLE READ snapshot distinguishes missing parents, deleted history and registration collisions. Hidden references disclose only their existence; conflicting registration IDs remain redacted. At most 200 visible diagnosis entries are returned, with an explicit truncation flag; more than 20,000 scanned references, cancellation or database/commit failure returns no conclusion. Normal active references are outside this diagnostic list. Missing-parent gallery images that are themselves deleted are explicitly classified as retained history.

The first reviewed repair supports only a live gallery_images row whose old gallery is absent even with soft-deleted parents included. The operator selects an existing, active **draft** destination. The image must select this exact active public reusable original asset, with the exact registered original URL and either no thumbnail or that same original URL. Aliases, derivatives, independent thumbnails, private/customer/conversation/dedicated objects, unbound business-draft receipts, restricted references and conflicting registrations remain outside this repair. Products, swatches, workbench media, tickets, deleted images/parents and duplicate registrations are diagnosed but not repaired here. This does not recover the missing parent's metadata or claim that the selected draft was its original owner. Gallery cover/image scanning now reports nonpublished parents as retained_unpublished_parent, so a repaired draft remains in the retained usage filter until publication.

POST /assets/:id/missing-gallery-parent-repair-previews is read-only. POST /assets/:id/missing-gallery-parent-repairs requires media:configure and gallery:view/create/edit, a fingerprint of the current asset/image/destination snapshot, and a 3–1000 character review reason. Both inherit media:view. Authentication supplies audit identity. The mutation rereads every record, rechecks authoritative references and trusted exact object identity, verifies provider existence, and commits reattachment and audit together. Only gallery_id and normal update timestamps change; original URLs, thumbnail, order, descriptive metadata, classification and bytes stay intact. The audit stores persisted before/after image snapshots and the reviewed target/reason. A repeated or stale preview returns conflict without another mutation or audit. Audit failure or transaction cancellation rolls back the association and destination timestamp.

This rare manual operation uses PostgreSQL SHARE ROW EXCLUSIVE locks on galleries and gallery_images to protect the absence of the old parent against all ordinary insert/restore/update/delete writers. Acquisition/row-lock waits are limited to two seconds and the whole operation has a fifteen-second context. Ordinary gallery writers can briefly wait during review commit; lock failures require a fresh diagnosis. There is no bulk repair or physical cleanup. The image remains an authoritative reference throughout, including after attachment to a draft. Existing cleanup receipts are not reset, consumed or rewritten.

The media usage summary and asset reference detail both open diagnosis. The UI requires a target preview and explicit reason, clears previews on target changes, prevents duplicate clicks, and cancels/clears on close, disposal, permissions, identity, filters and list refresh. Uncertain mutation results require reinspection and audit review; aborting an HTTP request does not prove the server did not commit. There is no automatic resend or silent history removal.

Validation: 4151 distinct local Go leaves passed with 445 environment leaves skipped; ten tested packages passed and the outbox domain compiled. This combines the related-package run with final focused gallery/association/summary checks. The isolated PostgreSQL matrix passed 124 distinct leaves without skips, including 16 new cases; the final 15 repair cases were rechecked after the draft-state correction. Admin tests passed 134 cases (17 new), type/font/locale checks and production build passed, and 17 browser checks passed with desktop/mobile screenshots inspected. Nine Go files and all 21 delivery files passed formatting/whitespace checks. The exact step-48 tmpfs container, credentials, browser session and isolated server/listener were verified removed.

No new migration (highest remains 428). This is local release preparation only: no business database, actual uploads/history/tasks/runtime configuration, commit, deployment or existing-service restart was involved. Full-library usage indexing/performance, other anomaly repair contracts, historical storage identities and governance, public customer-service attachment retention/revocation, attachment conversion, private-video Range, release organization and production/original-visitor/real-cloud acceptance remain open. Final local verification and file hashes are indexed by media-center-association-repair-final-validation-and-delivery-manifest-2026-10-19.json in the external evidence directory. Browser evidence uses actual source components and production CSS with synthetic endpoints/identity.

## Media-center follow-up 49: dangling customer-service attachment owners remain protective (2026-10-19)

Authoritative attachment bindings now survive a missing message in the reference scan. A missing message is represented as ticket_message_attachment_binding using the existing binding ID, missing_parent state and no guessed conversation ID. Existing messages retain their actual ticket identity. Legacy attachment arrays, structured metadata, managed bindings and registered upload receipts now share batched parent-state inspection, including soft-deleted tickets. Global conversation metadata permission is required for the new binding type in diagnosis, usage summaries and reference details; other roles see only the existing restricted-reference indication.

Every registered customer-service upload receipt independently retains its asset, including bound/expired receipts and receipts with missing or mismatched bindings. Bound receipts are retained_upload_receipt, rather than sendable pending uploads. Usage summaries preserve missing/deleted parent states instead of overriding every receipt as pending; missing parents enter both anomalous and retained filters. Reference counts can therefore include both a message binding and its upload receipt: these are separate authoritative owners, not duplicate messages. Receipt status, keys, actor identities and URLs are not rewritten.

Unknown historical assets with managed binding/receipt owners continue to require conversation access even when message projections are absent. This read-only privacy projection does not infer a customer or restore message ownership. The existing cleanup worker rechecks the same full reference set: a dangling binding blocks a delayed original cleanup event; after the last unlink that same event can retry without a replacement payload. No new repair or purge endpoint is introduced. Product media still requires its own variant, primary-image, role and locale contract; ticket inbox restore is not recovery of a missing conversation.

No new migration (highest remains 428). Local Go, isolated PostgreSQL, admin tests, typecheck/build and complete delivery-file hashes are indexed by media-center-customer-service-dangling-reference-retention-step49-final-validation-and-delivery-manifest-2026-10-19.json in the external evidence directory. This step adds no browser acceptance claim. It remains local release preparation: no business database, live uploads/history/tasks/runtime configuration, commit, deployment or existing-service restart. The remaining scope listed after follow-up 48 is unchanged.

## Media-center follow-up 50: authorized remote video byte ranges with pinned object validators (2026-10-19)

The existing visitor/member/agent conversation attachment routes now support bounded original-video reads from S3 and OSS. Conversation ownership, current agent assignment, message membership and the registered exact object identity are checked before provider metadata access. Query parameters cannot override the object key, container or scope. This also applies when a caller supplies an old ETag: revocation is checked before HTTP preconditions. Images, local seekable files and video derivative selection retain their existing read paths.

Opening a remote video source performs HEAD only, verifies a strong ETag and unencoded representation, and fixes size and modification time for that request. GET sends an exact byte interval plus If-Match. Content-Range, response length, ETag and encoding must match before any video bytes or 200/206 status are exposed. Unverifiable metadata or provider responses fail closed with 503 and no provider error body. Invalid responses close the provider body; request cancellation stops reads. If a validated stream later terminates early, its advertised Content-Length lets the client detect truncation; the response cannot be replaced with 503 after bytes have been sent.

Single closed/open/suffix ranges, integer overflow boundaries, 416 and If-Range validators are supported, along with the standard conditional GET preconditions. Multiple ranges and unknown units are ignored and receive a full 200 representation; multipart ranges are not implemented. The shared responder can answer HEAD without downloading a body, but this step does not add HEAD registrations to all existing GET routes. Every response reaching the attachment handler remains private, no-store, varies by Cookie and uses nosniff. No public redirect, presigned URL or full-object buffering is introduced. Existing service consumers can still read the lazy source as a complete stream.

Validation uses synthetic S3/OSS HTTP providers and real SQLite-backed visitor/member/agent conversation authorization. It covers private/public bucket routing, precise conditional reads, provider mismatches, object changes, cancellation, short bodies, empty objects, large ranges and assignment revocation. The related backend package regression and complete delivery-file hashes are recorded in media-center-remote-private-video-range-step50-final-validation-and-delivery-manifest-2026-10-19.json in the external evidence directory. Initial test import/permission-context fixture failures are preserved separately. No frontend code changed; follow-up 49 frontend results are prior evidence and were not rerun here. No PostgreSQL environment was provisioned for this read-only change; environment-dependent database tests remain explicitly skipped.

Private video upload is still rejected with 415. The private upload preparation path remains image-only, and MEDIA_UPLOAD_CUSTOMER_SERVICE_PRIVATE_ATTACHMENTS_ENABLED stays default false. Real video encoding/playback, real-cloud and production/cache acceptance, video upload receipts/capacity/activation and original-visitor acceptance remain outstanding. This is local release preparation with no new migration (latest 428), business database, live object/history/task/runtime change, commit, deployment or existing-service restart. Other open media-center work remains as documented: independent business repair contracts, global usage performance/indexing, historical storage identity and registration governance, public attachment retention/revocation, conversion to operational material and release organization.

## Media-center follow-up 51: agent private-video uploads preserve receipt, binding and cleanup ownership (2026-10-19)

With MEDIA_UPLOAD_CUSTOMER_SERVICE_PRIVATE_ATTACHMENTS_ENABLED explicitly enabled, the authorized agent attachment upload route accepts videos using the existing 80 MiB MP4/MOV/WebM rule and 82 MiB request limit. It derives image/video type from file content, rechecks current conversation permission and carries that detected type through private upload preparation instead of hardcoding image. Zero agent identities are rejected even with a privileged scope. Visitor/member upload routes remain image-only. The switch and its example value remain false; disabled deployments retain the existing public compatibility behavior. No runtime configuration was changed.

A conversation-scoped receipt with an allocated private original key, actor identity, expected bytes and 24-hour expiry commits before provider IO. Video writes reuse the existing bounded write deadline, private-root/private-bucket routing and no-store metadata. Asset creation, serialized registered-original quota recheck and ready receipt exchange remain atomic. Videos retain their actual MIME, SHA-256 and byte size, with no image dimensions or image derivatives. Lost provider responses and failed asset registration preserve the original failed receipt/key for expiry cleanup; they do not create another key or discard recovery ownership.

The existing agent UI already posts to this conversation route and sends the returned preview URL. HTTP validation now covers video receipt creation, private Range preview, sending the current UI metadata/attachments payload, removal of the preview URL from persisted metadata, atomic message binding, preview invalidation after binding and authorized Range reads from the message. Before sending, the receipt and current uploader scope protect the video; after sending, the authoritative message binding and retained upload receipt both protect it. Public upload access remains denied. No frontend implementation changed, and no new browser playback claim is made.

Local failure tests verify that invalid/oversized uploads, insufficient quota, foreign support, zero identities and customer video attempts create no receipt or object. Registered pending videos keep consuming registered-original quota after retirement and failed physical cleanup, releasing it only after successful object/metadata cleanup. Failed unregistered writes remain traceable, including expected bytes, but are not registered quota usage. This step does not introduce full physical-capacity reservation/admission, transcoding or video posters, codec validation, or customer video upload.

An independently labeled tmpfs PostgreSQL container validates video binding versus expiry in both lock orders using pg_blocking_pids, and concurrent video registrations against one actor quota. Only one registration succeeds; the losing physical write keeps its failed receipt. The relevant existing customer-service, actor-quota and late-write reconciliation cases also pass. The container and its exact credential file were verified removed. Local Go/config/upload/storage/HTTP regressions, preserved initial observer-fixture failure and complete delivery hashes are indexed by media-center-customer-service-agent-private-video-upload-step51-final-validation-and-delivery-manifest-2026-10-19.json in the external evidence directory. Frontend build/tests were not rerun; follow-up 49 remains prior frontend evidence.

No new media migration; media requirements remain through 428. Verify 420/421/428 and the existing expiry/cleanup workers before any future activation. An unrelated carbon-rim product/variant specification migration 429 appeared in the shared workspace during finalization; it is outside this step and is not covered by these validation results. This remains local release preparation: no business database, live upload/history/task/runtime change, commit, deployment or existing-service restart. Real-cloud/browser/production and original-visitor acceptance, full physical quota admission, global usage indexing/performance, historical storage identity/registration governance, independent business repair contracts, public attachment retention/revocation, conversion to operational material and release organization remain outstanding.

## Pre-Launch Checklist

- `go test ./...` passes.
- Database migrations are reviewed and reversible where practical.
- Logs do not print secrets or sensitive customer data.
- Cookie, CSRF, CORS, and WebSocket origin settings are configured for the real domains.
- Upload endpoints enforce type, size, and storage limits.
- Payment callbacks are verified and idempotent.
- Backups and restore drills exist for PostgreSQL and object storage.
