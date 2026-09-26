# Breadcrumb and page sub-navigation current source

Last updated: 2026-07-24

## Required breadcrumb contract

- A canonical second-level page registered in `pageSubNavigationEntries` owns
  its configured third-level navigation. Switching between second-level
  siblings must not remove that navigation from the destination breadcrumb.
- An exact third-level tab route resolves to the same owning entry and marks
  only the matching tab active.
- Locale prefixes, query strings, and trailing slashes must not change either
  result.
- Unknown tabs, deeper descendants, prefix lookalikes, and unrelated paths
  must not inherit another page's third-level navigation.
- Changes to breadcrumb ownership must keep
  `scripts/breadcrumb-navigation-contract.test.ts` passing.

This note defines how Nuxt storefront breadcrumbs should represent pages that use slash child routes for in-page tabs, such as `/guides/tireguides/choose`.

## Current problem

Some pages are one canonical page component with several tab routes:

- `/guides/tireguides/choose`
- `/support/warranty/damaged-lost`
- `/company/about/factory`

The tab segment is a Nuxt child route that reuses the owning page component. Breadcrumbs include the child route segment, so `/guides/tireguides/choose` renders as:

```text
Home / Guides / Tire Guides
```

That is technically correct for the route, but it hides the active tab and makes direct links feel confusing.

## Current source of truth

The single source for third-level page tab data is:

- `nuxt-i18n/app/utils/pageSubNavigationData.ts`

It is re-exported through `nuxt-i18n/app/utils/pageSubNavigation.ts`, which exports:

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
### Canonical and tab route contract

Breadcrumb sub-navigation uses an explicit two-way route classification:

1. If the current path exactly equals a registered PageSubNavigationEntry.path,
   it is a canonical page. Its expandable menu must come from the parent
   route's sibling pages (for example /guides/tireguides expands to Tire
   Guides and Wheelset Buyers Guide). It must not open its own tabs.
2. Only an exact match for a registered tab target (<entry.path>/<tab-id>) or
   a tab's explicit to is a tab route. That route may open the owning page's
   internal tab list (for example /guides/tireguides/installation).
3. Prefix or nested matches are not tab matches. Unknown tab IDs and deeper
   descendants must not open the internal tab list.

The pure resolver is app/utils/pageSubNavigationBreadcrumb.ts. Its contract
is covered by npm run test:breadcrumb-navigation.


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

1. Update the page tab array in `pageSubNavigationData.ts`.
2. Ensure `nuxt.config.ts` registers the same tab IDs for the page route.
3. Do not separately edit `SiteHeader.vue` or `HeaderMegaMenu.vue` for each tab.
4. Update this document if breadcrumb behavior changes.
