# Brand Rename Continuity Plan

This note defines how to finish removing legacy `TANZANITE` naming while keeping the current production stack online.

## Phase 1 Rule

- `learn.gripe` is the canonical public domain.
- Active deployment files must not contain `tanzanite.site`, `admin.tanzanite.site`, `www.tanzanite.site`, or `tanzanite-edge`.
- Do not replace the production project, database, Redis, or uploads volume just to remove old names.
- Keep the live gateway, Caddy, and Nginx paths running during the rename.

## What Must Stay Running

- The existing Hostinger VPS deployment.
- The shared edge gateway.
- The storefront, admin, and API services.
- The current health checks and cache purge flow.

## What Can Be Cleaned Now

- Default domain values in production compose and env templates.
- Edge hostname entries in Caddy and Nginx.
- File names and route fragments that still carry the old brand.
- Internal docs that describe the active deployment boundary.

## Phase Gates

1. Phase 1 passes when the deployment verifier reports no legacy public hostname references and `learn.gripe` serves normally.
2. Phase 2 can then retire the remaining artifact names and archived references that still carry `TANZANITE`.
3. Backend operations start only after Phase 1 and Phase 2 are both stable.

## What Must Wait

- Old-domain redirects and aliases should stay until the new domain is stable.
- Historical rollback notes can still mention `tanzanite` if they explain migration history.
- Backend admin tooling for cloud/provider control should be a later phase, not part of the rename cutover.

## Safe Order

1. Update canonical domain values to `learn.gripe`.
2. Deploy and verify `GET /`, `GET /healthz`, and `GET /api/v1/settings/site`.
3. Switch edge and Nginx hostnames to the new domain only.
4. Rename deployment files so the old brand does not remain in live entry points.
5. Keep legacy redirects or aliases only for the observation window.
6. Remove the old names from active routing after the new domain is stable.

## Acceptance Criteria

- Production serves `learn.gripe` without breaking the current stack.
- `/api/v1/settings/site` returns the current public brand and contains no
  storefront URL under the legacy domain.
- No active deployment file still depends on `tanzanite.site` or `tanzanite-edge`.
- Rollback remains possible during the observation window.
- The rename does not replace the database, Redis, uploads, or application
  data; persisted public settings are updated by an idempotent migration.

## Out Of Scope For Now

- A full admin panel for Cloudflare, Hostinger, and deployment orchestration.
- One-click provider login or auto-provisioning.
- Replacing the current release flow before the rename is finished.
