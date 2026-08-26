# URL Management Domain

## Status

Approved architecture and active implementation record. This document is the
source of truth for moving the storefront route catalog out of SEO and turning
it into an operational URL management domain.

Implemented:

- URL Management is a top-level Admin domain; the old SEO URL catalog route and
  API have been removed.
- Managed redirect rules are stored separately from route observations and are
  consumed by the storefront runtime.
- SQL migration `168_storefront_url_issues` adds persistent URL issue and
  immutable issue-event records.
- Catalog sync and route checks project detected failures into durable issues.
- The Issue Queue is a real lifecycle view with acknowledgement, self-claim,
  comments, redirect linking, resolution, suppression, recheck, and
  verification actions.

Still planned:

- generated typed static route registry;
- redirect policy version history, rollback, and runtime observation;
- source-owner deep links and cross-admin assignment;
- redirect publication events automatically linked to every affected issue.

## 1. Purpose

URL Management owns the operational lifecycle of storefront paths:

- route discovery and route inventory;
- redirect policy and publication;
- route health checks;
- duplicate and canonical diagnostics;
- triage, assignment, resolution, and verification of URL incidents.

It is not an SEO editor, a content editor, a product editor, or a domain/DNS
management screen.

The primary goal is to make every detected URL problem actionable. A list row
or a failed check is evidence, not a completed workflow.

## 2. Administration Information Architecture

URL Management is a top-level Admin domain. It uses the existing Admin sidebar
for the domain entry and the existing tab bar for its internal views.

```text
URL Management
|- Overview
|- Route Catalog
|- Issue Queue
|- Redirects
|- Canonical and Collisions
`- Sync and Checks
```

`SEO` remains responsible for Home, Articles, and Products. It must not retain
a second URL catalog entry after this migration.

The root route redirects to `Overview`. Each tab has a stable, shareable URL:

```text
/urls/overview
/urls/catalog
/urls/issues
/urls/redirects
/urls/canonical
/urls/operations
```

The Admin API moves from `/api/admin/seo/routes` to `/api/admin/urls`. The old
Admin route and API are removed atomically; there is no compatibility route
because the system has not entered production operation.

## 3. Ownership Boundary

| Fact or action | Owner |
| --- | --- |
| Product and blog canonical routes | Product and Content source domains |
| Static canonical paths, static aliases, and tab path declarations | Storefront route registry |
| Page-level redirect rules created for migrations | URL Management |
| Route snapshot and route check evidence | URL Management |
| Canonical HTML rendering | Storefront SEO renderer |
| Article canonical exception | SEO Article workflow |
| Domain-to-domain redirects, DNS, TLS, edge host bindings | Operations Domain Binding |
| URL issue lifecycle and audit trail | URL Management |

URL Management may link to a source owner and offer the exact next action, but
it must not create shadow copies of product slugs, article content, or
arbitrary canonical HTML values.

## 4. Four-Layer Data Model

### 4.1 Observation

`storefront_route_catalog_entries` remains a synchronized route snapshot.
`storefront_route_check_results` remains append-only check evidence.

These records are derived facts. Synchronization can overwrite the catalog
snapshot, and a health check can update the latest check projection. Neither
table stores human acknowledgement, assignment, a resolution note, or a manual
redirect decision.

### 4.2 Policy

`storefront_redirect_rules` is the durable source of truth for Admin-managed
page redirects. A rule records:

- exact source path and locale policy;
- internal target path or approved absolute target;
- status code;
- enabled state;
- owner type (`managed`, `migration`, or `system`);
- publication state and policy version;
- change reason and audit metadata.

Managed rules cannot shadow an active canonical route or a system-owned static
alias. This prevents a manual rule from hiding a route collision.

### 4.3 Work Item

`storefront_url_issues` is the deduplicated lifecycle record for a detected
problem. It stores the route entry, issue type, severity, current state,
assignee, resolution type, and the latest evidence reference.

`storefront_url_issue_events` stores the immutable timeline: detection,
acknowledgement, claim, comment, redirect linkage, resolution, suppression,
and verification result.

The current projection rules are:

- catalog synchronization detects `path_collision` and `stale_route`;
- route checks detect redirect, HTTP, canonical, and probe failures;
- an observation can create a new issue or reopen an already verified/resolved
  issue;
- a healthy observation never closes an issue automatically;
- only an explicit resolution followed by verification can move an issue to
  `verified`;
- suppression requires both a reason and a future review time.

### 4.4 Delivery

Redirect policy is published as a versioned storefront route-control snapshot.
The storefront runtime is the only page-level redirect executor. Operations
may distribute the compiled policy to edge infrastructure, but Operations does
not own the redirect records.

The UI must always show whether a rule is draft, published, observed, or
drifted. A successful database write alone is not a successful redirect
change.

## 5. Issue Taxonomy and Required Resolution

| Issue type | Detection | Required resolution |
| --- | --- | --- |
| `redirect_chain` | Alias reaches expected target after more than one redirect | Replace with a direct redirect, publish, then recheck |
| `redirect_target_mismatch` | Alias reaches a different target | Correct the rule or source route owner, publish, then recheck |
| `redirect_status_mismatch` | Alias has an unexpected redirect status | Correct expected permanent redirect behavior, publish, then recheck |
| `not_found` | Route returns HTTP 404 | Restore the active source, or retire it in its source domain; create a managed redirect only after the next sync marks the old path stale |
| `server_error` | Route returns HTTP 5xx | Hand off to runtime/deployment diagnostics, then verify by recheck |
| `canonical_mismatch` | HTML canonical differs from expected canonical path | Open the source owner that controls canonical output, apply the owner fix, then recheck |
| `path_collision` | Multiple active sources claim one locale/path | Select the retained source and change or retire the other source; redirects do not resolve an active collision |
| `stale_route` | Previously cataloged route disappears from the latest snapshot | Restore source, redirect it, or explicitly retire it |
| `check_error` | Probe cannot complete | Retry or resolve the underlying network/runtime failure |

Alias behavior is not automatically an incident. A legacy alias is healthy only
when it reaches its declared canonical target with the declared status code and
within the allowed redirect-hop policy. The normal permanent-alias policy is a
single hop to a 200 canonical URL.

## 6. UI Behavior

### Overview

Shows operational counts, oldest unverified issues, recent policy publication,
and direct links into filtered queues. It does not duplicate the full catalog.

### Route Catalog

Shows every discovered route and its source, expected behavior, latest
observation, and related issue state. It is the inventory view, not the primary
place to resolve work.

### Issue Queue

Shows durable issue records, with active work (`open`, `acknowledged`, and
`resolved`) as the default queue. Every row has:

- evidence;
- expected versus observed behavior;
- owner;
- prescribed action;
- assignee and lifecycle state;
- a link to the source owner or redirect editor;
- recheck and verification controls.

### Redirects

Shows policy records rather than inferred check rows. It supports create,
disable, and publish today. Editing, rollback, policy versions, and runtime
publication observation remain planned work.

### Canonical and Collisions

Shows only canonical mismatches, duplicate paths, and route ownership
conflicts. Canonical values remain read-only except where the source domain
already provides a constrained canonical exception workflow.

### Sync and Checks

Runs discovery and checks, displays their execution history, and explains
which evidence generated or verified issues. It does not silently close an
issue without a successful verification check.

### Runtime Access Boundary

`STOREFRONT_BASE_URL` is the public canonical origin. It is used for
canonical links, public URLs, and external site-quality workflows.

`STOREFRONT_INTERNAL_ORIGIN` is the backend-to-storefront origin used by URL
catalog synchronization and route checks. In production it must point to the
private Compose service, normally `http://storefront:3000`. This keeps manifest
reads and route probes away from Cloudflare or another public edge challenge.

