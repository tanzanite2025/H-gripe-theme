# Tanzanite Admin Console

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

- Admin app: `http://localhost:3000`
- Backend API base: `/api/admin`
- Local backend: `http://localhost:9000`

## Build

```powershell
npm run build
```

## Project Map

```text
web/admin/
|-- src/
|   |-- api/          # API clients
|   |-- assets/       # static assets
|   |-- components/
|   |   |-- admin/    # admin-specific composed components
|   |   `-- ui/       # shadcn-vue registry components
|   |-- router/       # route definitions and guards
|   |-- stores/       # Pinia stores
|   |-- styles/       # Tailwind theme and design tokens
|   |-- utils/        # HTTP/client helpers
|   `-- views/        # page-level views
|-- public/
|-- index.html
|-- vite.config.js
`-- package.json
```

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

- Backend guide: `../../README.md`
- Root project guide: `../../../README.md`
