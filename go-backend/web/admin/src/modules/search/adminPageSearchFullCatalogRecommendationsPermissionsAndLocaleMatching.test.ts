import { describe, expect, it } from 'vitest'
import { adminNavigationItems, filterNavigationItems } from '@/lib/adminNavigation'
import zhCN from '@/i18n/locales/zh-CN'
import enUS from '@/i18n/locales/en-US'
import { buildPermissionScopedAdminPageSearchCatalog, filterAdminPageSearchCatalogByQuery, type AdminPageSearchTranslator } from './adminPermissionScopedPageSearchCatalog'
import { adminCommonPageRecommendationIds, selectPermittedCommonAdminPageRecommendations } from './adminCommonPageRecommendationSelection'
import { createAdminNavigationPageSearchSource } from './adminNavigationPageSearchSource'

function createAdminPageSearchTranslatorForLocale(messages: typeof zhCN | typeof enUS): AdminPageSearchTranslator {
  return (key, params = {}, fallback = key) => {
    const message = key.split('.').reduce<unknown>((value, segment) => (
      value && typeof value === 'object' ? (value as Record<string, unknown>)[segment] : undefined
    ), messages)
    return (typeof message === 'string' ? message : fallback).replace(/\{(\w+)\}/g, (_, name) => String(params[name] ?? `{${name}}`))
  }
}

function buildAdminPageSearchCatalogForPermissions(permissions?: string[], messages: typeof zhCN | typeof enUS = zhCN) {
  const permittedNavigation = filterNavigationItems(adminNavigationItems, (permission) => permissions === undefined || permissions.includes(permission))
  return buildPermissionScopedAdminPageSearchCatalog(permittedNavigation, createAdminPageSearchTranslatorForLocale(messages))
}

describe('admin page search catalog and recommendations', () => {
  it('keeps ten ordered common links while searching every permitted leaf tab', () => {
    const catalog = buildAdminPageSearchCatalogForPermissions()
    const recommendations = selectPermittedCommonAdminPageRecommendations(catalog)
    expect(recommendations.map((entry) => entry.id)).toEqual(adminCommonPageRecommendationIds)
    expect(new Set(recommendations.map((entry) => entry.path)).size).toBe(10)
    expect(catalog.length).toBeGreaterThan(10)
    expect(filterAdminPageSearchCatalogByQuery(catalog, '商品品牌').map((entry) => entry.id)).toEqual(['catalog-brands'])
    expect(recommendations.some((entry) => entry.id === 'catalog-brands')).toBe(false)
    expect(catalog.find((entry) => entry.id === 'orders-list')).toMatchObject({
      title: '订单列表', path: '/orders/list', routeName: 'OrdersList', moduleTitle: '订单管理',
    })
    expect(catalog.every((entry) => entry.title && entry.description && entry.routeName)).toBe(true)
  })

  it('inherits parent permissions and respects permissions of individual tabs', () => {
    const noPermissions = buildAdminPageSearchCatalogForPermissions([])
    expect(selectPermittedCommonAdminPageRecommendations(noPermissions).map((entry) => entry.id)).toEqual(['dashboard'])
    expect(filterAdminPageSearchCatalogByQuery(noPermissions, '商品')).toEqual([])
    const productPages = buildAdminPageSearchCatalogForPermissions(['product:view'])
    expect(productPages.some((entry) => entry.id === 'catalog-brands')).toBe(true)
    expect(productPages.some((entry) => entry.id === 'catalog-home-main-products')).toBe(false)
    expect(buildAdminPageSearchCatalogForPermissions(['content:view']).some((entry) => entry.id === 'catalog-home-main-products')).toBe(false)
    expect(buildAdminPageSearchCatalogForPermissions(['product:view', 'content:view']).some((entry) => entry.id === 'catalog-home-main-products')).toBe(true)
  })

  it('keeps search available without customer-service permissions and does not fill missing recommendations with unauthorized pages', () => {
    const catalog = buildAdminPageSearchCatalogForPermissions(['settings:view'])
    expect(selectPermittedCommonAdminPageRecommendations(catalog).map((entry) => entry.id)).toEqual(['dashboard', 'settings-markets', 'settings-api'])
    expect(catalog.some((entry) => entry.id === 'support-analytics')).toBe(false)
    expect(catalog.some((entry) => entry.id === 'support-public-chat')).toBe(true)
    expect(filterAdminPageSearchCatalogByQuery(catalog, '/settings/backup-recovery').map((entry) => entry.id)).toEqual(['settings-backup-recovery'])
  })

  it('searches titles, descriptions, canonical paths, bilingual aliases and normalized multiword queries', () => {
    const catalog = buildAdminPageSearchCatalogForPermissions()
    for (const [query, expectedId] of [
      ['订单列表', 'orders-list'], ['上架', 'catalog-products'], ['/settings/api', 'settings-api'],
      ['快递', 'shipping-tracking-shipments'], ['  API   TIMEZONE  ', 'settings-api'],
    ]) expect(filterAdminPageSearchCatalogByQuery(catalog, query).map((entry) => entry.id)).toEqual([expectedId])
    expect(filterAdminPageSearchCatalogByQuery(catalog, 'ＡＰＩ')).toEqual(filterAdminPageSearchCatalogByQuery(catalog, 'api'))
    expect(filterAdminPageSearchCatalogByQuery(catalog, 'ＡＰＩ').some((entry) => entry.id === 'settings-api')).toBe(true)
    expect(filterAdminPageSearchCatalogByQuery(catalog, 'orders media')).toEqual([])
    expect(filterAdminPageSearchCatalogByQuery(catalog, 'does-not-exist')).toEqual([])
  })

  it('localizes metadata and descriptions while preserving Chinese canonical labels as search aliases', () => {
    const englishCatalog = buildAdminPageSearchCatalogForPermissions(undefined, enUS)
    const brandPage = englishCatalog.find((entry) => entry.id === 'catalog-brands')!
    expect(brandPage.title).toBe('商品品牌') // Same canonical-label fallback as the sidebar.
    expect(brandPage.description).toBe(`Open ${brandPage.title} in ${brandPage.moduleTitle}.`)
    expect(filterAdminPageSearchCatalogByQuery(englishCatalog, '商品品牌').map((entry) => entry.id)).toEqual(['catalog-brands'])
    expect(filterAdminPageSearchCatalogByQuery(englishCatalog, brandPage.title).some((entry) => entry.id === brandPage.id)).toBe(true)
    const crawlerPage = englishCatalog.find((entry) => entry.id === 'settings-commercial-crawler')!
    expect(crawlerPage.title).toBe(enUS.nav['settings-commercial-crawler'])
    expect(crawlerPage.description).toBe(`Open ${crawlerPage.title} in System Settings.`)
    expect(filterAdminPageSearchCatalogByQuery(englishCatalog, '商业爬虫防护').map((entry) => entry.id)).toEqual(['settings-commercial-crawler'])
    expect(englishCatalog.find((entry) => entry.id === 'orders-list')!.description).toBe(enUS.adminPageSearch.descriptions['orders-list'])
  })

  it('adapts pages to the shared source contract and keeps recommendations separate from query results', () => {
    const source = createAdminNavigationPageSearchSource(adminNavigationItems, createAdminPageSearchTranslatorForLocale(zhCN))
    const context = { signal: new AbortController().signal }
    expect(source.getRecommendations(context)).toHaveLength(10)
    expect(source.search('商品品牌', context)).toMatchObject([{
      id: 'catalog-brands', sourceId: 'pages', title: '商品品牌', routeTarget: { name: 'CatalogProductBrands' }, linkLabel: '/catalog/brands',
    }])
  })
})
