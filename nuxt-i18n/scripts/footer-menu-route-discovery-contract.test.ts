import assert from 'node:assert/strict'
import type { RouteRecordNormalized } from 'vue-router'
import {
  createFooterMenusFromRoutes,
  createFooterShopLinksFromProductCategories,
  type FooterNavigationItem,
} from '../app/utils/footerMenus.js'

const createNormalizedFooterTestRoute = (
  path: string,
  name: string,
  meta: Record<string, unknown> = {},
  redirect?: RouteRecordNormalized['redirect'],
) => ({
  path,
  name,
  meta,
  redirect,
}) as unknown as RouteRecordNormalized

const flattenFooterNavigationItems = (
  items: readonly FooterNavigationItem[],
): FooterNavigationItem[] => items.flatMap(item => [
  item,
  ...(item.children ? flattenFooterNavigationItems(item.children) : []),
])

const sections = createFooterMenusFromRoutes([
  createNormalizedFooterTestRoute('/shop', 'shop', {
    footerLabelKey: 'products.nav.shop',
    footerLabelFallback: 'Shop',
  }),
  createNormalizedFooterTestRoute('/shop/:slug', 'shop-slug'),
  createNormalizedFooterTestRoute('/shop/:category(.*)*', 'shop-category'),
  createNormalizedFooterTestRoute('/guides', 'guides'),
  createNormalizedFooterTestRoute('/guides/tireguides', 'guides-tireguides', {
    footerLabelKey: 'products.nav.tireSizeCharts',
    footerLabelFallback: 'Tire Guides',
  }),
  createNormalizedFooterTestRoute('/guides/tireguides/tire-pressure', 'guides-tire-pressure', {
    footerLabelKey: 'products.nav.tireSizeCharts',
    footerLabelFallback: 'Tire Guides',
  }),
  createNormalizedFooterTestRoute('/guides/wheelset-buyers', 'guides-wheelset-buyers', {
    footerLabelKey: 'products.nav.wheelsetBuyersGuide',
    footerLabelFallback: 'Wheelset Guide',
  }),
  createNormalizedFooterTestRoute('/guides/spokeguides', 'guides-spokeguides', {
    footerLabelKey: 'products.nav.spokeGuides',
    footerLabelFallback: 'Spoke Guides',
  }),
  createNormalizedFooterTestRoute(
    '/guides/spokeguides/new-spoke-engineering-guide',
    'guides-spokeguides-new-guide',
    {
      footerLabelFallback: 'New Spoke Engineering Guide',
      footerGroupLabelFallback: 'Spoke Guides',
    },
  ),
  createNormalizedFooterTestRoute(
    '/zh_cn/guides/spokeguides/new-spoke-engineering-guide',
    'zh-guides-spokeguides-new-guide',
  ),
  createNormalizedFooterTestRoute('/guides/internal-draft', 'guides-internal-draft', { footer: false }),
  createNormalizedFooterTestRoute('/guides/redirect', 'guides-redirect', {}, '/guides/tireguides'),
  createNormalizedFooterTestRoute('/guides/:slug', 'guides-dynamic'),
  createNormalizedFooterTestRoute('/guides-archive/old-guide', 'guides-archive-old-guide'),
])

const shopSection = sections.find(section => section.id === 'shop')
assert.ok(shopSection, 'SHOP column should remain visible when its category routes are dynamic')
assert.deepEqual(shopSection.links.map(link => link.to), ['/shop'])

const guidesSection = sections.find(section => section.id === 'guides')
assert.ok(guidesSection, 'GUIDES column should be discovered from its static descendant pages')

const tireGuidesGroup = guidesSection.links.find(link => link.to === '/guides/tireguides')
assert.ok(tireGuidesGroup, 'Tire Guides should retain its real landing page')
assert.equal(tireGuidesGroup.fallback, 'Tire Guides')
assert.equal(tireGuidesGroup.children?.length, 9, 'All tire guide tabs should be listed')
assert.deepEqual(
  tireGuidesGroup.children?.map(link => link.to),
  [
    '/guides/tireguides/tire-size-markings',
    '/guides/tireguides/tire-frame-clearance',
    '/guides/tireguides/schwalbe-tire-circumference',
    '/guides/tireguides/tubeless',
    '/guides/tireguides/choose',
    '/guides/tireguides/tire-pressure',
    '/guides/tireguides/tire-pressure-calculator',
    '/guides/tireguides/choose-inner-tube',
    '/guides/tireguides/schwalbe-tire-selector',
  ],
)

