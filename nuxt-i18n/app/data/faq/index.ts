/**
 * Public FAQ storefront data API.
 *
 * The Go backend owns editable page/category metadata and FAQ content for
 * storefront rendering. Legacy static FAQ source data intentionally stays out
 * of this public barrel so storefront code cannot accidentally bypass backend
 * multilingual content.
 */

export {
  fetchAllFaqData,
  fetchFaqData,
  fetchFaqDataByRoutePath
} from './backend'
export {
  getPageFaqId,
  resolvePageFaqData,
  resolvePageFaqDataList,
  type ResolvedPageFaqData
} from './helpers'
export {
  filterGlobalAllFaqItems,
  flattenGlobalAllFaqItems,
  groupGlobalAllFaqItemsByPage,
  groupGlobalAllFaqTopics,
  type GlobalAllFaqFlatItem,
  type GlobalAllFaqSearchTopic,
  type GlobalAllFaqsDisplayGroup,
} from './global'
export { normalizeFaqRoutePath, resolveFaqRouteLookupPath, shouldAutoInsertFaqForRoute } from './routing'
export * from './types'
