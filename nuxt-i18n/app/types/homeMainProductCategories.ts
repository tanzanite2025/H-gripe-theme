export interface HomeMainProductCategoryItem {
  id: string
  src: string
  altText: string
  title: string
  caption: string
  width: number
  height: number
  desktopOrder: number
  targetUrl?: string
  targetLabel?: string
}

export interface HomeMainProductCategoryApiItem {
  id?: number | string
  image_url?: string
  thumbnail_url?: string
  alt_text?: string
  title?: string
  caption?: string
  width?: number | string
  height?: number | string
  desktop_order?: number | string
  target_url?: string
  target_label?: string
}

export interface HomeMainProductCategoriesApiEnvelope {
  data?: {
    locale?: string
    requested_locale?: string
    fallback?: boolean
    configured_count?: number | string
    items?: HomeMainProductCategoryApiItem[]
  }
}
