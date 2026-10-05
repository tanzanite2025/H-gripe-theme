import assert from 'node:assert/strict'
import {
  breadcrumbRoutePatternMatches,
  groupBreadcrumbRoutePathsAtLevel,
  isExactBreadcrumbPageSubNavigationOwner,
  normalizeBreadcrumbRouteSegments,
  resolveBreadcrumbSubNavigationOwner,
  resolveBreadcrumbSiblingTarget,
} from '../app/utils/breadcrumbRouteNavigation.js'
import { resolvePageSubNavigationBreadcrumb } from '../app/utils/pageSubNavigationBreadcrumb.js'
import type { PageSubNavigationEntry } from '../app/utils/pageSubNavigationData.js'
import { spokeGuideTabs, tireGuideTabs } from '../app/utils/pageSubNavigationData.js'

const entries: PageSubNavigationEntry[] = [
  {
    path: '/guides/tireguides',
    tabs: [
      { id: 'tire-size-markings', fallback: 'Tire size markings' },
      { id: 'tubeless', fallback: 'Tubeless tires & installation' },
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

const tab = match('/guides/tireguides/tubeless')
assert.equal(tab?.kind, 'tab')
assert.equal(tab?.entry.path, '/guides/tireguides')
assert.equal(tab?.tab.id, 'tubeless')

const englishDirectTab = match('/guides/tireguides/tubeless?source=breadcrumb')
assert.equal(englishDirectTab?.kind, 'tab')
assert.equal(englishDirectTab?.tab.id, 'tubeless')

const localizedTabWithTrailingSlash = match('/zh_cn/guides/tireguides/tubeless/')
assert.equal(localizedTabWithTrailingSlash?.kind, 'tab')
assert.equal(localizedTabWithTrailingSlash?.tab.id, 'tubeless')

// Canonical page paths and exact tab paths remain distinct. The canonical
// crumb owns its same-level route menu; the selected tab crumb owns tab peers.
assert.equal(canonical?.kind, 'canonical')

// Exact matching is intentional: deeper descendants and unknown tabs do not
// open the owning page's internal tab menu.
assert.equal(match('/guides/tireguides/tubeless/details'), null)
assert.equal(match('/guides/tireguides/installation'), null)
assert.equal(match('/guides/tireguides/unknown'), null)
assert.equal(match('/guides'), null)
assert.equal(match('/guides/tireguides-installation'), null)
assert.equal(match('/unknown/guides/tireguides'), null)

// On a calculator route, the Tire Guides crumb is an ancestor and must keep
// its /guides sibling menu; only an exact canonical/tab crumb owns page tabs.
assert.equal(isExactBreadcrumbPageSubNavigationOwner(
  '/guides/tireguides',
  '/guides/tireguides/tire-pressure-calculator',
  ['en', 'zh_cn'],
), false)
assert.equal(isExactBreadcrumbPageSubNavigationOwner(
  '/guides/tireguides/tire-pressure-calculator?source=breadcrumb',
  '/guides/tireguides/tire-pressure-calculator',
  ['en', 'zh_cn'],
), true)
assert.deepEqual(normalizeBreadcrumbRouteSegments(
  '/guides/tireguides/tire-pressure-calculator',
  ['en', 'zh_cn'],
), ['guides', 'tireguides', 'tire-pressure-calculator'])
assert.deepEqual(normalizeBreadcrumbRouteSegments(
  '/zh_cn/guides/tireguides/tire-pressure-calculator?source=breadcrumb',
  ['en', 'zh_cn'],
), ['guides', 'tireguides', 'tire-pressure-calculator'])
assert.deepEqual(normalizeBreadcrumbRouteSegments(
  '/ZH_CN/guides/tireguides/tire-pressure-calculator#calculator',
  ['en', 'zh_cn'],
), ['guides', 'tireguides', 'tire-pressure-calculator'])
assert.equal(resolveBreadcrumbSubNavigationOwner({
  breadcrumbPath: '/guides/tireguides',
  currentRoutePath: '/guides/tireguides/tire-pressure-calculator',
  siblingPaths: ['/guides/tireguides', '/guides/wheelset-buyers'],
  hasPageSubNavigation: true,
  localeCodes: ['en', 'zh_cn'],
}), 'same-level-route-siblings')
assert.equal(resolveBreadcrumbSubNavigationOwner({
  breadcrumbPath: '/zh_cn/guides/tireguides',
  currentRoutePath: '/zh_cn/guides/tireguides/tire-pressure-calculator',
  siblingPaths: ['/guides/tireguides', '/guides/wheelset-buyers'],
  hasPageSubNavigation: true,
  localeCodes: ['en', 'zh_cn'],
}), 'same-level-route-siblings')
assert.equal(resolveBreadcrumbSubNavigationOwner({
  breadcrumbPath: '/guides/tireguides',
  currentRoutePath: '/guides/tireguides',
  siblingPaths: ['/guides/tireguides'],
  hasPageSubNavigation: true,
  localeCodes: ['en', 'zh_cn'],
}), 'page-sub-navigation')
assert.equal(resolveBreadcrumbSubNavigationOwner({
  breadcrumbPath: '/zh_cn/guides/tireguides',
  currentRoutePath: '/zh_cn/guides/tireguides/tire-pressure-calculator',
  siblingPaths: ['/guides/tireguides'],
  hasPageSubNavigation: true,
  localeCodes: ['en', 'zh_cn'],
}), null)

const tireGuideRouteTab = resolvePageSubNavigationBreadcrumb(
  '/guides/tireguides/schwalbe-tire-selector',
  [{ path: '/guides/tireguides', tabs: tireGuideTabs }],
)
assert.equal(tireGuideRouteTab?.kind, 'tab')
assert.equal(tireGuideRouteTab?.entry.tabs.length, 9)

const tirePressureCalculatorRouteTab = resolvePageSubNavigationBreadcrumb(
  '/guides/tireguides/tire-pressure-calculator',
  [{ path: '/guides/tireguides', tabs: tireGuideTabs }],
)
assert.equal(tirePressureCalculatorRouteTab?.kind, 'tab')
assert.equal(tirePressureCalculatorRouteTab?.tab.id, 'tire-pressure-calculator')

const clearanceRouteTab = resolvePageSubNavigationBreadcrumb(
  '/guides/tireguides/tire-frame-clearance',
  [{ path: '/guides/tireguides', tabs: tireGuideTabs }],
)
assert.equal(clearanceRouteTab?.kind, 'tab')
assert.equal(clearanceRouteTab?.tab.id, 'tire-frame-clearance')

const circumferenceRouteTab = resolvePageSubNavigationBreadcrumb(
  '/guides/tireguides/schwalbe-tire-circumference',
  [{ path: '/guides/tireguides', tabs: tireGuideTabs }],
)
assert.equal(circumferenceRouteTab?.kind, 'tab')
assert.equal(circumferenceRouteTab?.tab.id, 'schwalbe-tire-circumference')

const wheelsetCanonical = match('/guides/wheelset-buyers')
assert.equal(wheelsetCanonical?.kind, 'canonical')
assert.equal(wheelsetCanonical?.entry.path, '/guides/wheelset-buyers')
assert.notEqual(wheelsetCanonical?.entry, canonical?.entry)

const spokeGuideEntries: PageSubNavigationEntry[] = [
  { path: '/guides/spokeguides', tabs: spokeGuideTabs },
]
const spokeCanonical = resolvePageSubNavigationBreadcrumb('/guides/spokeguides', spokeGuideEntries)
assert.equal(spokeCanonical?.kind, 'canonical')
assert.equal(spokeCanonical?.entry.path, '/guides/spokeguides')
const spokeCalculatorTab = resolvePageSubNavigationBreadcrumb(
  '/zh_cn/guides/spokeguides/spoke-length-calculator',
  spokeGuideEntries,
  ['en', 'zh_cn'],
)
assert.equal(spokeCalculatorTab?.kind, 'tab')
assert.equal(spokeCalculatorTab?.entry.path, '/guides/spokeguides')
assert.equal(spokeCalculatorTab?.tab.id, 'spoke-length-calculator')

const switchSibling = (options: {
  currentPath: string
  breadcrumbPath: string
  siblingPath: string
  fallbackPath: string
  routePatterns: string[]
}) => resolveBreadcrumbSiblingTarget({ ...options, localeCodes: ['en', 'zh_cn'] })

// Breadcrumb levels are grouped by the same parent prefix and exact depth.
// Descendant pages confirm a level but never leak into its sibling list.
const guideSiblings = groupBreadcrumbRoutePathsAtLevel(
  ['guides'],
  2,
  [
    '/guides/tireguides',
    '/guides/tireguides/tire-pressure',
    '/guides/wheelset-buyers',
    '/guides/wheelset-buyers/overview',
    '/resources/blog',
  ],
  ['en', 'zh_cn'],
)
assert.deepEqual(guideSiblings.map(group => group.path), [
  '/guides/tireguides',
  '/guides/wheelset-buyers',
])

const localizedGuideSiblings = groupBreadcrumbRoutePathsAtLevel(
  ['zh_cn', 'guides', 'tireguides'],
  3,
  [
    '/guides/tireguides/tubeless',
    '/zh_cn/guides/tireguides/tire-pressure-calculator',
    '/zh_cn/guides/tireguides/schwalbe-tire-selector',
    '/fr/guides/wheelset-buyers/overview',
  ],
  ['en', 'zh_cn', 'fr'],
)
assert.deepEqual(localizedGuideSiblings.map(group => group.path), [
  '/guides/tireguides/tubeless',
  '/guides/tireguides/tire-pressure-calculator',
  '/guides/tireguides/schwalbe-tire-selector',
])

// A level with one registered path resolves to one group, so it has no sibling
// menu to render.
const singleGuideBranch = groupBreadcrumbRoutePathsAtLevel(
  ['guides', 'tireguides'],
  3,
  [
    '/guides/tireguides/schwalbe-tire-selector',
    '/guides/tireguides/schwalbe-tire-selector/details',
  ],
  ['en', 'zh_cn'],
)
assert.deepEqual(singleGuideBranch.map(group => group.path), [
  '/guides/tireguides/schwalbe-tire-selector',
])

// Switching an intermediate breadcrumb preserves every lower segment when
// the destination branch registers that route, including newly added levels.
assert.equal(switchSibling({
  currentPath: '/shop/wheels/gravel/fitment/road?source=header',
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
  currentPath: '/guides/tireguides/tubeless',
  breadcrumbPath: '/guides/tireguides',
  siblingPath: '/guides/wheelset-buyers',
  fallbackPath: '/guides/wheelset-buyers',
  routePatterns: ['/guides/wheelset-buyers/:tab(overview|safety-instructions)'],
}), '/guides/wheelset-buyers')

// Prefixed locale URLs normalize to the same route depth, then re-localize at
// the component boundary. A sibling route only keeps a suffix it actually has.
assert.equal(switchSibling({
  currentPath: '/zh_cn/guides/tireguides/tire-pressure-calculator',
  breadcrumbPath: '/zh_cn/guides/tireguides',
  siblingPath: '/guides/wheelset-buyers',
  fallbackPath: '/guides/wheelset-buyers',
  routePatterns: [
    '/guides/wheelset-buyers/:tab(overview|safety-instructions)',
  ],
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

console.log('Breadcrumb navigation contract checks passed: route ownership, same-level grouping, singleton levels, and registered descendant preservation are correct.')
