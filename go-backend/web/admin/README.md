# Tanzanite Admin Console

Vue 3 + Vite admin console for the Go backend.

## Stack

- Vue 3
- Vite
- Pinia
- Vue Router
- Element Plus
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
|   |-- components/   # shared components
|   |-- router/       # route definitions and guards
|   |-- stores/       # Pinia stores
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

## Related Docs

- Backend guide: `../../README.md`
- Root project guide: `../../../README.md`
- Dark theme guide: `DARK_THEME_GUIDE.md`
