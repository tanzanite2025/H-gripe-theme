import assert from 'node:assert/strict'
import { breadcrumbRoutePatternMatches, resolveBreadcrumbSiblingTarget } from '../app/utils/breadcrumbRouteNavigation.js'
import { resolvePageSubNavigationBreadcrumb } from '../app/utils/pageSubNavigationBreadcrumb.js'
import type { PageSubNavigationEntry } from '../app/utils/pageSubNavigationData.js'

const entries: PageSubNavigationEntry[] = [
  {
    path: '/guides/tireguides',
    tabs: [
      { id: 'size', fallback: 'Tire size' },
      { id: 'installation', fallback: 'Installation' },
    ],
  },
  {
    path: '/guides/wheelset-buyers',
    tabs: [
      { id: 'overview', fallback: 'Buying overview' },
    ],
  },
]

const match = (path: string) => resolvePageSubNavigationBreadcrumb(path, entries, ['en', 'zh_cn'])

const canonical = match('/guides/tireguides')
assert.equal(canonical?.kind, 'canonical')
assert.equal(canonical?.entry.path, '/guides/tireguides')

const localizedCanonical = match('/zh_cn/guides/tireguides/?from=header')
assert.equal(localizedCanonical?.kind, 'canonical')
assert.equal(localizedCanonical?.entry.path, '/guides/tireguides')

const canonicalWithTrailingSlash = match('/guides/tireguides/')
assert.equal(canonicalWithTrailingSlash?.kind, 'canonical')
assert.equal(canonicalWithTrailingSlash?.entry.path, '/guides/tireguides')

const tab = match('/guides/tireguides/installation')
assert.equal(tab?.kind, 'tab')
assert.equal(tab?.entry.path, '/guides/tireguides')
assert.equal(tab?.tab.id, 'installation')

const localizedTab = match('/en/guides/tireguides/installation?source=breadcrumb')
assert.equal(localizedTab?.kind, 'tab')
assert.equal(localizedTab?.tab.id, 'installation')

const localizedTabWithTrailingSlash = match('/zh_cn/guides/tireguides/installation/')
assert.equal(localizedTabWithTrailingSlash?.kind, 'tab')
assert.equal(localizedTabWithTrailingSlash?.tab.id, 'installation')

// Canonical second-level pages retain their entry so the header can attach the
// page's third-level navigation immediately after switching siblings.
assert.equal(canonical?.kind, 'canonical')

// Exact matching is intentional: deeper descendants and unknown tabs do not
// open the owning page's internal tab menu.
assert.equal(match('/guides/tireguides/installation/details'), null)
assert.equal(match('/guides/tireguides/unknown'), null)
assert.equal(match('/guides'), null)
assert.equal(match('/guides/tireguides-installation'), null)
assert.equal(match('/unknown/guides/tireguides'), null)

const wheelsetCanonical = match('/guides/wheelset-buyers')
assert.equal(wheelsetCanonical?.kind, 'canonical')
assert.equal(wheelsetCanonical?.entry.path, '/guides/wheelset-buyers')
assert.notEqual(wheelsetCanonical?.entry, canonical?.entry)

const switchSibling = (options: {
  currentPath: string
  breadcrumbPath: string
  siblingPath: string
  fallbackPath: string
  routePatterns: string[]
}) => resolveBreadcrumbSiblingTarget({ ...options, localeCodes: ['en', 'zh_cn'] })

// Switching an intermediate breadcrumb preserves every lower segment when
// the destination branch registers that route, including newly added levels.
assert.equal(switchSibling({
  currentPath: '/en/shop/wheels/gravel/fitment/road?source=header',
  breadcrumbPath: '/shop/wheels',
  siblingPath: '/shop/tires',
  fallbackPath: '/shop/tires',
  routePatterns: [
    '/shop/wheels/:wheelType/fitment/:fitmentType',
    '/shop/tires/:tireType/fitment/:fitmentType',
  ],
}), '/shop/tires/gravel/fitment/road')

// Dynamic route constraints are honored, so switching between pages with
// different tab sets does not create an invalid child URL.
assert.equal(breadcrumbRoutePatternMatches(
  '/guides/wheelset-buyers/:tab(overview|safety-instructions)',
  '/guides/wheelset-buyers/overview',
), true)
assert.equal(breadcrumbRoutePatternMatches(
  '/guides/wheelset-buyers/:tab(overview|safety-instructions)',
  '/guides/wheelset-buyers/installation',
), false)
assert.equal(switchSibling({
  currentPath: '/guides/tireguides/installation',
  breadcrumbPath: '/guides/tireguides',
  siblingPath: '/guides/wheelset-buyers',
  fallbackPath: '/guides/wheelset-buyers',
  routePatterns: ['/guides/wheelset-buyers/:tab(overview|safety-instructions)'],
}), '/guides/wheelset-buyers')

// The same rule covers routes nested below a page tab, so newly added lower
// pages survive tab switches without adding tab-specific header logic.
assert.equal(switchSibling({
  currentPath: '/guides/wheelset-buyers/overview/details',
  breadcrumbPath: '/guides/wheelset-buyers/overview',
  siblingPath: '/guides/wheelset-buyers/safety-instructions',
  fallbackPath: '/guides/wheelset-buyers/safety-instructions',
  routePatterns: [
    '/guides/wheelset-buyers/:tab(overview|safety-instructions)',
    '/guides/wheelset-buyers/:tab(overview|safety-instructions)/details',
  ],
}), '/guides/wheelset-buyers/safety-instructions/details')

// A same-family root switch can retain its selected descendants; a different
// branch falls back to its registered representative route when no match exists.
assert.equal(switchSibling({
  currentPath: '/guides/tireguides/choose',
  breadcrumbPath: '/guides',
  siblingPath: '/guides',
  fallbackPath: '/guides/tireguides',
  routePatterns: ['/guides/tireguides/:tab(size|choose)'],
}), '/guides/tireguides/choose')
assert.equal(switchSibling({
  currentPath: '/guides/tireguides/choose',
  breadcrumbPath: '/guides',
  siblingPath: '/support',
  fallbackPath: '/support/faqs',
  routePatterns: ['/support/faqs/:tab(overview|contact)'],
}), '/support/faqs')

console.log('Breadcrumb navigation contract checks passed: route ownership, constrained tab routes, and registered descendant preservation are correct.')
