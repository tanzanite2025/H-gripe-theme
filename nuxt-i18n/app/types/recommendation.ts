export interface RecommendationProductCard {
  id: number | string
  title: string
  url: string
  thumbnail?: string
  priceLabel?: string
  slot?: string
  reason?: string
}

export interface RecommendationRequestContext {
  product_id?: number | null
  category_id?: number | null
  query?: string
  route?: string
}

export interface RecommendationRequest {
  surface: string
  locale?: string
  anonymous_id?: string
  session_id?: string
  context?: RecommendationRequestContext
  limit?: number
  exclude_product_ids?: number[]
}

export interface RecommendationAPIItem {
  product_id: number
  title: string
  url: string
  thumbnail?: string
  price_label?: string
  slot?: string
  reason?: string
}

export interface RecommendationAPIResult {
  request_id: string
  algorithm_version: string
  expires_at: string
  items: RecommendationAPIItem[]
}
