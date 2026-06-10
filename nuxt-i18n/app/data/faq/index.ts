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

/**
 * FAQ Registry - maps pageId to FAQ data
 */
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

/**
 * Get FAQ data for a specific page (Static fallback)
 * @param pageId - The page identifier
 * @returns The FAQ data for the page, or undefined if not found
 */
export function getFaqData(pageId: string): PageFaqData | undefined {
  return faqRegistry[pageId]
}

/**
 * Get all registered FAQ data (Static fallback)
 * @returns Array of all page FAQ data
 */
export function getAllFaqData(): PageFaqData[] {
  return Object.values(faqRegistry)
}

/**
 * Fetch FAQ data for a specific page from Go backend
 */
export async function fetchFaqData(pageId: string): Promise<PageFaqData | undefined> {
  try {
    const res = await $fetch<{ data: any[] }>(`http://localhost:8080/api/v1/content/faqs`, {
      query: { page_id: pageId, page_size: 100 }
    })
    
    const items = res.data || []
    if (items.length === 0) return getFaqData(pageId) // Fallback

    const categoriesMap = new Map<string, any>()
    items.forEach(item => {
      const catName = item.category || 'General'
      if (!categoriesMap.has(catName)) {
        categoriesMap.set(catName, {
          id: catName.toLowerCase().replace(/\s+/g, '-'),
          name: catName,
          items: []
        })
      }
      categoriesMap.get(catName).items.push({
        id: item.id.toString(),
        question: item.question,
        answer: item.answer,
        tags: []
      })
    })

    return {
      pageId,
      title: `${pageId.split('-').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ')} FAQs`,
      categories: Array.from(categoriesMap.values())
    }
  } catch (error) {
    console.error("Failed to fetch FAQs from Go backend:", error)
    return getFaqData(pageId)
  }
}

/**
 * Fetch all FAQ data from Go backend
 */
export async function fetchAllFaqData(): Promise<PageFaqData[]> {
  try {
    const res = await $fetch<{ data: any[] }>(`http://localhost:8080/api/v1/content/faqs`, {
      query: { page_size: 1000 }
    })
    
    const items = res.data || []
    if (items.length === 0) return getAllFaqData()

    const pagesMap = new Map<string, PageFaqData>()
    
    items.forEach(item => {
      const pid = item.page_id || 'general'
      if (!pagesMap.has(pid)) {
        pagesMap.set(pid, {
          pageId: pid,
          title: `${pid.split('-').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ')} FAQs`,
          categories: []
        })
      }
      
      const pageData = pagesMap.get(pid)!
      const catName = item.category || 'General'
      
      let category = pageData.categories.find(c => c.name === catName)
      if (!category) {
        category = {
          id: catName.toLowerCase().replace(/\s+/g, '-'),
          name: catName,
          items: []
        }
        pageData.categories.push(category)
      }
      
      category.items.push({
        id: item.id.toString(),
        question: item.question,
        answer: item.answer,
        tags: []
      })
    })

    return Array.from(pagesMap.values())
  } catch (error) {
    console.error("Failed to fetch all FAQs from Go backend:", error)
    return getAllFaqData()
  }
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