The backend must not use `STOREFRONT_BASE_URL` as a substitute for the internal
origin during synchronization. A public edge returning a challenge or access
policy response is an infrastructure boundary result, not a missing or broken
storefront route.

## 7. Storefront Route Registry

Static route facts currently exist in several places:

- `nuxt-i18n/public/storefront-route-manifest.json`;
- `nuxt-i18n/config/storefront/route-rules.ts`;
- redirect page components such as `app/pages/faq.vue`;
- `nuxt.config.ts` tabbed route declarations;
- `app/utils/pageSubNavigation.ts`.

They must converge into one typed storefront route registry. The registry
defines canonical static routes, aliases, route metadata, and tab child paths.
It generates:

- Nuxt tab child routes;
- static build-time redirect rules;
- the public route manifest;
- the static portion of the URL catalog;
- static-route validation fixtures.

The public JSON manifest is a generated artifact, not a hand-maintained
parallel source.

All locale expansion derives from the fixed shared storefront locale registry.
No URL Management page, route type, or route generator may add a local language
array.

## 8. Source-Owned Versus Managed Redirects

System-owned aliases are defined by the storefront route registry. They are
visible in URL Management but cannot be edited as ad hoc overrides.

Managed redirects are created for retired paths and approved migrations. They
must use an exact source path, cannot mask an active canonical route, and must
be published through the route-control delivery workflow.

This split ensures that an issue such as `/faq -> /support/faqs` is resolved at
the registry when it is a static alias, while a retired product slug can be
resolved by a managed migration redirect without editing application code.

## 9. Audit and Verification

Every policy and issue mutation records:

- actor;
- request path and source address;
- before and after values;
- reason;
- linked route entry and issue;
- publication version when relevant.

An issue reaches `resolved` only after a corrective action is recorded. It
reaches `verified` only after a later successful check matches the required
policy. Suppression requires a reason, expiry or review date, and remains
visible in history.

## 10. Migration Plan

1. Completed: move the existing route catalog API and UI from SEO to URL
   Management.
2. Completed: add the top-level Admin domain and the six route tabs.
3. Completed: retain catalog snapshots and check history while adding managed
   redirect policy through SQL migration `167`.
4. Completed: add issue projection and lifecycle records through SQL migration
   `168`, without changing source-owned canonical behavior.
5. Next: converge static route definitions into the typed registry and
   generate the manifest.
6. Next: add redirect policy versioning, rollback, and runtime publication
   observation.
7. Completed: remove the old SEO URL routes and duplicate Admin URL catalog
   entry.

No step may retain a second editable URL source of truth.
