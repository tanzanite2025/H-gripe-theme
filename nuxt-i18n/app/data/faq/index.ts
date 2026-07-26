/**
 * FAQ Data Registry
 *
 * The Go backend owns editable page/category metadata and FAQ content for
 * storefront rendering. Static FAQ files are retained only as legacy source
 * data and are not used as a storefront fallback.
 */

export {
  faqRegistry,
  faqRoutePathByPageId,
  getAllFaqData,
  getAllFaqItems,
  getFaqData,
  getFaqDataByRoutePath,
  getPageFaqId,
  resolvePageFaqData,
  resolvePageFaqDataList,
  type ResolvedPageFaqData
} from './registry'
export {
  fetchAllFaqData,
  fetchFaqData,
  fetchFaqDataByRoutePath
} from './backend'
export { normalizeFaqRoutePath, resolveFaqRouteLookupPath, shouldAutoInsertFaqForRoute } from './routing'
export * from './types'
