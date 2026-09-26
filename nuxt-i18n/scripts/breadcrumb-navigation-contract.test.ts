import assert from 'node:assert/strict'
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

console.log('Breadcrumb navigation contract checks passed: canonical pages retain third-level navigation entries and exact tab routes resolve correctly.')
