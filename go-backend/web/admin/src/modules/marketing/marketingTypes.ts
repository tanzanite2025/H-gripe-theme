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
  value?: number | string
  description?: string
  min_amount?: number | string
  max_discount?: number | string
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

export interface GiftCardRecord {
  id: string | number
  code: string
  initial_value?: number | string
  balance?: number | string
  currency?: string
  recipient_name?: string
  recipient_email?: string
  status?: string
  expires_at?: string
  created_at?: string
}

export interface GiftCardFilters {
  status: string
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
  tz_loyalty_purchase_earn_points_per_currency_unit: number | string
  tz_loyalty_referral_referrer_points: number | string
  tz_loyalty_referral_referee_points: number | string
  tz_loyalty_checkin_base_points: number | string
  tz_loyalty_checkin_streak_interval_days: number | string
  tz_loyalty_checkin_streak_bonus_points: number | string
  tz_loyalty_checkin_max_points: number | string
}

export interface GiftCardRedeemOption {
  key?: string
  value: number | string
  currency: string
  stock_quantity: number | string
  redeemed_quantity?: number | string
  remaining_quantity?: number | string
}

export interface GiftCardRedeemSettings {
  tz_redeem_enabled: boolean
  tz_redeem_currency: string
  tz_redeem_exchange_rate: number | string
  tz_redeem_min_points: number | string
  tz_redeem_max_value_per_day: number | string
  tz_redeem_card_expiry_days: number | string
  options?: GiftCardRedeemOption[]
}

export interface MemberLevel {
  id?: string | number | null
  name?: string
  color?: string
  min_points?: number | string
  max_points?: number | string
  discount_rate?: number | string
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
  max_redeem_gift_card_value?: number
}

export interface PromotionRiskItem {
  severity?: string
  kind?: string
  scenario?: string
  coupon_id?: string | number
  coupon_code?: string
  coupon_type?: string
  coupon_status?: string
  coupon_value?: number
  coupon_min_amount?: number
  coupon_max_discount?: number
  member_level_id?: string | number
  member_level_name?: string
  member_discount_rate?: number
  points_discount_rate?: number
  full_cover_subtotal_threshold?: number
  gateway_minimum_threshold?: number
  estimated_subtotal?: number
  estimated_coupon_discount?: number
  estimated_member_discount?: number
  estimated_points_discount?: number
  estimated_discount_amount?: number
  estimated_payable_amount?: number
  factors?: string[]
  recommendation?: string
  starts_at?: string
  ends_at?: string
}

export interface PromotionRiskAnalysis {
  generated_at?: string
  currency?: string
  gateway_minimum_amount?: number
  summary?: PromotionRiskSummary
  items?: PromotionRiskItem[]
}
