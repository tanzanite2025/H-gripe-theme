/**
 * FAQ System Type Definitions
 * 
 * This file defines the data structures for the page-embedded FAQ system.
 * The Go backend is the editable source for page metadata and FAQ content.
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
  /** Optional dedicated FAQ answer image. Must be a backend-validated 800x800 WebP. */
  answerImageUrl?: string
  /** Alt text for the dedicated FAQ answer image */
  answerImageAlt?: string
  /** Validated image width */
  answerImageWidth?: number
  /** Validated image height */
  answerImageHeight?: number
  /** Optional tags for categorization and filtering */
  tags?: string[]
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
  /** FAQ items displayed in page order */
  items: FaqItem[]
}

/**
 * Props for the PageFaq component
 */
export interface PageFaqProps {
  /** Page identifier to load FAQ data */
  pageId: string
  /** Optional resolved FAQ data, used by the automatic route slot */
  data?: PageFaqData
  /** Optional custom title override */
  title?: string
  /** Theme variant */
  theme?: 'light' | 'dark'
  /** Maximum number of items to display (for preview mode) */
  maxItems?: number
  /** Whether to show "View All" link */
  showViewAllLink?: boolean
}

