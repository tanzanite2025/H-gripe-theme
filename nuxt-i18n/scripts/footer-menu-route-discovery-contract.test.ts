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
  createNormalizedFooterTestRoute('/technology/test-report', 'technology-test-report', {
    footerLabelKey: 'support.nav.testReport',
    footerLabelFallback: 'Test Report',
  }),
  createNormalizedFooterTestRoute('/support/faqs', 'support-faqs'),
  createNormalizedFooterTestRoute('/tools', 'tools', {
    footerLabelKey: 'toolsNavigation.label',
    footerLabelFallback: 'Tools',
  }),
  createNormalizedFooterTestRoute('/tools/schwalbe-tire-selector', 'tools-schwalbe-tire-selector', {
    footerLabelKey: 'toolsNavigation.schwalbeTireSelectorLabel',
    footerLabelFallback: 'Schwalbe tire selector',
  }),
  createNormalizedFooterTestRoute('/tools/brand-wheelset-spoke-specs', 'tools-brand-wheelset-spoke-specs', {
    footerLabelKey: 'toolsNavigation.brandWheelsetSpokeSpecsLabel',
    footerLabelFallback: 'Wheelset spoke and nipple specifications',
  }),
  createNormalizedFooterTestRoute('/tools/spoke-length-calculator', 'tools-spoke-length-calculator', {
    footerLabelKey: 'toolsNavigation.spokeLengthCalculatorLabel',
    footerLabelFallback: 'Spoke length calculator',
  }),
  createNormalizedFooterTestRoute('/tools/spoke-lacing-topology-simulator', 'tools-spoke-lacing-topology-simulator', {
    footerLabelKey: 'toolsNavigation.spokeLacingTopologySimulatorLabel',
    footerLabelFallback: 'Spoke lacing topology simulator',
  }),
  createNormalizedFooterTestRoute('/tools/dynamic-tire-pressure-calculator', 'tools-dynamic-tire-pressure-calculator', {
    footerLabelKey: 'toolsNavigation.dynamicTirePressureCalculatorLabel',
    footerLabelFallback: 'Dynamic tire pressure calculator',
  }),
  createNormalizedFooterTestRoute('/tools/tire-rim-width-compatibility-engine', 'tools-tire-rim-width-compatibility-engine', {
    footerLabelKey: 'toolsNavigation.tireRimWidthCompatibilityEngineLabel',
    footerLabelFallback: 'Tire and rim width compatibility engine',
  }),
  createNormalizedFooterTestRoute('/tools/tire-size-marking-converter', 'tools-tire-size-marking-converter', {
    footerLabelKey: 'toolsNavigation.tireSizeMarkingConverterLabel',
    footerLabelFallback: 'Tire size marking converter',
  }),
  createNormalizedFooterTestRoute('/tools/spoke-stress-relief-mechanics', 'tools-spoke-stress-relief-mechanics', {
    footerLabelKey: 'toolsNavigation.spokeStressReliefMechanicsLabel',
    footerLabelFallback: 'Stainless steel spoke stress-relief mechanics',
  }),
  createNormalizedFooterTestRoute('/tools/valve-length-fitment-calculator', 'tools-valve-length-fitment-calculator', {
    footerLabelKey: 'toolsNavigation.valveLengthFitmentCalculatorLabel',
    footerLabelFallback: 'Inner tube valve-length fitment calculator',
  }),
  createNormalizedFooterTestRoute('/tools/freehub-cassette-fitment-engine', 'tools-freehub-cassette-fitment-engine', {
    footerLabelKey: 'toolsNavigation.freehubCassetteFitmentEngineLabel',
    footerLabelFallback: 'Freehub and cassette fitment engine',
  }),
  createNormalizedFooterTestRoute('/guides/tireguides', 'guides-tireguides', {
    footerLabelKey: 'products.nav.tireSizeCharts',
    footerLabelFallback: 'Tire Guides',
  }),
  createNormalizedFooterTestRoute('/guides/wheelset-buyers', 'guides-wheelset-buyers', {
    footerLabelKey: 'products.nav.wheelsetBuyersGuide',
    footerLabelFallback: 'Wheelset Guide',
  }),
  createNormalizedFooterTestRoute('/guides/internal-draft', 'guides-internal-draft', { footer: false }),
  createNormalizedFooterTestRoute('/guides/redirect', 'guides-redirect', {}, '/guides/tireguides'),
  createNormalizedFooterTestRoute('/guides/:slug', 'guides-dynamic'),
  createNormalizedFooterTestRoute('/guides-archive/old-guide', 'guides-archive-old-guide'),
])

const shopSection = sections.find(section => section.id === 'shop')
assert.ok(shopSection, 'SHOP column should remain visible when its category routes are dynamic')
assert.deepEqual(shopSection.links.map(link => link.to), ['/shop'])

