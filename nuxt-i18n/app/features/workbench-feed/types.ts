export interface WorkbenchFeedMedia {
  id: number
  file_path: string
  width: number
  height: number
  file_size_kb: number
  caption: string
  sort_order: number
}

export interface WorkbenchTaggedProduct {
  id: number
  product_id: number
  variant_id?: number | null
  product_slug: string
  display_title: string
  price_minor: number
  currency: string
  direct_action: 'detail_drawer'
  available: boolean
  sort_order: number
}

export interface WorkbenchFeedEntry {
  id: number
  entry_number: string
  published_at: string
  locale: string
  content: string
  tags: string[]
  status?: 'draft' | 'published' | 'archived'
  media: WorkbenchFeedMedia[]
  tagged_products: WorkbenchTaggedProduct[]
}

export interface WorkbenchFeedPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface WorkbenchFeedListResult {
  entries: WorkbenchFeedEntry[]
  pagination: WorkbenchFeedPagination
}