const wheelsetGuidesGroup = guidesSection.links.find(link => link.to === '/guides/wheelset-buyers')
assert.ok(wheelsetGuidesGroup, 'Wheelset Guide should retain its real landing page')
assert.equal(wheelsetGuidesGroup.fallback, 'Wheelset Guide')
assert.equal(wheelsetGuidesGroup.children?.length, 5, 'All wheelset guide tabs should be listed')

const spokeGuidesGroup = guidesSection.links.find(link => link.fallback === 'Spoke Guides')
assert.ok(spokeGuidesGroup, 'Spoke Guides should be represented as a group')
assert.equal(spokeGuidesGroup.to, '/guides/spokeguides', 'Spoke Guides should link to its category landing page')
assert.deepEqual(
  spokeGuidesGroup.children?.map(link => link.to),
  [
    '/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics',
    '/guides/spokeguides/brand-wheelset-spoke-specs',
    '/guides/spokeguides/spoke-length-calculator',
    '/guides/spokeguides/new-spoke-engineering-guide',
  ],
)

const allGuideLinks = flattenFooterNavigationItems(guidesSection.links)
const guidePaths = allGuideLinks
  .map(link => link.to)
  .filter((path): path is string => Boolean(path))
assert.equal(new Set(guidePaths).size, guidePaths.length, 'GUIDES must not emit duplicate canonical paths')

const guidesWithoutIndexPage = createFooterMenusFromRoutes([
  createNormalizedFooterTestRoute('/guides/tireguides/tire-pressure', 'guides-tire-pressure', {
    footerLabelKey: 'products.nav.tireSizeCharts',
    footerLabelFallback: 'Tire Guides',
  }),
])
const pageLessTireGuidesGroup = guidesWithoutIndexPage
  .find(section => section.id === 'guides')
  ?.links.find(link => link.fallback === 'Tire Guides')
assert.ok(pageLessTireGuidesGroup)
assert.equal(pageLessTireGuidesGroup.to, undefined)
assert.ok(pageLessTireGuidesGroup.children?.some(link => link.to === '/guides/tireguides/tire-pressure'))

const shopCategoryLinks = createFooterShopLinksFromProductCategories([
  {
    name: 'Wheel Parts',
    slug: 'wheel-parts',
    routePath: '/zh_cn/shop/wheel-parts',
    children: [{ name: 'Spokes', slug: 'spokes' }],
  },
  {
    name: 'Tires',
    slug: 'tires',
  },
])
assert.deepEqual(shopCategoryLinks, [
  {
    fallback: 'Wheel Parts',
    to: '/shop/wheel-parts',
    children: [{ fallback: 'Spokes', to: '/shop/spokes' }],
  },
  { fallback: 'Tires', to: '/shop/tires' },
])
assert.deepEqual(createFooterShopLinksFromProductCategories([]), [])

const sectionsWithProductCategories = createFooterMenusFromRoutes(
  [
    createNormalizedFooterTestRoute('/shop', 'shop'),
    createNormalizedFooterTestRoute('/shop/:slug', 'shop-slug'),
  ],
  { shopLinks: shopCategoryLinks },
)
assert.deepEqual(
  sectionsWithProductCategories.find(section => section.id === 'shop')?.links,
  shopCategoryLinks,
)

const sectionsWhileProductCategoriesAreLoading = createFooterMenusFromRoutes(
  [
    createNormalizedFooterTestRoute('/shop', 'shop'),
    createNormalizedFooterTestRoute('/shop/:slug', 'shop-slug'),
  ],
  { shopLinks: [] },
)
assert.equal(
  sectionsWhileProductCategoriesAreLoading.find(section => section.id === 'shop'),
  undefined,
  'SHOP must not fall back to a generic Shop link while categories are unavailable',
)

console.log('Footer menu route discovery contract passed.')
