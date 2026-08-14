# Breadcrumb and page sub-navigation current source

Last updated: 2026-07-24

This note defines how Nuxt storefront breadcrumbs should represent pages that use slash child routes for in-page tabs, such as `/guides/tireguides/choose`.

## Current problem

Some pages are one canonical page component with several tab routes:

- `/guides/tireguides/choose`
- `/support/warranty/returns`
- `/company/about/factory`

The tab segment is a Nuxt child route that reuses the owning page component. Breadcrumbs include the child route segment, so `/guides/tireguides/choose` renders as:

```text
Home / Guides / Tire Guides
```

That is technically correct for the route, but it hides the active tab and makes direct links feel confusing.

## Current source of truth

The single source for third-level page tabs is:

- `nuxt-i18n/app/utils/pageSubNavigation.ts`

It exports:

- per-page tab arrays, such as `tireGuideTabs`
- `pageSubNavigationEntries`
- `getPageSubNavigationForPath()`
- `getPrimaryMegaNavCardChildren()`

Header mega-menu child chips already derive from this registry. Breadcrumb tab display must use the same registry and must not hardcode `/choose`, `/tube`, or other tab IDs in header components.

## Display rule

For a page with registered sub-navigation and a valid current child route:

```text
Home / Guides / Tire Guides / How to choose
```

The last crumb is the active tab label. It should also act as a compact switcher that shows every tab belonging to the current page.

If there is no valid child route segment, breadcrumb stays at the canonical page:

```text
Home / Guides / Tire Guides
```

## Interaction rule

- Desktop and mobile use the same data source and same active tab detection.
- The active tab crumb opens a small menu/list of all same-page tabs.
- Selecting a tab navigates to the same canonical path plus `/<tab-id>`.
- The menu must localize paths with `localePath()`.
- The menu must close after navigation, outside click, or Escape.

## Route/file responsibility

Tab child routes are generated in `nuxt.config.ts` and intentionally reuse the same page component. Do not split every tab into a separate page file unless SEO/content ownership later requires independent pages.

Current cleanup status:

- Tire Guides now lives at `app/pages/guides/tireguides.vue`.
- It no longer relies on `definePageMeta({ path: '/guides/tireguides' })`; the filesystem route is the route responsibility.
- Company factory / appearance / hole-pattern tabs are owned only by `/company/about`. `/company/ourstory` remains a standalone story page.

## Maintenance rule

When adding/removing tabs for a page:

1. Update the page tab array in `pageSubNavigation.ts`.
2. Ensure `nuxt.config.ts` registers the same tab IDs for the page route.
3. Do not separately edit `SiteHeader.vue` or `HeaderMegaMenu.vue` for each tab.
4. Update this document if breadcrumb behavior changes.
