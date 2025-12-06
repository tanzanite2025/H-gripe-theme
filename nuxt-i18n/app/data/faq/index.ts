/**
 * FAQ Data Registry
 * 
 * Central registry for all page-specific FAQ data.
 * Import individual page FAQ data and register them here.
 */

import type { FaqRegistry, PageFaqData } from './types'

// Import page-specific FAQ data
import { supportPaymentFaq } from './pages/support-payment'
import { supportShippingFaq } from './pages/support-shipping'
import { supportAfterSalesFaq } from './pages/support-after-sales'
import { supportWarrantyFaq } from './pages/support-warranty'
import { supportWarrantyCheckFaq } from './pages/support-warranty-check'
import { supportProductFeedbackFaq } from './pages/support-product-feedback'
import { supportTestReportFaq } from './pages/support-test-report'
import { companyMembershipFaq } from './pages/company-membership'
import { guidesWheelsetBuyersFaq } from './pages/guides-wheelset-buyers'

/**
 * FAQ Registry - maps pageId to FAQ data
 */
export const faqRegistry: FaqRegistry = {
  'support-payment': supportPaymentFaq,
  'support-shipping': supportShippingFaq,
  'support-after-sales': supportAfterSalesFaq,
  'support-warranty': supportWarrantyFaq,
  'support-warranty-check': supportWarrantyCheckFaq,
  'support-product-feedback': supportProductFeedbackFaq,
  'support-test-report': supportTestReportFaq,
  'company-membership': companyMembershipFaq,
  'guides-wheelset-buyers': guidesWheelsetBuyersFaq,
}

/**
 * Get FAQ data for a specific page
 * @param pageId - The page identifier
 * @returns The FAQ data for the page, or undefined if not found
 */
export function getFaqData(pageId: string): PageFaqData | undefined {
  return faqRegistry[pageId]
}

/**
 * Get all registered FAQ data
 * @returns Array of all page FAQ data
 */
export function getAllFaqData(): PageFaqData[] {
  return Object.values(faqRegistry)
}

/**
 * Get all FAQ items across all pages (for the aggregated FAQ page)
 * @returns Flattened array of all FAQ categories with page context
 */
export function getAllFaqItems() {
  const allItems: Array<{
    pageId: string
    pageTitle: string
    category: string
    categoryIcon?: string
    question: string
    answer: string
    tags?: string[]
  }> = []

  for (const [pageId, pageData] of Object.entries(faqRegistry)) {
    for (const category of pageData.categories) {
      for (const item of category.items) {
        allItems.push({
          pageId,
          pageTitle: pageData.title || pageId,
          category: category.name,
          categoryIcon: category.icon,
          question: item.question,
          answer: item.answer,
          tags: item.tags,
        })
      }
    }
  }

  return allItems
}

// Re-export types
export * from './types'
