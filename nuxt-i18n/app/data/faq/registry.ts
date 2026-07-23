import type { FaqRegistry, PageFaqData } from './types'
import { normalizeFaqRoutePath } from './routing'

import { supportPaymentFaq } from './pages/support-payment'
import { supportShippingFaq } from './pages/support-shipping'
import { supportWarrantyFaq } from './pages/support-warranty'
import { supportWarrantyCheckFaq } from './pages/support-warranty-check'
import { supportProductFeedbackFaq } from './pages/support-product-feedback'
import { supportTestReportFaq } from './pages/support-test-report'
import { companyMembershipFaq } from './pages/company-membership'
import { guidesWheelsetBuyersFaq } from './pages/guides-wheelset-buyers'
import { guidesTireguidesFaq } from './pages/guides-tireguides'
import { productsSpokeCalculatorFaq } from './pages/products-spoke-calculator'
import { companyOemOdmFaq } from './pages/company-oem-odm'
import { companyCertificatesFaq } from './pages/company-certificates'
import { companyContactFaq } from './pages/company-contact'
import { companyGlobalPartnersFaq } from './pages/company-global-partners'
import { companyOurStoryFaq } from './pages/company-ourstory'

export const faqRegistry: FaqRegistry = {
  'support-payment': supportPaymentFaq,
  'support-shipping': supportShippingFaq,
  'support-warranty': supportWarrantyFaq,
  'support-warranty-check': supportWarrantyCheckFaq,
  'support-product-feedback': supportProductFeedbackFaq,
  'support-test-report': supportTestReportFaq,
  'company-membership': companyMembershipFaq,
  'guides-wheelset-buyers': guidesWheelsetBuyersFaq,
  'guides-tireguides': guidesTireguidesFaq,
  'products-spoke-calculator': productsSpokeCalculatorFaq,
  'company-oem-odm': companyOemOdmFaq,
  'company-certificates': companyCertificatesFaq,
  'company-contact': companyContactFaq,
  'company-global-partners': companyGlobalPartnersFaq,
  'company-ourstory': companyOurStoryFaq,
}

export const faqRoutePathByPageId: Record<string, string> = {
  'products-spoke-calculator': '/spoke-calculator',
  'guides-tireguides': '/guides/tireguides',
  'guides-wheelset-buyers': '/guides/wheelset-buyers',
  'support-payment': '/support/payment',
  'support-shipping': '/support/shipping',
  'support-warranty': '/support/warranty',
  'support-warranty-check': '/support/warranty-check',
  'support-product-feedback': '/support/product-feedback',
  'support-test-report': '/support/test-report',
  'company-membership': '/membershipandpoints',
  'company-oem-odm': '/company/oem-odm',
  'company-certificates': '/company/certificates',
  'company-contact': '/company/contact',
  'company-global-partners': '/company/global-partners',
  'company-ourstory': '/company/ourstory',
}

/**
 * Get FAQ data for a specific page (Static fallback)
 */
export function getFaqData(pageId: string): PageFaqData | undefined {
  return faqRegistry[pageId]
}

/**
 * Get all registered FAQ data (Static fallback)
 */
export function getAllFaqData(): PageFaqData[] {
  return Object.values(faqRegistry)
}

export function getFaqDataByRoutePath(routePath: string): PageFaqData | undefined {
  const normalizedPath = normalizeFaqRoutePath(routePath)
  return getAllFaqData().find(page => normalizeFaqRoutePath(faqRoutePathByPageId[page.pageId] || '') === normalizedPath)
}

/**
 * Get all FAQ items across all pages (for the aggregated FAQ page)
 */
export function getAllFaqItems() {
  const allItems: Array<{
    pageId: string
    pageTitle: string
    category: string
    categoryIcon?: string
    question: string
    answer: string
    answerImageUrl?: string
    answerImageAlt?: string
    answerImageWidth?: number
    answerImageHeight?: number
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
          answerImageUrl: item.answerImageUrl,
          answerImageAlt: item.answerImageAlt,
          answerImageWidth: item.answerImageWidth,
          answerImageHeight: item.answerImageHeight,
          tags: item.tags,
        })
      }
    }
  }

  return allItems
}
