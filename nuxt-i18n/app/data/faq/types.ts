/**
 * FAQ System Type Definitions
 * 
 * This file defines the data structures for the page-embedded FAQ system.
 * All FAQ content is managed statically on the frontend.
 */

/**
 * Single FAQ item
 */
export interface FaqItem {
  /** Unique identifier for the FAQ item */
  id: string
  /** The question text */
  question: string
  /** The answer text (supports HTML for rich formatting) */
  answer: string
  /** Optional tags for categorization and filtering */
  tags?: string[]
}

/**
 * FAQ category containing multiple FAQ items
 */
export interface FaqCategory {
  /** Unique identifier for the category */
  id: string
  /** Category display name */
  name: string
  /** Optional icon (emoji or icon class) */
  icon?: string
  /** FAQ items in this category */
  items: FaqItem[]
}

/**
 * Page-specific FAQ data structure
 */
export interface PageFaqData {
  /** Page identifier (e.g., 'home', 'shop', 'product-detail') */
  pageId: string
  /** Page display title for the FAQ section */
  title?: string
  /** Optional subtitle or description */
  subtitle?: string
  /** Categories of FAQs for this page */
  categories: FaqCategory[]
}

/**
 * Props for the PageFaq component
 */
export interface PageFaqProps {
  /** Page identifier to load FAQ data */
  pageId: string
  /** Optional custom title override */
  title?: string
  /** Theme variant */
  theme?: 'light' | 'dark'
  /** Whether to show category headers */
  showCategories?: boolean
  /** Maximum number of items to display (for preview mode) */
  maxItems?: number
  /** Whether to show "View All" link */
  showViewAllLink?: boolean
}

/**
 * Registry of all page FAQ data
 * Key is the pageId, value is the FAQ data for that page
 */
export type FaqRegistry = Record<string, PageFaqData>
