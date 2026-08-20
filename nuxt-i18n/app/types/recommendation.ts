import type { ShopProduct } from '~/composables/useShopProducts'

export interface RecommendationProductCard extends ShopProduct {
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
  weight_grams?: number
  review_summary?: RecommendationAPIReviewSummary
  slot?: string
  reason?: string
}

export interface RecommendationAPIReviewSummary {
  product_id: number
  total_reviews: number
  average_rating: number
  rating_5_count: number
  rating_4_count: number
  rating_3_count: number
  rating_2_count: number
  rating_1_count: number
}

export interface RecommendationAPIResult {
  request_id: string
  algorithm_version: string
  expires_at: string
  items: RecommendationAPIItem[]
}
