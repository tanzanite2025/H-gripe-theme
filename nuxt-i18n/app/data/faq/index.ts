/**
 * FAQ Data Registry
 *
 * The Go backend owns editable page/category metadata and FAQ content for
 * production rendering. Static files remain the storefront fallback.
 */

export {
  faqRegistry,
  faqRoutePathByPageId,
  getAllFaqData,
  getAllFaqItems,
  getFaqData,
  getFaqDataByRoutePath
} from './registry'
export {
  fetchAllFaqData,
  fetchFaqData,
  fetchFaqDataByRoutePath
} from './backend'
export { normalizeFaqRoutePath } from './routing'
export * from './types'
