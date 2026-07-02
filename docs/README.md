# Tanzanite Documentation

This directory is a project-level documentation hub. It is not the source of truth for runtime behavior; the current code and root `README.md` win when documents disagree.

## Current Entry Points

- Project overview: `../README.md`
- Backend guide: `../go-backend/README.md`
- Backend API notes: `../go-backend/API.md`
- Backend quick start: `../go-backend/QUICK_START.md`
- Backend deployment notes: `../go-backend/DEPLOYMENT.md`
- Backend module notes: `../go-backend/docs/`
- Admin console guide: `../go-backend/web/admin/README.md`
- Storefront notes: `../nuxt-i18n/docs/notes/`

## Archive

- `archive/project/` contains old project status, completion, deployment-ready, payment, storage, monitoring, and fix reports.
- `archive/backend/` contains old backend security, quality, frontend-page, and completion reports.
- `archive/audit/` contains old audit and optimization reports.
- `archive/refactoring/` contains old handler refactoring completion reports.
- `archive/bugfix/` contains old bugfix reports.
- `archive/splitting/` contains old file-splitting reports.
- `archive/optimization/` contains old optimization and cleanup reports.

Archived files are historical context only. They should not be used to claim production readiness, feature completeness, benchmark numbers, or current architecture.

## Maintenance Rules

- Keep active docs short, factual, and tied to current code.
- Move one-off completion reports into `archive/` after the work is done.
- Avoid claims like "production ready", "100% complete", or exact performance gains unless they are backed by current tests or measurements.
- Prefer one source of truth for each area: backend docs under `go-backend/`, storefront notes under `nuxt-i18n/`, project-level docs under `docs/`.
- Remove legacy WordPress compatibility docs unless they describe an explicit migration-only tool.

Last updated: 2026-07-02.
