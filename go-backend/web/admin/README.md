# Admin Console

Vue 3 + Vite admin console for the Go backend.

## Stack

- Vue 3
- Vite
- Pinia
- Vue Router
- Tailwind CSS 4
- shadcn-vue (Nova style)
- Reka UI
- Lucide icons
- Axios

## Local Development

```powershell
cd go-backend/web/admin
npm install
npm run dev
```

Default local address:

- Admin app: `http://localhost:9300`
- Backend API base: `/api/admin`
- Local backend: `http://localhost:9200`

## Build

```powershell
npm run build
```

Generate a production bundle report with `npm run build:analyze`. It preserves production mode and font checks, and writes `output/bundle-analysis/admin-client-treemap.html` outside `dist/`. Using `--mode analyze` would change environment-file loading, including `.env.production`.

The installed admin toolchain currently uses Vite 8 / Rolldown, while the Nuxt storefront builder uses Vite 7 / Rollup. Check the actual builder resolution before sharing tree-shaking options or plugins. See the [storefront loading and production bundler architecture](../../../nuxt-i18n/config/storefront/storefront-loading-hydration-and-production-bundler-architecture.md) for version-specific behavior and upgrade checks.

## Project Map

```text
web/admin/
|-- src/
|   |-- api/          # API clients
|   |-- assets/       # static assets
|   |-- components/
|   |   |-- admin/    # reusable admin composition components
|   |   `-- ui/       # shadcn-vue registry components
|   |-- modules/      # feature/domain logic, API wrappers, types, pure helpers
|   |-- router/       # route definitions and guards
|   |-- stores/       # Pinia stores
|   |-- styles/       # Tailwind theme and design tokens
|   |-- utils/        # HTTP/client helpers
|   `-- views/        # route entry pages and page-level orchestration
|-- public/
|-- index.html
|-- vite.config.ts
`-- package.json
```

## Layer Boundaries

- `src/modules/<feature>/` owns reusable domain logic: types, enums, API wrappers, route builders, pure helpers, and feature constants.
- `src/views/` owns route-level pages. It can keep page-local state, loading, filters, and submit handlers, but should not duplicate reusable domain logic.
- `src/components/admin/` owns reusable admin composition components such as tables, dialogs, editors, and panels.
- Use `selection-configuration` as the canonical feature name. Do not introduce new `selection-config` folders.

## API Rules

- Admin requests should use the shared Axios client.
- Admin API paths are under `/api/admin`.
- Cookie-authenticated unsafe requests must include CSRF headers where required by the backend.
- Do not duplicate backend pricing, payment, or permission business rules in the frontend.

## UI Architecture

- Generated primitives stay in `src/components/ui`.
- Product-specific composition belongs in `src/components/admin`.
- Shared design tokens live in `src/styles/admin.css`.
- Page views use the shared admin table, form, dialog, status, statistics, and pagination patterns.

## Related Docs

- Admin global search boundaries and extension guide: [admin-global-search-module-boundaries-and-extension-guide.md](admin-global-search-module-boundaries-and-extension-guide.md)
- Backend guide: `../../README.md`
- Root project guide: [README.md](../../../README.md)
- Frontend bundle measurements and optimization history: [bundle analysis report](../../../frontend-bundle-tree-shaking-analysis-and-safe-optimization-guide.md)