const guidesSection = sections.find(section => section.id === 'guides')
const technologySection = sections.find(section => section.id === 'technology')
assert.ok(technologySection, 'TECHNOLOGY must discover the migrated test report')
assert.deepEqual(technologySection.links.map(link => link.to), ['/technology/test-report'])
assert.ok(!sections.find(section => section.id === 'support')?.links.some(link => link.to?.includes('test-report')),
  'SUPPORT must not retain ownership of the migrated test report')
const toolsSection = sections.find(section => section.id === 'tools')
assert.ok(toolsSection, 'TOOLS must discover the hub and migrated selector')
assert.equal(toolsSection.titleLinkTo, '/tools', 'TOOLS title must link to the hub')
assert.deepEqual(flattenFooterNavigationItems(toolsSection.links).map(link => link.to), [
  '/tools/schwalbe-tire-selector',
  '/tools/brand-wheelset-spoke-specs',
  '/tools/spoke-length-calculator',
  '/tools/spoke-lacing-topology-simulator',
  '/tools/dynamic-tire-pressure-calculator',
  '/tools/tire-rim-width-compatibility-engine',
  '/tools/tire-size-marking-converter',
  '/tools/spoke-stress-relief-mechanics',
  '/tools/valve-length-fitment-calculator',
  '/tools/freehub-cassette-fitment-engine',
], 'TOOLS child list must contain only concrete tools, without a duplicate hub')
assert.ok(guidesSection, 'GUIDES column should be discovered from its static descendant pages')

const tireGuidesGroup = guidesSection.links.find(link => link.to === '/guides/tireguides')
assert.ok(tireGuidesGroup, 'Tire Guides should retain its real landing page')
assert.equal(tireGuidesGroup.fallback, 'Tire Guides')
assert.equal(tireGuidesGroup.children?.length, 3, 'Only remaining tire guide tabs should be listed')
assert.deepEqual(
  tireGuidesGroup.children?.map(link => link.to),
  [
    '/guides/tireguides/tire-frame-clearance',
    '/guides/tireguides/schwalbe-tire-circumference',
    '/guides/tireguides/tubeless',
  ],
)

const wheelsetGuidesGroup = guidesSection.links.find(link => link.to === '/guides/wheelset-buyers')
assert.ok(wheelsetGuidesGroup, 'Wheelset Guide should retain its real landing page')
assert.equal(wheelsetGuidesGroup.fallback, 'Wheelset Guide')
assert.equal(wheelsetGuidesGroup.children?.length, 4, 'Only remaining wheelset guide tabs should be listed')
assert.ok(!flattenFooterNavigationItems(guidesSection.links).some(link => link.to?.includes('wheelset-spoke-lacing-topology-and-geometry-reference')))

assert.ok(!flattenFooterNavigationItems(guidesSection.links).some(link => link.to?.startsWith('/guides/spokeguides')), 'Empty Spoke Guides category must not remain in the footer')

const allGuideLinks = flattenFooterNavigationItems(guidesSection.links)
const guidePaths = allGuideLinks
  .map(link => link.to)
  .filter((path): path is string => Boolean(path))
assert.equal(new Set(guidePaths).size, guidePaths.length, 'GUIDES must not emit duplicate canonical paths')
assert.ok(!guidePaths.includes('/guides/tireguides/choose'), 'The retired tire width route must not remain as a footer alias')

const guidesWithoutIndexPage = createFooterMenusFromRoutes([
  createNormalizedFooterTestRoute('/guides/tireguides/tubeless', 'guides-tireguides-tubeless', {
    footerLabelKey: 'products.nav.tireSizeCharts',
    footerLabelFallback: 'Tire Guides',
  }),
])
const pageLessTireGuidesGroup = guidesWithoutIndexPage
  .find(section => section.id === 'guides')
  ?.links.find(link => link.fallback === 'Tire Guides')
assert.ok(pageLessTireGuidesGroup)
assert.equal(pageLessTireGuidesGroup.to, undefined)
assert.ok(pageLessTireGuidesGroup.children?.some(link => link.to === '/guides/tireguides/tubeless'))

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
assert.deepEqual(
  createFooterShopLinksFromProductCategories([]),
  [{ labelKey: 'products.nav.shop', fallback: 'Shop', to: '/shop' }],
)

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
const shopSectionWhileProductCategoriesAreLoading = sectionsWhileProductCategoriesAreLoading
  .find(section => section.id === 'shop')
assert.ok(shopSectionWhileProductCategoriesAreLoading, 'SHOP must remain visible while categories are unavailable')
assert.deepEqual(
  shopSectionWhileProductCategoriesAreLoading.links.map(link => link.to),
  ['/shop'],
)

console.log('Footer menu route discovery contract passed.')
