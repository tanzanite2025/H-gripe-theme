export type BehaviorEventType =
  | 'page_view'
  | 'product_view'
  | 'product_dwell'
  | 'search_submit'
  | 'filter_apply'
  | 'category_navigation_click'
  | 'calculator_use'
  | 'recommendation_impression'
  | 'recommendation_click'
  | 'add_to_cart'
  | 'wishlist_add'
  | 'begin_checkout'
  | 'ad_landing'
  | 'quiz_completed'

export type BehaviorEventMetadataValue = string | number | boolean | null

export interface BehaviorEventMetadata {
  [key: string]: BehaviorEventMetadataValue
}

export interface TrackBehaviorEventInput {
  eventType: BehaviorEventType
  productId?: number | null
  categoryId?: number | null
  locale?: string
  path?: string
  referrer?: string
  metadata?: BehaviorEventMetadata
  occurredAt?: string
}
