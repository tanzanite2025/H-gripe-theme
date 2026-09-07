# FAQ management source of truth

FAQ is organized by page only. Each page owns a flat ordered `items` list:

```text
{ page_id, title, subtitle, items[] }
```

There is no FAQ category table, category field, category API, category route, or
legacy category fallback. This project has not shipped, so old category-shaped
data is intentionally unsupported.

## Current ownership

- `faq_pages`: page metadata including `page_id`, `route_path`, `domain`,
  `locale`, `title`, `subtitle`, `sort_order`, and `status`.
- `faqs`: FAQ items. `page_id` is the only structural relationship; `order`
  controls the item order within a page.
- `cmd/import/faqs`: imports page-owned items and does not accept category data.
- Admin `/faqs`: manages FAQ pages and flat FAQ items.
- Nuxt `PageFaq`: renders the page title and a flat list of FAQ items.
- Nuxt `PageFaqSlot`: resolves the current route to a page and inserts that
  page's FAQ block.

## SSR and publishing flow

FAQ content is stored in the Go backend database. Nuxt reads the public FAQ
page endpoint during SSR, so the rendered page does not depend on a static FAQ
registry or local fallback data.

The normal publishing flow is:

```text
Admin publish
  -> Go FAQ service writes the database
  -> purge Nuxt HTML cache
  -> next SSR request reads the page and its items
```

## Admin responsibilities

- `web/admin/src/views/FAQs.vue`: composes the page.
- `web/admin/src/composables/faq/useFaqStructure.ts`: page structure state.
- `web/admin/src/composables/faq/useFaqEditor.ts`: item editing state.
- `web/admin/src/components/admin/faq/FAQPageEditorDialog.vue`: page metadata.
- `web/admin/src/components/admin/faq/FAQEditorDialog.vue`: FAQ item content.
- `web/admin/src/components/admin/faq/FAQFilterPanel.vue`: list filters.
- `web/admin/src/api/faq.ts`: the admin FAQ HTTP protocol entry point.

The admin may filter the list by page. That is page ownership, not an FAQ
classification layer.

## Go responsibilities

- `internal/service/faq_service.go`: item CRUD, sorting, search, and deletion.
- `internal/service/faq_admin_structure.go`: page structure and page ownership.
- `internal/service/faq_public.go`: public page lookup and response assembly.
- `internal/service/faq_service_types.go`: FAQ page and item DTOs.
- `internal/repository/faq_items_repository.go`: item queries.
- `internal/repository/faq_structure_repository.go`: page queries.

All FAQ APIs return page-owned flat items. No handler, repository, service, or
request accepts a category argument.

## Nuxt storefront responsibilities

- `app/components/PageFaqSlot.vue`: route-based page lookup and insertion.
- `app/components/PageFaq.vue`: page title, item list, and empty state.
- `app/composables/usePageFaq.ts`: loading, expansion, item limits, and
  `hasMoreItems`.
- `app/components/FaqAnswerContent.vue`: answer and optional image rendering.
- `app/data/faq/backend.ts`: backend FAQ requests and normalization.
- `app/data/faq/index.ts`: FAQ data-layer exports.

FAQ aggregation pages may keep a page selector because it selects page
ownership. They do not expose or emulate item categories.

## Answer content boundary

- `faqs.answer` is the cleaned lightweight HTML answer.
- `faqs.answer_image_url` is the optional FAQ answer image.
- Each FAQ item may have at most one answer image.
- The backend remains the final validator for answer HTML and image constraints.

## Migration rule

The versioned SQL migrations are the source of the initial database shape.
They create `faq_pages` and page-owned `faqs` items only. They do not create
`faq_categories`, the `faqs.category` column, category indexes, category seed
data, or compatibility migrations.

When adding a storefront page that needs FAQ content:

1. Add or update its `faq_pages` record.
2. Add page-owned `faqs` items with the intended `order`.
3. Update the route resolver if the page uses a dynamic route.
4. Keep translations in backend FAQ records rather than restoring static
   category-shaped data in Nuxt.
