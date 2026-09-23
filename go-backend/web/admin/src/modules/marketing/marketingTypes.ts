import type { AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'

export interface MarketingPagination {
  page: number
  pageSize: number
  total: number
}

export interface CouponRecord {
  id: string | number
  code: string
  type: string
  currency?: string
  value_minor?: number | string
  value_rate_decimal?: number | string
  description?: string
  min_amount_minor?: number | string
  max_discount_minor?: number | string
  used_count?: number | string
  usage_limit?: number | string
  start_date?: string
  end_date?: string
  enabled?: boolean
}

export interface CouponFilters {
  status: string
}

export interface MarketingStatusDisplay {
  label: string
  tone: AdminStatusTone
}

export interface LoyaltyTransaction {
  id: string | number
  type?: string | null
  points?: number | string | null
  balance?: number | string | null
  source?: string | null
  source_id?: string | number | null
  description?: string | null
  created_at?: string | number | Date | null
}

export interface LoyaltyFilters {
  user_id: string | number
}

export interface LoyaltyAdjustmentForm {
  user_id: string | number
  points: number | string
  description: string
}

export type LoyaltyErrors = Partial<Record<keyof LoyaltyAdjustmentForm, string>>

export interface LoyaltySettings {
  points_redemption_enabled: boolean
  points_redemption_currency: string
  points_exchange_rate: number | string
  tz_loyalty_purchase_earn_points_per_currency_unit: number | string
  tz_loyalty_referral_referrer_points: number | string
  tz_loyalty_referral_referee_points: number | string
  tz_loyalty_checkin_base_points: number | string
  tz_loyalty_checkin_streak_interval_days: number | string
  tz_loyalty_checkin_streak_bonus_points: number | string
  tz_loyalty_checkin_max_points: number | string
}

export interface MemberLevel {
  id?: string | number | null
  name?: string
  color?: string
  min_points?: number | string
  max_points?: number | string
  discount_rate_decimal?: string
  benefits?: string
  sort_order?: number | string
}

export interface PromotionRiskSummary {
  severity?: string
  candidate_coupon_count?: number
  risk_item_count?: number
  zero_total_risk_count?: number
  gateway_minimum_risk_count?: number
  member_level_count?: number
  max_member_discount_rate?: number
  max_member_discount_level_name?: string
  points_redemption_enabled?: boolean
  direct_points_discount_cap_rate?: number
}

export interface PromotionRiskItem {
  severity?: string
  kind?: string
  scenario?: string
  coupon_id?: string | number
  coupon_code?: string
  coupon_type?: string
  coupon_status?: string
  coupon_value?: string
  coupon_min_amount?: string
  coupon_max_discount?: string
  member_level_id?: string | number
  member_level_name?: string
  member_discount_rate?: number
  points_discount_rate?: number
  full_cover_subtotal_threshold?: string
  gateway_minimum_threshold?: string
  estimated_subtotal?: string
  estimated_coupon_discount?: string
  estimated_member_discount?: string
  estimated_points_discount?: string
  estimated_discount_amount?: string
  estimated_payable_amount?: string
  factors?: string[]
  recommendation?: string
  starts_at?: string
  ends_at?: string
}

export interface PromotionRiskAnalysis {
  generated_at?: string
  currency?: string
  gateway_minimum_amount?: string
  summary?: PromotionRiskSummary
  items?: PromotionRiskItem[]
}
